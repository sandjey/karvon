package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"karvon/internal/dto"
	"karvon/internal/model"
	"karvon/internal/repository"
	"karvon/internal/service"
	"karvon/pkg/i18n"
)

type WarehouseHandler struct {
	svc *service.WarehouseService
}

func NewWarehouseHandler(svc *service.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{svc: svc}
}

func (h *WarehouseHandler) RegisterRoutes(rg *gin.RouterGroup, auth, verified gin.HandlerFunc) {
	g := rg.Group("/warehouses")
	g.GET("", h.List)
	g.GET("/:id", h.GetOne)
	g.GET("/:id/stats", auth, h.Stats)
	g.POST("", auth, verified, h.Create)
	g.PUT("/:id", auth, h.Update)
	g.PATCH("/:id/status", auth, h.SetStatus)
	g.DELETE("/:id", auth, h.Delete)
}

type warehouseResp struct {
	*model.WarehouseListing
	OccupancyPercent *float64 `json:"occupancy_percent"`
	Color            string   `json:"color"`
}

// цветовая кодировка типов складов (BRD)
var warehouseColors = map[string]string{
	"regular": "#1A56A0",
	"cold":    "#0288D1",
	"customs": "#6A1B9A",
}

func wrapWarehouse(w *model.WarehouseListing) warehouseResp {
	var occ *float64
	if w.AreaTotalM2 != nil && *w.AreaTotalM2 > 0 && w.AreaFreeM2 != nil {
		v := (*w.AreaTotalM2 - *w.AreaFreeM2) / *w.AreaTotalM2 * 100
		occ = &v
	}
	return warehouseResp{WarehouseListing: w, OccupancyPercent: occ, Color: warehouseColors[w.WarehouseType]}
}

func maskWarehouse(w *model.WarehouseListing) {
	w.User = nil
	w.PhoneMain = nil
	w.PhoneExtra = nil
	w.Email = nil
	w.ContactPerson = nil
	if w.Company != nil {
		w.Company.Phone = ""
		w.Company.Email = ""
	}
}

func (h *WarehouseHandler) Create(c *gin.Context) {
	var req dto.WarehouseUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	w, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		handleWarehouseErr(c, err)
		return
	}
	CreatedMsg(c, wrapWarehouse(w), "WAREHOUSE_CREATED")
}

func (h *WarehouseHandler) List(c *gin.Context) {
	page, perPage := parsePagination(c)
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
	list, total, err := h.svc.List(c.Request.Context(), f)
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
}

func (h *WarehouseHandler) GetOne(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	viewerID := optionalUserID(c)
	w, err := h.svc.GetByID(c.Request.Context(), id, viewerID)
	if err != nil {
		handleWarehouseErr(c, err)
		return
	}
	if w.UserID != viewerID {
		maskWarehouse(w)
	}
	OK(c, wrapWarehouse(w))
}

func (h *WarehouseHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.WarehouseUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	w, err := h.svc.Update(c.Request.Context(), id, userID, req)
	if err != nil {
		handleWarehouseErr(c, err)
		return
	}
	OKMsg(c, wrapWarehouse(w), "WAREHOUSE_UPDATED")
}

func (h *WarehouseHandler) SetStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.WarehouseStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.SetStatus(c.Request.Context(), id, userID, req.Status); err != nil {
		handleWarehouseErr(c, err)
		return
	}
	OKMsg(c, nil, "WAREHOUSE_UPDATED")
}

func (h *WarehouseHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		handleWarehouseErr(c, err)
		return
	}
	OKMsg(c, nil, "WAREHOUSE_DELETED")
}

func (h *WarehouseHandler) Stats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	stats, err := h.svc.Stats(c.Request.Context(), id, userID)
	if err != nil {
		handleWarehouseErr(c, err)
		return
	}
	OK(c, stats)
}

func handleWarehouseErr(c *gin.Context, err error) {
	switch {
	case isErr(err, service.ErrListingNotFound):
		FailCode(c, http.StatusNotFound, "LISTING_NOT_FOUND")
	case isErr(err, service.ErrNotOwner):
		Forbidden(c)
	case isErr(err, service.ErrValidation):
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
	case isErr(err, service.ErrCompanyNotFound), isErr(err, service.ErrCompanyNotOwned):
		FailCode(c, http.StatusForbidden, "COMPANY_NOT_VERIFIED")
	default:
		InternalError(c)
	}
}
