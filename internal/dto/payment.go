package dto

type CreatePaymentRequest struct {
	PaymentType string `json:"payment_type" binding:"required,oneof=tokens subscription listing"`
	PricingKey  string `json:"pricing_key" binding:"required"`
	Currency    string `json:"currency" binding:"omitempty,oneof=UZS USD"`
	ListingType string `json:"listing_type" binding:"omitempty,oneof=cargo warehouse"`
	ListingID   string `json:"listing_id"`
}

type BoostRequest struct {
	PricingKey string `json:"pricing_key" binding:"required"`
	Currency   string `json:"currency" binding:"omitempty,oneof=UZS USD"`
}

type WebhookRequest struct {
	OrderID   string `json:"order_id" binding:"required"`
	Method    string `json:"method"`
	Signature string `json:"signature"`
}
