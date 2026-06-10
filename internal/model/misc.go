package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ListingMedia struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EntityType   string    `gorm:"type:media_entity_type_enum;not null" json:"entity_type"`
	EntityID     uuid.UUID `gorm:"type:uuid;not null" json:"entity_id"`
	FileURL      string    `gorm:"type:text;not null" json:"file_url"`
	FileType     string    `gorm:"type:media_file_type_enum;not null" json:"file_type"`
	OriginalName *string   `gorm:"type:varchar(200)" json:"original_name"`
	SortOrder    int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
}

func (ListingMedia) TableName() string { return "listing_media" }

type ContactView struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	ListingType string    `gorm:"type:contact_listing_type_enum;not null" json:"listing_type"`
	ListingID   uuid.UUID `gorm:"type:uuid;not null" json:"listing_id"`
	TokensSpent int       `gorm:"not null;default:0" json:"tokens_spent"`
	ViewedAt    time.Time `gorm:"not null;index;autoCreateTime" json:"viewed_at"`
}

type SavedRoute struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID               uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	FromCity             string    `gorm:"type:varchar(100);not null" json:"from_city"`
	ToCity               string    `gorm:"type:varchar(100);not null" json:"to_city"`
	NotificationsEnabled bool      `gorm:"not null;default:true" json:"notifications_enabled"`
	CreatedAt            time.Time `json:"created_at"`
}

type Favorite struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	ListingType string    `gorm:"type:favorite_listing_type_enum;not null" json:"listing_type"`
	ListingID   uuid.UUID `gorm:"type:uuid;not null" json:"listing_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Notification struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Type      string         `gorm:"type:varchar(50);not null" json:"type"`
	Title     string         `gorm:"type:varchar(200);not null" json:"title"`
	Body      *string        `gorm:"type:text" json:"body"`
	IsRead    bool           `gorm:"not null;default:false" json:"is_read"`
	Meta      datatypes.JSON `gorm:"type:jsonb" json:"meta,omitempty"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
}
