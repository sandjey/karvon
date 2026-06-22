package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctm/internal/model"
)

type PricingRepo struct {
	db *gorm.DB
}

func NewPricingRepo(db *gorm.DB) *PricingRepo { return &PricingRepo{db: db} }

func (r *PricingRepo) ListActive(ctx context.Context) ([]model.PricingConfig, error) {
	var list []model.PricingConfig
	err := r.db.WithContext(ctx).Where("is_active = true").Order("key ASC").Find(&list).Error
	return list, err
}

func (r *PricingRepo) ListAll(ctx context.Context) ([]model.PricingConfig, error) {
	var list []model.PricingConfig
	err := r.db.WithContext(ctx).Order("key ASC").Find(&list).Error
	return list, err
}

// ListByKeys возвращает активные тарифы по точным ключам.
func (r *PricingRepo) ListByKeys(ctx context.Context, keys []string) ([]model.PricingConfig, error) {
	var list []model.PricingConfig
	err := r.db.WithContext(ctx).
		Where("is_active = true AND key IN ?", keys).
		Order("key ASC").Find(&list).Error
	return list, err
}

// ListByPrefix возвращает активные тарифы, ключ которых начинается с префикса (tokens_, sub_).
func (r *PricingRepo) ListByPrefix(ctx context.Context, prefix string) ([]model.PricingConfig, error) {
	var list []model.PricingConfig
	err := r.db.WithContext(ctx).
		Where("is_active = true AND key LIKE ?", prefix+"%").
		Order("key ASC").Find(&list).Error
	return list, err
}

func (r *PricingRepo) FindByKey(ctx context.Context, key string) (*model.PricingConfig, error) {
	var p model.PricingConfig
	err := r.db.WithContext(ctx).First(&p, "key = ?", key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *PricingRepo) Update(ctx context.Context, key string, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.PricingConfig{}).
		Where("key = ?", key).Updates(fields).Error
}

func (r *PricingRepo) Create(ctx context.Context, p *model.PricingConfig) error {
	// Явно указываем колонки, иначе GORM пропускает is_active=false (zero value + default:true в теге).
	return r.db.WithContext(ctx).
		Select("key", "label", "price_uzs", "price_usd", "tokens_amount", "duration_days", "is_active", "updated_by").
		Create(p).Error
}

func (r *PricingRepo) Delete(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("key = ?", key).Delete(&model.PricingConfig{}).Error
}

// TokensAmount возвращает кол-во токенов по ключу тарифа, либо fallback если ключа нет.
func (r *PricingRepo) TokensAmount(ctx context.Context, key string, fallback int) int {
	p, err := r.FindByKey(ctx, key)
	if err != nil || p == nil {
		return fallback
	}
	return p.TokensAmount
}

// IsFreeMode возвращает true если глобальный бесплатный режим включён.
func (r *PricingRepo) IsFreeMode(ctx context.Context) bool {
	return r.TokensAmount(ctx, "system:free_mode", 0) > 0
}

// IsFreeCargo returns true if cargo creation is free (no quota limit).
func (r *PricingRepo) IsFreeCargo(ctx context.Context) bool {
	return r.IsFreeMode(ctx) || r.TokensAmount(ctx, "system:free_cargo", 0) > 0
}

// IsFreeWarehouse returns true if warehouse creation is free (no quota limit).
func (r *PricingRepo) IsFreeWarehouse(ctx context.Context) bool {
	return r.IsFreeMode(ctx) || r.TokensAmount(ctx, "system:free_warehouse", 0) > 0
}

// IsFreeContacts returns true if contact viewing is free (no token debit).
func (r *PricingRepo) IsFreeContacts(ctx context.Context) bool {
	return r.IsFreeMode(ctx) || r.TokensAmount(ctx, "system:free_contacts", 0) > 0
}
