package service

import (
	"context"

	"github.com/google/uuid"

	"karvon/internal/dto"
	"karvon/internal/model"
	"karvon/internal/repository"
)

type RouteService struct {
	repo *repository.RouteRepo
}

func NewRouteService(repo *repository.RouteRepo) *RouteService {
	return &RouteService{repo: repo}
}

func (s *RouteService) Create(ctx context.Context, userID uuid.UUID, req dto.SaveRouteRequest) (*model.SavedRoute, error) {
	route := &model.SavedRoute{
		UserID:               userID,
		FromCity:             req.FromCity,
		ToCity:               req.ToCity,
		NotificationsEnabled: true,
	}
	if err := s.repo.Create(ctx, route); err != nil {
		return nil, err
	}
	return route, nil
}

func (s *RouteService) List(ctx context.Context, userID uuid.UUID) ([]model.SavedRoute, error) {
	return s.repo.FindByUser(ctx, userID)
}

func (s *RouteService) owned(ctx context.Context, id, userID uuid.UUID) (*model.SavedRoute, error) {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ErrNotFound
	}
	if r.UserID != userID {
		return nil, ErrNotOwner
	}
	return r, nil
}

func (s *RouteService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.owned(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *RouteService) SetNotifications(ctx context.Context, id, userID uuid.UUID, enabled bool) error {
	if _, err := s.owned(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.SetNotifications(ctx, id, enabled)
}
