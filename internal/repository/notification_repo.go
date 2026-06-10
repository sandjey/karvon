package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"karvon/internal/model"
)

type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(ctx context.Context, n *model.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

// List — уведомления пользователя: сначала непрочитанные, затем по дате.
func (r *NotificationRepo) List(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.Notification, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ?", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Notification
	err := q.Order("is_read ASC, created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *NotificationRepo) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).Count(&count).Error
	return count, err
}

func (r *NotificationRepo) MarkRead(ctx context.Context, userID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).Update("is_read", true).Error
}

func (r *NotificationRepo) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).Update("is_read", true).Error
}
