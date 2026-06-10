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

// MulticardCallback — тело callback-запроса от Multicard при успешной оплате инвойса.
type MulticardCallback struct {
	StoreInvoiceID string `json:"store_invoice_id"` // наш UUID заказа
	Status         string `json:"status"`            // success|error|revert|draft|progress|billing
	PS             string `json:"ps"`                // uzcard|humo|visa|mastercard|...
	UUID           string `json:"uuid"`              // UUID транзакции в Multicard
}
