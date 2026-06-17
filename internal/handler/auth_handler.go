package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"karvon/internal/dto"
	"karvon/internal/service"
	"karvon/pkg/i18n"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	g := rg.Group("/auth")
	g.POST("/send-otp", h.SendOTP)
	g.POST("/verify-otp", h.VerifyOTP)
	g.POST("/refresh", h.Refresh)
	g.POST("/logout", auth, h.Logout)
	g.POST("/complete-registration", auth, h.CompleteRegistration)
}

func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req dto.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	resp, err := h.svc.SendOTP(c.Request.Context(), req)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	OKMsg(c, resp, "OTP_SENT")
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	pair, err := h.svc.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	OKMsg(c, pair, "AUTH_SUCCESS")
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	resp, err := h.svc.Refresh(c.Request.Context(), req)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	OKMsg(c, resp, "AUTH_SUCCESS")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	_ = h.svc.Logout(c.Request.Context(), req.RefreshToken, userID)
	OKMsg(c, nil, "LOGOUT_SUCCESS")
}

func (h *AuthHandler) CompleteRegistration(c *gin.Context) {
	var req dto.CompleteRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.CompleteRegistration(c.Request.Context(), userID, req); err != nil {
		h.handleErr(c, err)
		return
	}
	OKMsg(c, nil, "REGISTRATION_COMPLETE")
}

func (h *AuthHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrOTPRateLimit):
		FailCode(c, http.StatusTooManyRequests, "OTP_MAX_ATTEMPTS")
	case errors.Is(err, service.ErrOTPInvalid):
		FailCode(c, http.StatusBadRequest, "OTP_INVALID")
	case errors.Is(err, service.ErrOTPMaxAttempts):
		FailCode(c, http.StatusBadRequest, "OTP_MAX_ATTEMPTS")
	case errors.Is(err, service.ErrOTPExpired):
		FailCode(c, http.StatusBadRequest, "OTP_EXPIRED")
	case errors.Is(err, service.ErrUserBlocked):
		FailCode(c, http.StatusForbidden, "USER_BLOCKED")
	case errors.Is(err, service.ErrTokenInvalid):
		Unauthorized(c)
	case errors.Is(err, service.ErrNotFound):
		FailCode(c, http.StatusNotFound, "USER_NOT_FOUND")
	case errors.Is(err, service.ErrNameAlreadySet):
		FailCode(c, http.StatusConflict, "NAME_ALREADY_SET")
	case errors.Is(err, service.ErrTelegramNotConfigured):
		FailCode(c, http.StatusServiceUnavailable, "TELEGRAM_NOT_CONFIGURED")
	case errors.Is(err, service.ErrWhatsAppNotConfigured):
		FailCode(c, http.StatusServiceUnavailable, "WHATSAPP_NOT_CONFIGURED")
	case errors.Is(err, service.ErrPhoneNotOnTelegram):
		FailCode(c, http.StatusUnprocessableEntity, "PHONE_NOT_ON_TELEGRAM")
	case errors.Is(err, service.ErrPhoneNotOnWhatsApp):
		FailCode(c, http.StatusUnprocessableEntity, "PHONE_NOT_ON_WHATSAPP")
	default:
		InternalError(c)
	}
}
