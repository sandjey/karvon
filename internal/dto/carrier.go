package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateCarrierRequest struct {
	OrgType       string   `json:"org_type"        binding:"required,max=50"`
	Name          string   `json:"name"            binding:"required,min=2,max=200"`
	INN           string   `json:"inn"             binding:"required,min=3,max=30"`
	Country       string   `json:"country"         binding:"required,max=100"`
	City          string   `json:"city"            binding:"required,max=100"`
	Region        *string  `json:"region"          binding:"omitempty,max=100"`
	Phone         string   `json:"phone"           binding:"required,max=20"`
	Email         *string  `json:"email"           binding:"omitempty,email,max=100"`
	Website       *string  `json:"website"         binding:"omitempty,max=300"`
	TransportType string   `json:"transport_type"  binding:"required,oneof=railway auto aviation sea"`
	WorkCountries []string `json:"work_countries"  binding:"required,min=1,max=100"`
	Description   *string  `json:"description"     binding:"omitempty,max=2000"`
	LogoURL       *string  `json:"logo_url"        binding:"omitempty,max=500"`

	Media []MediaItem `json:"media"`
}

type UpdateCarrierRequest struct {
	OrgType       *string  `json:"org_type"        binding:"omitempty,max=50"`
	Name          *string  `json:"name"            binding:"omitempty,min=2,max=200"`
	INN           *string  `json:"inn"             binding:"omitempty,min=3,max=30"`
	Country       *string  `json:"country"         binding:"omitempty,max=100"`
	City          *string  `json:"city"            binding:"omitempty,max=100"`
	Region        *string  `json:"region"          binding:"omitempty,max=100"`
	Phone         *string  `json:"phone"           binding:"omitempty,max=20"`
	Email         *string  `json:"email"           binding:"omitempty,email,max=100"`
	Website       *string  `json:"website"         binding:"omitempty,max=300"`
	TransportType *string  `json:"transport_type"  binding:"omitempty,oneof=railway auto aviation sea"`
	WorkCountries []string `json:"work_countries"  binding:"omitempty,min=1,max=100"`
	Description   *string  `json:"description"     binding:"omitempty,max=2000"`
	LogoURL       *string  `json:"logo_url"        binding:"omitempty,max=500"`
	Status        *string  `json:"status"          binding:"omitempty,oneof=active archived"`

	Media []MediaItem `json:"media"`
}

type CarrierResponse struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	OrgType         string     `json:"org_type"`
	Name            string     `json:"name"`
	INN             string     `json:"inn"`
	Country         string     `json:"country"`
	City            string     `json:"city"`
	Region          *string    `json:"region"`
	Phone           string     `json:"phone"`
	Email           *string    `json:"email"`
	EmailVerified   bool       `json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	Website         *string    `json:"website"`
	TransportType   string     `json:"transport_type"`
	WorkCountries   []string   `json:"work_countries"`
	Description     *string    `json:"description"`
	LogoURL         *string    `json:"logo_url"`
	Media           []MediaItem `json:"media"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
