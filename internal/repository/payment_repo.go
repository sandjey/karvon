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

// SavePaymentDetails сохраняет данные транзакции из webhook-callback Multicard.
func (r *PaymentRepo) SavePaymentDetails(ctx context.Context, id uuid.UUID, paymentUUID, phone, receiptURL string, totalTiyin, commissionTiyin int64) error {
	fields := map[string]interface{}{}
	if paymentUUID != "" {
		fields["payment_uuid"] = paymentUUID
	}
	if phone != "" {
		fields["phone"] = phone
	}
	if receiptURL != "" {
		fields["receipt_url"] = receiptURL
	}
	if totalTiyin > 0 {
		fields["total_amount_tiyin"] = totalTiyin
	}
	if commissionTiyin > 0 {
		fields["commission_tiyin"] = commissionTiyin
	}
	if len(fields) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.PaymentOrder{}).Where("id = ?", id).Updates(fields).Error
}

// MarkReverted помечает заказ как возвращённый (revert от Multicard).
func (r *PaymentRepo) MarkReverted(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.PaymentOrder{}).
		Where("id = ?", id).
		Update("status", "reverted").Error
}

// DeactivateSubscriptionByOrder деактивирует подписку, созданную по данному заказу.
func (r *PaymentRepo) DeactivateSubscriptionByOrder(ctx context.Context, orderID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Subscription{}).
		Where("payment_order_id = ?", orderID).
		Update("is_active", false).Error
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

// FindExpiringSoon возвращает активные подписки, истекающие в течение [from, to].
func (r *PaymentRepo) FindExpiringSoon(ctx context.Context, from, to time.Time) ([]model.Subscription, error) {
	var list []model.Subscription
	err := r.db.WithContext(ctx).
		Where("is_active = true AND expires_at >= ? AND expires_at < ?", from, to).
		Find(&list).Error
	return list, err
}
