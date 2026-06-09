package model

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	User            User       `gorm:"foreignKey:UserID"`
	Name            string     `gorm:"type:varchar(200);not null"`
	INN             string     `gorm:"type:varchar(20);not null;column:inn"`
	Region          *string    `gorm:"type:varchar(100)"`
	Phone           *string    `gorm:"type:varchar(20)"`
	INNDocURL       *string    `gorm:"type:text;column:inn_doc_url"`
	RegDocURL       *string    `gorm:"type:text;column:reg_doc_url"`
	Status          string     `gorm:"type:company_status;not null;default:'pending';index"`
	RejectionReason *string    `gorm:"type:text"`
	ModeratorID     *uuid.UUID `gorm:"type:uuid"`
	VerifiedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
