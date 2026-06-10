package dto

import (
	"time"

	"github.com/google/uuid"
)

type CargoWaypointRequest struct {
	SortOrder    int     `json:"sort_order"`
	WaypointType string  `json:"waypoint_type" binding:"required,oneof=border customs reload passthrough"`
	City         *string `json:"city"`
	Country      *string `json:"country"`
}

// CargoUpsertRequest — тело для создания/обновления груза. Все опциональные поля — указатели.
type CargoUpsertRequest struct {
	Type          string   `json:"type" binding:"required,oneof=domestic international"`
	CargoName     *string  `json:"cargo_name"`
	CargoType     *string  `json:"cargo_type"`
	IsADR         *bool    `json:"is_adr"`
	HasTempRegime *bool    `json:"has_temp_regime"`
	TempMin       *float64 `json:"temp_min"`
	TempMax       *float64 `json:"temp_max"`
	WeightTon     *float64 `json:"weight_ton"`
	VolumeM3      *float64 `json:"volume_m3"`
	PlacesCount   *int     `json:"places_count"`
	LengthM       *float64 `json:"length_m"`
	WidthM        *float64 `json:"width_m"`
	HeightM       *float64 `json:"height_m"`
	LoadingType   *string  `json:"loading_type" binding:"omitempty,oneof=ftl ltl partial"`
	BodyTypes     []string `json:"body_types"`
	VehicleType   *string  `json:"vehicle_type"`
	TractorAxles  *int     `json:"tractor_axles"`
	TrailerAxles  *int     `json:"trailer_axles"`
	OnlyRecoupling *bool   `json:"only_recoupling"`

	FromCity      *string    `json:"from_city"`
	FromCountry   *string    `json:"from_country"`
	FromDate      *time.Time `json:"from_date"`
	LoadingMethod *string    `json:"loading_method"`
	ToCity        *string    `json:"to_city"`
	ToCountry     *string    `json:"to_country"`
	ToDate        *time.Time `json:"to_date"`

	UnloadingMethod  *string  `json:"unloading_method"`
	TransitCountries []string `json:"transit_countries"`
	BorderCrossing   *string  `json:"border_crossing"`
	CustomsType      *string  `json:"customs_type" binding:"omitempty,oneof=export import transit"`
	Incoterms        *string  `json:"incoterms"`
	Permits          []string `json:"permits"`

	RateMode          *string  `json:"rate_mode" binding:"omitempty,oneof=negotiable fixed announcement"`
	RateAmount        *float64 `json:"rate_amount"`
	RateCurrency      *string  `json:"rate_currency" binding:"omitempty,oneof=UZS USD"`
	RateVAT           *bool    `json:"rate_vat"`
	PaymentTerms      *string  `json:"payment_terms"`
	PaymentMethod     *string  `json:"payment_method"`
	PrepaymentPercent *int     `json:"prepayment_percent"`

	DurationHours *int                   `json:"duration_hours"`
	Notes         *string                `json:"notes"`
	CompanyID     *uuid.UUID             `json:"company_id"`
	Waypoints     []CargoWaypointRequest `json:"waypoints"`
}

type SaveTemplateRequest struct {
	TemplateName string `json:"template_name" binding:"required,min=2,max=100"`
}

type CargoStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active archived completed"`
}

type CargoStatsResponse struct {
	ViewsCount          int            `json:"views_count"`
	ContactsBoughtCount int            `json:"contacts_bought_count"`
	Favorites           []FavoriteUser `json:"favorites"`
}

type FavoriteUser struct {
	UserName  *string   `json:"user_name"`
	AddedAt   time.Time `json:"added_at"`
}
