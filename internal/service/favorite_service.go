package service

import (
	"context"

	"github.com/google/uuid"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/repository"
)

type FavoriteService struct {
	repo      *repository.FavoriteRepo
	cargo     *repository.CargoRepo
	warehouse *repository.WarehouseRepo
	media     *repository.MediaRepo
}

func NewFavoriteService(repo *repository.FavoriteRepo, cargo *repository.CargoRepo, warehouse *repository.WarehouseRepo, media *repository.MediaRepo) *FavoriteService {
	return &FavoriteService{repo: repo, cargo: cargo, warehouse: warehouse, media: media}
}

func (s *FavoriteService) Add(ctx context.Context, userID uuid.UUID, req dto.FavoriteRequest) error {
	exists, err := s.repo.Exists(ctx, userID, req.ListingType, req.ListingID)
	if err != nil {
		return err
	}
	if exists {
		return nil // идемпотентно
	}
	return s.repo.Add(ctx, &model.Favorite{
		UserID:      userID,
		ListingType: req.ListingType,
		ListingID:   req.ListingID,
	})
}

func (s *FavoriteService) Remove(ctx context.Context, userID uuid.UUID, req dto.FavoriteRequest) error {
	return s.repo.Remove(ctx, userID, req.ListingType, req.ListingID)
}

func (s *FavoriteService) List(ctx context.Context, userID uuid.UUID, listingType string, offset, limit int) ([]dto.FavoriteItemResponse, int64, error) {
	favs, total, err := s.repo.ListByUser(ctx, userID, listingType, offset, limit)
	if err != nil || len(favs) == 0 {
		return nil, total, err
	}

	var cargoIDs, warehouseIDs []uuid.UUID
	for _, f := range favs {
		if f.ListingType == "cargo" {
			cargoIDs = append(cargoIDs, f.ListingID)
		} else {
			warehouseIDs = append(warehouseIDs, f.ListingID)
		}
	}

	// Batch load listings
	cargoList, _ := s.cargo.FindByIDs(ctx, cargoIDs)
	warehouseList, _ := s.warehouse.FindByIDs(ctx, warehouseIDs)

	cargoMap := make(map[uuid.UUID]*model.CargoListing, len(cargoList))
	for _, c := range cargoList {
		cargoMap[c.ID] = c
	}
	warehouseMap := make(map[uuid.UUID]*model.WarehouseListing, len(warehouseList))
	for _, w := range warehouseList {
		warehouseMap[w.ID] = w
	}

	// Batch load first media (preview_image)
	cargoMedia, _ := s.media.ListByEntityIDs(ctx, "cargo", cargoIDs)
	warehouseMedia, _ := s.media.ListByEntityIDs(ctx, "warehouse", warehouseIDs)

	result := make([]dto.FavoriteItemResponse, 0, len(favs))
	for _, f := range favs {
		item := dto.FavoriteItemResponse{
			ID:          f.ID,
			ListingType: f.ListingType,
			ListingID:   f.ListingID,
			AddedAt:     f.CreatedAt,
		}

		if f.ListingType == "cargo" {
			if c, ok := cargoMap[f.ListingID]; ok {
				companyName := ""
				if c.Company != nil {
					companyName = c.Company.Name
				}
				status := c.Status
				if c.IsAdminBlocked {
					status = "blocked"
				}
				var previewImage *string
				if medias, ok := cargoMedia[f.ListingID]; ok && len(medias) > 0 {
					previewImage = &medias[0].FileURL
				}
				item.Listing = dto.FavoriteItemListing{
					ID:           c.ID,
					Name:         c.CargoName,
					CompanyName:  companyName,
					City:         c.FromCity,
					PricePerUnit: c.PricePerUnit,
					Currency:     c.Currency,
					Status:       status,
					PreviewImage: previewImage,
				}
			}
		} else {
			if w, ok := warehouseMap[f.ListingID]; ok {
				companyName := ""
				if w.Company != nil {
					companyName = w.Company.Name
				}
				status := w.Status
				if w.IsAdminBlocked {
					status = "blocked"
				}
				var previewImage *string
				if medias, ok := warehouseMedia[f.ListingID]; ok && len(medias) > 0 {
					previewImage = &medias[0].FileURL
				}
				item.Listing = dto.FavoriteItemListing{
					ID:           w.ID,
					Name:         w.Name,
					CompanyName:  companyName,
					City:         w.City,
					Status:       status,
					PreviewImage: previewImage,
				}
			}
		}

		result = append(result, item)
	}
	return result, total, nil
}
