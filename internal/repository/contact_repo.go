package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"karvon/internal/model"
)

// ErrNoTokens возвращается из транзакции списания, когда баланс недостаточен.
var ErrNoTokens = errors.New("insufficient tokens")

type ContactRepo struct {
	db *gorm.DB
}

func NewContactRepo(db *gorm.DB) *ContactRepo { return &ContactRepo{db: db} }

// HasActiveSubscription — есть ли у пользователя активная подписка.
func (r *ContactRepo) HasActiveSubscription(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Subscription{}).
		Where("user_id = ? AND is_active = true AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count > 0, err
}

// RecentView возвращает запись просмотра контакта за последние `within`, если есть.
func (r *ContactRepo) RecentView(ctx context.Context, userID uuid.UUID, listingType string, listingID uuid.UUID, within time.Duration) (*model.ContactView, error) {
	var v model.ContactView
	since := time.Now().Add(-within)
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND listing_type = ? AND listing_id = ? AND viewed_at >= ?", userID, listingType, listingID, since).
		Order("viewed_at DESC").First(&v).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// RecordFree фиксирует бесплатный просмотр (подписка/повтор), без списания.
func (r *ContactRepo) RecordFree(ctx context.Context, userID uuid.UUID, listingType string, listingID uuid.UUID) error {
	return r.db.WithContext(ctx).Create(&model.ContactView{
		UserID:      userID,
		ListingType: listingType,
		ListingID:   listingID,
		TokensSpent: 0,
	}).Error
}

// DebitAndRecord атомарно: списывает 1 токен, пишет contact_view + token_transaction,
// инкрементирует contacts_bought_count объявления. Возвращает новый баланс.
func (r *ContactRepo) DebitAndRecord(ctx context.Context, userID uuid.UUID, listingType string, listingID uuid.UUID) (int, error) {
	var newBalance int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			return err
		}
		if user.TokenBalance < 1 {
			return ErrNoTokens
		}
		newBalance = user.TokenBalance - 1
		if err := tx.Model(&model.User{}).Where("id = ?", userID).
			Update("token_balance", newBalance).Error; err != nil {
			return err
		}
		view := &model.ContactView{
			UserID:      userID,
			ListingType: listingType,
			ListingID:   listingID,
			TokensSpent: 1,
		}
		if err := tx.Create(view).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.TokenTransaction{
			UserID:       userID,
			Type:         "debit",
			Amount:       1,
			Reason:       "contact_view",
			ReferenceID:  &view.ID,
			BalanceAfter: newBalance,
		}).Error; err != nil {
			return err
		}
		table := "cargo_listings"
		if listingType == "warehouse" {
			table = "warehouse_listings"
		}
		return tx.Table(table).Where("id = ?", listingID).
			UpdateColumn("contacts_bought_count", gorm.Expr("contacts_bought_count + 1")).Error
	})
	return newBalance, err
}

func (r *ContactRepo) History(ctx context.Context, userID uuid.UUID, within time.Duration) ([]model.ContactView, error) {
	var list []model.ContactView
	since := time.Now().Add(-within)
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND viewed_at >= ?", userID, since).
		Order("viewed_at DESC").Find(&list).Error
	return list, err
}

func (r *ContactRepo) TokenTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]model.TokenTransaction, error) {
	var list []model.TokenTransaction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").Limit(limit).Find(&list).Error
	return list, err
}
