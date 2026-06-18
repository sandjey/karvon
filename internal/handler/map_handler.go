package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"ctm/internal/repository"
)

type MapHandler struct {
	cargoRepo     *repository.CargoRepo
	warehouseRepo *repository.WarehouseRepo
}

func NewMapHandler(cargoRepo *repository.CargoRepo, warehouseRepo *repository.WarehouseRepo) *MapHandler {
	return &MapHandler{cargoRepo: cargoRepo, warehouseRepo: warehouseRepo}
}

func (h *MapHandler) RegisterRoutes(v1 *gin.RouterGroup) {
	g := v1.Group("/map")
	g.GET("/markers", h.Markers)
}

type geoPoint struct {
	Type        string     `json:"type"`
	Coordinates [2]float64 `json:"coordinates"` // [lng, lat] — GeoJSON standard
}

type geoFeature struct {
	Type       string         `json:"type"`
	ID         string         `json:"id"`
	Geometry   geoPoint       `json:"geometry"`
	Properties map[string]any `json:"properties"`
}

type geoCollection struct {
	Type     string       `json:"type"`
	Features []geoFeature `json:"features"`
}

// GET /api/v1/map/markers?kind=cargo|warehouse
// Returns raw GeoJSON FeatureCollection (no wrapper) — compatible with MapLibre, Leaflet, mobile SDKs.
func (h *MapHandler) Markers(c *gin.Context) {
	ctx := context.Background()
	kind := c.Query("kind") // optional filter: "cargo" | "warehouse" | "" (all)

	features := make([]geoFeature, 0)

	if kind == "" || kind == "cargo" {
		rows, err := h.cargoRepo.ListForMap(ctx)
		if err == nil {
			for _, row := range rows {
				features = append(features, geoFeature{
					Type: "Feature",
					ID:   row.ID,
					Geometry: geoPoint{
						Type:        "Point",
						Coordinates: [2]float64{row.Lng, row.Lat},
					},
					Properties: map[string]any{
						"id":    row.ID,
						"kind":  "cargo",
						"title": row.CargoName,
						"city":  row.FromCity,
					},
				})
			}
		}
	}

	if kind == "" || kind == "warehouse" {
		rows, err := h.warehouseRepo.ListForMap(ctx)
		if err == nil {
			for _, row := range rows {
				features = append(features, geoFeature{
					Type: "Feature",
					ID:   row.ID,
					Geometry: geoPoint{
						Type:        "Point",
						Coordinates: [2]float64{row.Lng, row.Lat},
					},
					Properties: map[string]any{
						"id":      row.ID,
						"kind":    "warehouse",
						"title":   row.Name,
						"address": row.Address,
					},
				})
			}
		}
	}

	c.JSON(http.StatusOK, geoCollection{Type: "FeatureCollection", Features: features})
}
