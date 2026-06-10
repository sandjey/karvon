package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"karvon/internal/model"
)

// DashboardEvent — событие ленты дашборда (кто купил контакт / добавил в избранное).
type DashboardEvent struct {
	Type        string    `json:"type"`
	ActorName   *string   `json:"actor_name"`
	ListingType string    `json:"listing_type"`
	ListingID   uuid.UUID `json:"listing_id"`
	At          time.Time `json:"at"`
}

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

// DashboardEvents — лента событий по объявлениям пользователя: кто купил контакт и кто добавил в избранное.
func (r *UserRepo) DashboardEvents(ctx context.Context, userID uuid.UUID, limit int) ([]DashboardEvent, error) {
	q := `
		WITH my_listings AS (
			SELECT id, 'cargo' AS lt FROM cargo_listings WHERE user_id = ?
			UNION ALL
			SELECT id, 'warehouse' AS lt FROM warehouse_listings WHERE user_id = ?
		)
		SELECT 'contact_purchased' AS type, u.name AS actor_name, cv.listing_type, cv.listing_id, cv.viewed_at AS at
		FROM contact_views cv
		JOIN users u ON u.id = cv.user_id
		WHERE cv.listing_id IN (SELECT id FROM my_listings) AND cv.user_id <> ?
		UNION ALL
		SELECT 'favorited' AS type, u.name AS actor_name, f.listing_type, f.listing_id, f.created_at AS at
		FROM favorites f
		JOIN users u ON u.id = f.user_id
		WHERE f.listing_id IN (SELECT id FROM my_listings) AND f.user_id <> ?
		ORDER BY at DESC
		LIMIT ?`
	var events []DashboardEvent
	err := r.db.WithContext(ctx).Raw(q, userID, userID, userID, userID, limit).Scan(&events).Error
	return events, err
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
