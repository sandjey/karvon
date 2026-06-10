package dto

type AdminLoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateModeratorRequest struct {
	Phone string  `json:"phone" binding:"required"`
	Name  *string `json:"name" binding:"omitempty,max=100"`
}

type TopupRequest struct {
	Amount int `json:"amount" binding:"required,min=1"`
}

type BlockRequest struct {
	Blocked bool `json:"blocked"`
}

type UpdatePricingRequest struct {
	Label        *string  `json:"label"`
	PriceUZS     *float64 `json:"price_uzs"`
	PriceUSD     *float64 `json:"price_usd"`
	TokensAmount *int     `json:"tokens_amount"`
	DurationDays *int     `json:"duration_days"`
	IsActive     *bool    `json:"is_active"`
}

type CreatePricingRequest struct {
	Key          string  `json:"key" binding:"required,max=100"`
	Label        string  `json:"label" binding:"required,max=200"`
	PriceUZS     float64 `json:"price_uzs"`
	PriceUSD     float64 `json:"price_usd"`
	TokensAmount int     `json:"tokens_amount"`
	DurationDays int     `json:"duration_days"`
	IsActive     *bool   `json:"is_active"`
}
