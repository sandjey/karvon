package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/repository"
)

type CompanyService struct {
	repo     *repository.CompanyRepo
	emailSvc *EmailService
	pricing  *repository.PricingRepo
	media    *repository.MediaRepo
}

func NewCompanyService(repo *repository.CompanyRepo, emailSvc *EmailService, pricing *repository.PricingRepo, media *repository.MediaRepo) *CompanyService {
	return &CompanyService{repo: repo, emailSvc: emailSvc, pricing: pricing, media: media}
}

func (s *CompanyService) saveMedia(ctx context.Context, id uuid.UUID, items []dto.MediaItem) {
	if items == nil {
		return
	}
	media := make([]model.ListingMedia, len(items))
	for i, it := range items {
		media[i] = model.ListingMedia{FileURL: it.FileURL, FileType: it.FileType, OriginalName: it.OriginalName, SortOrder: it.SortOrder}
	}
	_ = s.media.Replace(ctx, "company", id, media)
}

func (s *CompanyService) loadMedia(ctx context.Context, c *model.Company) {
	if c == nil {
		return
	}
	c.Media, _ = s.media.ListByEntity(ctx, "company", c.ID)
}

func (s *CompanyService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateCompanyRequest) (*model.Company, error) {
	if countPhotos(req.Media) > 5 {
		return nil, ErrTooManyPhotos
	}
	exists, err := s.repo.INNExistsApproved(ctx, req.INN, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	status := "pending"
	if s.pricing.IsCompanyModerationOff(ctx) {
		status = "approved"
	}

	c := &model.Company{
		ID:         uuid.New(),
		UserID:     userID,
		Country:    req.Country,
		OrgType:    req.OrgType,
		Name:       req.Name,
		INN:        req.INN,
		Phone:      req.Phone,
		Email:      req.Email,
		City:       req.City,
		Region:     req.Region,
		Street:     req.Street,
		PostalCode: req.PostalCode,
		LogoURL:    req.LogoURL,
		RegDocURL:  req.RegDocURL,
		InnDocURL:  req.InnDocURL,
		Status:     status,
	}
	if req.Email != "" && s.emailSvc.IsVerified(ctx, req.Email) {
		now := time.Now()
		c.EmailVerified = true
		c.EmailVerifiedAt = &now
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	s.saveMedia(ctx, c.ID, req.Media)
	s.loadMedia(ctx, c)
	return c, nil
}

func (s *CompanyService) GetAll(ctx context.Context, userID uuid.UUID) ([]model.Company, error) {
	list, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		s.loadMedia(ctx, &list[i])
	}
	return list, nil
}

func (s *CompanyService) GetByID(ctx context.Context, userID, companyID uuid.UUID) (*model.Company, error) {
	c, err := s.repo.FindByID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCompanyNotFound
	}
	if c.UserID != userID {
		return nil, ErrCompanyNotOwned
	}
	s.loadMedia(ctx, c)
	return c, nil
}

func (s *CompanyService) Update(ctx context.Context, userID, companyID uuid.UUID, req dto.UpdateCompanyRequest) error {
	if countPhotos(req.Media) > 5 {
		return ErrTooManyPhotos
	}
	c, err := s.repo.FindByID(ctx, companyID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCompanyNotFound
	}
	if c.UserID != userID {
		return ErrCompanyNotOwned
	}
	// pending — нельзя редактировать ничего, кроме docs_requested
	if c.Status == "pending" {
		return ErrCompanyNotEditable
	}

	fields := map[string]interface{}{}
	if req.Country != nil {
		fields["country"] = *req.Country
	}
	if req.OrgType != nil {
		fields["org_type"] = *req.OrgType
	}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Phone != nil {
		fields["phone"] = *req.Phone
	}
	if req.Email != nil {
		fields["email"] = *req.Email
		if s.emailSvc.IsVerified(ctx, *req.Email) {
			fields["email_verified"] = true
			now := time.Now()
			fields["email_verified_at"] = &now
		}
	}
	if req.City != nil {
		fields["city"] = *req.City
	}
	if req.Region != nil {
		fields["region"] = *req.Region
	}
	if req.Street != nil {
		fields["street"] = *req.Street
	}
	if req.PostalCode != nil {
		fields["postal_code"] = *req.PostalCode
	}
	if req.LogoURL != nil {
		fields["logo_url"] = *req.LogoURL
	}
	// документы можно обновить только при docs_requested
	if c.Status == "docs_requested" {
		if req.RegDocURL != nil {
			fields["reg_doc_url"] = *req.RegDocURL
		}
		if req.InnDocURL != nil {
			fields["inn_doc_url"] = *req.InnDocURL
		}
		// при обновлении после запроса документов — вернуть статус в pending
		if len(fields) > 0 {
			fields["status"] = "pending"
		}
	}
	if req.Media != nil {
		s.saveMedia(ctx, companyID, req.Media)
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(ctx, companyID, fields)
}
