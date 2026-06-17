package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"karvon/internal/dto"
	"karvon/internal/model"
	"karvon/internal/repository"
	"karvon/internal/service"
	"karvon/pkg/i18n"
)

type CargoHandler struct {
	svc *service.CargoService
}

func NewCargoHandler(svc *service.CargoService) *CargoHandler {
	return &CargoHandler{svc: svc}
}

func (h *CargoHandler) RegisterRoutes(rg *gin.RouterGroup, auth, verified gin.HandlerFunc) {
	g := rg.Group("/listings/cargo")
	// public
	g.GET("", h.List)
	g.GET("/templates", auth, h.Templates)
	g.GET("/:id", h.GetOne)
	g.GET("/:id/stats", auth, h.Stats)
	g.GET("/:id/similar", h.Similar)
	// owner / verified
	g.POST("", auth, verified, h.Create)
	g.PATCH("/:id", auth, h.Update)
	g.PATCH("/:id/status", auth, h.SetStatus)
	g.DELETE("/:id", auth, h.Delete)
	g.POST("/:id/duplicate", auth, verified, h.Duplicate)
	g.POST("/:id/template", auth, h.SaveTemplate)
	g.POST("/from-template/:id", auth, verified, h.FromTemplate)
}

func (h *CargoHandler) Create(c *gin.Context) {
	var req dto.CargoUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	listing, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		handleCargoErr(c, err)
		return
	}
	CreatedMsg(c, listing, "CARGO_CREATED")
}

func (h *CargoHandler) List(c *gin.Context) {
	page, perPage := parsePagination(c)
	f := buildCargoFilter(c, page, perPage)
	list, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		InternalError(c)
		return
	}
	for i := range list {
		maskCargo(&list[i])
	}
	Paginated(c, list, int(total), page, perPage)
}

// buildCargoFilter собирает фильтр листинга товаров из query-параметров.
func buildCargoFilter(c *gin.Context, page, perPage int) repository.CargoFilter {
	f := repository.CargoFilter{
		Category:      c.Query("category"),
		FromCity:      c.Query("from_city"),
		FromCountry:   c.Query("from_country"),
		ToCity:        c.Query("to_city"),
		ToCountry:     c.Query("to_country"),
		BodyType:      c.Query("body_type"),
		Divisibility:  c.Query("divisibility"),
		Packaging:     c.Query("packaging"),
		MinOrderMax:   parseFloatPtr(c.Query("min_order_max")),
		QtyMin:        parseFloatPtr(c.Query("qty_min")),
		QtyMax:        parseFloatPtr(c.Query("qty_max")),
		PriceMin:      parseFloatPtr(c.Query("price_min")),
		PriceMax:      parseFloatPtr(c.Query("price_max")),
		HasTempRegime: c.Query("temp_regime") == "true",
		VerifiedOnly:  c.Query("verified_only") == "true",
		Sort:          c.Query("sort"),
		Offset:        (page - 1) * perPage,
		Limit:         perPage,
	}
	if v := c.Query("adr"); v == "true" {
		t := true
		f.IsADR = &t
	} else if v == "false" {
		fl := false
		f.IsADR = &fl
	}
	return f
}

func (h *CargoHandler) GetOne(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	viewerID := optionalUserID(c)
	listing, err := h.svc.GetByID(c.Request.Context(), id, viewerID)
	if err != nil {
		handleCargoErr(c, err)
		return
	}
	if listing.UserID != viewerID {
		maskCargo(listing)
	}
	OK(c, listing)
}

func (h *CargoHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.CargoUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	listing, err := h.svc.Update(c.Request.Context(), id, userID, req)
	if err != nil {
		handleCargoErr(c, err)
		return
	}
	OKMsg(c, listing, "CARGO_UPDATED")
}

func (h *CargoHandler) SetStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.CargoStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.SetStatus(c.Request.Context(), id, userID, req.Status, req.InStock); err != nil {
		handleCargoErr(c, err)
		return
	}
	OKMsg(c, nil, "CARGO_STATUS_UPDATED")
}

func (h *CargoHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		handleCargoErr(c, err)
		return
	}
	OKMsg(c, nil, "CARGO_DELETED")
}

func (h *CargoHandler) Duplicate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	listing, err := h.svc.Duplicate(c.Request.Context(), id, userID)
	if err != nil {
		handleCargoErr(c, err)
		return
	}
	CreatedMsg(c, listing, "CARGO_CREATED")
}

func (h *CargoHandler) SaveTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.SaveTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	tpl, err := h.svc.SaveAsTemplate(c.Request.Context(), id, userID, req.TemplateName)
	if err != nil {
		handleCargoErr(c, err)
		return
	}
	CreatedMsg(c, tpl, "TEMPLATE_SAVED")
}

func (h *CargoHandler) Templates(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	list, err := h.svc.Templates(c.Request.Context(), userID)
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, list)
}

func (h *CargoHandler) FromTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	listing, err := h.svc.FromTemplate(c.Request.Context(), id, userID)
	if err != nil {
		handleCargoErr(c, err)
		return
	}
	CreatedMsg(c, listing, "CARGO_CREATED")
}

func (h *CargoHandler) Stats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	stats, err := h.svc.Stats(c.Request.Context(), id, userID)
	if err != nil {
		handleCargoErr(c, err)
		return
	}
	OK(c, stats)
}

func (h *CargoHandler) Similar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	listing, err := h.svc.GetByID(c.Request.Context(), id, optionalUserID(c))
	if err != nil {
		handleCargoErr(c, err)
		return
	}
	similar, err := h.svc.Similar(c.Request.Context(), id, listing.Category)
	if err != nil {
		InternalError(c)
		return
	}
	for i := range similar {
		maskCargo(&similar[i])
	}
	OK(c, similar)
}

// maskCargo прячет контактные данные компании в публичной выдаче (контакты — через /contacts/view).
func maskCargo(c *model.CargoListing) {
	c.User = nil
	if c.Company != nil {
		c.Company.Phone = ""
		c.Company.Email = ""
	}
}

func handleCargoErr(c *gin.Context, err error) {
	switch {
	case isErr(err, service.ErrListingNotFound):
		FailCode(c, http.StatusNotFound, "LISTING_NOT_FOUND")
	case isErr(err, service.ErrNotOwner):
		Forbidden(c)
	case isErr(err, service.ErrCompanyNotFound), isErr(err, service.ErrCompanyNotOwned):
		FailCode(c, http.StatusForbidden, "COMPANY_NOT_VERIFIED")
	case isErr(err, service.ErrPhotoLimitExceeded):
		FailCode(c, http.StatusBadRequest, "PHOTO_LIMIT_EXCEEDED")
	default:
		InternalError(c)
	}
}

func optionalUserID(c *gin.Context) uuid.UUID {
	if v, ok := c.Get("user_id"); ok {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

func parseFloatPtr(s string) *float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}
