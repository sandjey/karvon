package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"karvon/internal/repository"
	"karvon/internal/service"
	"karvon/pkg/geo"
)

type SearchHandler struct {
	cargo     *service.CargoService
	warehouse *service.WarehouseService
}

func NewSearchHandler(cargo *service.CargoService, warehouse *service.WarehouseService) *SearchHandler {
	return &SearchHandler{cargo: cargo, warehouse: warehouse}
}

func (h *SearchHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/search")
	g.GET("", h.Search)
	g.GET("/cities", h.Cities)
}

// GET /search?type=cargo|warehouse + соответствующие фильтры.
func (h *SearchHandler) Search(c *gin.Context) {
	page, perPage := parsePagination(c)
	switch c.DefaultQuery("type", "cargo") {
	case "warehouse":
		f := repository.WarehouseFilter{
			Region:        c.Query("region"),
			WarehouseType: c.Query("warehouse_type"),
			AreaMin:       parseFloatPtr(c.Query("area_min")),
			AreaMax:       parseFloatPtr(c.Query("area_max")),
			TempMin:       parseFloatPtr(c.Query("temp_min")),
			TempMax:       parseFloatPtr(c.Query("temp_max")),
			Services:      c.QueryArray("services"),
			Sort:          c.Query("sort"),
			Offset:        (page - 1) * perPage,
			Limit:         perPage,
		}
		list, total, err := h.warehouse.List(c.Request.Context(), f)
		if err != nil {
			InternalError(c)
			return
		}
		out := make([]warehouseResp, len(list))
		for i := range list {
			maskWarehouse(&list[i])
			out[i] = wrapWarehouse(&list[i])
		}
		Paginated(c, out, int(total), page, perPage)
	default:
		f := repository.CargoFilter{
			FromCity:     c.Query("from_city"),
			ToCity:       c.Query("to_city"),
			Type:         c.Query("cargo_type_kind"),
			LoadingType:  c.Query("loading_type"),
			CargoType:    c.Query("cargo_type"),
			BodyTypes:    c.QueryArray("body_types"),
			WeightMin:    parseFloatPtr(c.Query("weight_min")),
			WeightMax:    parseFloatPtr(c.Query("weight_max")),
			VerifiedOnly: c.Query("verified_only") == "true",
			Sort:         c.Query("sort"),
			Offset:       (page - 1) * perPage,
			Limit:        perPage,
		}
		if v := c.Query("date_from"); v != "" {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				f.DateFrom = &t
			}
		}
		list, total, err := h.cargo.List(c.Request.Context(), f)
		if err != nil {
			InternalError(c)
			return
		}
		for i := range list {
			maskCargo(&list[i])
		}
		Paginated(c, list, int(total), page, perPage)
	}
}

// GET /search/cities?q=
func (h *SearchHandler) Cities(c *gin.Context) {
	q := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	OK(c, geo.SearchCities(q, "", limit))
}
