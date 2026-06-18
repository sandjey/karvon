package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ctm/internal/dto"
	"ctm/internal/service"
	"ctm/pkg/i18n"
)

type RouteHandler struct {
	svc *service.RouteService
}

func NewRouteHandler(svc *service.RouteService) *RouteHandler {
	return &RouteHandler{svc: svc}
}

func (h *RouteHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	g := rg.Group("/routes", auth)
	g.POST("", h.Create)
	g.GET("", h.List)
	g.DELETE("/:id", h.Delete)
	g.PATCH("/:id/notifications", h.SetNotifications)
}

func (h *RouteHandler) Create(c *gin.Context) {
	var req dto.SaveRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	route, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		InternalError(c)
		return
	}
	CreatedMsg(c, route, "ROUTE_SAVED")
}

func (h *RouteHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	list, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, list)
}

func (h *RouteHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		handleRouteErr(c, err)
		return
	}
	OKMsg(c, nil, "ROUTE_DELETED")
}

func (h *RouteHandler) SetNotifications(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.RouteNotificationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.SetNotifications(c.Request.Context(), id, userID, req.Enabled); err != nil {
		handleRouteErr(c, err)
		return
	}
	OKMsg(c, nil, "ROUTE_UPDATED")
}

func handleRouteErr(c *gin.Context, err error) {
	switch {
	case isErr(err, service.ErrNotFound):
		NotFound(c)
	case isErr(err, service.ErrNotOwner):
		Forbidden(c)
	default:
		InternalError(c)
	}
}
