package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"karvon/internal/dto"
	"karvon/internal/service"
	"karvon/pkg/i18n"
)

type PaymentsHandler struct {
	pricing  *service.PricingService
	payments *service.PaymentService
}

func NewPaymentsHandler(pricing *service.PricingService, payments *service.PaymentService) *PaymentsHandler {
	return &PaymentsHandler{pricing: pricing, payments: payments}
}

func (h *PaymentsHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	g := rg.Group("/payments")
	g.GET("/packages", h.Packages)
	g.POST("/create", auth, h.Create)
	g.POST("/webhook", h.Webhook) // публичный, проверка подписи внутри
	g.GET("/history", auth, h.History)

	rg.GET("/subscriptions/active", auth, h.ActiveSubscription)
	rg.POST("/listings/:type/:id/boost", auth, h.Boost)
}

func (h *PaymentsHandler) Packages(c *gin.Context) {
	packages, subs, err := h.pricing.Tariffs(c.Request.Context())
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, gin.H{"packages": packages, "subscriptions": subs})
}

func (h *PaymentsHandler) Create(c *gin.Context) {
	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	order, url, err := h.payments.Create(c.Request.Context(), service.CreatePaymentInput{
		UserID:      userID,
		PaymentType: req.PaymentType,
		PricingKey:  req.PricingKey,
		Currency:    req.Currency,
		ListingType: req.ListingType,
		ListingID:   req.ListingID,
	})
	if err != nil {
		if isErr(err, service.ErrNotFound) {
			FailCode(c, http.StatusNotFound, "NOT_FOUND")
			return
		}
		InternalError(c)
		return
	}
	CreatedMsg(c, gin.H{"order_id": order.ID, "payment_url": url, "amount": order.Amount, "currency": order.Currency}, "PAYMENT_CREATED")
}

func (h *PaymentsHandler) Webhook(c *gin.Context) {
	payload, _ := c.GetRawData()
	var req dto.WebhookRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	if err := h.payments.Webhook(c.Request.Context(), orderID, req.Method, payload, req.Signature); err != nil {
		// Rahmat будет повторять при 500
		InternalError(c)
		return
	}
	OK(c, gin.H{"ok": true})
}

func (h *PaymentsHandler) History(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	page, perPage := parsePagination(c)
	list, total, err := h.payments.History(c.Request.Context(), userID, (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	Paginated(c, list, int(total), page, perPage)
}

func (h *PaymentsHandler) ActiveSubscription(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	sub, err := h.payments.ActiveSubscription(c.Request.Context(), userID)
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, sub)
}

func (h *PaymentsHandler) Boost(c *gin.Context) {
	var req dto.BoostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	order, url, err := h.payments.Boost(c.Request.Context(), userID, c.Param("type"), c.Param("id"), req.PricingKey, req.Currency)
	if err != nil {
		switch {
		case isErr(err, service.ErrListingNotFound):
			FailCode(c, http.StatusNotFound, "LISTING_NOT_FOUND")
		case isErr(err, service.ErrNotOwner):
			Forbidden(c)
		case isErr(err, service.ErrNotFound):
			FailCode(c, http.StatusNotFound, "NOT_FOUND")
		default:
			InternalError(c)
		}
		return
	}
	CreatedMsg(c, gin.H{"order_id": order.ID, "payment_url": url, "amount": order.Amount, "currency": order.Currency}, "PAYMENT_CREATED")
}
