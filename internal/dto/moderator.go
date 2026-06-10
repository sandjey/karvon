package dto

import (
	"time"

	"github.com/google/uuid"
)

type QueueItemResponse struct {
	ID              uuid.UUID `json:"id"`
	CompanyName     string    `json:"company_name"`
	Country         string    `json:"country"`
	OrgType         string    `json:"org_type"`
	INN             string    `json:"inn"`
	INNVerified     bool      `json:"inn_verified"`
	Status          string    `json:"status"`
	Phone           string    `json:"phone"`
	Email           string    `json:"email"`
	City            string    `json:"city"`
	Region          *string   `json:"region,omitempty"`
	Street          string    `json:"street"`
	RegDocURL       string    `json:"reg_doc_url,omitempty"`
	RejectionReason *string   `json:"rejection_reason,omitempty"`
	DocsRequestNote *string   `json:"docs_request_note,omitempty"`
	UserName        *string   `json:"user_name"`
	UserPhone       string    `json:"user_phone"`
	CreatedAt       time.Time `json:"created_at"`
	Deadline        time.Time `json:"deadline"`
	IsUrgent        bool      `json:"is_urgent"`
}

type RejectRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type RequestDocsRequest struct {
	Message string `json:"message" binding:"required"`
}
