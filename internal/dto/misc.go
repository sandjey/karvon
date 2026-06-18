package dto

import (
	"github.com/google/uuid"

	"ctm/internal/model"
)

type FavoriteRequest struct {
	ListingType string    `json:"listing_type" binding:"required,oneof=cargo warehouse"`
	ListingID   uuid.UUID `json:"listing_id" binding:"required"`
}

type FavoriteItem struct {
	model.Favorite
	IsExpired bool `json:"is_expired"`
}

type SaveRouteRequest struct {
	FromCity string `json:"from_city" binding:"required,max=100"`
	ToCity   string `json:"to_city" binding:"required,max=100"`
}

type RouteNotificationsRequest struct {
	Enabled bool `json:"enabled"`
}
