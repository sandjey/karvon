package service

import (
	"context"
	"errors"
	"fmt"

	"ctm/pkg/email"
	"ctm/pkg/emailotp"
)

var ErrEmailOTPInvalid = errors.New("EMAIL_OTP_INVALID")
var ErrEmailOTPExpired = errors.New("EMAIL_OTP_EXPIRED")

type EmailService struct {
	store  *emailotp.Store
	sender *email.Sender
}

func NewEmailService(store *emailotp.Store, sender *email.Sender) *EmailService {
	return &EmailService{store: store, sender: sender}
}

func (s *EmailService) SendOTP(ctx context.Context, toEmail string) error {
	code, err := emailotp.GenerateCode()
	if err != nil { return err }
	if err := s.store.Save(ctx, toEmail, code); err != nil { return err }
	subject := "Код подтверждения / Tasdiqlash kodi"
	body := fmt.Sprintf("Ваш код подтверждения: %s\nTasdiqlash kodingiz: %s\n\nКод действителен 10 минут.", code, code)
	return s.sender.Send(toEmail, subject, body)
}

func (s *EmailService) VerifyOTP(ctx context.Context, emailAddr, code string) error {
	err := s.store.Verify(ctx, emailAddr, code)
	if errors.Is(err, emailotp.ErrInvalid) { return ErrEmailOTPInvalid }
	if errors.Is(err, emailotp.ErrExpired) { return ErrEmailOTPExpired }
	return err
}

func (s *EmailService) IsVerified(ctx context.Context, emailAddr string) bool {
	return s.store.IsVerified(ctx, emailAddr)
}
