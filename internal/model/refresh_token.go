package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	User      User      `gorm:"foreignKey:UserID"`
	Token     string    `gorm:"type:text;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
}
