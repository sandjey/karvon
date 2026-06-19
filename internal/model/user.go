package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Phone        string    `gorm:"type:varchar(30);uniqueIndex;not null"          json:"phone"`
	Name         *string   `gorm:"type:varchar(100)"                              json:"name"`
	Email        *string   `gorm:"type:varchar(100)"                              json:"email"`
	ExtraPhone   *string   `gorm:"type:varchar(30);column:extra_phone"            json:"extra_phone"`
	WhatsApp     *string   `gorm:"type:varchar(20);column:whatsapp"               json:"whatsapp"`
	Telegram     *string   `gorm:"type:varchar(50)"                               json:"telegram"`
	City         *string   `gorm:"type:varchar(100)"                              json:"city"`
	Country      *string   `gorm:"type:varchar(100)"                              json:"country"`
	TokenBalance int       `gorm:"not null;default:5"                             json:"token_balance"`
	Role         string    `gorm:"type:user_role;not null;default:'user'"         json:"role"`
	IsBlocked    bool      `gorm:"not null;default:false"                         json:"is_blocked"`
	TokenVersion int       `gorm:"not null;default:0"                             json:"-"`
	// Admin panel credentials (moderators only)
	AdminLogin        *string `gorm:"type:varchar(100);uniqueIndex" json:"admin_login,omitempty"`
	AdminPasswordHash *string `gorm:"type:varchar(200)"             json:"-"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type OTPCode struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Phone     string    `gorm:"type:varchar(20);not null;index"`
	Code      string    `gorm:"type:varchar(6);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"not null;default:false"`
	Attempts  int       `gorm:"not null;default:0"`
	CreatedAt time.Time
}
