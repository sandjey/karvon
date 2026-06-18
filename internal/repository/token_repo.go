package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ctm/internal/model"
)

type TokenRepo struct {
	db *gorm.DB
}

func NewTokenRepo(db *gorm.DB) *TokenRepo {
	return &TokenRepo{db: db}
}

func (r *TokenRepo) Save(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	return r.db.WithContext(ctx).Create(&model.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}).Error
}

func (r *TokenRepo) Find(ctx context.Context, token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&rt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &rt, err
}

func (r *TokenRepo) Delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Where("token = ?", token).
		Delete(&model.RefreshToken{}).Error
}

func (r *TokenRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.RefreshToken{}).Error
}
