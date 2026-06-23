package service

import (
	"context"
	"errors"
	"fmt"
	"html"

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

// otpEmailStrings — локализованные строки письма с OTP-кодом.
type otpEmailStrings struct {
	subject  string
	preview  string // короткий превью-текст (виден в списке писем)
	heading  string
	intro    string // «Ваш код подтверждения для CentralTradeMarket:»
	validity string // «Код действителен 10 минут.»
	warning  string // «Никому не передавайте этот код.»
	footer   string
}

var otpEmailLocales = map[string]otpEmailStrings{
	"ru": {
		subject:  "Код подтверждения — CentralTradeMarket",
		preview:  "Ваш код подтверждения для CentralTradeMarket",
		heading:  "Подтверждение входа",
		intro:    "Ваш код подтверждения для CentralTradeMarket:",
		validity: "Код действителен 10 минут.",
		warning:  "Никому не передавайте этот код.",
		footer:   "Если вы не запрашивали код, просто проигнорируйте это письмо.",
	},
	"en": {
		subject:  "Verification code — CentralTradeMarket",
		preview:  "Your CentralTradeMarket verification code",
		heading:  "Sign-in verification",
		intro:    "Your verification code for CentralTradeMarket:",
		validity: "The code is valid for 10 minutes.",
		warning:  "Do not share this code with anyone.",
		footer:   "If you didn't request this code, please ignore this email.",
	},
	"uz": {
		subject:  "Tasdiqlash kodi — CentralTradeMarket",
		preview:  "CentralTradeMarket tasdiqlash kodingiz",
		heading:  "Kirishni tasdiqlash",
		intro:    "CentralTradeMarket uchun tasdiqlash kodingiz:",
		validity: "Kod 10 daqiqa davomida amal qiladi.",
		warning:  "Bu kodni hech kimga bermang.",
		footer:   "Agar siz kod so'ramagan bo'lsangiz, bu xatga e'tibor bermang.",
	},
}

func (s *EmailService) SendOTP(ctx context.Context, toEmail, lang string) error {
	code, err := emailotp.GenerateCode()
	if err != nil {
		return err
	}
	if err := s.store.Save(ctx, toEmail, code); err != nil {
		return err
	}

	loc, ok := otpEmailLocales[lang]
	if !ok {
		loc = otpEmailLocales["ru"]
	}

	body := renderOTPEmail(loc, code)
	return s.sender.SendHTML(toEmail, loc.subject, body)
}

// renderOTPEmail строит брендированное HTML-письмо с OTP-кодом.
// Вёрстка на таблицах с инлайн-стилями для совместимости с email-клиентами.
func renderOTPEmail(loc otpEmailStrings, code string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%[1]s</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f6f9; font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
<span style="display:none; max-height:0; overflow:hidden; opacity:0;">%[2]s</span>
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f6f9; padding:24px 0;">
<tr>
<td align="center">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:480px; width:100%%; background-color:#ffffff; border-radius:16px; overflow:hidden; box-shadow:0 4px 24px rgba(16,24,40,0.06);">
<tr>
<td style="background:linear-gradient(135deg,#0d47a1 0%%,#1976d2 100%%); padding:32px 40px; text-align:center;">
<div style="font-size:22px; font-weight:700; color:#ffffff; letter-spacing:0.3px;">CentralTradeMarket</div>
</td>
</tr>
<tr>
<td style="padding:36px 40px 8px 40px;">
<div style="font-size:20px; font-weight:600; color:#101828; margin-bottom:12px;">%[3]s</div>
<div style="font-size:15px; line-height:22px; color:#475467;">%[4]s</div>
</td>
</tr>
<tr>
<td style="padding:24px 40px;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0">
<tr>
<td align="center" style="background-color:#f0f5ff; border:1px solid #d6e4ff; border-radius:12px; padding:22px 16px;">
<div style="font-family:'Courier New',Courier,monospace; font-size:38px; font-weight:700; letter-spacing:10px; color:#0d47a1;">%[5]s</div>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td style="padding:0 40px 8px 40px;">
<div style="font-size:14px; line-height:21px; color:#475467; margin-bottom:6px;">%[6]s</div>
<div style="font-size:14px; line-height:21px; color:#b42318; font-weight:600;">%[7]s</div>
</td>
</tr>
<tr>
<td style="padding:24px 40px 36px 40px; border-top:1px solid #eaecf0;">
<div style="font-size:12px; line-height:18px; color:#98a2b3;">%[8]s</div>
<div style="font-size:12px; line-height:18px; color:#98a2b3; margin-top:8px;">© CentralTradeMarket</div>
</td>
</tr>
</table>
</td>
</tr>
</table>
</body>
</html>`,
		html.EscapeString(loc.subject),
		html.EscapeString(loc.preview),
		html.EscapeString(loc.heading),
		html.EscapeString(loc.intro),
		html.EscapeString(code),
		html.EscapeString(loc.validity),
		html.EscapeString(loc.warning),
		html.EscapeString(loc.footer),
	)
}

func (s *EmailService) VerifyOTP(ctx context.Context, emailAddr, code string) error {
	err := s.store.Verify(ctx, emailAddr, code)
	if errors.Is(err, emailotp.ErrInvalid) {
		return ErrEmailOTPInvalid
	}
	if errors.Is(err, emailotp.ErrExpired) {
		return ErrEmailOTPExpired
	}
	return err
}

func (s *EmailService) IsVerified(ctx context.Context, emailAddr string) bool {
	return s.store.IsVerified(ctx, emailAddr)
}
