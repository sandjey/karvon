package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CarrierCompany struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User   *User     `gorm:"foreignKey:UserID" json:"-"`

	// Юридические данные
	OrgType string  `gorm:"type:company_org_type;not null" json:"org_type"`
	Name    string  `gorm:"type:varchar(200);not null" json:"name"`
	INN     string  `gorm:"type:varchar(30);not null;column:inn" json:"inn"`
	Country string  `gorm:"type:varchar(100);not null" json:"country"`
	City    string  `gorm:"type:varchar(100);not null" json:"city"`
	Region  *string `gorm:"type:varchar(100)" json:"region"`

	// Контакты
	Phone           string     `gorm:"type:varchar(20);not null" json:"phone"`
	Email           *string    `gorm:"type:varchar(100)" json:"email"`
	EmailVerified   bool       `gorm:"not null;default:false;column:email_verified" json:"email_verified"`
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at" json:"email_verified_at"`
	Website         *string    `gorm:"type:varchar(300)" json:"website"`

	// Специфика перевозчика
	TransportType string         `gorm:"type:carrier_transport_type_enum;not null" json:"transport_type"`
	WorkCountries pq.StringArray `gorm:"type:text[];column:work_countries;not null" json:"work_countries"`
	Description   *string        `gorm:"type:text" json:"description"`
	LogoURL       *string        `gorm:"type:varchar(500);column:logo_url" json:"logo_url"`

	// Статус
	Status    string    `gorm:"type:varchar(20);not null;default:'active';index" json:"status"` // active|archived
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
