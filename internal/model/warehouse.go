package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type WarehouseListing struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Company   *Company  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"-"`

	WarehouseType string  `gorm:"type:warehouse_type_enum;not null;index" json:"warehouse_type"`
	Name          string  `gorm:"type:varchar(200);not null" json:"name"`
	Region        *string `gorm:"type:varchar(100)" json:"region"`
	Address       *string `gorm:"type:text" json:"address"`
	Lat           *float64 `gorm:"type:decimal" json:"lat"`
	Lng           *float64 `gorm:"type:decimal" json:"lng"`

	PhoneMain     *string `gorm:"type:varchar(20)" json:"phone_main"`
	PhoneExtra    *string `gorm:"type:varchar(20)" json:"phone_extra"`
	ContactPerson *string `gorm:"type:varchar(100)" json:"contact_person"`
	Email         *string `gorm:"type:varchar(100)" json:"email"`
	Website       *string `gorm:"type:varchar(200)" json:"website"`

	Specialization pq.StringArray `gorm:"type:text[]" json:"specialization"`
	AreaTotalM2    *float64       `gorm:"type:decimal;column:area_total_m2" json:"area_total_m2"`
	AreaFreeM2     *float64       `gorm:"type:decimal;column:area_free_m2" json:"area_free_m2"`
	CeilingHeightM *float64       `gorm:"type:decimal;column:ceiling_height_m" json:"ceiling_height_m"`
	HeatingType    *string        `gorm:"type:heating_type_enum" json:"heating_type"`
	StorageType    pq.StringArray `gorm:"type:text[]" json:"storage_type"`

	TempMin          *float64       `gorm:"type:decimal" json:"temp_min"`
	TempMax          *float64       `gorm:"type:decimal" json:"temp_max"`
	ColdChamberTypes pq.StringArray `gorm:"type:text[]" json:"cold_chamber_types"`

	CustomsLicenseNumber   *string        `gorm:"type:varchar(100)" json:"customs_license_number"`
	CustomsLicenseIssued   *time.Time     `json:"customs_license_issued"`
	CustomsLicenseExpires  *time.Time     `json:"customs_license_expires"`
	CustomsSpecialServices pq.StringArray `gorm:"type:text[]" json:"customs_special_services"`

	Infrastructure pq.StringArray `gorm:"type:text[]" json:"infrastructure"`
	Services       pq.StringArray `gorm:"type:text[]" json:"services"`
	WorkingHours   datatypes.JSON `gorm:"type:jsonb" json:"working_hours,omitempty"`

	IsBoosted          bool       `gorm:"not null;default:false" json:"is_boosted"`
	BoostExpiresAt     *time.Time `json:"boost_expires_at"`
	IsPaid             bool       `gorm:"not null;default:false" json:"is_paid"`
	IsAdminBlocked     bool       `gorm:"not null;default:false;column:is_admin_blocked" json:"is_admin_blocked,omitempty"`
	Status             string     `gorm:"type:warehouse_status_enum;not null;default:'active';index" json:"status"`
	ViewsCount          int        `gorm:"not null;default:0" json:"views_count"`
	ContactsBoughtCount int        `gorm:"not null;default:0" json:"contacts_bought_count"`
	CreatedAt      time.Time  `json:"created_at"`

	Media []ListingMedia `gorm:"-" json:"media,omitempty"` // фото/видео (entity=warehouse)
}
