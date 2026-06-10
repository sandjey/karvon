package service

import (
	"context"

	"github.com/google/uuid"

	"karvon/internal/model"
	"karvon/internal/repository"
)

type NotificationService struct {
	repo *repository.NotificationRepo
}

func NewNotificationService(repo *repository.NotificationRepo) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) List(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.Notification, int64, error) {
	return s.repo.List(ctx, userID, offset, limit)
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.UnreadCount(ctx, userID)
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.MarkRead(ctx, userID, id)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllRead(ctx, userID)
}
