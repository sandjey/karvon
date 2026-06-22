package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateCarrierRequest struct {
	Name          string     `json:"name"           binding:"required,min=2,max=200"`
	TransportType string     `json:"transport_type" binding:"required,oneof=railway auto aviation sea"`
	Countries     []string   `json:"countries"      binding:"required,min=1,max=100"`
	Description   *string    `json:"description"    binding:"omitempty,max=2000"`
	LogoURL       *string    `json:"logo_url"       binding:"omitempty,max=500"`
	Website       *string    `json:"website"        binding:"omitempty,max=300"`
	ContactPhone  *string    `json:"contact_phone"  binding:"omitempty,max=50"`
	ContactEmail  *string    `json:"contact_email"  binding:"omitempty,email,max=100"`
	CompanyID     *uuid.UUID `json:"company_id"     binding:"omitempty"`
}

type UpdateCarrierRequest struct {
	Name          *string  `json:"name"           binding:"omitempty,min=2,max=200"`
	TransportType *string  `json:"transport_type" binding:"omitempty,oneof=railway auto aviation sea"`
	Countries     []string `json:"countries"      binding:"omitempty,min=1,max=100"`
	Description   *string  `json:"description"    binding:"omitempty,max=2000"`
	LogoURL       *string  `json:"logo_url"       binding:"omitempty,max=500"`
	Website       *string  `json:"website"        binding:"omitempty,max=300"`
	ContactPhone  *string  `json:"contact_phone"  binding:"omitempty,max=50"`
	ContactEmail  *string  `json:"contact_email"  binding:"omitempty,email,max=100"`
	IsActive      *bool    `json:"is_active"`
}

type CarrierResponse struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	CompanyID     *uuid.UUID `json:"company_id"`
	Name          string     `json:"name"`
	TransportType string     `json:"transport_type"`
	Countries     []string   `json:"countries"`
	Description   *string    `json:"description"`
	LogoURL       *string    `json:"logo_url"`
	Website       *string    `json:"website"`
	ContactPhone  *string    `json:"contact_phone"`
	ContactEmail  *string    `json:"contact_email"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
