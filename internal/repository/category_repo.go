package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ctm/internal/model"
)

type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) ListActive(ctx context.Context) ([]model.CargoCategory, error) {
	var list []model.CargoCategory
	err := r.db.WithContext(ctx).Where("is_active = true").Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *CategoryRepo) ListAll(ctx context.Context) ([]model.CargoCategory, error) {
	var list []model.CargoCategory
	err := r.db.WithContext(ctx).Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *CategoryRepo) FindByKey(ctx context.Context, key string) (*model.CargoCategory, error) {
	var c model.CargoCategory
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *CategoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.CargoCategory, error) {
	var c model.CargoCategory
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *CategoryRepo) Create(ctx context.Context, c *model.CargoCategory) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CategoryRepo) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.CargoCategory{}).Where("id = ?", id).Updates(fields).Error
}

func (r *CategoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.CargoCategory{}, "id = ?", id).Error
}
