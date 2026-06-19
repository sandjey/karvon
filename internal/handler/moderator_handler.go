package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/service"
	"ctm/pkg/i18n"
)

type ModeratorHandler struct {
	svc *service.ModeratorService
}

func NewModeratorHandler(svc *service.ModeratorService) *ModeratorHandler {
	return &ModeratorHandler{svc: svc}
}

func (h *ModeratorHandler) RegisterRoutes(rg *gin.RouterGroup, auth, moderatorRole gin.HandlerFunc) {
	g := rg.Group("/moderator", auth, moderatorRole)
	g.GET("/queue", h.GetQueue)
	g.GET("/queue/:id", h.GetQueueItem)
	g.POST("/queue/:id/approve", h.Approve)
	g.POST("/queue/:id/reject", h.Reject)
	g.POST("/queue/:id/request-docs", h.RequestDocs)
	g.GET("/history", h.GetHistory)
}

func (h *ModeratorHandler) GetQueue(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	page, perPage := parsePagination(c)
	list, total, err := h.svc.GetQueue(c.Request.Context(), status, page, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	items := make([]dto.QueueItemResponse, len(list))
	for i := range list {
		items[i] = toQueueItem(&list[i])
	}
	Paginated(c, items, int(total), page, perPage)
}

func (h *ModeratorHandler) GetQueueItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	comp, err := h.svc.GetQueueItem(c.Request.Context(), id)
	if err != nil {
		handleModeratorErr(c, err)
		return
	}
	OK(c, toQueueItem(comp))
}

func (h *ModeratorHandler) Approve(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	modID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.Approve(c.Request.Context(), modID, id); err != nil {
		handleModeratorErr(c, err)
		return
	}
	OKMsg(c, nil, "COMPANY_APPROVED")
}

func (h *ModeratorHandler) Reject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	modID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.Reject(c.Request.Context(), modID, id, req.Reason); err != nil {
		handleModeratorErr(c, err)
		return
	}
	OKMsg(c, nil, "COMPANY_REJECTED")
}

func (h *ModeratorHandler) RequestDocs(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.RequestDocsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	modID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.RequestDocs(c.Request.Context(), modID, id, req.Message); err != nil {
		handleModeratorErr(c, err)
		return
	}
	OKMsg(c, nil, "DOCS_REQUESTED")
}

func (h *ModeratorHandler) GetHistory(c *gin.Context) {
	modID := c.MustGet("user_id").(uuid.UUID)
	page, perPage := parsePagination(c)
	list, total, err := h.svc.GetHistory(c.Request.Context(), modID, page, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	items := make([]dto.QueueItemResponse, len(list))
	for i := range list {
		items[i] = toQueueItem(&list[i])
	}
	Paginated(c, items, int(total), page, perPage)
}

func toQueueItem(c *model.Company) dto.QueueItemResponse {
	deadline := c.CreatedAt.Add(24 * time.Hour)
	return dto.QueueItemResponse{
		ID:              c.ID,
		CompanyName:     c.Name,
		Country:         c.Country,
		OrgType:         c.OrgType,
		INN:             c.INN,
		INNVerified:     c.INNVerified,
		Status:          c.Status,
		Phone:           c.Phone,
		Email:           c.Email,
		City:            c.City,
		Region:          c.Region,
		Street:          c.Street,
		RegDocURL:       c.RegDocURL,
		InnDocURL:       c.InnDocURL,
		RejectionReason: c.RejectionReason,
		DocsRequestNote: c.DocsRequestNote,
		UserName:        c.User.Name,
		UserPhone:       c.User.Phone,
		CreatedAt:       c.CreatedAt,
		Deadline:        deadline,
		IsUrgent:        time.Until(deadline) < 2*time.Hour,
	}
}

func handleModeratorErr(c *gin.Context, err error) {
	switch {
	case isErr(err, service.ErrCompanyNotFound):
		FailCode(c, http.StatusNotFound, "COMPANY_NOT_FOUND")
	default:
		InternalError(c)
	}
}
