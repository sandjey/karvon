package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateCompanyRequest struct {
	Name      string  `json:"name"       binding:"required,min=2,max=200"`
	INN       string  `json:"inn"        binding:"required,min=9,max=20"`
	Region    *string `json:"region"     binding:"omitempty,max=100"`
	Phone     *string `json:"phone"      binding:"omitempty,max=20"`
	INNDocURL *string `json:"inn_doc_url"`
	RegDocURL *string `json:"reg_doc_url"`
}

type UpdateCompanyRequest struct {
	Name      *string `json:"name"       binding:"omitempty,min=2,max=200"`
	Region    *string `json:"region"     binding:"omitempty,max=100"`
	Phone     *string `json:"phone"      binding:"omitempty,max=20"`
	INNDocURL *string `json:"inn_doc_url"`
	RegDocURL *string `json:"reg_doc_url"`
}

type CompanyResponse struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	INN             string     `json:"inn"`
	Region          *string    `json:"region"`
	Phone           *string    `json:"phone"`
	INNDocURL       *string    `json:"inn_doc_url"`
	RegDocURL       *string    `json:"reg_doc_url"`
	Status          string     `json:"status"`
	RejectionReason *string    `json:"rejection_reason"`
	VerifiedAt      *time.Time `json:"verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
