package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"karvon/internal/dto"
	"karvon/internal/model"
	"karvon/internal/repository"
)

type WarehouseService struct {
	repo    *repository.WarehouseRepo
	company *repository.CompanyRepo
}

func NewWarehouseService(repo *repository.WarehouseRepo, company *repository.CompanyRepo) *WarehouseService {
	return &WarehouseService{repo: repo, company: company}
}

func (s *WarehouseService) resolveCompany(ctx context.Context, userID uuid.UUID, explicit *uuid.UUID) (uuid.UUID, error) {
	companies, err := s.company.FindByUserID(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}
	if explicit != nil {
		for _, c := range companies {
			if c.ID == *explicit && c.Status == "approved" {
				return c.ID, nil
			}
		}
		return uuid.Nil, ErrCompanyNotOwned
	}
	for _, c := range companies {
		if c.Status == "approved" {
			return c.ID, nil
		}
	}
	return uuid.Nil, ErrCompanyNotFound
}

// validateByType проверяет обязательные поля по типу склада.
func validateByType(req *dto.WarehouseUpsertRequest) error {
	switch req.WarehouseType {
	case "cold":
		if req.TempMin == nil || req.TempMax == nil || len(req.ColdChamberTypes) == 0 {
			return ErrValidation
		}
	case "customs":
		if req.CustomsLicenseNumber == nil || req.CustomsLicenseIssued == nil || req.CustomsLicenseExpires == nil {
			return ErrValidation
		}
	}
	// свободная площадь не больше общей
	if req.AreaTotalM2 != nil && req.AreaFreeM2 != nil && *req.AreaFreeM2 > *req.AreaTotalM2 {
		return ErrValidation
	}
	return nil
}

func (s *WarehouseService) Create(ctx context.Context, userID uuid.UUID, req dto.WarehouseUpsertRequest) (*model.WarehouseListing, error) {
	if err := validateByType(&req); err != nil {
		return nil, err
	}
	companyID, err := s.resolveCompany(ctx, userID, req.CompanyID)
	if err != nil {
		return nil, err
	}
	w := &model.WarehouseListing{
		ID:        uuid.New(),
		CompanyID: companyID,
		UserID:    userID,
		Status:    "active",
	}
	applyWarehouse(&req, w)
	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, w.ID)
}

func (s *WarehouseService) GetByID(ctx context.Context, id, viewerID uuid.UUID) (*model.WarehouseListing, error) {
	w, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, ErrListingNotFound
	}
	if w.UserID != viewerID {
		_ = s.repo.IncrementViews(ctx, id)
	}
	return w, nil
}

func (s *WarehouseService) List(ctx context.Context, f repository.WarehouseFilter) ([]model.WarehouseListing, int64, error) {
	return s.repo.List(ctx, f)
}

func (s *WarehouseService) owned(ctx context.Context, id, userID uuid.UUID) (*model.WarehouseListing, error) {
	w, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, ErrListingNotFound
	}
	if w.UserID != userID {
		return nil, ErrNotOwner
	}
	return w, nil
}

func (s *WarehouseService) Update(ctx context.Context, id, userID uuid.UUID, req dto.WarehouseUpsertRequest) (*model.WarehouseListing, error) {
	w, err := s.owned(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if err := validateByType(&req); err != nil {
		return nil, err
	}
	applyWarehouse(&req, w)
	if err := s.repo.Save(ctx, w); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *WarehouseService) SetStatus(ctx context.Context, id, userID uuid.UUID, status string) error {
	if _, err := s.owned(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *WarehouseService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.owned(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func applyWarehouse(req *dto.WarehouseUpsertRequest, w *model.WarehouseListing) {
	w.WarehouseType = req.WarehouseType
	w.Name = req.Name
	w.Region = req.Region
	w.Address = req.Address
	w.Lat = req.Lat
	w.Lng = req.Lng
	w.PhoneMain = req.PhoneMain
	w.PhoneExtra = req.PhoneExtra
	w.ContactPerson = req.ContactPerson
	w.Email = req.Email
	w.Website = req.Website
	if req.Specialization != nil {
		w.Specialization = pq.StringArray(req.Specialization)
	}
	w.AreaTotalM2 = req.AreaTotalM2
	w.AreaFreeM2 = req.AreaFreeM2
	w.CeilingHeightM = req.CeilingHeightM
	w.HeatingType = req.HeatingType
	if req.StorageType != nil {
		w.StorageType = pq.StringArray(req.StorageType)
	}
	w.TempMin = req.TempMin
	w.TempMax = req.TempMax
	if req.ColdChamberTypes != nil {
		w.ColdChamberTypes = pq.StringArray(req.ColdChamberTypes)
	}
	w.CustomsLicenseNumber = req.CustomsLicenseNumber
	w.CustomsLicenseIssued = req.CustomsLicenseIssued
	w.CustomsLicenseExpires = req.CustomsLicenseExpires
	if req.Infrastructure != nil {
		w.Infrastructure = pq.StringArray(req.Infrastructure)
	}
	if req.Services != nil {
		w.Services = pq.StringArray(req.Services)
	}
	if req.WorkingHours != nil {
		w.WorkingHours = req.WorkingHours
	}
}
