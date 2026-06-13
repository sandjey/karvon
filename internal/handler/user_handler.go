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

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	g := rg.Group("/users", auth)
	g.GET("/me", h.GetProfile)
	g.PUT("/me", h.UpdateProfile)
	g.GET("/me/stats", h.GetStats)
	g.GET("/me/events", h.GetEvents)
}

func (h *UserHandler) GetEvents(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	events, err := h.svc.GetEvents(c.Request.Context(), userID)
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, events)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	u, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		handleUserErr(c, err)
		return
	}
	OK(c, toProfileResponse(u))
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.UpdateProfile(c.Request.Context(), userID, req); err != nil {
		handleUserErr(c, err)
		return
	}
	u, _ := h.svc.GetProfile(c.Request.Context(), userID)
	OKMsg(c, toProfileResponse(u), "PROFILE_UPDATED")
}

func (h *UserHandler) GetStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	stats, err := h.svc.GetStats(c.Request.Context(), userID)
	if err != nil {
		handleUserErr(c, err)
		return
	}
	OK(c, stats)
}

func toProfileResponse(u *model.User) dto.ProfileResponse {
	return dto.ProfileResponse{
		ID:           u.ID,
		Phone:        u.Phone,
		Name:         u.Name,
		Email:        u.Email,
		ExtraPhone:   u.ExtraPhone,
		WhatsApp:     u.WhatsApp,
		Telegram:     u.Telegram,
		City:         u.City,
		Country:      u.Country,
		TokenBalance: u.TokenBalance,
		Role:         u.Role,
		CreatedAt:    u.CreatedAt,
	}
}

func handleUserErr(c *gin.Context, err error) {
	switch {
	case isErr(err, service.ErrNotFound):
		FailCode(c, http.StatusNotFound, "USER_NOT_FOUND")
	default:
		InternalError(c)
	}
}
