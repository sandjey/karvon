package notifier

import "context"

type Channel string

const (
	ChannelWhatsApp Channel = "whatsapp"
	ChannelTelegram Channel = "telegram"
)

// Notifier отправляет OTP-коды пользователю.
type Notifier interface {
	// Send отправляет сообщение.
	// to — номер телефона (WhatsApp) или chat_id (Telegram).
	Send(ctx context.Context, to, message string) error
}
