package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"karvon/internal/model"
)

type CargoRepo struct {
	db *gorm.DB
}

func NewCargoRepo(db *gorm.DB) *CargoRepo { return &CargoRepo{db: db} }

// CargoFilter описывает параметры фильтрации списка грузов.
type CargoFilter struct {
	FromCity     string
	ToCity       string
	Type         string
	LoadingType  string
	CargoType    string
	BodyTypes    []string
	WeightMin    *float64
	WeightMax    *float64
	DateFrom     *time.Time
	VerifiedOnly bool
	Sort         string // newest | date | weight
	Offset       int
	Limit        int
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
		Preload("Waypoints", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
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

func (r *CargoRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&model.CargoListing{}).
		Where("id = ?", id).Update("status", status).Error
}

func (r *CargoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.CargoListing{}, "id = ?", id).Error
}

// List возвращает активные (не шаблонные) объявления с фильтрами. Boosted — сверху.
func (r *CargoRepo) List(ctx context.Context, f CargoFilter) ([]model.CargoListing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.CargoListing{}).
		Where("status = 'active' AND is_template = false")

	if f.FromCity != "" {
		q = q.Where("from_city ILIKE ?", "%"+f.FromCity+"%")
	}
	if f.ToCity != "" {
		q = q.Where("to_city ILIKE ?", "%"+f.ToCity+"%")
	}
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if f.LoadingType != "" {
		q = q.Where("loading_type = ?", f.LoadingType)
	}
	if f.CargoType != "" {
		q = q.Where("cargo_type ILIKE ?", "%"+f.CargoType+"%")
	}
	if len(f.BodyTypes) > 0 {
		q = q.Where("body_types && ?", pq.StringArray(f.BodyTypes))
	}
	if f.WeightMin != nil {
		q = q.Where("weight_ton >= ?", *f.WeightMin)
	}
	if f.WeightMax != nil {
		q = q.Where("weight_ton <= ?", *f.WeightMax)
	}
	if f.DateFrom != nil {
		q = q.Where("from_date >= ?", *f.DateFrom)
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
	case "date":
		order = "is_boosted DESC, from_date ASC NULLS LAST"
	case "weight":
		order = "is_boosted DESC, weight_ton DESC NULLS LAST"
	}

	var list []model.CargoListing
	err := q.Preload("Company").
		Order(order).
		Offset(f.Offset).Limit(f.Limit).
		Find(&list).Error
	return list, total, err
}

// ListByUser — объявления пользователя (включая архив, без шаблонов).
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

// ListTemplates — шаблоны пользователя.
func (r *CargoRepo) ListTemplates(ctx context.Context, userID uuid.UUID) ([]model.CargoListing, error) {
	var list []model.CargoListing
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_template = true", userID).
		Order("created_at DESC").Find(&list).Error
	return list, err
}

// ExpireDue переводит истёкшие активные объявления в archived. Возвращает их ID.
func (r *CargoRepo) ExpireDue(ctx context.Context) ([]model.CargoListing, error) {
	var due []model.CargoListing
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Where("status = 'active' AND is_template = false AND expires_at IS NOT NULL AND expires_at < ?", now).
		Find(&due).Error; err != nil {
		return nil, err
	}
	if len(due) == 0 {
		return nil, nil
	}
	ids := make([]uuid.UUID, len(due))
	for i, c := range due {
		ids[i] = c.ID
	}
	err := r.db.WithContext(ctx).Model(&model.CargoListing{}).
		Where("id IN ?", ids).Update("status", "archived").Error
	return due, err
}
