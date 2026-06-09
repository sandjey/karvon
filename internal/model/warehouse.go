package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type WarehouseListing struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID             uuid.UUID      `gorm:"type:uuid;not null;index"`
	Company               Company        `gorm:"foreignKey:CompanyID"`
	UserID                uuid.UUID      `gorm:"type:uuid;not null;index"`
	User                  User           `gorm:"foreignKey:UserID"`
	WarehouseType         string         `gorm:"type:warehouse_type_enum;not null;index"`
	Name                  string         `gorm:"type:varchar(200);not null"`
	Region                *string        `gorm:"type:varchar(100)"`
	Address               *string        `gorm:"type:text"`
	Lat                   *float64       `gorm:"type:decimal"`
	Lng                   *float64       `gorm:"type:decimal"`
	PhoneMain             *string        `gorm:"type:varchar(20)"`
	PhoneExtra            *string        `gorm:"type:varchar(20)"`
	ContactPerson         *string        `gorm:"type:varchar(100)"`
	Email                 *string        `gorm:"type:varchar(100)"`
	Website               *string        `gorm:"type:varchar(200)"`
	Specialization        pq.StringArray `gorm:"type:text[]"`
	AreaTotalM2           *float64       `gorm:"type:decimal;column:area_total_m2"`
	AreaFreeM2            *float64       `gorm:"type:decimal;column:area_free_m2"`
	CeilingHeightM        *float64       `gorm:"type:decimal;column:ceiling_height_m"`
	HeatingType           *string        `gorm:"type:heating_type_enum"`
	StorageType           pq.StringArray `gorm:"type:text[]"`
	TempMin               *float64       `gorm:"type:decimal"`
	TempMax               *float64       `gorm:"type:decimal"`
	ColdChamberTypes      pq.StringArray `gorm:"type:text[]"`
	CustomsLicenseNumber  *string        `gorm:"type:varchar(100)"`
	CustomsLicenseIssued  *time.Time
	CustomsLicenseExpires *time.Time
	Infrastructure        pq.StringArray `gorm:"type:text[]"`
	Services              pq.StringArray `gorm:"type:text[]"`
	WorkingHours          datatypes.JSON `gorm:"type:jsonb"`
	IsBoosted             bool           `gorm:"not null;default:false"`
	BoostExpiresAt        *time.Time
	IsPaid                bool           `gorm:"not null;default:false"`
	Status                string         `gorm:"type:warehouse_status_enum;not null;default:'active';index"`
	ViewsCount            int            `gorm:"not null;default:0"`
	CreatedAt             time.Time
}
