package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// CargoContact — контактные данные, доступные после покупки просмотра.
type CargoContact struct {
	IsUnlocked  bool    `json:"is_unlocked"`
	Phone       *string `json:"phone,omitempty"`
	WhatsApp    *string `json:"whatsapp,omitempty"`
	Telegram    *string `json:"telegram,omitempty"`
	Email       *string `json:"email,omitempty"`
	CompanyName *string `json:"company_name,omitempty"`
	City        *string `json:"city,omitempty"`
	Country     *string `json:"country,omitempty"`
}

// CargoListing — карточка ТОВАРА (маркетплейс-модель). Грузовладелец публикует товар,
// перевозчик находит и покупает контакт. Карточка постоянная (не истекает по таймеру).
type CargoListing struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Company   *Company  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"-"`

	// ── О товаре ──────────────────────────────────────────────
	CargoName   string  `gorm:"type:varchar(200);not null" json:"cargo_name"`             // Наименование товара
	Category    string  `gorm:"type:varchar(50);not null;index" json:"category"` // Категория из таблицы cargo_categories
	Description *string `gorm:"type:text" json:"description"`                            // Описание/состав/характеристики

	// ── Объём ─────────────────────────────────────────────────
	QuantityAvailable *float64 `gorm:"type:decimal" json:"quantity_available"`            // Всего в наличии
	QuantityUnit      *string  `gorm:"type:quantity_unit_enum" json:"quantity_unit"`      // ton|places|pallet|m3
	MinOrder          *float64 `gorm:"type:decimal" json:"min_order"`                     // Мин. заказ
	MinOrderUnit      *string  `gorm:"type:quantity_unit_enum" json:"min_order_unit"`
	Divisibility      *string  `gorm:"type:divisibility_enum" json:"divisibility"`        // ftl|ltl|dogruz

	// ── Упаковка ──────────────────────────────────────────────
	Packaging     *string  `gorm:"type:packaging_enum" json:"packaging"` // bulk|pallets|bags|barrels|rolls|boxes|liquid|oversized
	WeightPerPlace *float64 `gorm:"type:decimal" json:"weight_per_place"`
	LengthM       *float64 `gorm:"type:decimal;column:length_m" json:"length_m"`
	WidthM        *float64 `gorm:"type:decimal;column:width_m" json:"width_m"`
	HeightM       *float64 `gorm:"type:decimal;column:height_m" json:"height_m"`

	// ── Цена (видна в листинге без покупки контакта) ──────────
	PriceMode     *string  `gorm:"type:rate_mode_enum;column:rate_mode" json:"price_mode"`  // negotiable|fixed|announcement(=on_request)
	PricePerUnit  *float64 `gorm:"type:decimal;column:rate_amount" json:"price_per_unit"`
	Currency      *string  `gorm:"type:currency_enum;column:rate_currency" json:"currency"`
	VatMode       *string  `gorm:"type:vat_mode_enum" json:"vat_mode"`                      // yes|no|unspecified
	PaymentTerms  *string  `gorm:"type:varchar(50)" json:"payment_terms"`                  // prepaid|on_fact|deferred
	PaymentMethod *string  `gorm:"type:varchar(50)" json:"payment_method"`                 // cash|bank|agreement

	// ── Место забора ──────────────────────────────────────────
	FromCountry    *string        `gorm:"type:varchar(100)" json:"from_country"`
	FromCity       *string        `gorm:"type:varchar(100);index" json:"from_city"`
	ToCountry      *string        `gorm:"type:varchar(100)" json:"to_country"`
	ToCity         *string        `gorm:"type:varchar(100);index" json:"to_city"`
	PickupAddress  *string        `gorm:"type:text" json:"pickup_address"`
	Lat            *float64       `gorm:"type:decimal" json:"lat"`
	Lng            *float64       `gorm:"type:decimal" json:"lng"`
	LoadingMethods pq.StringArray `gorm:"type:text[]" json:"loading_methods"` // self|loader|crane
	WorkingHours   datatypes.JSON `gorm:"type:jsonb" json:"working_hours,omitempty"`

	// ── Параметры ─────────────────────────────────────────────
	IsADR            bool           `gorm:"not null;default:false;column:is_adr" json:"is_adr"`
	ADRClass         *int           `gorm:"column:adr_class" json:"adr_class"` // 1..9
	HasTempRegime    bool           `gorm:"not null;default:false" json:"has_temp_regime"`
	TempMin          *float64       `gorm:"type:decimal" json:"temp_min"`
	TempMax          *float64       `gorm:"type:decimal" json:"temp_max"`
	RequiredBodyTypes pq.StringArray `gorm:"type:text[];column:body_types" json:"required_body_types"` // tent|ref|container|cistern|carcarrier

	// ── Логистика (для международных) ─────────────────────────
	Incoterms          *string        `gorm:"type:varchar(20)" json:"incoterms"`
	DeliveryGeography  pq.StringArray `gorm:"type:text[]" json:"delivery_geography"` // uz|cis|world
	Permits            pq.StringArray `gorm:"type:text[]" json:"permits"`            // TIR|CMR|phyto|CITES|EKMT

	CommentForCarrier *string `gorm:"type:text;column:notes" json:"comment_for_carrier"`

	// ── Статус / служебные ────────────────────────────────────
	InStock      bool    `gorm:"not null;default:true" json:"in_stock"` // false = «Нет в наличии»
	Status       string  `gorm:"type:cargo_status_enum;not null;default:'active';index" json:"status"`
	IsTemplate   bool    `gorm:"not null;default:false" json:"is_template"`
	TemplateName *string `gorm:"type:varchar(100)" json:"template_name"`

	IsBoosted       bool       `gorm:"not null;default:false" json:"is_boosted"`
	BoostExpiresAt  *time.Time `json:"boost_expires_at"`
	IsPaid          bool       `gorm:"not null;default:false" json:"is_paid"`
	IsAdminBlocked  bool       `gorm:"not null;default:false;column:is_admin_blocked" json:"is_admin_blocked,omitempty"`

	ViewsCount          int `gorm:"not null;default:0" json:"views_count"`
	ContactsBoughtCount int `gorm:"not null;default:0" json:"contacts_bought_count"`

	Media   []ListingMedia `gorm:"-" json:"media,omitempty"`    // фото/видео/сертификаты (entity=cargo)
	Contact *CargoContact  `gorm:"-" json:"contact,omitempty"` // контакты (после покупки)

	// Legacy-колонки (NOT NULL в БД) — заполняются дефолтами, в API не используются.
	Type          string `gorm:"type:cargo_type_enum;not null;default:'domestic'" json:"-"`
	DurationHours int    `gorm:"not null;default:24" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
