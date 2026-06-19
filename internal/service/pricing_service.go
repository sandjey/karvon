package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"ctm/internal/model"
	"ctm/internal/repository"
)

type PricingService struct {
	repo *repository.PricingRepo
}

func NewPricingService(repo *repository.PricingRepo) *PricingService {
	return &PricingService{repo: repo}
}

// Tariffs возвращает все публичные активные тарифы, сгруппированные по префиксу ключа.
// tokens_* → packages, sub_* → subscriptions, listing_* → listings, boost_* → boosts.
// Ключи free_* считаются системными и в публичный список не попадают.
func (s *PricingService) Tariffs(ctx context.Context) (packages, subscriptions, listings, boosts []model.PricingConfig, err error) {
	all, err := s.repo.ListActive(ctx)
	if err != nil {
		return
	}
	for _, p := range all {
		switch {
		case strings.HasPrefix(p.Key, "tokens_"):
			packages = append(packages, p)
		case strings.HasPrefix(p.Key, "sub_"):
			subscriptions = append(subscriptions, p)
		case strings.HasPrefix(p.Key, "listing_"):
			listings = append(listings, p)
		case strings.HasPrefix(p.Key, "boost_"):
			boosts = append(boosts, p)
		}
	}
	return
}

func (s *PricingService) ListAll(ctx context.Context) ([]model.PricingConfig, error) {
	return s.repo.ListAll(ctx)
}

func (s *PricingService) Update(ctx context.Context, key string, fields map[string]interface{}, updatedBy uuid.UUID) error {
	p, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		return err
	}
	if p == nil {
		return ErrNotFound
	}
	fields["updated_by"] = updatedBy
	return s.repo.Update(ctx, key, fields)
}

func (s *PricingService) Create(ctx context.Context, p *model.PricingConfig, by uuid.UUID) error {
	existing, err := s.repo.FindByKey(ctx, p.Key)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrAlreadyExists
	}
	p.UpdatedBy = &by
	return s.repo.Create(ctx, p)
}

func (s *PricingService) Delete(ctx context.Context, key string) error {
	return s.repo.Delete(ctx, key)
}
