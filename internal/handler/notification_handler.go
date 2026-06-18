package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ctm/internal/service"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	g := rg.Group("/notifications", auth)
	g.GET("", h.List)
	g.GET("/unread-count", h.UnreadCount)
	g.PATCH("/:id/read", h.MarkRead)
	g.PATCH("/read-all", h.MarkAllRead)
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	page, perPage := parsePagination(c)
	list, total, err := h.svc.List(c.Request.Context(), userID, (page-1)*perPage, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	Paginated(c, list, int(total), page, perPage)
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	count, err := h.svc.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		InternalError(c)
		return
	}
	OK(c, gin.H{"unread_count": count})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.MarkRead(c.Request.Context(), userID, id); err != nil {
		InternalError(c)
		return
	}
	OKMsg(c, nil, "NOTIFICATION_READ")
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.MarkAllRead(c.Request.Context(), userID); err != nil {
		InternalError(c)
		return
	}
	OKMsg(c, nil, "NOTIFICATIONS_READ")
}
