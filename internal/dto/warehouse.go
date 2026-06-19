package dto

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)


type WarehouseUpsertRequest struct {
	CompanyID *uuid.UUID `json:"company_id"`

	WarehouseType *string `json:"warehouse_type" binding:"omitempty,oneof=regular cold customs"`
	Name          *string `json:"name"           binding:"omitempty,min=2,max=200"`
	Description   *string `json:"description"`
	Country       *string `json:"country"`
	City          *string `json:"city"`
	Region        *string `json:"region"`
	Address       *string `json:"address"`
	Lat           *float64 `json:"lat"`
	Lng           *float64 `json:"lng"`

	PhoneMain     *string `json:"phone_main"`
	PhoneExtra    *string `json:"phone_extra"`
	ContactPerson *string `json:"contact_person" binding:"omitempty"`
	Email         *string `json:"email" binding:"omitempty,email"`
	Website       *string `json:"website"`

	Specialization []string `json:"specialization"`
	AreaTotalM2    *float64 `json:"area_total_m2"`
	AreaFreeM2     *float64 `json:"area_free_m2"`
	CeilingHeightM *float64 `json:"ceiling_height_m"`
	HeatingType    *string  `json:"heating_type" binding:"omitempty,oneof=heated unheated open closed"`
	StorageType    []string `json:"storage_type"`

	TempMin          *float64 `json:"temp_min"`
	TempMax          *float64 `json:"temp_max"`
	ColdChamberTypes []string `json:"cold_chamber_types"`

	CustomsLicenseNumber   *string    `json:"customs_license_number"`
	CustomsLicenseIssued   *time.Time `json:"customs_license_issued"`
	CustomsLicenseExpires  *time.Time `json:"customs_license_expires"`
	CustomsSpecialServices []string   `json:"customs_special_services"`

	Infrastructure []string       `json:"infrastructure"`
	Services       []string       `json:"services"`
	WorkingHours   datatypes.JSON `json:"working_hours"`

	PricePerM2        *float64 `json:"price_per_m2"`
	PriceCurrency     *string  `json:"price_currency"`
	MinRentPeriodDays *int     `json:"min_rent_period_days"`

	Media []MediaItem `json:"media"`
}

type WarehouseStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active archived"`
}

type WarehouseStatsResponse struct {
	ViewsCount          int            `json:"views_count"`
	ContactsBoughtCount int            `json:"contacts_bought_count"`
	Favorites           []FavoriteUser `json:"favorites"`
}
