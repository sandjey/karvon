package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/service"
	"ctm/pkg/i18n"
)

type CarrierHandler struct {
	svc *service.CarrierService
}

func NewCarrierHandler(svc *service.CarrierService) *CarrierHandler {
	return &CarrierHandler{svc: svc}
}

func (h *CarrierHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	g := rg.Group("/carriers")
	g.GET("", h.List)
	g.GET("/:id", h.GetOne)
	g.POST("", auth, h.Create)
	g.PATCH("/:id", auth, h.Update)
	g.DELETE("/:id", auth, h.Delete)
}

func toCarrierResponse(c *model.CarrierCompany) dto.CarrierResponse {
	return dto.CarrierResponse{
		ID:              c.ID,
		UserID:          c.UserID,
		OrgType:         c.OrgType,
		Name:            c.Name,
		INN:             c.INN,
		Country:         c.Country,
		City:            c.City,
		Region:          c.Region,
		Phone:           c.Phone,
		Email:           c.Email,
		EmailVerified:   c.EmailVerified,
		EmailVerifiedAt: c.EmailVerifiedAt,
		Website:         c.Website,
		TransportType:   c.TransportType,
		WorkCountries:   []string(c.WorkCountries),
		Description:     c.Description,
		LogoURL:         c.LogoURL,
		Status:          c.Status,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

func handleCarrierErr(c *gin.Context, err error) {
	switch {
	case isErr(err, service.ErrCarrierNotFound):
		FailCode(c, http.StatusNotFound, "CARRIER_NOT_FOUND")
	case isErr(err, service.ErrCarrierNotOwned):
		FailCode(c, http.StatusForbidden, "FORBIDDEN")
	case isErr(err, service.ErrCarrierCountriesLimit):
		FailCode(c, http.StatusBadRequest, "CARRIER_COUNTRIES_LIMIT")
	default:
		InternalError(c)
	}
}

func (h *CarrierHandler) Create(c *gin.Context) {
	var req dto.CreateCarrierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	carrier, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		handleCarrierErr(c, err)
		return
	}
	CreatedMsg(c, toCarrierResponse(carrier), "CARRIER_CREATED")
}

func (h *CarrierHandler) List(c *gin.Context) {
	page, perPage := parsePagination(c)
	transportType := c.Query("transport_type")
	country := c.Query("country")
	list, total, err := h.svc.List(c.Request.Context(), transportType, country, page, perPage)
	if err != nil {
		InternalError(c)
		return
	}
	out := make([]dto.CarrierResponse, len(list))
	for i := range list {
		out[i] = toCarrierResponse(&list[i])
	}
	Paginated(c, out, int(total), page, perPage)
}

func (h *CarrierHandler) GetOne(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	carrier, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleCarrierErr(c, err)
		return
	}
	OK(c, toCarrierResponse(carrier))
}

func (h *CarrierHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	var req dto.UpdateCarrierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.Update(c.Request.Context(), userID, id, req); err != nil {
		handleCarrierErr(c, err)
		return
	}
	OKMsg(c, nil, "CARRIER_UPDATED")
}

func (h *CarrierHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		FailCode(c, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.svc.Delete(c.Request.Context(), userID, id); err != nil {
		handleCarrierErr(c, err)
		return
	}
	OKMsg(c, nil, "CARRIER_DELETED")
}
