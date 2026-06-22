package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateCompanyRequest struct {
	Country    string  `json:"country"     binding:"required,max=100"`
	OrgType    string  `json:"org_type"    binding:"required,max=50"`
	Name       string  `json:"name"        binding:"required,min=2,max=200"`
	INN        string  `json:"inn"         binding:"required,min=3,max=30"`
	Phone      string  `json:"phone"       binding:"required,max=30"`
	Email      string  `json:"email"       binding:"omitempty,email,max=100"`
	City       string  `json:"city"        binding:"required,max=100"`
	Region     *string `json:"region"      binding:"omitempty,max=100"`
	Street     string  `json:"street"      binding:"required,max=200"`
	PostalCode *string `json:"postal_code" binding:"omitempty,max=20"`
	RegDocURL  string  `json:"reg_doc_url" binding:"required"`
	InnDocURL  *string `json:"inn_doc_url" binding:"omitempty"`
}

type UpdateCompanyRequest struct {
	Country    *string `json:"country"     binding:"omitempty,max=100"`
	OrgType    *string `json:"org_type"    binding:"omitempty,max=50"`
	Name       *string `json:"name"        binding:"omitempty,min=2,max=200"`
	Phone      *string `json:"phone"       binding:"omitempty,max=30"`
	Email      *string `json:"email"       binding:"omitempty,email,max=100"`
	City       *string `json:"city"        binding:"omitempty,max=100"`
	Region     *string `json:"region"      binding:"omitempty,max=100"`
	Street     *string `json:"street"      binding:"omitempty,max=200"`
	PostalCode *string `json:"postal_code" binding:"omitempty,max=20"`
	RegDocURL  *string `json:"reg_doc_url"`
	InnDocURL  *string `json:"inn_doc_url"`
}

type CompanyResponse struct {
	ID              uuid.UUID  `json:"id"`
	Country         string     `json:"country"`
	OrgType         string     `json:"org_type"`
	Name            string     `json:"name"`
	INN             string     `json:"inn"`
	INNVerified     bool       `json:"inn_verified"`
	Phone           string     `json:"phone"`
	Email           string     `json:"email"`
	EmailVerified   bool       `json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	City            string     `json:"city"`
	Region          *string    `json:"region"`
	Street          string     `json:"street"`
	PostalCode      *string    `json:"postal_code"`
	RegDocURL       string     `json:"reg_doc_url"`
	InnDocURL       *string    `json:"inn_doc_url"`
	Status          string     `json:"status"`
	RejectionReason *string    `json:"rejection_reason"`
	DocsRequestNote *string    `json:"docs_request_note"`
	VerifiedAt      *time.Time `json:"verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
