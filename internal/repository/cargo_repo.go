package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"karvon/internal/model"
)

type CargoRepo struct {
	db *gorm.DB
}

func NewCargoRepo(db *gorm.DB) *CargoRepo { return &CargoRepo{db: db} }

// CargoFilter — фильтры листинга товаров.
type CargoFilter struct {
	Category      string
	FromCity      string
	FromCountry   string
	Divisibility  string
	Packaging     string
	MinOrderMax   *float64
	QtyMin        *float64
	QtyMax        *float64
	PriceMin      *float64
	PriceMax      *float64
	HasTempRegime bool
	IsADR         *bool
	VerifiedOnly  bool
	Sort          string // newest | price_asc | price_desc | quantity
	Offset        int
	Limit         int
}

func (r *CargoRepo) Create(ctx context.Context, c *model.CargoListing) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CargoRepo) Save(ctx context.Context, c *model.CargoListing) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *CargoRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.CargoListing, error) {
	var c model.CargoListing
	err := r.db.WithContext(ctx).
		Preload("Company").
		Preload("User").
		First(&c, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *CargoRepo) IncrementViews(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.CargoListing{}).
		Where("id = ?", id).
		UpdateColumn("views_count", gorm.Expr("views_count + 1")).Error
}

func (r *CargoRepo) UpdateStatus(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.CargoListing{}).
		Where("id = ?", id).Updates(fields).Error
}

func (r *CargoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.CargoListing{}, "id = ?", id).Error
}

// List — активные не-шаблонные карточки товаров. Boosted — сверху.
func (r *CargoRepo) List(ctx context.Context, f CargoFilter) ([]model.CargoListing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.CargoListing{}).
		Where("status = 'active' AND is_template = false")

	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	if f.FromCity != "" {
		q = q.Where("from_city ILIKE ?", "%"+f.FromCity+"%")
	}
	if f.FromCountry != "" {
		q = q.Where("from_country ILIKE ?", "%"+f.FromCountry+"%")
	}
	if f.Divisibility != "" {
		q = q.Where("divisibility = ?", f.Divisibility)
	}
	if f.Packaging != "" {
		q = q.Where("packaging = ?", f.Packaging)
	}
	if f.MinOrderMax != nil {
		q = q.Where("min_order <= ?", *f.MinOrderMax)
	}
	if f.QtyMin != nil {
		q = q.Where("quantity_available >= ?", *f.QtyMin)
	}
	if f.QtyMax != nil {
		q = q.Where("quantity_available <= ?", *f.QtyMax)
	}
	if f.PriceMin != nil {
		q = q.Where("rate_amount >= ?", *f.PriceMin)
	}
	if f.PriceMax != nil {
		q = q.Where("rate_amount <= ?", *f.PriceMax)
	}
	if f.HasTempRegime {
		q = q.Where("has_temp_regime = true")
	}
	if f.IsADR != nil {
		q = q.Where("is_adr = ?", *f.IsADR)
	}
	if f.VerifiedOnly {
		q = q.Joins("JOIN companies ON companies.id = cargo_listings.company_id").
			Where("companies.status = 'approved'")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "is_boosted DESC, created_at DESC"
	switch f.Sort {
	case "price_asc":
		order = "is_boosted DESC, rate_amount ASC NULLS LAST"
	case "price_desc":
		order = "is_boosted DESC, rate_amount DESC NULLS LAST"
	case "quantity":
		order = "is_boosted DESC, quantity_available DESC NULLS LAST"
	}

	var list []model.CargoListing
	err := q.Preload("Company").
		Order(order).
		Offset(f.Offset).Limit(f.Limit).
		Find(&list).Error
	return list, total, err
}

func (r *CargoRepo) ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.CargoListing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.CargoListing{}).
		Where("user_id = ? AND is_template = false", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.CargoListing
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *CargoRepo) ListTemplates(ctx context.Context, userID uuid.UUID) ([]model.CargoListing, error) {
	var list []model.CargoListing
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_template = true", userID).
		Order("created_at DESC").Find(&list).Error
	return list, err
}

// ExpireBoosts снимает истёкшие бусты. Возвращает число затронутых.
func (r *CargoRepo) ExpireBoosts(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.CargoListing{}).
		Where("is_boosted = true AND boost_expires_at IS NOT NULL AND boost_expires_at < now()").
		Update("is_boosted", false)
	return res.RowsAffected, res.Error
}

// SetBoost включает буст до указанного времени.
func (r *CargoRepo) SetBoost(ctx context.Context, id uuid.UUID, until interface{}) error {
	return r.db.WithContext(ctx).Model(&model.CargoListing{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"is_boosted": true, "boost_expires_at": until}).Error
}
