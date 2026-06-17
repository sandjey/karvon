package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Phone        string    `gorm:"type:varchar(30);uniqueIndex;not null"`
	Name         *string   `gorm:"type:varchar(100)"`
	Email        *string   `gorm:"type:varchar(100)"`
	ExtraPhone   *string   `gorm:"type:varchar(30);column:extra_phone"`
	WhatsApp     *string   `gorm:"type:varchar(20);column:whatsapp"`
	Telegram     *string   `gorm:"type:varchar(50)"`
	City         *string   `gorm:"type:varchar(100)"`
	Country      *string   `gorm:"type:varchar(100)"`
	TokenBalance int       `gorm:"not null;default:5"`
	Role         string    `gorm:"type:user_role;not null;default:'user'"`
	IsBlocked    bool      `gorm:"not null;default:false"`
	TokenVersion int       `gorm:"not null;default:0"`
	// Admin panel credentials (moderators only)
	AdminLogin        *string `gorm:"type:varchar(100);uniqueIndex"`
	AdminPasswordHash *string `gorm:"type:varchar(200)"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
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
