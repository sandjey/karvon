package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"karvon/internal/dto"
	"karvon/internal/service"
	"karvon/pkg/i18n"
)

type FavoriteHandler struct {
	svc *service.FavoriteService
}

func NewFavoriteHandler(svc *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{svc: svc}
}

func (h *FavoriteHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	g := rg.Group("/favorites", auth)
	g.POST("", h.Add)
	g.DELETE("", h.Remove)
	g.GET("", h.List)
}

func (h *FavoriteHandler) Add(c *gin.Context) {
	var req dto.FavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.Add(c.Request.Context(), userID, req); err != nil {
		InternalError(c)
		return
	}
	OKMsg(c, nil, "FAVORITE_ADDED")
}

func (h *FavoriteHandler) Remove(c *gin.Context) {
	var req dto.FavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.Remove(c.Request.Context(), userID, req); err != nil {
		InternalError(c)
		return
	}
	OKMsg(c, nil, "FAVORITE_REMOVED")
}

func (h *FavoriteHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	list, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, list)
}
