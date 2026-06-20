package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ctm/internal/model"
)

type MediaRepo struct {
	db *gorm.DB
}

func NewMediaRepo(db *gorm.DB) *MediaRepo { return &MediaRepo{db: db} }

func (r *MediaRepo) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]model.ListingMedia, error) {
	var list []model.ListingMedia
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("sort_order ASC").Find(&list).Error
	for i := range list {
		list[i].FileURL = strings.ReplaceAll(list[i].FileURL, "/uploads/uploads/", "/uploads/")
	}
	return list, err
}

// Replace удаляет старые медиа сущности и вставляет новые (в транзакции).
func (r *MediaRepo) Replace(ctx context.Context, entityType string, entityID uuid.UUID, items []model.ListingMedia) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
			Delete(&model.ListingMedia{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].EntityType = entityType
			items[i].EntityID = entityID
			if items[i].SortOrder == 0 {
				items[i].SortOrder = i
			}
		}
		return tx.Create(&items).Error
	})
}

// ListByEntityIDs загружает медиа для нескольких entity за один SQL-запрос.
func (r *MediaRepo) ListByEntityIDs(ctx context.Context, entityType string, ids []uuid.UUID) (map[uuid.UUID][]model.ListingMedia, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var items []model.ListingMedia
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id IN ?", entityType, ids).
		Order("sort_order ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID][]model.ListingMedia, len(ids))
	for _, item := range items {
		item.FileURL = strings.ReplaceAll(item.FileURL, "/uploads/uploads/", "/uploads/")
		result[item.EntityID] = append(result[item.EntityID], item)
	}
	return result, nil
}

func (r *MediaRepo) DeleteByEntity(ctx context.Context, entityType string, entityID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Delete(&model.ListingMedia{}).Error
}
