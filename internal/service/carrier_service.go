package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/repository"
)

var ErrCarrierNotFound = errors.New("CARRIER_NOT_FOUND")
var ErrCarrierNotOwned = errors.New("FORBIDDEN")
var ErrCarrierCountriesLimit = errors.New("CARRIER_COUNTRIES_LIMIT")

type CarrierService struct {
	repo     *repository.CarrierRepo
	emailSvc *EmailService
	media    *repository.MediaRepo
}

func NewCarrierService(repo *repository.CarrierRepo, emailSvc *EmailService, media *repository.MediaRepo) *CarrierService {
	return &CarrierService{repo: repo, emailSvc: emailSvc, media: media}
}

func (s *CarrierService) saveMedia(ctx context.Context, id uuid.UUID, items []dto.MediaItem) {
	if items == nil {
		return
	}
	media := make([]model.ListingMedia, len(items))
	for i, it := range items {
		media[i] = model.ListingMedia{FileURL: it.FileURL, FileType: it.FileType, OriginalName: it.OriginalName, SortOrder: it.SortOrder}
	}
	_ = s.media.Replace(ctx, "carrier", id, media)
}

func (s *CarrierService) loadMedia(ctx context.Context, c *model.CarrierCompany) {
	if c == nil {
		return
	}
	c.Media, _ = s.media.ListByEntity(ctx, "carrier", c.ID)
}

func (s *CarrierService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateCarrierRequest) (*model.CarrierCompany, error) {
	if len(req.WorkCountries) > 100 {
		return nil, ErrCarrierCountriesLimit
	}
	if countPhotos(req.Media) > 5 {
		return nil, ErrTooManyPhotos
	}
	c := &model.CarrierCompany{
		ID:            uuid.New(),
		UserID:        userID,
		OrgType:       req.OrgType,
		Name:          req.Name,
		INN:           req.INN,
		Country:       req.Country,
		City:          req.City,
		Region:        req.Region,
		Phone:         req.Phone,
		Email:         req.Email,
		Website:       req.Website,
		TransportType: req.TransportType,
		WorkCountries: pq.StringArray(req.WorkCountries),
		Description:   req.Description,
		LogoURL:       req.LogoURL,
		Status:        "active",
	}
	if req.Email != nil && s.emailSvc.IsVerified(ctx, *req.Email) {
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

func (s *CarrierService) List(ctx context.Context, transportType, country string, page, perPage int) ([]model.CarrierCompany, int64, error) {
	offset := (page - 1) * perPage
	list, total, err := s.repo.List(ctx, transportType, country, true, offset, perPage)
	if err == nil {
		for i := range list {
			s.loadMedia(ctx, &list[i])
		}
	}
	return list, total, err
}

func (s *CarrierService) ListMine(ctx context.Context, userID uuid.UUID, page, perPage int) ([]model.CarrierCompany, int64, error) {
	offset := (page - 1) * perPage
	list, total, err := s.repo.ListByUser(ctx, userID, offset, perPage)
	if err == nil {
		for i := range list {
			s.loadMedia(ctx, &list[i])
		}
	}
	return list, total, err
}

func (s *CarrierService) GetByID(ctx context.Context, id uuid.UUID) (*model.CarrierCompany, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCarrierNotFound
	}
	s.loadMedia(ctx, c)
	return c, nil
}

func (s *CarrierService) Update(ctx context.Context, userID, id uuid.UUID, req dto.UpdateCarrierRequest) error {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCarrierNotFound
	}
	if c.UserID != userID {
		return ErrCarrierNotOwned
	}
	if len(req.WorkCountries) > 100 {
		return ErrCarrierCountriesLimit
	}
	if countPhotos(req.Media) > 5 {
		return ErrTooManyPhotos
	}
	fields := map[string]interface{}{}
	if req.OrgType != nil {
		fields["org_type"] = *req.OrgType
	}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.INN != nil {
		fields["inn"] = *req.INN
	}
	if req.Country != nil {
		fields["country"] = *req.Country
	}
	if req.City != nil {
		fields["city"] = *req.City
	}
	if req.Region != nil {
		fields["region"] = *req.Region
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
	if req.Website != nil {
		fields["website"] = *req.Website
	}
	if req.TransportType != nil {
		fields["transport_type"] = *req.TransportType
	}
	if req.WorkCountries != nil {
		fields["work_countries"] = pq.StringArray(req.WorkCountries)
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.LogoURL != nil {
		fields["logo_url"] = *req.LogoURL
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Media != nil {
		s.saveMedia(ctx, id, req.Media)
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(ctx, id, fields)
}

func (s *CarrierService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCarrierNotFound
	}
	if c.UserID != userID {
		return ErrCarrierNotOwned
	}
	_ = s.media.DeleteByEntity(ctx, "carrier", id)
	return s.repo.Delete(ctx, id)
}
