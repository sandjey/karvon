package dto

import "github.com/google/uuid"

type OpenContactRequest struct {
	ListingType string    `json:"listing_type" binding:"required,oneof=cargo warehouse"`
	ListingID   uuid.UUID `json:"listing_id" binding:"required"`
}

type ContactInfo struct {
	Phone         *string `json:"phone"`
	PhoneExtra    *string `json:"phone_extra"`
	WhatsApp      *string `json:"whatsapp"`
	Telegram      *string `json:"telegram"`
	Email         *string `json:"email"`
	ContactPerson *string `json:"contact_person"`
	CompanyName   *string `json:"company_name"`
	City          *string `json:"city"`
	Country       *string `json:"country"`
	TokensSpent   int     `json:"tokens_spent"`
	TokenBalance  int     `json:"token_balance"`
}
