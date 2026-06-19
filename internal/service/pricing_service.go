package service

import (
	"context"

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

// Tariffs возвращает все публичные тарифы: токены, подписки, листинги, бусты.
func (s *PricingService) Tariffs(ctx context.Context) (packages, subscriptions, listings, boosts []model.PricingConfig, err error) {
	if packages, err = s.repo.ListByPrefix(ctx, "tokens_"); err != nil {
		return
	}
	if subscriptions, err = s.repo.ListByPrefix(ctx, "sub_"); err != nil {
		return
	}
	if listings, err = s.repo.ListByKeys(ctx, []string{"listing_paid"}); err != nil {
		return
	}
	boosts, err = s.repo.ListByPrefix(ctx, "boost_")
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
