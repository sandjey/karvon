package email

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
)

// Sender отправляет email через SMTP.
type Sender struct {
	host     string
	port     string
	user     string
	password string
	from     string
	fromName string
}

// NewSender создаёт отправителя. Если host пустой — отправка отключена (no-op).
func NewSender(host, port, user, password, from, fromName string) *Sender {
	return &Sender{host: host, port: port, user: user, password: password, from: from, fromName: fromName}
}

// Enabled возвращает true, если SMTP настроен.
func (s *Sender) Enabled() bool {
	return s.host != "" && s.user != ""
}

// Send отправляет plain-text письмо на указанный адрес. Если Sender не настроен — no-op.
func (s *Sender) Send(to, subject, body string) error {
	return s.send(to, subject, body, "text/plain; charset=UTF-8")
}

// SendHTML отправляет HTML-письмо на указанный адрес. Если Sender не настроен — no-op.
func (s *Sender) SendHTML(to, subject, htmlBody string) error {
	return s.send(to, subject, htmlBody, "text/html; charset=UTF-8")
}

// send — общая логика отправки письма с заданным Content-Type.
func (s *Sender) send(to, subject, body, contentType string) error {
	if !s.Enabled() || to == "" {
		return nil
	}

	addr := net.JoinHostPort(s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.password, s.host)

	// Заголовок From: с отображаемым именем, если задано. Имя кодируется
	// для поддержки не-ASCII (для ASCII возвращается само имя).
	// В SMTP-конверт (client.Mail) ниже всегда идёт голый адрес s.from.
	fromHeader := s.from
	if s.fromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", s.fromName), s.from)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s\r\n\r\n%s",
		fromHeader, to, subject, contentType, body,
	)

	// Пробуем STARTTLS, при ошибке — plain TLS (порт 465).
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := fmt.Fprint(wc, msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	return wc.Close()
}
