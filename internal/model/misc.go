package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ListingMedia struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EntityType   string    `gorm:"type:media_entity_type_enum;not null"`
	EntityID     uuid.UUID `gorm:"type:uuid;not null"`
	FileURL      string    `gorm:"type:text;not null"`
	FileType     string    `gorm:"type:media_file_type_enum;not null"`
	OriginalName *string   `gorm:"type:varchar(200)"`
	SortOrder    int       `gorm:"not null;default:0"`
	CreatedAt    time.Time
}

func (ListingMedia) TableName() string { return "listing_media" }

type ContactView struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	User        User      `gorm:"foreignKey:UserID"`
	ListingType string    `gorm:"type:contact_listing_type_enum;not null"`
	ListingID   uuid.UUID `gorm:"type:uuid;not null"`
	TokensSpent int       `gorm:"not null;default:0"`
	ViewedAt    time.Time `gorm:"not null;index;autoCreateTime"`
}

type SavedRoute struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID               uuid.UUID `gorm:"type:uuid;not null;index"`
	User                 User      `gorm:"foreignKey:UserID"`
	FromCity             string    `gorm:"type:varchar(100);not null"`
	ToCity               string    `gorm:"type:varchar(100);not null"`
	NotificationsEnabled bool      `gorm:"not null;default:true"`
	CreatedAt            time.Time
}

type Favorite struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	User        User      `gorm:"foreignKey:UserID"`
	ListingType string    `gorm:"type:favorite_listing_type_enum;not null"`
	ListingID   uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time
}

type Notification struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	User      User           `gorm:"foreignKey:UserID"`
	Type      string         `gorm:"type:varchar(50);not null"`
	Title     string         `gorm:"type:varchar(200);not null"`
	Body      *string        `gorm:"type:text"`
	IsRead    bool           `gorm:"not null;default:false"`
	Meta      datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt time.Time      `gorm:"index"`
}
