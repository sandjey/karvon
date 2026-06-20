package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ctm/internal/dto"
	"ctm/internal/service"
	"ctm/pkg/i18n"
)

type ContactHandler struct {
	svc     *service.ContactService
	pricing *service.PricingService
}

func NewContactHandler(svc *service.ContactService, pricing *service.PricingService) *ContactHandler {
	return &ContactHandler{svc: svc, pricing: pricing}
}

func (h *ContactHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	g := rg.Group("/contacts", auth)
	g.POST("/view", h.View)
	g.GET("/history", h.History)
	rg.GET("/users/me/tokens", auth, h.Tokens)
}

func (h *ContactHandler) View(c *gin.Context) {
	var req dto.OpenContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	info, err := h.svc.Open(c.Request.Context(), userID, req)
	if err != nil {
		if isErr(err, service.ErrInsufficientTokens) {
			packages, subs, listings, boosts, _ := h.pricing.Tariffs(c.Request.Context())
			c.JSON(http.StatusPaymentRequired, gin.H{
				"success": false,
				"error":   gin.H{"code": "INSUFFICIENT_TOKENS", "message": i18n.T(c, "INSUFFICIENT_TOKENS")},
				"data":    gin.H{"packages": packages, "subscriptions": subs, "listings": listings, "boosts": boosts},
			})
			return
		}
		if isErr(err, service.ErrListingNotFound) {
			FailCode(c, http.StatusNotFound, "LISTING_NOT_FOUND")
			return
		}
		InternalError(c)
		return
	}
	OKMsg(c, info, "CONTACT_OPENED")
}

func (h *ContactHandler) History(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	page, perPage := parsePagination(c)
	list, total, err := h.svc.History(c.Request.Context(), userID, (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	Paginated(c, list, int(total), page, perPage)
}

func (h *ContactHandler) Tokens(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	balance, txs, err := h.svc.TokenInfo(c.Request.Context(), userID)
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, gin.H{"token_balance": balance, "transactions": txs})
}
