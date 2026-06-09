package model

import (
	"time"

	"github.com/google/uuid"
)

type TokenTransaction struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	User         User       `gorm:"foreignKey:UserID"`
	Type         string     `gorm:"type:token_transaction_type_enum;not null"`
	Amount       int        `gorm:"not null"`
	Reason       string     `gorm:"type:token_reason_enum;not null"`
	ReferenceID  *uuid.UUID `gorm:"type:uuid"`
	BalanceAfter int        `gorm:"not null"`
	CreatedAt    time.Time  `gorm:"index"`
}

type PaymentOrder struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	User           User       `gorm:"foreignKey:UserID"`
	PaymentType    string     `gorm:"type:payment_type_enum;not null"`
	ItemID         *string    `gorm:"type:varchar(100)"`
	Amount         float64    `gorm:"type:decimal;not null"`
	Currency       string     `gorm:"type:payment_currency_enum;not null;default:'UZS'"`
	Status         string     `gorm:"type:payment_status_enum;not null;default:'pending';index"`
	RahmatOrderID  *string    `gorm:"type:varchar(200);index"`
	PaymentMethod  *string    `gorm:"type:varchar(50)"`
	PaidAt         *time.Time
	CreatedAt      time.Time
}

type Subscription struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID    `gorm:"type:uuid;not null;index"`
	User           User         `gorm:"foreignKey:UserID"`
	Plan           string       `gorm:"type:subscription_plan_enum;not null"`
	StartsAt       time.Time    `gorm:"not null"`
	ExpiresAt      time.Time    `gorm:"not null;index"`
	IsActive       bool         `gorm:"not null;default:true;index"`
	PaymentOrderID *uuid.UUID   `gorm:"type:uuid"`
	PaymentOrder   *PaymentOrder `gorm:"foreignKey:PaymentOrderID"`
	CreatedAt      time.Time
}

type PricingConfig struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Key          string     `gorm:"type:varchar(100);uniqueIndex;not null"`
	Label        string     `gorm:"type:varchar(200);not null"`
	PriceUZS     float64    `gorm:"type:decimal;not null;default:0;column:price_uzs"`
	PriceUSD     float64    `gorm:"type:decimal;not null;default:0;column:price_usd"`
	TokensAmount int        `gorm:"not null;default:0"`
	DurationDays int        `gorm:"not null;default:0"`
	IsActive     bool       `gorm:"not null;default:true"`
	UpdatedBy    *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt    time.Time
}
