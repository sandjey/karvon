package service

import "errors"

var (
	ErrOTPRateLimit          = errors.New("OTP_MAX_ATTEMPTS")
	ErrOTPInvalid            = errors.New("OTP_INVALID")
	ErrOTPMaxAttempts        = errors.New("OTP_MAX_ATTEMPTS")
	ErrOTPExpired            = errors.New("OTP_EXPIRED")
	ErrUserBlocked           = errors.New("USER_BLOCKED")
	ErrTokenInvalid          = errors.New("UNAUTHORIZED")
	ErrNotFound              = errors.New("NOT_FOUND")
	ErrNameAlreadySet        = errors.New("NAME_ALREADY_SET")
	ErrTelegramNotConfigured = errors.New("TELEGRAM_NOT_CONFIGURED")
	ErrWhatsAppNotConfigured = errors.New("WHATSAPP_NOT_CONFIGURED")
	ErrAlreadyExists         = errors.New("ALREADY_EXISTS")
	ErrCompanyNotFound       = errors.New("COMPANY_NOT_FOUND")
	ErrCompanyNotOwned       = errors.New("FORBIDDEN")
	ErrCompanyNotEditable    = errors.New("COMPANY_NOT_EDITABLE")
)
