package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"karvon/pkg/geo"
)

type GeoHandler struct{}

func NewGeoHandler() *GeoHandler { return &GeoHandler{} }

func (h *GeoHandler) RegisterRoutes(v1 *gin.RouterGroup) {
	g := v1.Group("/geo")
	g.GET("/countries", h.Countries)
	g.GET("/cities", h.Cities)
}

// GET /api/v1/geo/countries?q=&limit=
func (h *GeoHandler) Countries(c *gin.Context) {
	q := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "250"))
	if limit <= 0 || limit > 500 {
		limit = 250
	}
	OK(c, geo.SearchCountries(q, limit))
}

// GET /api/v1/geo/cities?q=&country_code=&limit=
func (h *GeoHandler) Cities(c *gin.Context) {
	q := c.Query("q")
	cc := c.Query("country_code")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	OK(c, geo.SearchCities(q, cc, limit))
}
