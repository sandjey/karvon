package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"karvon/internal/dto"
	"karvon/internal/model"
	"karvon/internal/service"
	"karvon/pkg/i18n"
)

type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// RegisterRoutes: login публичен, остальное — только super_admin.
func (h *AdminHandler) RegisterRoutes(rg *gin.RouterGroup, auth, superAdmin gin.HandlerFunc) {
	rg.POST("/admin/login", h.Login)

	g := rg.Group("/admin", auth, superAdmin)
	g.GET("/dashboard", h.Dashboard)
	g.GET("/users", h.Users)
	g.GET("/users/:id", h.User)
	g.PATCH("/users/:id/block", h.Block)
	g.POST("/users/:id/topup", h.Topup)
	g.GET("/listings", h.Listings)
	g.DELETE("/listings/:type/:id", h.DeleteListing)
	g.PATCH("/listings/:type/:id/block", h.BlockListing)
	g.GET("/companies", h.Companies)
	g.GET("/payments", h.Payments)
	g.GET("/pricing", h.Pricing)
	g.POST("/pricing", h.CreatePricing)
	g.PUT("/pricing/:key", h.UpdatePricing)
	g.DELETE("/pricing/:key", h.DeletePricing)
	g.POST("/moderators", h.CreateModerator)
	g.DELETE("/moderators/:id", h.DeleteModerator)
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	tokens, err := h.svc.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		FailCode(c, http.StatusUnauthorized, "INVALID_CREDENTIALS")
		return
	}
	OKMsg(c, tokens, "AUTH_SUCCESS")
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	m, err := h.svc.Dashboard(c.Request.Context(), c.DefaultQuery("period", "30d"))
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, m)
}

func (h *AdminHandler) Users(c *gin.Context) {
	page, perPage := parsePagination(c)
	list, total, err := h.svc.Users(c.Request.Context(), c.Query("q"), c.Query("role"), (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	Paginated(c, list, int(total), page, perPage)
}

func (h *AdminHandler) User(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	u, err := h.svc.User(c.Request.Context(), id)
	if err != nil {
		FailCode(c, http.StatusNotFound, "USER_NOT_FOUND")
		return
	}
	OK(c, u)
}

func (h *AdminHandler) Block(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.BlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	if err := h.svc.SetBlocked(c.Request.Context(), id, req.Blocked); err != nil {
		InternalError(c)
		return
	}
	OKMsg(c, nil, "USER_BLOCK_UPDATED")
}

func (h *AdminHandler) Topup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.TopupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	if err := h.svc.TopupTokens(c.Request.Context(), id, req.Amount); err != nil {
		FailCode(c, http.StatusNotFound, "USER_NOT_FOUND")
		return
	}
	OKMsg(c, nil, "TOKENS_TOPPED_UP")
}

func (h *AdminHandler) Listings(c *gin.Context) {
	page, perPage := parsePagination(c)
	if c.DefaultQuery("type", "cargo") == "warehouse" {
		list, total, err := h.svc.Warehouses(c.Request.Context(), (page-1)*perPage, perPage)
		if err != nil {
			InternalError(c)
			return
		}
		Paginated(c, list, int(total), page, perPage)
		return
	}
	list, total, err := h.svc.Cargo(c.Request.Context(), (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	Paginated(c, list, int(total), page, perPage)
}

func (h *AdminHandler) DeleteListing(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	if err := h.svc.DeleteListing(c.Request.Context(), c.Param("type"), id); err != nil {
		FailCode(c, http.StatusNotFound, "LISTING_NOT_FOUND")
		return
	}
	OKMsg(c, nil, "LISTING_DELETED")
}

func (h *AdminHandler) BlockListing(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	if err := h.svc.BlockListing(c.Request.Context(), c.Param("type"), id); err != nil {
		FailCode(c, http.StatusNotFound, "LISTING_NOT_FOUND")
		return
	}
	OKMsg(c, nil, "LISTING_BLOCK_UPDATED")
}

func (h *AdminHandler) Companies(c *gin.Context) {
	page, perPage := parsePagination(c)
	list, total, err := h.svc.Companies(c.Request.Context(), c.Query("status"), (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	Paginated(c, list, int(total), page, perPage)
}

func (h *AdminHandler) Payments(c *gin.Context) {
	page, perPage := parsePagination(c)
	list, total, err := h.svc.Payments(c.Request.Context(), (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	Paginated(c, list, int(total), page, perPage)
}

func (h *AdminHandler) Pricing(c *gin.Context) {
	list, err := h.svc.Pricing(c.Request.Context())
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, list)
}

func (h *AdminHandler) UpdatePricing(c *gin.Context) {
	var req dto.UpdatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	fields := map[string]interface{}{}
	if req.Label != nil {
		fields["label"] = *req.Label
	}
	if req.PriceUZS != nil {
		fields["price_uzs"] = *req.PriceUZS
	}
	if req.PriceUSD != nil {
		fields["price_usd"] = *req.PriceUSD
	}
	if req.TokensAmount != nil {
		fields["tokens_amount"] = *req.TokensAmount
	}
	if req.DurationDays != nil {
		fields["duration_days"] = *req.DurationDays
	}
	if req.IsActive != nil {
		fields["is_active"] = *req.IsActive
	}
	adminID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.UpdatePricing(c.Request.Context(), c.Param("key"), fields, adminID); err != nil {
		FailCode(c, http.StatusNotFound, "NOT_FOUND")
		return
	}
	OKMsg(c, nil, "PRICING_UPDATED")
}

func (h *AdminHandler) CreatePricing(c *gin.Context) {
	var req dto.CreatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	adminID := c.MustGet("user_id").(uuid.UUID)
	p := &model.PricingConfig{
		Key:          req.Key,
		Label:        req.Label,
		PriceUZS:     req.PriceUZS,
		PriceUSD:     req.PriceUSD,
		TokensAmount: req.TokensAmount,
		DurationDays: req.DurationDays,
		IsActive:     active,
	}
	if err := h.svc.CreatePricing(c.Request.Context(), p, adminID); err != nil {
		if isErr(err, service.ErrAlreadyExists) {
			FailCode(c, http.StatusConflict, "ALREADY_EXISTS")
			return
		}
		InternalError(c)
		return
	}
	CreatedMsg(c, p, "PRICING_UPDATED")
}

func (h *AdminHandler) DeletePricing(c *gin.Context) {
	if err := h.svc.DeletePricing(c.Request.Context(), c.Param("key")); err != nil {
		InternalError(c)
		return
	}
	OKMsg(c, nil, "PRICING_UPDATED")
}

func (h *AdminHandler) CreateModerator(c *gin.Context) {
	var req dto.CreateModeratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	u, err := h.svc.CreateModerator(c.Request.Context(), req.Phone, req.Name, req.Login, req.Password)
	if err != nil {
		InternalError(c)
		return
	}
	CreatedMsg(c, u, "MODERATOR_CREATED")
}

func (h *AdminHandler) DeleteModerator(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	if err := h.svc.DeleteModerator(c.Request.Context(), id); err != nil {
		FailCode(c, http.StatusNotFound, "USER_NOT_FOUND")
		return
	}
	OKMsg(c, nil, "MODERATOR_DELETED")
}
