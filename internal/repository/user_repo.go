package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"karvon/internal/model"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepo) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("name", name).Error
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Updates(fields).Error
}

func (r *UserRepo) CountCompanies(ctx context.Context, userID uuid.UUID) (total int64, verified int64, err error) {
	if err = r.db.WithContext(ctx).Model(&model.Company{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return
	}
	err = r.db.WithContext(ctx).Model(&model.Company{}).Where("user_id = ? AND status = 'approved'", userID).Count(&verified).Error
	return
}

func (r *UserRepo) CreditTokens(ctx context.Context, userID uuid.UUID, amount int, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.Clauses().Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}

		newBalance := user.TokenBalance + amount
		if err := tx.Model(&user).Update("token_balance", newBalance).Error; err != nil {
			return err
		}

		return tx.Create(&model.TokenTransaction{
			UserID:       userID,
			Type:         "credit",
			Amount:       amount,
			Reason:       reason,
			BalanceAfter: newBalance,
		}).Error
	})
}
