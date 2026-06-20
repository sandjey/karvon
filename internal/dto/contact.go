package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

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

type ContactHistoryListing struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	CompanyName string    `json:"company_name"`
	FromCity    *string   `json:"from_city"`
	ToCity      *string   `json:"to_city"`
	Status      string    `json:"status"`
}

type ContactHistoryContact struct {
	Phone    *string `json:"phone"`
	WhatsApp *string `json:"whatsapp"`
	Telegram *string `json:"telegram"`
}

type ContactHistoryItem struct {
	ID          uuid.UUID             `json:"id"`
	ListingType string                `json:"listing_type"`
	Listing     ContactHistoryListing `json:"listing"`
	Contact     ContactHistoryContact `json:"contact"`
	TokensSpent int                   `json:"tokens_spent"`
	ViewReason  string                `json:"view_reason"`
	ViewedAt    time.Time             `json:"viewed_at"`
	FreeUntil   time.Time             `json:"free_until"`
}

type TokenTxListing struct {
	ID       uuid.UUID `json:"id"`
	Type     string    `json:"type"`
	Name     string    `json:"name"`
	FromCity *string   `json:"from_city"`
}

type TokenTransactionResponse struct {
	ID            uuid.UUID       `json:"id"`
	Type          string          `json:"type"`
	Amount        int             `json:"amount"`
	Reason        string          `json:"reason"`
	BalanceBefore int             `json:"balance_before"`
	BalanceAfter  int             `json:"balance_after"`
	Listing       *TokenTxListing `json:"listing"`
	Meta          json.RawMessage `json:"meta"`
	Description   string          `json:"description"`
	CreatedAt     time.Time       `json:"created_at"`
}
