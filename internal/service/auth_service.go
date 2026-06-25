package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/repository"
	jwtpkg "ctm/pkg/jwt"
	"ctm/pkg/notifier"
	"ctm/pkg/otpstore"
)

const otpTTL = 5 * time.Minute

var phoneRegexp = regexp.MustCompile(`[\s\-\(\)]`)

type AuthService struct {
	userRepo    *repository.UserRepo
	otpStore    *otpstore.OTPStore
	tokenRepo   *repository.TokenRepo
	pricingRepo *repository.PricingRepo
	jwtMgr      *jwtpkg.Manager
	telegram    notifier.Notifier
	emailSvc    *EmailService
}

func NewAuthService(
	userRepo *repository.UserRepo,
	otpStore *otpstore.OTPStore,
	tokenRepo *repository.TokenRepo,
	pricingRepo *repository.PricingRepo,
	jwtMgr *jwtpkg.Manager,
	telegram notifier.Notifier,
	emailSvc *EmailService,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		otpStore:    otpStore,
		tokenRepo:   tokenRepo,
		pricingRepo: pricingRepo,
		jwtMgr:      jwtMgr,
		telegram:    telegram,
		emailSvc:    emailSvc,
	}
}

// SendOTP генерирует и отправляет OTP-код через Telegram или Email.
func (s *AuthService) SendOTP(ctx context.Context, req dto.SendOTPRequest, lang string) (*dto.SendOTPResponse, error) {
	channel := notifier.Channel(req.Channel)
	if channel == "" {
		channel = notifier.ChannelTelegram
	}

	switch channel {
	case "email":
		if strings.TrimSpace(req.Email) == "" {
			return nil, ErrEmailRequired
		}
		if s.emailSvc == nil {
			return nil, ErrInvalidChannel
		}
		if err := s.emailSvc.SendOTP(ctx, req.Email, lang); err != nil {
			return nil, err
		}
		return &dto.SendOTPResponse{
			ExpiresIn:       600,
			CooldownSeconds: 0,
		}, nil

	case notifier.ChannelTelegram:
		if strings.TrimSpace(req.Phone) == "" {
			return nil, ErrPhoneRequired
		}
		if s.telegram == nil {
			return nil, ErrTelegramNotConfigured
		}
		phone := normalizePhone(req.Phone)

		// Pre-flight: cooldown + rate limit check (без изменения счётчиков).
		if err := s.otpStore.CheckSendAllowed(ctx, phone); err != nil {
			return nil, err
		}

		code, err := generateOTP()
		if err != nil {
			return nil, err
		}

		message := fmt.Sprintf("🔐 Ваш код подтверждения Central Trade Market: *%s*\nКод действует 5 минут.", code)

		if err := s.telegram.Send(ctx, phone, message); err != nil {
			if errors.Is(err, notifier.ErrRateLimit) {
				return nil, ErrOTPRateLimit
			}
			if errors.Is(err, notifier.ErrNumberNotRegistered) {
				return nil, ErrPhoneNotOnTelegram
			}
			return nil, err
		}

		// Сохранить в Redis (хэшированно), запустить cooldown.
		cooldown, err := s.otpStore.SaveOTP(ctx, phone, code, "")
		if err != nil {
			return nil, err
		}

		return &dto.SendOTPResponse{
			ExpiresIn:       int(otpTTL.Seconds()),
			CooldownSeconds: int(cooldown.Seconds()),
		}, nil

	default:
		return nil, ErrInvalidChannel
	}
}

// VerifyOTP проверяет OTP (по email или phone) и возвращает токены.
func (s *AuthService) VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.TokenPair, error) {
	hasEmail := req.Email != nil && strings.TrimSpace(*req.Email) != ""
	hasPhone := strings.TrimSpace(req.Phone) != ""

	if hasEmail == hasPhone {
		// должен быть задан ровно один идентификатор
		return nil, ErrValidation
	}

	if hasEmail {
		return s.verifyOTPEmail(ctx, strings.TrimSpace(*req.Email), req.Code)
	}
	return s.verifyOTPPhone(ctx, req.Phone, req.Code)
}

func (s *AuthService) verifyOTPPhone(ctx context.Context, rawPhone, code string) (*dto.TokenPair, error) {
	phone := normalizePhone(rawPhone)

	_, err := s.otpStore.Verify(ctx, phone, code)
	if err != nil {
		switch {
		case errors.Is(err, otpstore.ErrOTPExpired):
			return nil, ErrOTPExpired
		case errors.Is(err, otpstore.ErrOTPInvalid):
			return nil, ErrOTPInvalid
		case errors.Is(err, otpstore.ErrOTPMaxAttempts), errors.Is(err, otpstore.ErrOTPVerifyRateLimited):
			return nil, ErrOTPMaxAttempts
		default:
			return nil, err
		}
	}

	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	isNew := user == nil || user.Name == nil
	if user == nil {
		isNew = true
		p := phone
		newUser := &model.User{
			ID:    uuid.New(),
			Phone: &p,
			Role:  "user",
		}
		if err := s.createWithBonus(ctx, newUser); err != nil {
			return nil, err
		}
		user = newUser
	}

	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	return s.issueTokens(ctx, user, isNew)
}

