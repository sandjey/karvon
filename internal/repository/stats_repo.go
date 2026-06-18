package repository

import (
	"context"

	"gorm.io/gorm"

	"ctm/internal/model"
)

type StatsRepo struct {
	db *gorm.DB
}

func NewStatsRepo(db *gorm.DB) *StatsRepo { return &StatsRepo{db: db} }

type PublicStats struct {
	ActiveListings     int64   `json:"active_listings"`
	VerifiedCompanies  int64   `json:"verified_companies"`
	CountriesCovered   int64   `json:"countries_covered"`
	WarehouseSpacesM2  float64 `json:"warehouse_spaces_m2"`
}

func (r *StatsRepo) PublicStats(ctx context.Context) (*PublicStats, error) {
	var cargoCnt, warehouseCnt, companies, countries int64
	var spaces float64

	if err := r.db.WithContext(ctx).Model(&model.CargoListing{}).
		Where("status = 'active' AND is_admin_blocked = false AND is_template = false").
		Count(&cargoCnt).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("status = 'active' AND is_admin_blocked = false").
		Count(&warehouseCnt).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&model.Company{}).
		Where("status = 'approved'").
		Count(&companies).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&model.CargoListing{}).
		Where("status = 'active' AND is_admin_blocked = false AND from_country IS NOT NULL AND from_country <> ''").
		Distinct("from_country").
		Count(&countries).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&model.WarehouseListing{}).
		Where("status = 'active' AND is_admin_blocked = false").
		Select("COALESCE(SUM(area_free_m2), 0)").
		Scan(&spaces).Error; err != nil {
		return nil, err
	}

	return &PublicStats{
		ActiveListings:    cargoCnt + warehouseCnt,
		VerifiedCompanies: companies,
		CountriesCovered:  countries,
		WarehouseSpacesM2: spaces,
	}, nil
}
