package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type FleetVehicle struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;index"`
	User             User           `gorm:"foreignKey:UserID"`
	Composition      *string        `gorm:"type:vehicle_composition_enum"`
	Brand            *string        `gorm:"type:varchar(100)"`
	Model            *string        `gorm:"type:varchar(100)"`
	PlateNumber      *string        `gorm:"type:varchar(20)"`
	Year             *int
	CountryReg       *string        `gorm:"type:varchar(100)"`
	BodyType         *string        `gorm:"type:varchar(50)"`
	VolumeM3         *float64       `gorm:"type:decimal;column:volume_m3"`
	LoadCapacityTon  *float64       `gorm:"type:decimal"`
	LengthM          *float64       `gorm:"type:decimal;column:length_m"`
	WidthM           *float64       `gorm:"type:decimal;column:width_m"`
	HeightM          *float64       `gorm:"type:decimal;column:height_m"`
	TractorAxles     *int
	TrailerAxles     *int
	HasTIR           bool           `gorm:"not null;default:false;column:has_tir"`
	HasCMR           bool           `gorm:"not null;default:false;column:has_cmr"`
	HasADR           bool           `gorm:"not null;default:false;column:has_adr"`
	HasEKMT          bool           `gorm:"not null;default:false;column:has_ekmt"`
	AllowedCountries pq.StringArray `gorm:"type:text[]"`
	DriverLanguage   pq.StringArray `gorm:"type:text[]"`
	CreatedAt        time.Time
}

type TransportListing struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	VehicleID       uuid.UUID `gorm:"type:uuid;not null;index"`
	Vehicle         FleetVehicle `gorm:"foreignKey:VehicleID"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index"`
	User            User      `gorm:"foreignKey:UserID"`
	CurrentLocation *string   `gorm:"type:varchar(150)"`
	FromCity        *string   `gorm:"type:varchar(100)"`
	ToCity          *string   `gorm:"type:varchar(100)"`
	ReadyDate       *time.Time
	Status          string    `gorm:"type:transport_status_enum;not null;default:'active';index"`
	IsBoosted       bool      `gorm:"not null;default:false"`
	BoostExpiresAt  *time.Time
	IsPaid          bool      `gorm:"not null;default:false"`
	ViewsCount      int       `gorm:"not null;default:0"`
	CreatedAt       time.Time
}
