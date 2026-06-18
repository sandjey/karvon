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
}

func NewFavoriteService(repo *repository.FavoriteRepo, cargo *repository.CargoRepo, warehouse *repository.WarehouseRepo) *FavoriteService {
	return &FavoriteService{repo: repo, cargo: cargo, warehouse: warehouse}
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

func (s *FavoriteService) List(ctx context.Context, userID uuid.UUID) ([]dto.FavoriteItem, error) {
	favs, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.FavoriteItem, len(favs))
	for i, f := range favs {
		out[i] = dto.FavoriteItem{Favorite: f, IsExpired: s.isExpired(ctx, f.ListingType, f.ListingID)}
	}
	return out, nil
}

func (s *FavoriteService) isExpired(ctx context.Context, listingType string, listingID uuid.UUID) bool {
	switch listingType {
	case "cargo":
		c, err := s.cargo.FindByID(ctx, listingID)
		if err != nil || c == nil {
			return true
		}
		return c.Status != "active"
	case "warehouse":
		w, err := s.warehouse.FindByID(ctx, listingID)
		if err != nil || w == nil {
			return true
		}
		return w.Status != "active"
	}
	return false
}
