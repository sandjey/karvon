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
	whatsapp    notifier.Notifier
	telegram    notifier.Notifier
}

func NewAuthService(
	userRepo *repository.UserRepo,
	otpStore *otpstore.OTPStore,
	tokenRepo *repository.TokenRepo,
	pricingRepo *repository.PricingRepo,
	jwtMgr *jwtpkg.Manager,
	whatsapp notifier.Notifier,
	telegram notifier.Notifier,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		otpStore:    otpStore,
		tokenRepo:   tokenRepo,
		pricingRepo: pricingRepo,
		jwtMgr:      jwtMgr,
		whatsapp:    whatsapp,
		telegram:    telegram,
	}
}

// SendOTP генерирует и отправляет OTP-код.
func (s *AuthService) SendOTP(ctx context.Context, req dto.SendOTPRequest) (*dto.SendOTPResponse, error) {
	phone := normalizePhone(req.Phone)

	// Pre-flight: cooldown + rate limit check (без изменения счётчиков).
	if err := s.otpStore.CheckSendAllowed(ctx, phone); err != nil {
		return nil, err
	}

	channel := notifier.Channel(req.Channel)
	if channel == "" {
		channel = notifier.ChannelWhatsApp
	}

	switch channel {
	case notifier.ChannelTelegram:
		if s.telegram == nil {
			return nil, ErrTelegramNotConfigured
		}
	default:
		if s.whatsapp == nil {
			return nil, ErrWhatsAppNotConfigured
		}
	}

	code, err := generateOTP()
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("🔐 Ваш код подтверждения Central Trade Market: *%s*\nКод действует 5 минут.", code)

	var requestID string

	switch channel {
	case notifier.ChannelTelegram:
		if err := s.telegram.Send(ctx, phone, message); err != nil {
			if errors.Is(err, notifier.ErrRateLimit) {
				return nil, ErrOTPRateLimit
			}
			if errors.Is(err, notifier.ErrNumberNotRegistered) {
				if s.whatsapp != nil {
					if err2 := s.whatsapp.Send(ctx, phone, message); err2 != nil {
						if errors.Is(err2, notifier.ErrRateLimit) {
							return nil, ErrOTPRateLimit
						}
						return nil, ErrPhoneNotOnTelegram
					}
					break
				}
				return nil, ErrPhoneNotOnTelegram
			}
			return nil, err
		}
	default:
		if err := s.whatsapp.Send(ctx, phone, message); err != nil {
			if errors.Is(err, notifier.ErrRateLimit) {
				return nil, ErrOTPRateLimit
			}
			return nil, err
		}
	}

	// Сохранить в Redis (хэшированно), запустить cooldown.
	cooldown, err := s.otpStore.SaveOTP(ctx, phone, code, requestID)
	if err != nil {
		return nil, err
	}

	cd := int(cooldown.Seconds())
	return &dto.SendOTPResponse{
		ExpiresIn:       int(otpTTL.Seconds()),
		CooldownSeconds: cd,
	}, nil
}

// VerifyOTP проверяет OTP и возвращает токены.
func (s *AuthService) VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.TokenPair, error) {
	phone := normalizePhone(req.Phone)

	_, err := s.otpStore.Verify(ctx, phone, req.Code)
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
		newUser := &model.User{
			ID:    uuid.New(),
			Phone: phone,
			Role:  "user",
		}
		if err := s.userRepo.Create(ctx, newUser); err != nil {
			return nil, err
		}
		user = newUser

		bonus := s.pricingRepo.TokensAmount(ctx, "tokens_registration", 5)
		if bonus > 0 {
			if err := s.userRepo.CreditTokens(ctx, user.ID, bonus, "registration"); err != nil {
				return nil, err
			}
		}
	}

	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	return s.issueTokens(ctx, user, isNew)
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
	if req.Email != nil {
		fields["email"] = req.Email
	}
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
