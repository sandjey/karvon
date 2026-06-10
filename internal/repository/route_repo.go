package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"karvon/internal/model"
)

type RouteRepo struct {
	db *gorm.DB
}

func NewRouteRepo(db *gorm.DB) *RouteRepo { return &RouteRepo{db: db} }

func (r *RouteRepo) Create(ctx context.Context, route *model.SavedRoute) error {
	return r.db.WithContext(ctx).Create(route).Error
}

func (r *RouteRepo) FindByUser(ctx context.Context, userID uuid.UUID) ([]model.SavedRoute, error) {
	var list []model.SavedRoute
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *RouteRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.SavedRoute, error) {
	var route model.SavedRoute
	err := r.db.WithContext(ctx).First(&route, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &route, err
}

func (r *RouteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.SavedRoute{}, "id = ?", id).Error
}

func (r *RouteRepo) SetNotifications(ctx context.Context, id uuid.UUID, enabled bool) error {
	return r.db.WithContext(ctx).Model(&model.SavedRoute{}).
		Where("id = ?", id).Update("notifications_enabled", enabled).Error
}

// FindMatching возвращает маршруты с включёнными уведомлениями, совпадающие по from/to.
func (r *RouteRepo) FindMatching(ctx context.Context, fromCity, toCity string) ([]model.SavedRoute, error) {
	var list []model.SavedRoute
	q := r.db.WithContext(ctx).Where("notifications_enabled = true")
	if fromCity != "" {
		q = q.Where("? ILIKE '%' || from_city || '%'", fromCity)
	}
	if toCity != "" {
		q = q.Where("? ILIKE '%' || to_city || '%'", toCity)
	}
	err := q.Find(&list).Error
	return list, err
}
