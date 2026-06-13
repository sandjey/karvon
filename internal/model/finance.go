package model

import (
	"time"

	"github.com/google/uuid"
)

type TokenTransaction struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Type         string     `gorm:"type:token_transaction_type_enum;not null" json:"type"`
	Amount       int        `gorm:"not null" json:"amount"`
	Reason       string     `gorm:"type:token_reason_enum;not null" json:"reason"`
	ReferenceID  *uuid.UUID `gorm:"type:uuid" json:"reference_id"`
	BalanceAfter int        `gorm:"not null" json:"balance_after"`
	CreatedAt    time.Time  `gorm:"index" json:"created_at"`
}

type PaymentOrder struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	PaymentType      string     `gorm:"type:payment_type_enum;not null" json:"payment_type"`
	ItemID           *string    `gorm:"type:varchar(100)" json:"item_id"`
	Amount           float64    `gorm:"type:decimal;not null" json:"amount"`
	Currency         string     `gorm:"type:payment_currency_enum;not null;default:'UZS'" json:"currency"`
	Status           string     `gorm:"type:payment_status_enum;not null;default:'pending';index" json:"status"`
	// RahmatOrderID хранит UUID инвойса Multicard (возвращается при создании инвойса)
	RahmatOrderID    *string    `gorm:"type:varchar(200);index" json:"rahmat_order_id"`
	// PaymentUUID хранит UUID транзакции Multicard из webhook-callback
	PaymentUUID      *string    `gorm:"type:varchar(200)" json:"payment_uuid,omitempty"`
	PaymentMethod    *string    `gorm:"type:varchar(50)" json:"payment_method"`
	Phone            *string    `gorm:"type:varchar(30)" json:"phone,omitempty"`
	ReceiptURL       *string    `gorm:"type:varchar(500)" json:"receipt_url,omitempty"`
	TotalAmountTiyin *int64     `gorm:"type:bigint" json:"total_amount_tiyin,omitempty"`
	CommissionTiyin  *int64     `gorm:"type:bigint" json:"commission_tiyin,omitempty"`
	PaidAt           *time.Time `json:"paid_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

type Subscription struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID     `gorm:"type:uuid;not null;index" json:"user_id"`
	Plan           string        `gorm:"type:subscription_plan_enum;not null" json:"plan"`
	StartsAt       time.Time     `gorm:"not null" json:"starts_at"`
	ExpiresAt      time.Time     `gorm:"not null;index" json:"expires_at"`
	IsActive       bool          `gorm:"not null;default:true;index" json:"is_active"`
	PaymentOrderID *uuid.UUID    `gorm:"type:uuid" json:"payment_order_id"`
	PaymentOrder   *PaymentOrder `gorm:"foreignKey:PaymentOrderID" json:"-"`
	CreatedAt      time.Time     `json:"created_at"`
}

type PricingConfig struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Key          string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Label        string     `gorm:"type:varchar(200);not null" json:"label"`
	PriceUZS     float64    `gorm:"type:decimal;not null;default:0;column:price_uzs" json:"price_uzs"`
	PriceUSD     float64    `gorm:"type:decimal;not null;default:0;column:price_usd" json:"price_usd"`
	TokensAmount int        `gorm:"not null;default:0" json:"tokens_amount"`
	DurationDays int        `gorm:"not null;default:0" json:"duration_days"`
	IsActive     bool       `gorm:"not null;default:true" json:"is_active"`
	UpdatedBy    *uuid.UUID `gorm:"type:uuid" json:"updated_by"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
