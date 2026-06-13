package dto

type CreatePaymentRequest struct {
	PaymentType    string `json:"payment_type" binding:"required,oneof=tokens subscription listing"`
	PricingKey     string `json:"pricing_key" binding:"required"`
	Currency       string `json:"currency" binding:"omitempty,oneof=UZS USD"`
	ListingType    string `json:"listing_type" binding:"omitempty,oneof=cargo warehouse"`
	ListingID      string `json:"listing_id"`
	ReturnURL      string `json:"return_url"`       // куда вернуть после оплаты (опционально)
	ReturnErrorURL string `json:"return_error_url"` // куда вернуть при ошибке (опционально)
}

type BoostRequest struct {
	PricingKey     string `json:"pricing_key" binding:"required"`
	Currency       string `json:"currency" binding:"omitempty,oneof=UZS USD"`
	ReturnURL      string `json:"return_url"`
	ReturnErrorURL string `json:"return_error_url"`
}

// MulticardCallback — полное тело callback-запроса (PaymentModel) от Multicard.
// Статусы: draft | progress | billing | success | error | revert.
type MulticardCallback struct {
	UUID             string `json:"uuid"`              // UUID транзакции в Multicard
	StoreInvoiceID   string `json:"store_invoice_id"`  // наш UUID заказа
	Status           string `json:"status"`            // draft|progress|billing|success|error|revert
	PS               string `json:"ps"`                // uzcard|humo|visa|mastercard|...
	PaymentAmount    int64  `json:"payment_amount"`    // сумма оплаты в тийинах
	TotalAmount      int64  `json:"total_amount"`      // итого с комиссией в тийинах
	CommissionAmount int64  `json:"commission_amount"` // комиссия в тийинах
	CardPan          string `json:"card_pan"`          // маскированный номер карты
	Phone            string `json:"phone"`             // телефон плательщика
	ReceiptURL       string `json:"receipt_url"`       // URL ОФД-чека
	OtpHash          string `json:"otp_hash"`          // хэш OTP-подтверждения
	PaymentTime      string `json:"payment_time"`      // время оплаты
}
