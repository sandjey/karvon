package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"karvon/internal/model"
)

type PaymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepo(db *gorm.DB) *PaymentRepo { return &PaymentRepo{db: db} }

func (r *PaymentRepo) Create(ctx context.Context, o *model.PaymentOrder) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *PaymentRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.PaymentOrder, error) {
	var o model.PaymentOrder
	err := r.db.WithContext(ctx).First(&o, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &o, err
}

func (r *PaymentRepo) MarkPaid(ctx context.Context, id uuid.UUID, method string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.PaymentOrder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": "paid", "paid_at": now, "payment_method": method}).Error
}

func (r *PaymentRepo) SaveInvoiceUUID(ctx context.Context, id uuid.UUID, invoiceUUID string) error {
	return r.db.WithContext(ctx).Model(&model.PaymentOrder{}).
		Where("id = ?", id).
		Update("rahmat_order_id", invoiceUUID).Error
}

func (r *PaymentRepo) History(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.PaymentOrder, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.PaymentOrder{}).Where("user_id = ?", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PaymentOrder
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *PaymentRepo) CreateSubscription(ctx context.Context, s *model.Subscription) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *PaymentRepo) ActiveSubscription(ctx context.Context, userID uuid.UUID) (*model.Subscription, error) {
	var s model.Subscription
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true AND expires_at > ?", userID, time.Now()).
		Order("expires_at DESC").First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}