func (s *AuthService) verifyOTPEmail(ctx context.Context, emailAddr, code string) (*dto.TokenPair, error) {
	if s.emailSvc == nil {
		return nil, ErrInvalidChannel
	}

	if err := s.emailSvc.VerifyOTP(ctx, emailAddr, code); err != nil {
		switch {
		case errors.Is(err, ErrEmailOTPExpired):
			return nil, ErrOTPExpired
		case errors.Is(err, ErrEmailOTPInvalid):
			return nil, ErrOTPInvalid
		default:
			return nil, err
		}
	}

	user, err := s.userRepo.FindByEmail(ctx, emailAddr)
	if err != nil {
		return nil, err
	}

	isNew := user == nil || user.Name == nil
	if user == nil {
		isNew = true
		e := emailAddr
		newUser := &model.User{
			ID:    uuid.New(),
			Email: &e,
			Role:  "user",
		}
		if err := s.createWithBonus(ctx, newUser); err != nil {
			return nil, err
		}
		user = newUser
	}

	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	return s.issueTokens(ctx, user, isNew)
}

// createWithBonus создаёт пользователя и начисляет регистрационный бонус.
func (s *AuthService) createWithBonus(ctx context.Context, newUser *model.User) error {
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return err
	}
	bonus := s.pricingRepo.TokensAmount(ctx, "tokens_registration", 5)
	if bonus > 0 {
		if err := s.userRepo.CreditTokens(ctx, newUser.ID, bonus, "registration"); err != nil {
			return err
		}
	}
	return nil
}

// Refresh выдаёт новую пару токенов.
func (s *AuthService) Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.RefreshResponse, error) {
	rt, err := s.tokenRepo.Find(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		return nil, ErrTokenInvalid
	}

	_ = s.tokenRepo.Delete(ctx, req.RefreshToken)

	user, err := s.userRepo.FindByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return nil, ErrTokenInvalid
	}
	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	access, err := s.jwtMgr.GenerateAccess(user.ID, user.Role, user.TokenVersion)
	if err != nil {
		return nil, err
	}
	refresh, expiresAt, err := s.jwtMgr.GenerateRefresh(user.ID)
	if err != nil {
		return nil, err
	}
	if err := s.tokenRepo.Save(ctx, user.ID, refresh, expiresAt); err != nil {
		return nil, err
	}

	return &dto.RefreshResponse{AccessToken: access, RefreshToken: refresh}, nil
}

// Logout инвалидирует все токены пользователя.
func (s *AuthService) Logout(ctx context.Context, refreshToken string, userID uuid.UUID) error {
	_ = s.tokenRepo.DeleteByUserID(ctx, userID)
	return s.userRepo.IncrementTokenVersion(ctx, userID)
}

// CompleteRegistration устанавливает имя и опциональные поля профиля.
func (s *AuthService) CompleteRegistration(ctx context.Context, userID uuid.UUID, req dto.CompleteRegistrationRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ErrNotFound
	}
	if user.Name != nil && *user.Name != "" {
		return ErrNameAlreadySet
	}
	fields := map[string]interface{}{"name": req.Name}
	if req.ExtraPhone != nil {
		fields["extra_phone"] = req.ExtraPhone
	}
	if req.WhatsApp != nil {
		fields["whatsapp"] = req.WhatsApp
	}
	if req.Telegram != nil {
		fields["telegram"] = req.Telegram
	}
	if req.City != nil {
		fields["city"] = req.City
	}
	if req.Country != nil {
		fields["country"] = req.Country
	}
	return s.userRepo.UpdateProfile(ctx, userID, fields)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (s *AuthService) issueTokens(ctx context.Context, user *model.User, isNew bool) (*dto.TokenPair, error) {
	access, err := s.jwtMgr.GenerateAccess(user.ID, user.Role, user.TokenVersion)
	if err != nil {
		return nil, err
	}
	refresh, expiresAt, err := s.jwtMgr.GenerateRefresh(user.ID)
	if err != nil {
		return nil, err
	}
	if err := s.tokenRepo.Save(ctx, user.ID, refresh, expiresAt); err != nil {
		return nil, err
	}
	return &dto.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		IsNewUser:    isNew,
	}, nil
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func normalizePhone(phone string) string {
	phone = phoneRegexp.ReplaceAllString(phone, "")
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}
	return phone
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
