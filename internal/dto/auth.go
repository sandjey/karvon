package dto

// ── Requests ─────────────────────────────────────────────────────────────────

type SendOTPRequest struct {
	Phone   string `json:"phone" binding:"required"`
	Channel string `json:"channel"` // "whatsapp" | "telegram", default: whatsapp
}

type VerifyOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required,len=6"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type CompleteRegistrationRequest struct {
	Name       string  `json:"name"        binding:"required,min=2,max=100"`
	Email      *string `json:"email"       binding:"omitempty,email"`
	ExtraPhone *string `json:"extra_phone" binding:"omitempty,max=30"`
	WhatsApp   *string `json:"whatsapp"    binding:"omitempty,max=20"`
	Telegram   *string `json:"telegram"    binding:"omitempty,max=50"`
	City       *string `json:"city"        binding:"omitempty,max=100"`
	Country    *string `json:"country"     binding:"omitempty,max=100"`
}

// ── Responses ─────────────────────────────────────────────────────────────────

type SendOTPResponse struct {
	ExpiresIn       int `json:"expires_in"`       // секунды до истечения OTP
	CooldownSeconds int `json:"cooldown_seconds"` // когда можно запросить следующий OTP
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IsNewUser    bool   `json:"is_new_user"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
