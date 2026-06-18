package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ctm/internal/model"
	"ctm/internal/repository"
	"ctm/pkg/email"
)

type EmailSender interface {
	Send(to, subject, body string) error
	Enabled() bool
}

type ModeratorService struct {
	companyRepo *repository.CompanyRepo
	notifRepo   *repository.NotificationRepo
	mailer      EmailSender
}

func NewModeratorService(companyRepo *repository.CompanyRepo, notifRepo *repository.NotificationRepo, mailer EmailSender) *ModeratorService {
	return &ModeratorService{companyRepo: companyRepo, notifRepo: notifRepo, mailer: mailer}
}

func (s *ModeratorService) sendEmail(to, subject, body string) {
	if s.mailer == nil || !s.mailer.Enabled() || to == "" {
		return
	}
	if err := s.mailer.Send(to, subject, body); err != nil {
		log.Warn().Err(err).Str("to", to).Msg("failed to send email")
	}
}

// ensure *email.Sender satisfies EmailSender
var _ EmailSender = (*email.Sender)(nil)

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
	c, err := s.companyRepo.FindByIDWithUser(ctx, companyID)
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
	s.sendEmail(c.Email,
		"Компания одобрена — Central Trade Market",
		"Ваша компания «"+c.Name+"» прошла проверку и одобрена. Теперь вы можете публиковать объявления.",
	)
	return nil
}

func (s *ModeratorService) Reject(ctx context.Context, moderatorID, companyID uuid.UUID, reason string) error {
	c, err := s.companyRepo.FindByIDWithUser(ctx, companyID)
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
	s.sendEmail(c.Email,
		"Компания отклонена — Central Trade Market",
		"Ваша компания «"+c.Name+"» отклонена.\nПричина: "+reason+"\n\nПодайте заявку повторно с исправленными данными.",
	)
	return nil
}

func (s *ModeratorService) RequestDocs(ctx context.Context, moderatorID, companyID uuid.UUID, message string) error {
	c, err := s.companyRepo.FindByIDWithUser(ctx, companyID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCompanyNotFound
	}
	if err := s.companyRepo.Update(ctx, companyID, map[string]interface{}{
		"status":            "docs_requested",
		"moderator_id":      moderatorID,
		"docs_request_note": message,
	}); err != nil {
		return err
	}
	_ = s.notifRepo.Create(ctx, &model.Notification{
		UserID: c.UserID,
		Type:   "docs_requested",
		Title:  "Требуются документы",
		Body:   &message,
	})
	s.sendEmail(c.Email,
		"Требуются дополнительные документы — Central Trade Market",
		"По компании «"+c.Name+"» запрошены дополнительные документы.\n\n"+message+
			"\n\nОтправьте документы через Telegram на аккаунт поддержки.",
	)
	return nil
}

func (s *ModeratorService) GetHistory(ctx context.Context, moderatorID uuid.UUID, page, perPage int) ([]model.Company, int64, error) {
	offset := (page - 1) * perPage
	return s.companyRepo.FindModeratorHistory(ctx, moderatorID, offset, perPage)
}
