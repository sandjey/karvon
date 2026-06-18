package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/repository"
)

type WarehouseService struct {
	repo      *repository.WarehouseRepo
	company   *repository.CompanyRepo
	media     *repository.MediaRepo
	cargoRepo *repository.CargoRepo
	favorites *repository.FavoriteRepo
}

func NewWarehouseService(
	repo *repository.WarehouseRepo,
	company *repository.CompanyRepo,
	media *repository.MediaRepo,
	cargoRepo *repository.CargoRepo,
	favorites *repository.FavoriteRepo,
) *WarehouseService {
	return &WarehouseService{repo: repo, company: company, media: media, cargoRepo: cargoRepo, favorites: favorites}
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

// validateCreate проверяет обязательные поля при создании склада.
func validateCreate(req *dto.WarehouseUpsertRequest) error {
	if req.WarehouseType == nil || *req.WarehouseType == "" {
		return ErrValidation
	}
	if req.Name == nil || *req.Name == "" {
		return ErrValidation
	}
	if req.ContactPerson == nil || *req.ContactPerson == "" {
		return ErrValidation
	}
	switch *req.WarehouseType {
	case "cold":
		if req.TempMin == nil || req.TempMax == nil || len(req.ColdChamberTypes) == 0 {
			return ErrValidation
		}
	case "customs":
		if req.CustomsLicenseNumber == nil || req.CustomsLicenseIssued == nil || req.CustomsLicenseExpires == nil {
			return ErrValidation
		}
	}
	if req.AreaTotalM2 != nil && req.AreaFreeM2 != nil && *req.AreaFreeM2 > *req.AreaTotalM2 {
		return ErrValidation
	}
	return nil
}

func (s *WarehouseService) Similar(ctx context.Context, id uuid.UUID, warehouseType string, region *string) ([]model.WarehouseListing, error) {
	r := ""
	if region != nil {
		r = *region
	}
	return s.repo.Similar(ctx, id, warehouseType, r, 5)
}

func (s *WarehouseService) Create(ctx context.Context, userID uuid.UUID, req dto.WarehouseUpsertRequest) (*model.WarehouseListing, error) {
	if err := validateCreate(&req); err != nil {
		return nil, err
	}
	if countPhotos(req.Media) > 10 {
		return nil, ErrPhotoLimitExceeded
	}
	companyID, err := s.resolveCompany(ctx, userID, req.CompanyID)
	if err != nil {
		return nil, err
	}

	// Проверяем: есть ли уже бесплатное активное объявление (груз или склад).
	freeCargo, err := s.cargoRepo.CountFreeActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	freeWarehouse, err := s.repo.CountFreeActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	initialStatus := "active"
	isPaid := false
	if freeCargo+freeWarehouse > 0 {
		initialStatus = "archived"
		isPaid = false
	}

	w := &model.WarehouseListing{
		ID:        uuid.New(),
		CompanyID: companyID,
		UserID:    userID,
		Status:    initialStatus,
		IsPaid:    isPaid,
	}
	applyWarehouse(&req, w)
	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}
	s.saveMedia(ctx, w.ID, req.Media)
	return s.loadFull(ctx, w.ID)
}

func (s *WarehouseService) GetByID(ctx context.Context, id, viewerID uuid.UUID) (*model.WarehouseListing, error) {
	w, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, ErrListingNotFound
	}
	if w.IsAdminBlocked && w.UserID != viewerID {
		return nil, ErrListingNotFound
	}
	if w.UserID != viewerID {
		_ = s.repo.IncrementViews(ctx, id)
	}
	w.Media, _ = s.media.ListByEntity(ctx, "warehouse", id)
	return w, nil
}

func (s *WarehouseService) loadFull(ctx context.Context, id uuid.UUID) (*model.WarehouseListing, error) {
	w, err := s.repo.FindByID(ctx, id)
	if err != nil || w == nil {
		return w, err
	}
	w.Media, _ = s.media.ListByEntity(ctx, "warehouse", id)
	return w, nil
}

func (s *WarehouseService) saveMedia(ctx context.Context, id uuid.UUID, items []dto.MediaItem) {
	if items == nil {
		return
	}
	media := make([]model.ListingMedia, len(items))
	for i, it := range items {
		media[i] = model.ListingMedia{FileURL: it.FileURL, FileType: it.FileType, OriginalName: it.OriginalName, SortOrder: it.SortOrder}
	}
	_ = s.media.Replace(ctx, "warehouse", id, media)
}

func (s *WarehouseService) List(ctx context.Context, f repository.WarehouseFilter) ([]model.WarehouseListing, int64, error) {
	list, total, err := s.repo.List(ctx, f)
	if err == nil {
		s.attachMedia(ctx, list)
	}
	return list, total, err
}

func (s *WarehouseService) attachMedia(ctx context.Context, list []model.WarehouseListing) {
	if len(list) == 0 {
		return
	}
	ids := make([]uuid.UUID, len(list))
	for i := range list {
		ids[i] = list[i].ID
	}
	mediaMap, _ := s.media.ListByEntityIDs(ctx, "warehouse", ids)
	for i := range list {
		list[i].Media = mediaMap[list[i].ID]
	}
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
	if countPhotos(req.Media) > 10 {
		return nil, ErrPhotoLimitExceeded
	}
	w, err := s.owned(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	applyWarehouse(&req, w)
	if err := s.repo.Save(ctx, w); err != nil {
		return nil, err
	}
	if req.Media != nil {
		s.saveMedia(ctx, id, req.Media)
	}
	return s.loadFull(ctx, id)
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
	_ = s.media.DeleteByEntity(ctx, "warehouse", id)
	return s.repo.Delete(ctx, id)
}

func (s *WarehouseService) Stats(ctx context.Context, id, userID uuid.UUID) (*dto.WarehouseStatsResponse, error) {
	w, err := s.owned(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.favorites.UsersByListing(ctx, "warehouse", id)
	if err != nil {
		return nil, err
	}
	favs := make([]dto.FavoriteUser, len(rows))
	for i, r := range rows {
		favs[i] = dto.FavoriteUser{UserName: r.UserName, CompanyName: r.CompanyName, AddedAt: r.CreatedAt}
	}
	return &dto.WarehouseStatsResponse{
		ViewsCount:          w.ViewsCount,
		ContactsBoughtCount: w.ContactsBoughtCount,
		Favorites:           favs,
	}, nil
}

func applyWarehouse(req *dto.WarehouseUpsertRequest, w *model.WarehouseListing) {
	if req.WarehouseType != nil {
		w.WarehouseType = *req.WarehouseType
	}
	if req.Name != nil {
		w.Name = *req.Name
	}
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
	if req.CustomsSpecialServices != nil {
		w.CustomsSpecialServices = pq.StringArray(req.CustomsSpecialServices)
	}
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
