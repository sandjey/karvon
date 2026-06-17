package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"karvon/internal/model"
)

type WarehouseRepo struct {
	db *gorm.DB
}

func NewWarehouseRepo(db *gorm.DB) *WarehouseRepo { return &WarehouseRepo{db: db} }

type WarehouseFilter struct {
	Region        string
	WarehouseType string
	AreaMin       *float64
	AreaMax       *float64
	TempMin       *float64
	TempMax       *float64
	Services      []string
	Sort          string
	Offset        int
	Limit         int
}

func (r *WarehouseRepo) Create(ctx context.Context, w *model.WarehouseListing) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *WarehouseRepo) Save(ctx context.Context, w *model.WarehouseListing) error {
	return r.db.WithContext(ctx).Save(w).Error
}

func (r *WarehouseRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.WarehouseListing, error) {
	var w model.WarehouseListing
	err := r.db.WithContext(ctx).Preload("Company").First(&w, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &w, err
}

func (r *WarehouseRepo) IncrementViews(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("id = ?", id).
		UpdateColumn("views_count", gorm.Expr("views_count + 1")).Error
}

func (r *WarehouseRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("id = ?", id).Update("status", status).Error
}

// MarkPaid активирует оплаченное объявление (is_paid=true, status=active).
func (r *WarehouseRepo) MarkPaid(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"is_paid": true, "status": "active"}).Error
}

// RevertPaid откатывает оплату (is_paid=false, status=archived).
func (r *WarehouseRepo) RevertPaid(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"is_paid": false, "status": "archived"}).Error
}

func (r *WarehouseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.WarehouseListing{}, "id = ?", id).Error
}

// CountFreeActive считает активные бесплатные склады пользователя.
func (r *WarehouseRepo) CountFreeActive(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("user_id = ? AND is_paid = false AND status = 'active'", userID).
		Count(&count).Error
	return count, err
}

// IncrementContactsBought увеличивает счётчик купленных контактов склада.
func (r *WarehouseRepo) IncrementContactsBought(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("id = ?", id).
		UpdateColumn("contacts_bought_count", gorm.Expr("contacts_bought_count + 1")).Error
}

func (r *WarehouseRepo) UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("id = ?", id).Updates(fields).Error
}

func (r *WarehouseRepo) List(ctx context.Context, f WarehouseFilter) ([]model.WarehouseListing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.WarehouseListing{}).Where("status = 'active' AND is_admin_blocked = false")
	if f.Region != "" {
		q = q.Where("region ILIKE ?", "%"+f.Region+"%")
	}
	if f.WarehouseType != "" {
		q = q.Where("warehouse_type = ?", f.WarehouseType)
	}
	if f.AreaMin != nil {
		q = q.Where("area_total_m2 >= ?", *f.AreaMin)
	}
	if f.AreaMax != nil {
		q = q.Where("area_total_m2 <= ?", *f.AreaMax)
	}
	if f.TempMin != nil {
		q = q.Where("temp_min <= ?", *f.TempMin)
	}
	if f.TempMax != nil {
		q = q.Where("temp_max >= ?", *f.TempMax)
	}
	if len(f.Services) > 0 {
		q = q.Where("services && ?", pq.StringArray(f.Services))
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "is_boosted DESC, created_at DESC"
	if f.Sort == "area" {
		order = "is_boosted DESC, area_total_m2 DESC NULLS LAST"
	}
	var list []model.WarehouseListing
	err := q.Preload("Company").Order(order).Offset(f.Offset).Limit(f.Limit).Find(&list).Error
	return list, total, err
}

func (r *WarehouseRepo) ExpireBoosts(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("is_boosted = true AND boost_expires_at IS NOT NULL AND boost_expires_at < now()").
		Update("is_boosted", false)
	return res.RowsAffected, res.Error
}

func (r *WarehouseRepo) SetBoost(ctx context.Context, id uuid.UUID, until interface{}) error {
	return r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"is_boosted": true, "boost_expires_at": until}).Error
}

// RemoveBoost снимает буст (используется при revert-платеже).
func (r *WarehouseRepo) RemoveBoost(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"is_boosted": false, "boost_expires_at": nil}).Error
}

// Similar возвращает похожие активные склады того же типа и региона.
func (r *WarehouseRepo) Similar(ctx context.Context, excludeID uuid.UUID, warehouseType, region string, limit int) ([]model.WarehouseListing, error) {
	var list []model.WarehouseListing
	q := r.db.WithContext(ctx).
		Where("status = 'active' AND is_admin_blocked = false AND id <> ?", excludeID)
	if warehouseType != "" {
		q = q.Where("warehouse_type = ?", warehouseType)
	}
	if region != "" {
		q = q.Where("region ILIKE ?", "%"+region+"%")
	}
	err := q.Preload("Company").
		Order("is_boosted DESC, created_at DESC").
		Limit(limit).Find(&list).Error
	return list, err
}

// CountActiveByUser считает активные склады пользователя.
func (r *WarehouseRepo) CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("user_id = ? AND status = 'active'", userID).
		Count(&count).Error
	return count, err
}

func (r *WarehouseRepo) ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.WarehouseListing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.WarehouseListing{}).Where("user_id = ?", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.WarehouseListing
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}
