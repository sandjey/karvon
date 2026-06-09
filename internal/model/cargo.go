package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CargoListing struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID           uuid.UUID      `gorm:"type:uuid;not null;index"`
	Company             Company        `gorm:"foreignKey:CompanyID"`
	UserID              uuid.UUID      `gorm:"type:uuid;not null;index"`
	User                User           `gorm:"foreignKey:UserID"`
	Type                string         `gorm:"type:cargo_type_enum;not null"`
	CargoName           *string        `gorm:"type:varchar(200)"`
	CargoType           *string        `gorm:"type:varchar(50)"`
	IsADR               bool           `gorm:"not null;default:false;column:is_adr"`
	HasTempRegime       bool           `gorm:"not null;default:false"`
	TempMin             *float64       `gorm:"type:decimal"`
	TempMax             *float64       `gorm:"type:decimal"`
	WeightTon           *float64       `gorm:"type:decimal"`
	VolumeM3            *float64       `gorm:"type:decimal;column:volume_m3"`
	PlacesCount         *int           `gorm:"type:int"`
	LengthM             *float64       `gorm:"type:decimal;column:length_m"`
	WidthM              *float64       `gorm:"type:decimal;column:width_m"`
	HeightM             *float64       `gorm:"type:decimal;column:height_m"`
	LoadingType         *string        `gorm:"type:loading_type_enum"`
	BodyTypes           pq.StringArray `gorm:"type:text[]"`
	VehicleType         *string        `gorm:"type:varchar(50)"`
	TractorAxles        *int
	TrailerAxles        *int
	OnlyRecoupling      bool    `gorm:"not null;default:false"`
	FromCity            *string `gorm:"type:varchar(100);index"`
	FromCountry         *string `gorm:"type:varchar(100)"`
	FromDate            *time.Time
	LoadingMethod       *string `gorm:"type:varchar(50)"`
	ToCity              *string `gorm:"type:varchar(100);index"`
	ToCountry           *string `gorm:"type:varchar(100)"`
	ToDate              *time.Time
	UnloadingMethod     *string        `gorm:"type:varchar(50)"`
	TransitCountries    pq.StringArray `gorm:"type:text[]"`
	BorderCrossing      *string        `gorm:"type:varchar(100)"`
	CustomsType         *string        `gorm:"type:customs_type_enum"`
	Incoterms           *string        `gorm:"type:varchar(20)"`
	Permits             pq.StringArray `gorm:"type:text[]"`
	RateMode            *string        `gorm:"type:rate_mode_enum"`
	RateAmount          *float64       `gorm:"type:decimal"`
	RateCurrency        *string        `gorm:"type:currency_enum"`
	RateVAT             bool           `gorm:"not null;default:false;column:rate_vat"`
	PaymentTerms        *string        `gorm:"type:text"`
	PaymentMethod       *string        `gorm:"type:varchar(50)"`
	PrepaymentPercent   *int
	DurationHours       int        `gorm:"not null;default:24"`
	ExpiresAt           *time.Time `gorm:"index"`
	Notes               *string    `gorm:"type:text"`
	Status              string     `gorm:"type:cargo_status_enum;not null;default:'active';index"`
	IsTemplate          bool       `gorm:"not null;default:false"`
	TemplateName        *string    `gorm:"type:varchar(100)"`
	IsBoosted           bool       `gorm:"not null;default:false"`
	BoostExpiresAt      *time.Time
	IsPaid              bool `gorm:"not null;default:false"`
	ViewsCount          int  `gorm:"not null;default:0"`
	ContactsBoughtCount int  `gorm:"not null;default:0"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type CargoWaypoint struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CargoListingID uuid.UUID `gorm:"type:uuid;not null;index"`
	SortOrder      int       `gorm:"not null;default:0"`
	WaypointType   string    `gorm:"type:waypoint_type_enum;not null"`
	City           *string   `gorm:"type:varchar(100)"`
	Country        *string   `gorm:"type:varchar(100)"`
}
