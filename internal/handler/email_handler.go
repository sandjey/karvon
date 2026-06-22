package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ctm/internal/service"
	"ctm/pkg/i18n"
)

type EmailHandler struct{ svc *service.EmailService }

func NewEmailHandler(svc *service.EmailService) *EmailHandler {
	return &EmailHandler{svc: svc}
}

func (h *EmailHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	g := rg.Group("/email", auth)
	g.POST("/send-otp", h.SendOTP)
	g.POST("/verify-otp", h.VerifyOTP)
}

func (h *EmailHandler) SendOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	if err := h.svc.SendOTP(c.Request.Context(), req.Email); err != nil {
		InternalError(c)
		return
	}
	OKMsg(c, nil, "EMAIL_OTP_SENT")
}

func (h *EmailHandler) VerifyOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code"  binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	if err := h.svc.VerifyOTP(c.Request.Context(), req.Email, req.Code); err != nil {
		switch {
		case isErr(err, service.ErrEmailOTPInvalid):
			FailCode(c, http.StatusBadRequest, "EMAIL_OTP_INVALID")
		case isErr(err, service.ErrEmailOTPExpired):
			FailCode(c, http.StatusBadRequest, "EMAIL_OTP_EXPIRED")
		default:
			InternalError(c)
		}
		return
	}
	OKMsg(c, nil, "EMAIL_VERIFIED")
}
