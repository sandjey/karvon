package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"karvon/internal/model"
)

type OTPRepo struct {
	db *gorm.DB
}

func NewOTPRepo(db *gorm.DB) *OTPRepo {
	return &OTPRepo{db: db}
}

// CountRecentByPhone считает все OTP-запросы (включая использованные) за последние 10 минут.
// Используется для rate limiting: не зависит от статуса used, чтобы накопленные старые
// неиспользованные записи не блокировали пользователя навсегда.
func (r *OTPRepo) CountRecentByPhone(ctx context.Context, phone string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.OTPCode{}).
		Where("phone = ? AND created_at > ?", phone, time.Now().Add(-10*time.Minute)).
		Count(&count).Error
	return count, err
}

// InvalidateByPhone помечает все активные OTP для телефона как использованные.
func (r *OTPRepo) InvalidateByPhone(ctx context.Context, phone string) error {
	return r.db.WithContext(ctx).
		Model(&model.OTPCode{}).
		Where("phone = ? AND used = false", phone).
		Update("used", true).Error
}

func (r *OTPRepo) Create(ctx context.Context, otp *model.OTPCode) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

// FindActive ищет действующий (не истёкший, не использованный) OTP.
func (r *OTPRepo) FindActive(ctx context.Context, phone, code string) (*model.OTPCode, error) {
	var otp model.OTPCode
	err := r.db.WithContext(ctx).
		Where("phone = ? AND code = ? AND used = false AND expires_at > ?", phone, code, time.Now()).
		First(&otp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &otp, err
}

func (r *OTPRepo) IncrementAttempts(ctx context.Context, id interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.OTPCode{}).
		Where("id = ?", id).
		UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error
}

func (r *OTPRepo) MarkUsed(ctx context.Context, id interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.OTPCode{}).
		Where("id = ?", id).
		Update("used", true).Error
}

// DeleteOlderThan удаляет OTP-коды старше указанного времени (фоновая очистка).
func (r *OTPRepo) DeleteOlderThan(ctx context.Context, age time.Duration) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("created_at < ?", time.Now().Add(-age)).
		Delete(&model.OTPCode{})
	return res.RowsAffected, res.Error
}
