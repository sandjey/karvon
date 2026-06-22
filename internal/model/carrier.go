package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CarrierCompany struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User      *User      `gorm:"foreignKey:UserID" json:"-"`
	CompanyID *uuid.UUID `gorm:"type:uuid;index" json:"company_id"`
	Company   *Company   `gorm:"foreignKey:CompanyID" json:"company,omitempty"`

	Name          string         `gorm:"type:varchar(200);not null" json:"name"`
	TransportType string         `gorm:"type:carrier_transport_type_enum;not null" json:"transport_type"` // railway|auto|aviation|sea
	Countries     pq.StringArray `gorm:"type:text[];not null" json:"countries"`                           // ISO-коды стран, макс 100
	Description   *string        `gorm:"type:text" json:"description"`
	LogoURL       *string        `gorm:"type:varchar(500)" json:"logo_url"`
	Website       *string        `gorm:"type:varchar(300)" json:"website"`
	ContactPhone  *string        `gorm:"type:varchar(50)" json:"contact_phone"`
	ContactEmail  *string        `gorm:"type:varchar(100)" json:"contact_email"`
	IsActive      bool           `gorm:"not null;default:true" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
