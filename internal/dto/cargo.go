package dto

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type MediaItem struct {
	FileURL      string  `json:"file_url" binding:"required"`
	FileType     string  `json:"file_type" binding:"required,oneof=photo video document"`
	OriginalName *string `json:"original_name"`
	SortOrder    int     `json:"sort_order"`
}

// CargoUpsertRequest — тело создания/обновления карточки ТОВАРА.
type CargoUpsertRequest struct {
	CompanyID *uuid.UUID `json:"company_id"`

	// О товаре
	CargoName   string  `json:"cargo_name" binding:"required,min=2,max=200"`
	Category    string  `json:"category" binding:"required,oneof=stroymat food textile metal chemical wood electronics other"`
	Description *string `json:"description"`

	// Объём
	QuantityAvailable *float64 `json:"quantity_available"`
	QuantityUnit      *string  `json:"quantity_unit" binding:"omitempty,oneof=ton places pallet m3"`
	MinOrder          *float64 `json:"min_order"`
	MinOrderUnit      *string  `json:"min_order_unit" binding:"omitempty,oneof=ton places pallet m3"`
	Divisibility      *string  `json:"divisibility" binding:"omitempty,oneof=ftl ltl dogruz"`

	// Упаковка
	Packaging      *string  `json:"packaging" binding:"omitempty,oneof=bulk pallets bags barrels rolls boxes liquid oversized"`
	WeightPerPlace *float64 `json:"weight_per_place"`
	LengthM        *float64 `json:"length_m"`
	WidthM         *float64 `json:"width_m"`
	HeightM        *float64 `json:"height_m"`

	// Цена
	PriceMode     *string  `json:"price_mode" binding:"omitempty,oneof=negotiable fixed announcement"`
	PricePerUnit  *float64 `json:"price_per_unit"`
	Currency      *string  `json:"currency" binding:"omitempty,oneof=UZS USD"`
	VatMode       *string  `json:"vat_mode" binding:"omitempty,oneof=yes no unspecified"`
	PaymentTerms  *string  `json:"payment_terms"`
	PaymentMethod *string  `json:"payment_method"`

	// Место забора
	FromCountry    *string        `json:"from_country"`
	FromCity       *string        `json:"from_city"`
	PickupAddress  *string        `json:"pickup_address"`
	Lat            *float64       `json:"lat"`
	Lng            *float64       `json:"lng"`
	LoadingMethods []string       `json:"loading_methods"`
	WorkingHours   datatypes.JSON `json:"working_hours"`

	// Параметры
	IsADR             *bool    `json:"is_adr"`
	ADRClass          *int     `json:"adr_class" binding:"omitempty,min=1,max=9"`
	HasTempRegime     *bool    `json:"has_temp_regime"`
	TempMin           *float64 `json:"temp_min"`
	TempMax           *float64 `json:"temp_max"`
	RequiredBodyTypes []string `json:"required_body_types"`

	// Логистика
	Incoterms         *string  `json:"incoterms"`
	DeliveryGeography []string `json:"delivery_geography"`
	Permits           []string `json:"permits"`

	CommentForCarrier *string `json:"comment_for_carrier"`

	Media []MediaItem `json:"media"`
}

type SaveTemplateRequest struct {
	TemplateName string `json:"template_name" binding:"required,min=2,max=100"`
}

// CargoStatusRequest — статус (active|archived) и/или наличие.
type CargoStatusRequest struct {
	Status  *string `json:"status" binding:"omitempty,oneof=active archived"`
	InStock *bool   `json:"in_stock"`
}

type CargoStatsResponse struct {
	ViewsCount          int            `json:"views_count"`
	ContactsBoughtCount int            `json:"contacts_bought_count"`
	Favorites           []FavoriteUser `json:"favorites"`
}

type FavoriteUser struct {
	UserName    *string   `json:"user_name"`
	CompanyName *string   `json:"company_name"`
	AddedAt     time.Time `json:"added_at"`
}
