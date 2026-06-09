package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"karvon/internal/model"
	"karvon/internal/repository"
)

type ModeratorService struct {
	companyRepo *repository.CompanyRepo
	notifRepo   *repository.NotificationRepo
}

func NewModeratorService(companyRepo *repository.CompanyRepo, notifRepo *repository.NotificationRepo) *ModeratorService {
	return &ModeratorService{companyRepo: companyRepo, notifRepo: notifRepo}
}

func (s *ModeratorService) GetQueue(ctx context.Context, status string, page, perPage int) ([]model.Company, int64, error) {
	if status == "" {
		status = "pending"
	}
	offset := (page - 1) * perPage
	return s.companyRepo.FindQueue(ctx, status, offset, perPage)
}

func (s *ModeratorService) GetQueueItem(ctx context.Context, id uuid.UUID) (*model.Company, error) {
	c, err := s.companyRepo.FindByIDWithUser(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCompanyNotFound
	}
	return c, nil
}

func (s *ModeratorService) Approve(ctx context.Context, moderatorID, companyID uuid.UUID) error {
	c, err := s.companyRepo.FindByID(ctx, companyID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCompanyNotFound
	}
	now := time.Now()
	if err := s.companyRepo.Update(ctx, companyID, map[string]interface{}{
		"status":       "approved",
		"moderator_id": moderatorID,
		"verified_at":  now,
	}); err != nil {
		return err
	}
	body := "Ваша компания «" + c.Name + "» прошла проверку и одобрена."
	_ = s.notifRepo.Create(ctx, &model.Notification{
		UserID: c.UserID,
		Type:   "verification_approved",
		Title:  "Компания одобрена",
		Body:   &body,
	})
	return nil
}

func (s *ModeratorService) Reject(ctx context.Context, moderatorID, companyID uuid.UUID, reason string) error {
	c, err := s.companyRepo.FindByID(ctx, companyID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCompanyNotFound
	}
	if err := s.companyRepo.Update(ctx, companyID, map[string]interface{}{
		"status":           "rejected",
		"rejection_reason": reason,
		"moderator_id":     moderatorID,
	}); err != nil {
		return err
	}
	body := "Ваша компания «" + c.Name + "» отклонена. Причина: " + reason
	_ = s.notifRepo.Create(ctx, &model.Notification{
		UserID: c.UserID,
		Type:   "verification_rejected",
		Title:  "Компания отклонена",
		Body:   &body,
	})
	return nil
}

func (s *ModeratorService) RequestDocs(ctx context.Context, moderatorID, companyID uuid.UUID, message string) error {
	c, err := s.companyRepo.FindByID(ctx, companyID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCompanyNotFound
	}
	if err := s.companyRepo.Update(ctx, companyID, map[string]interface{}{
		"status":       "docs_requested",
		"moderator_id": moderatorID,
	}); err != nil {
		return err
	}
	_ = s.notifRepo.Create(ctx, &model.Notification{
		UserID: c.UserID,
		Type:   "docs_requested",
		Title:  "Требуются документы",
		Body:   &message,
	})
	return nil
}

func (s *ModeratorService) GetHistory(ctx context.Context, moderatorID uuid.UUID, page, perPage int) ([]model.Company, int64, error) {
	offset := (page - 1) * perPage
	return s.companyRepo.FindModeratorHistory(ctx, moderatorID, offset, perPage)
}
