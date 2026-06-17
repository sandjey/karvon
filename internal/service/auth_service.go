package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"karvon/internal/dto"
	"karvon/internal/model"
	"karvon/internal/repository"
	jwtpkg "karvon/pkg/jwt"
	"karvon/pkg/notifier"
)

const (
	otpTTL      = 5 * time.Minute
	otpMaxAttempts = 5
	otpRateLimit   = 3
)

var phoneRegexp = regexp.MustCompile(`[\s\-\(\)]`)

type AuthService struct {
	userRepo     *repository.UserRepo
	otpRepo      *repository.OTPRepo
	tokenRepo    *repository.TokenRepo
	pricingRepo  *repository.PricingRepo
	jwtMgr       *jwtpkg.Manager
	whatsapp     notifier.Notifier
	telegram     notifier.Notifier
	universalOTP string
}

func NewAuthService(
	userRepo *repository.UserRepo,
	otpRepo *repository.OTPRepo,
	tokenRepo *repository.TokenRepo,
	pricingRepo *repository.PricingRepo,
	jwtMgr *jwtpkg.Manager,
	whatsapp notifier.Notifier,
	telegram notifier.Notifier,
	universalOTP string,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		otpRepo:      otpRepo,
		tokenRepo:    tokenRepo,
		pricingRepo:  pricingRepo,
		jwtMgr:       jwtMgr,
		whatsapp:     whatsapp,
		telegram:     telegram,
		universalOTP: universalOTP,
	}
}

// SendOTP генерирует и отправляет OTP-код.
func (s *AuthService) SendOTP(ctx context.Context, req dto.SendOTPRequest) (*dto.SendOTPResponse, error) {
	phone := normalizePhone(req.Phone)

	// Проверка rate limit
	count, err := s.otpRepo.CountRecentByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if count >= otpRateLimit {
		return nil, ErrOTPRateLimit
	}

	channel := notifier.Channel(req.Channel)
	if channel == "" {
		channel = notifier.ChannelWhatsApp
	}

	// Проверить, зарегистрирован ли номер на выбранном канале
	switch channel {
	case notifier.ChannelTelegram:
		if s.telegram == nil {
			return nil, ErrTelegramNotConfigured
		}
		if checker, ok := s.telegram.(notifier.NumberChecker); ok {
			if err := checker.CheckNumber(ctx, phone); err != nil {
				return nil, ErrPhoneNotOnTelegram
			}
		}
	default:
		if s.whatsapp == nil {
			return nil, ErrWhatsAppNotConfigured
		}
		if checker, ok := s.whatsapp.(notifier.NumberChecker); ok {
			if err := checker.CheckNumber(ctx, phone); err != nil {
				return nil, ErrPhoneNotOnWhatsApp
			}
		}
	}

	// Деактивировать старые коды
	if err := s.otpRepo.InvalidateByPhone(ctx, phone); err != nil {
		return nil, err
	}

	// Генерация кода
	code, err := generateOTP()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(otpTTL)
	if err := s.otpRepo.Create(ctx, &model.OTPCode{
		Phone:     phone,
		Code:      code,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, err
	}

	message := fmt.Sprintf("🔐 Ваш код подтверждения KARVON: *%s*\nКод действует 5 минут.", code)

	// Отправка по выбранному каналу
	switch channel {
	case notifier.ChannelTelegram:
		if s.telegram == nil {
			return nil, ErrTelegramNotConfigured
		}
		if err := s.telegram.Send(ctx, phone, message); err != nil {
			return nil, ErrTelegramNotConfigured
		}
	default:
		if s.whatsapp == nil {
			return nil, ErrWhatsAppNotConfigured
		}
		if err := s.whatsapp.Send(ctx, phone, message); err != nil {
			return nil, ErrWhatsAppNotConfigured
		}
	}

	return &dto.SendOTPResponse{ExpiresIn: int(otpTTL.Seconds())}, nil
}

// VerifyOTP проверяет OTP и возвращает токены.
func (s *AuthService) VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.TokenPair, error) {
	phone := normalizePhone(req.Phone)

	// Универсальный OTP (QA): если задан в .env и совпадает — пропускаем проверку кода.
	universal := s.universalOTP != "" && req.Code == s.universalOTP

	if !universal {
		otp, err := s.otpRepo.FindActive(ctx, phone, req.Code)
		if err != nil {
			return nil, err
		}
		if otp == nil {
			return nil, ErrOTPInvalid
		}
		if otp.Attempts >= otpMaxAttempts {
			return nil, ErrOTPMaxAttempts
		}
		// Помечаем как использованный.
		_ = s.otpRepo.MarkUsed(ctx, otp.ID)
	}

	// Получить или создать пользователя
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

		// Начислить бонусные токены новому пользователю (сумма управляется супер-админом).
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

	// Удалить старый refresh token
	_ = s.tokenRepo.Delete(ctx, req.RefreshToken)

	user, err := s.userRepo.FindByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return nil, ErrTokenInvalid
	}
	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	access, err := s.jwtMgr.GenerateAccess(user.ID, user.Role)
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

// Logout инвалидирует refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.tokenRepo.Delete(ctx, refreshToken)
}

// CompleteRegistration устанавливает имя и опциональные поля профиля новому пользователю.
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
	access, err := s.jwtMgr.GenerateAccess(user.ID, user.Role)
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
