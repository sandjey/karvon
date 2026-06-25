package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/service"
	"ctm/pkg/i18n"
)

type CompanyHandler struct {
	svc *service.CompanyService
}

func NewCompanyHandler(svc *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{svc: svc}
}

func (h *CompanyHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	g := rg.Group("/companies", auth)
	g.POST("", h.Create)
	g.GET("", h.GetAll)
	g.GET("/:id", h.GetOne)
	g.PATCH("/:id", h.Update)
}

func (h *CompanyHandler) Create(c *gin.Context) {
	var req dto.CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	company, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		handleCompanyErr(c, err)
		return
	}
	CreatedMsg(c, toCompanyResponse(company), "COMPANY_CREATED")
}

func (h *CompanyHandler) GetAll(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	list, err := h.svc.GetAll(c.Request.Context(), userID)
	if err != nil {
		InternalError(c)
		return
	}
	resp := make([]dto.CompanyResponse, len(list))
	for i, c2 := range list {
		resp[i] = toCompanyResponse(&c2)
	}
	OK(c, resp)
}

func (h *CompanyHandler) GetOne(c *gin.Context) {
	companyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	company, err := h.svc.GetByID(c.Request.Context(), userID, companyID)
	if err != nil {
		handleCompanyErr(c, err)
		return
	}
	OK(c, toCompanyResponse(company))
}

func (h *CompanyHandler) Update(c *gin.Context) {
	companyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.Update(c.Request.Context(), userID, companyID, req); err != nil {
		handleCompanyErr(c, err)
		return
	}
	company, _ := h.svc.GetByID(c.Request.Context(), userID, companyID)
	OKMsg(c, toCompanyResponse(company), "COMPANY_UPDATED")
}

func toCompanyResponse(c *model.Company) dto.CompanyResponse {
	media := make([]dto.MediaItem, len(c.Media))
	for i, m := range c.Media {
		media[i] = dto.MediaItem{FileURL: m.FileURL, FileType: m.FileType, OriginalName: m.OriginalName, SortOrder: m.SortOrder}
	}
	return dto.CompanyResponse{
		ID:              c.ID,
		Country:         c.Country,
		OrgType:         c.OrgType,
		Name:            c.Name,
		INN:             c.INN,
		INNVerified:     c.INNVerified,
		Phone:           c.Phone,
		Email:           c.Email,
		EmailVerified:   c.EmailVerified,
		EmailVerifiedAt: c.EmailVerifiedAt,
		City:            c.City,
		Region:          c.Region,
		Street:          c.Street,
		PostalCode:      c.PostalCode,
		RegDocURL:       c.RegDocURL,
		InnDocURL:       c.InnDocURL,
		LogoURL:         c.LogoURL,
		Media:           media,
		Status:          c.Status,
		RejectionReason: c.RejectionReason,
		DocsRequestNote: c.DocsRequestNote,
		VerifiedAt:      c.VerifiedAt,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

func handleCompanyErr(c *gin.Context, err error) {
	switch {
	case isErr(err, service.ErrCompanyNotFound):
		FailCode(c, http.StatusNotFound, "COMPANY_NOT_FOUND")
	case isErr(err, service.ErrCompanyNotOwned):
		Forbidden(c)
	case isErr(err, service.ErrCompanyNotEditable):
		FailCode(c, http.StatusConflict, "COMPANY_NOT_EDITABLE")
	case isErr(err, service.ErrAlreadyExists):
		FailCode(c, http.StatusConflict, "ALREADY_EXISTS")
	case isErr(err, service.ErrTooManyPhotos):
		FailCode(c, http.StatusBadRequest, "TOO_MANY_PHOTOS")
	default:
		log.Printf("[company] unhandled error: %v", err)
		InternalError(c)
	}
}
