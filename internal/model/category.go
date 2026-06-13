package model

import (
	"time"

	"github.com/google/uuid"
)

// CargoCategory — категория товаров, управляется супер-админом (ru/uz/en).
type CargoCategory struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Key       string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"key"`
	LabelRu   string    `gorm:"type:varchar(200);not null" json:"label_ru"`
	LabelUz   string    `gorm:"type:varchar(200);not null" json:"label_uz"`
	LabelEn   string    `gorm:"type:varchar(200);not null" json:"label_en"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
