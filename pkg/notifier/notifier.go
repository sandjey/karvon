package notifier

import (
	"context"
	"errors"
)

type Channel string

const (
	ChannelWhatsApp Channel = "whatsapp"
	ChannelTelegram Channel = "telegram"
)

// ErrNumberNotRegistered возвращается CheckNumber когда номер не найден на платформе.
var ErrNumberNotRegistered = errors.New("phone number not registered on this platform")

// ErrRateLimit возвращается Send когда API вернул rate-limit (FLOOD_WAIT и подобные).
var ErrRateLimit = errors.New("rate limit exceeded")

// Notifier отправляет OTP-коды пользователю.
type Notifier interface {
	// Send отправляет сообщение.
	// to — номер телефона (WhatsApp) или chat_id (Telegram).
	Send(ctx context.Context, to, message string) error
}

// NumberChecker проверяет, зарегистрирован ли номер на платформе.
type NumberChecker interface {
	CheckNumber(ctx context.Context, phone string) error
}
