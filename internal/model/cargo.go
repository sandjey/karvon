package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CargoListing struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Company   *Company  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"-"`

	Type          string         `gorm:"type:cargo_type_enum;not null" json:"type"`
	CargoName     *string        `gorm:"type:varchar(200)" json:"cargo_name"`
	CargoType     *string        `gorm:"type:varchar(50)" json:"cargo_type"`
	IsADR         bool           `gorm:"not null;default:false;column:is_adr" json:"is_adr"`
	HasTempRegime bool           `gorm:"not null;default:false" json:"has_temp_regime"`
	TempMin       *float64       `gorm:"type:decimal" json:"temp_min"`
	TempMax       *float64       `gorm:"type:decimal" json:"temp_max"`
	WeightTon     *float64       `gorm:"type:decimal" json:"weight_ton"`
	VolumeM3      *float64       `gorm:"type:decimal;column:volume_m3" json:"volume_m3"`
	PlacesCount   *int           `gorm:"type:int" json:"places_count"`
	LengthM       *float64       `gorm:"type:decimal;column:length_m" json:"length_m"`
	WidthM        *float64       `gorm:"type:decimal;column:width_m" json:"width_m"`
	HeightM       *float64       `gorm:"type:decimal;column:height_m" json:"height_m"`
	LoadingType   *string        `gorm:"type:loading_type_enum" json:"loading_type"`
	BodyTypes     pq.StringArray `gorm:"type:text[]" json:"body_types"`
	VehicleType   *string        `gorm:"type:varchar(50)" json:"vehicle_type"`
	TractorAxles  *int           `json:"tractor_axles"`
	TrailerAxles  *int           `json:"trailer_axles"`
	OnlyRecoupling bool          `gorm:"not null;default:false" json:"only_recoupling"`

	FromCity      *string    `gorm:"type:varchar(100);index" json:"from_city"`
	FromCountry   *string    `gorm:"type:varchar(100)" json:"from_country"`
	FromDate      *time.Time `json:"from_date"`
	LoadingMethod *string    `gorm:"type:varchar(50)" json:"loading_method"`
	ToCity        *string    `gorm:"type:varchar(100);index" json:"to_city"`
	ToCountry     *string    `gorm:"type:varchar(100)" json:"to_country"`
	ToDate        *time.Time `json:"to_date"`

	UnloadingMethod  *string        `gorm:"type:varchar(50)" json:"unloading_method"`
	TransitCountries pq.StringArray `gorm:"type:text[]" json:"transit_countries"`
	BorderCrossing   *string        `gorm:"type:varchar(100)" json:"border_crossing"`
	CustomsType      *string        `gorm:"type:customs_type_enum" json:"customs_type"`
	Incoterms        *string        `gorm:"type:varchar(20)" json:"incoterms"`
	Permits          pq.StringArray `gorm:"type:text[]" json:"permits"`

	RateMode          *string  `gorm:"type:rate_mode_enum" json:"rate_mode"`
	RateAmount        *float64 `gorm:"type:decimal" json:"rate_amount"`
	RateCurrency      *string  `gorm:"type:currency_enum" json:"rate_currency"`
	RateVAT           bool     `gorm:"not null;default:false;column:rate_vat" json:"rate_vat"`
	PaymentTerms      *string  `gorm:"type:text" json:"payment_terms"`
	PaymentMethod     *string  `gorm:"type:varchar(50)" json:"payment_method"`
	PrepaymentPercent *int     `json:"prepayment_percent"`

	DurationHours int        `gorm:"not null;default:24" json:"duration_hours"`
	ExpiresAt     *time.Time `gorm:"index" json:"expires_at"`
	Notes         *string    `gorm:"type:text" json:"notes"`
	Status        string     `gorm:"type:cargo_status_enum;not null;default:'active';index" json:"status"`

	IsTemplate   bool    `gorm:"not null;default:false" json:"is_template"`
	TemplateName *string `gorm:"type:varchar(100)" json:"template_name"`

	IsBoosted      bool       `gorm:"not null;default:false" json:"is_boosted"`
	BoostExpiresAt *time.Time `json:"boost_expires_at"`
	IsPaid         bool       `gorm:"not null;default:false" json:"is_paid"`

	ViewsCount          int `gorm:"not null;default:0" json:"views_count"`
	ContactsBoughtCount int `gorm:"not null;default:0" json:"contacts_bought_count"`

	Waypoints []CargoWaypoint `gorm:"foreignKey:CargoListingID" json:"waypoints,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CargoWaypoint struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CargoListingID uuid.UUID `gorm:"type:uuid;not null;index" json:"cargo_listing_id"`
	SortOrder      int       `gorm:"not null;default:0" json:"sort_order"`
	WaypointType   string    `gorm:"type:waypoint_type_enum;not null" json:"waypoint_type"`
	City           *string   `gorm:"type:varchar(100)" json:"city"`
	Country        *string   `gorm:"type:varchar(100)" json:"country"`
}
