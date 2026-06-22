package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/repository"
)

var ErrCarrierNotFound = errors.New("CARRIER_NOT_FOUND")
var ErrCarrierNotOwned = errors.New("FORBIDDEN")
var ErrCarrierCountriesLimit = errors.New("CARRIER_COUNTRIES_LIMIT")

type CarrierService struct{ repo *repository.CarrierRepo }

func NewCarrierService(repo *repository.CarrierRepo) *CarrierService {
	return &CarrierService{repo: repo}
}

func (s *CarrierService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateCarrierRequest) (*model.CarrierCompany, error) {
	if len(req.Countries) > 100 {
		return nil, ErrCarrierCountriesLimit
	}
	c := &model.CarrierCompany{
		ID:            uuid.New(),
		UserID:        userID,
		CompanyID:     req.CompanyID,
		Name:          req.Name,
		TransportType: req.TransportType,
		Countries:     pq.StringArray(req.Countries),
		Description:   req.Description,
		LogoURL:       req.LogoURL,
		Website:       req.Website,
		ContactPhone:  req.ContactPhone,
		ContactEmail:  req.ContactEmail,
		IsActive:      true,
	}
	return c, s.repo.Create(ctx, c)
}

func (s *CarrierService) List(ctx context.Context, transportType, country string, page, perPage int) ([]model.CarrierCompany, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.List(ctx, transportType, country, true, offset, perPage)
}

func (s *CarrierService) GetByID(ctx context.Context, id uuid.UUID) (*model.CarrierCompany, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCarrierNotFound
	}
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
	if len(req.Countries) > 100 {
		return ErrCarrierCountriesLimit
	}
	fields := map[string]interface{}{}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.TransportType != nil {
		fields["transport_type"] = *req.TransportType
	}
	if req.Countries != nil {
		fields["countries"] = pq.StringArray(req.Countries)
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.LogoURL != nil {
		fields["logo_url"] = *req.LogoURL
	}
	if req.Website != nil {
		fields["website"] = *req.Website
	}
	if req.ContactPhone != nil {
		fields["contact_phone"] = *req.ContactPhone
	}
	if req.ContactEmail != nil {
		fields["contact_email"] = *req.ContactEmail
	}
	if req.IsActive != nil {
		fields["is_active"] = *req.IsActive
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
	return s.repo.Delete(ctx, id)
}
