package dto

import (
	"time"

	"github.com/google/uuid"
)

type UpdateProfileRequest struct {
	Name       *string `json:"name"        binding:"omitempty,min=2,max=100"`
	Email      *string `json:"email"       binding:"omitempty,email"`
	ExtraPhone *string `json:"extra_phone" binding:"omitempty,max=30"`
	WhatsApp   *string `json:"whatsapp"    binding:"omitempty,max=20"`
	Telegram   *string `json:"telegram"    binding:"omitempty,max=50"`
	City       *string `json:"city"        binding:"omitempty,max=100"`
	Country    *string `json:"country"     binding:"omitempty,max=100"`
}

type ProfileResponse struct {
	ID           uuid.UUID `json:"id"`
	Phone        string    `json:"phone"`
	Name         *string   `json:"name"`
	Email        *string   `json:"email"`
	ExtraPhone   *string   `json:"extra_phone"`
	WhatsApp     *string   `json:"whatsapp"`
	Telegram     *string   `json:"telegram"`
	City         *string   `json:"city"`
	Country      *string   `json:"country"`
	TokenBalance int       `json:"token_balance"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserStatsResponse struct {
	TokenBalance        int `json:"token_balance"`
	CompaniesCount      int `json:"companies_count"`
	VerifiedCompanies   int `json:"verified_companies"`
	ActiveCargo         int `json:"active_cargo"`
	ActiveWarehouses    int `json:"active_warehouses"`
	ContactsPurchased   int `json:"contacts_purchased"`
}

type ListingQuotaResponse struct {
	FreeUsed    int    `json:"free_used"`
	FreeTotal   int    `json:"free_total"`
	CanPostFree bool   `json:"can_post_free"`
	PricingKey  string `json:"pricing_key"`
}
