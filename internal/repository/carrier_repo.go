package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ctm/internal/model"
)

type CarrierRepo struct{ db *gorm.DB }

func NewCarrierRepo(db *gorm.DB) *CarrierRepo { return &CarrierRepo{db: db} }

func (r *CarrierRepo) Create(ctx context.Context, c *model.CarrierCompany) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CarrierRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.CarrierCompany, error) {
	var c model.CarrierCompany
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

// List с фильтрами: transport_type (опц.), country ISO-код (опц.), пагинация
func (r *CarrierRepo) List(ctx context.Context, transportType, country string, onlyActive bool, offset, limit int) ([]model.CarrierCompany, int64, error) {
	var list []model.CarrierCompany
	var total int64
	q := r.db.WithContext(ctx).Model(&model.CarrierCompany{})
	if onlyActive {
		q = q.Where("status = 'active'")
	}
	if transportType != "" {
		q = q.Where("transport_type = ?", transportType)
	}
	if country != "" {
		q = q.Where("? = ANY(work_countries)", country)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// ListByUser возвращает компании-перевозчики пользователя (все статусы) с пагинацией.
func (r *CarrierRepo) ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.CarrierCompany, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.CarrierCompany{}).Where("user_id = ?", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.CarrierCompany
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// FindByIDs — батч-загрузка перевозчиков по списку ID.
func (r *CarrierRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.CarrierCompany, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []*model.CarrierCompany
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error
	return list, err
}

func (r *CarrierRepo) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.CarrierCompany{}).Where("id = ?", id).Updates(fields).Error
}

func (r *CarrierRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.CarrierCompany{}, "id = ?", id).Error
}
