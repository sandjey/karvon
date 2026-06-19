package model

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	User   User      `gorm:"foreignKey:UserID"`

	// Компания
	Country       string     `gorm:"type:varchar(100);not null"`            // Страна регистрации
	OrgType       string     `gorm:"type:company_org_type;not null"`        // ООО/АО/ИП/Ltd/GmbH/Co.Ltd
	Name          string     `gorm:"type:varchar(200);not null"`            // Полное юридическое название
	INN           string     `gorm:"type:varchar(30);not null;column:inn"`  // ИНН/БИН/VAT/USCC
	INNVerified   bool       `gorm:"not null;default:false;column:inn_verified"` // Прошёл автопроверку по реестру
	INNVerifiedAt *time.Time `gorm:"column:inn_verified_at"`

	// Контакты
	Phone string `gorm:"type:varchar(20);not null"`
	Email string `gorm:"type:varchar(100);not null"`

	// Адрес
	City       string  `gorm:"type:varchar(100);not null"` // Город / район
	Region     *string `gorm:"type:varchar(100)"`          // Область / провинция (необяз.)
	Street     string  `gorm:"type:varchar(200);not null"` // Улица, дом, офис
	PostalCode *string `gorm:"type:varchar(20);column:postal_code"`

	// Документы
	RegDocURL string  `gorm:"type:text;not null;column:reg_doc_url"` // Свидетельство о регистрации
	InnDocURL *string `gorm:"type:text;column:inn_doc_url"`           // Скан ИНН / свидетельства налогоплательщика

	// Модерация
	Status          string     `gorm:"type:company_status;not null;default:'pending';index"`
	RejectionReason *string    `gorm:"type:text"`
	DocsRequestNote *string    `gorm:"type:text;column:docs_request_note"`
	ModeratorID     *uuid.UUID `gorm:"type:uuid"`
	VerifiedAt      *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
