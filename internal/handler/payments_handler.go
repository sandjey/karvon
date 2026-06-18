package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ctm/internal/dto"
	"ctm/internal/service"
	"ctm/pkg/i18n"
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
	g.POST("/webhook", h.Webhook) // публичный, верификация через store_invoice_id
	g.GET("/history", auth, h.History)
	g.GET("/:id", auth, h.GetOrder)

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
	order, checkoutURL, err := h.payments.Create(c.Request.Context(), service.CreatePaymentInput{
		UserID:      userID,
		PaymentType: req.PaymentType,
		PricingKey:  req.PricingKey,
		Currency:    req.Currency,
		ListingType: req.ListingType,
		ListingID:   req.ListingID,
		Lang:        i18n.Lang(c),
		ReturnURL:   req.ReturnURL,
	})
	if err != nil {
		if isErr(err, service.ErrNotFound) {
			FailCode(c, http.StatusNotFound, "NOT_FOUND")
			return
		}
		log.Error().Err(err).Str("pricing_key", req.PricingKey).Str("type", req.PaymentType).Msg("payment create failed")
		InternalError(c)
		return
	}
	CreatedMsg(c, gin.H{"order_id": order.ID, "payment_url": checkoutURL, "amount": order.Amount, "currency": order.Currency}, "PAYMENT_CREATED")
}

// Webhook принимает полный PaymentModel callback от Multicard.
// Обрабатывает статусы "success" (зачислить) и "revert" (откатить).
// Остальные статусы (draft/progress/billing/error) — квитируем 200 без обработки.
// При ошибке обработки возвращаем 500 — Multicard повторит запрос.
func (h *PaymentsHandler) Webhook(c *gin.Context) {
	var cb dto.MulticardCallback
	if err := c.ShouldBindJSON(&cb); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// Обрабатываем только финальные статусы
	if cb.Status != "success" && cb.Status != "revert" {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	if err := h.payments.Webhook(c.Request.Context(), cb); err != nil {
		if isErr(err, service.ErrNotFound) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
		log.Error().Err(err).Str("order_id", cb.StoreInvoiceID).Str("status", cb.Status).Msg("payment webhook failed")
		InternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetOrder возвращает статус платёжного заказа (для клиентского поллинга).
func (h *PaymentsHandler) GetOrder(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	order, err := h.payments.GetByID(c.Request.Context(), orderID, userID)
	if err != nil {
		switch {
		case isErr(err, service.ErrNotFound):
			FailCode(c, http.StatusNotFound, "NOT_FOUND")
		case isErr(err, service.ErrNotOwner):
			Forbidden(c)
		default:
			InternalError(c)
		}
		return
	}
	OK(c, order)
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
	order, checkoutURL, err := h.payments.Boost(
		c.Request.Context(),
		userID,
		c.Param("type"),
		c.Param("id"),
		req.PricingKey,
		req.Currency,
		i18n.Lang(c),
		req.ReturnURL,
	)
	if err != nil {
		switch {
		case isErr(err, service.ErrListingNotFound):
			FailCode(c, http.StatusNotFound, "LISTING_NOT_FOUND")
		case isErr(err, service.ErrNotOwner):
			Forbidden(c)
		case isErr(err, service.ErrNotFound):
			FailCode(c, http.StatusNotFound, "NOT_FOUND")
		default:
			log.Error().Err(err).Str("listing_type", c.Param("type")).Str("listing_id", c.Param("id")).Msg("boost payment failed")
			InternalError(c)
		}
		return
	}
	CreatedMsg(c, gin.H{"order_id": order.ID, "payment_url": checkoutURL, "amount": order.Amount, "currency": order.Currency}, "PAYMENT_CREATED")
}
