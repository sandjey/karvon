package dto

import (
	"time"

	"github.com/google/uuid"

	"ctm/internal/model"
)

// AdminUserDetailResponse — enriched user detail for admin panel.
type AdminUserDetailResponse struct {
	ID                 uuid.UUID            `json:"id"`
	Phone              string               `json:"phone"`
	Name               *string              `json:"name"`
	Email              *string              `json:"email"`
	ExtraPhone         *string              `json:"extra_phone"`
	WhatsApp           *string              `json:"whatsapp"`
	Telegram           *string              `json:"telegram"`
	City               *string              `json:"city"`
	Country            *string              `json:"country"`
	TokenBalance       int                  `json:"token_balance"`
	Role               string               `json:"role"`
	IsBlocked          bool                 `json:"is_blocked"`
	BlockedReason      *string              `json:"blocked_reason"`
	BlockedAt          *time.Time           `json:"blocked_at"`
	LastLoginAt        *time.Time           `json:"last_login_at"`
	RegistrationSource *string              `json:"registration_source"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
	// Activity stats
	Companies           []model.Company      `json:"companies"`
	ListingsCount       int64                `json:"listings_count"`
	ActiveSubscription  *model.Subscription  `json:"active_subscription"`
	TotalSpent          float64              `json:"total_spent"`
	PaymentsCount       int64                `json:"payments_count"`
	ContactsViewedCount int64                `json:"contacts_viewed_count"`
}

type AdminLoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateModeratorRequest struct {
	Phone    string  `json:"phone"    binding:"required"`
	Name     *string `json:"name"     binding:"omitempty,max=100"`
	Login    string  `json:"login"    binding:"required,min=3,max=50"`
	Password string  `json:"password" binding:"required,min=6,max=100"`
}

type TopupRequest struct {
	Amount int `json:"amount" binding:"required,min=1"`
}

type BlockRequest struct {
	Blocked bool   `json:"blocked"`
	Reason  string `json:"reason"`
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

type CreateCategoryRequest struct {
	Key      string `json:"key"      binding:"required,min=1,max=50"`
	LabelRu  string `json:"label_ru" binding:"required,min=1,max=200"`
	LabelUz  string `json:"label_uz" binding:"required,min=1,max=200"`
	LabelEn  string `json:"label_en" binding:"required,min=1,max=200"`
	IsActive *bool  `json:"is_active"`
}

type UpdateCategoryRequest struct {
	LabelRu  *string `json:"label_ru" binding:"omitempty,min=1,max=200"`
	LabelUz  *string `json:"label_uz" binding:"omitempty,min=1,max=200"`
	LabelEn  *string `json:"label_en" binding:"omitempty,min=1,max=200"`
	IsActive *bool   `json:"is_active"`
}
