package dto

import (
	"time"

	"github.com/google/uuid"
)

type QueueItemResponse struct {
	ID              uuid.UUID  `json:"id"`
	CompanyName     string     `json:"company_name"`
	INN             string     `json:"inn"`
	Status          string     `json:"status"`
	Region          *string    `json:"region,omitempty"`
	INNDocURL       *string    `json:"inn_doc_url,omitempty"`
	RegDocURL       *string    `json:"reg_doc_url,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	UserName        *string    `json:"user_name"`
	UserPhone       string     `json:"user_phone"`
	CreatedAt       time.Time  `json:"created_at"`
	Deadline        time.Time  `json:"deadline"`
	IsUrgent        bool       `json:"is_urgent"`
}

type RejectRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type RequestDocsRequest struct {
	Message string `json:"message" binding:"required"`
}
