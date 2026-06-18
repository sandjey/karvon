package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ctm/internal/model"
)

type FavoriteRepo struct {
	db *gorm.DB
}

func NewFavoriteRepo(db *gorm.DB) *FavoriteRepo { return &FavoriteRepo{db: db} }

func (r *FavoriteRepo) Exists(ctx context.Context, userID uuid.UUID, listingType string, listingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Favorite{}).
		Where("user_id = ? AND listing_type = ? AND listing_id = ?", userID, listingType, listingID).
		Count(&count).Error
	return count > 0, err
}

func (r *FavoriteRepo) Add(ctx context.Context, f *model.Favorite) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *FavoriteRepo) Remove(ctx context.Context, userID uuid.UUID, listingType string, listingID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND listing_type = ? AND listing_id = ?", userID, listingType, listingID).
		Delete(&model.Favorite{}).Error
}

func (r *FavoriteRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Favorite, error) {
	var list []model.Favorite
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error
	return list, err
}

// FavoriteUserRow — кто добавил объявление в избранное (имя + компания).
type FavoriteUserRow struct {
	UserName    *string
	CompanyName *string
	CreatedAt   time.Time
}

// UsersByListing возвращает пользователей, добавивших объявление в избранное (для статистики автора).
func (r *FavoriteRepo) UsersByListing(ctx context.Context, listingType string, listingID uuid.UUID) ([]FavoriteUserRow, error) {
	var rows []FavoriteUserRow
	err := r.db.WithContext(ctx).
		Table("favorites").
		Select("users.name AS user_name, (SELECT c.name FROM companies c WHERE c.user_id = favorites.user_id ORDER BY c.created_at LIMIT 1) AS company_name, favorites.created_at AS created_at").
		Joins("JOIN users ON users.id = favorites.user_id").
		Where("favorites.listing_type = ? AND favorites.listing_id = ?", listingType, listingID).
		Order("favorites.created_at DESC").
		Scan(&rows).Error
	return rows, err
}
