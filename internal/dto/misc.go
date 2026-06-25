package dto

import (
	"time"

	"github.com/google/uuid"

	"ctm/internal/model"
)

type FavoriteRequest struct {
	ListingType string    `json:"listing_type" binding:"required,oneof=cargo warehouse carrier"`
	ListingID   uuid.UUID `json:"listing_id" binding:"required"`
}

type FavoriteItem struct {
	model.Favorite
	IsExpired bool `json:"is_expired"`
}

type FavoriteItemListing struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	CompanyName  string    `json:"company_name"`
	City         *string   `json:"city"`
	PricePerUnit *float64  `json:"price_per_unit"`
	Currency     *string   `json:"currency"`
	Status       string    `json:"status"`
	PreviewImage *string   `json:"preview_image"`
}

type FavoriteItemResponse struct {
	ID          uuid.UUID           `json:"id"`
	ListingType string              `json:"listing_type"`
	ListingID   uuid.UUID           `json:"listing_id"`
	Listing     FavoriteItemListing `json:"listing"`
	AddedAt     time.Time           `json:"added_at"`
}

type SaveRouteRequest struct {
	FromCity string `json:"from_city" binding:"required,max=100"`
	ToCity   string `json:"to_city" binding:"required,max=100"`
}

type RouteNotificationsRequest struct {
	Enabled bool `json:"enabled"`
}
