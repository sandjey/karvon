package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"karvon/internal/model"
)

type CompanyRepo struct {
	db *gorm.DB
}

func NewCompanyRepo(db *gorm.DB) *CompanyRepo {
	return &CompanyRepo{db: db}
}

func (r *CompanyRepo) Create(ctx context.Context, c *model.Company) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CompanyRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Company, error) {
	var c model.Company
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *CompanyRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Company, error) {
	var list []model.Company
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// INNExistsApproved returns true if another approved company already uses this INN.
func (r *CompanyRepo) INNExistsApproved(ctx context.Context, inn string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Company{}).
		Where("inn = ? AND status = 'approved'", inn)
	if excludeID != nil {
		q = q.Where("id != ?", *excludeID)
	}
	var count int64
	err := q.Count(&count).Error
	return count > 0, err
}

func (r *CompanyRepo) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.Company{}).
		Where("id = ?", id).
		Updates(fields).Error
}

// FindByIDWithUser preloads the User association.
func (r *CompanyRepo) FindByIDWithUser(ctx context.Context, id uuid.UUID) (*model.Company, error) {
	var c model.Company
	err := r.db.WithContext(ctx).Preload("User").First(&c, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

// FindQueue returns companies with the given status, ordered by created_at ASC (oldest first).
func (r *CompanyRepo) FindQueue(ctx context.Context, status string, offset, limit int) ([]model.Company, int64, error) {
	var list []model.Company
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Company{}).Where("status = ?", status)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("User").Order("created_at ASC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// FindModeratorHistory returns companies processed by a given moderator.
func (r *CompanyRepo) FindModeratorHistory(ctx context.Context, modID uuid.UUID, offset, limit int) ([]model.Company, int64, error) {
	var list []model.Company
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Company{}).
		Where("moderator_id = ?", modID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("User").Order("updated_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}
