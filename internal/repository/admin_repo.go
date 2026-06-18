package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ctm/internal/model"
)

type AdminRepo struct {
	db *gorm.DB
}

func NewAdminRepo(db *gorm.DB) *AdminRepo { return &AdminRepo{db: db} }

type DashboardMetrics struct {
	NewUsers          int64   `json:"new_users"`
	NewCargo          int64   `json:"new_cargo"`
	NewWarehouses     int64   `json:"new_warehouses"`
	RevenueUZS        float64 `json:"revenue_uzs"`
	RevenueUSD        float64 `json:"revenue_usd"`
	ContactsBought    int64   `json:"contacts_bought"`
	ActiveSubs        int64   `json:"active_subscriptions"`
	UrgentVerifications int64 `json:"urgent_verifications"`
}

func (r *AdminRepo) Dashboard(ctx context.Context, since time.Time) (*DashboardMetrics, error) {
	m := &DashboardMetrics{}
	db := r.db.WithContext(ctx)

	db.Model(&model.User{}).Where("created_at >= ?", since).Count(&m.NewUsers)
	db.Model(&model.CargoListing{}).Where("created_at >= ? AND is_template = false", since).Count(&m.NewCargo)
	db.Model(&model.WarehouseListing{}).Where("created_at >= ?", since).Count(&m.NewWarehouses)
	db.Model(&model.ContactView{}).Where("viewed_at >= ? AND tokens_spent > 0", since).Count(&m.ContactsBought)
	db.Model(&model.Subscription{}).Where("is_active = true AND expires_at > ?", time.Now()).Count(&m.ActiveSubs)

	urgentBefore := time.Now().Add(2 * time.Hour).Add(-24 * time.Hour)
	db.Model(&model.Company{}).Where("status = 'pending' AND created_at <= ?", urgentBefore).Count(&m.UrgentVerifications)

	type sumRow struct {
		Currency string
		Total    float64
	}
	var rows []sumRow
	db.Model(&model.PaymentOrder{}).
		Select("currency, COALESCE(SUM(amount),0) AS total").
		Where("status = 'paid' AND paid_at >= ?", since).
		Group("currency").Scan(&rows)
	for _, row := range rows {
		if row.Currency == "USD" {
			m.RevenueUSD = row.Total
		} else {
			m.RevenueUZS = row.Total
		}
	}
	return m, nil
}

func (r *AdminRepo) Users(ctx context.Context, search, role string, offset, limit int) ([]model.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.User{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("phone ILIKE ? OR name ILIKE ?", like, like)
	}
	if role != "" {
		q = q.Where("role = ?", role)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.User
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *AdminRepo) SetBlocked(ctx context.Context, id uuid.UUID, blocked bool) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("is_blocked", blocked).Error
}

func (r *AdminRepo) AllCompanies(ctx context.Context, status string, offset, limit int) ([]model.Company, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Company{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Company
	err := q.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *AdminRepo) AllCargo(ctx context.Context, offset, limit int) ([]model.CargoListing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.CargoListing{}).Where("is_template = false")
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.CargoListing
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *AdminRepo) AllWarehouses(ctx context.Context, offset, limit int) ([]model.WarehouseListing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.WarehouseListing{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.WarehouseListing
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *AdminRepo) AllPayments(ctx context.Context, offset, limit int) ([]model.PaymentOrder, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.PaymentOrder{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PaymentOrder
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}
