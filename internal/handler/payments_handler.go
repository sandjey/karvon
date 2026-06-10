package handler

import (
	"github.com/gin-gonic/gin"

	"karvon/internal/service"
)

// PaymentsHandler — пока только витрина тарифов. Платёжный цикл Rahmat (День 8) не реализован.
type PaymentsHandler struct {
	pricing *service.PricingService
}

func NewPaymentsHandler(pricing *service.PricingService) *PaymentsHandler {
	return &PaymentsHandler{pricing: pricing}
}

func (h *PaymentsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/payments")
	g.GET("/packages", h.Packages)
}

func (h *PaymentsHandler) Packages(c *gin.Context) {
	packages, subs, err := h.pricing.Tariffs(c.Request.Context())
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, gin.H{"packages": packages, "subscriptions": subs})
}
