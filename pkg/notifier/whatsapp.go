package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/rs/zerolog/log"
)

var waOTPRegexp = regexp.MustCompile(`\*(\d{6})\*`)

type WhatsAppNotifier struct {
	serviceURL string
	token      string
	bypass     bool
	client     *http.Client
}

func NewWhatsApp(serviceURL, token string, bypass bool) *WhatsAppNotifier {
	return &WhatsAppNotifier{
		serviceURL: serviceURL,
		token:      token,
		bypass:     bypass,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WhatsAppNotifier) Send(_ context.Context, phone, message string) error {
	if w.bypass {
		code := ""
		if m := waOTPRegexp.FindStringSubmatch(message); len(m) >= 2 {
			code = m[1]
		}
		log.Info().Str("phone", phone).Str("code", code).Msg("[BYPASS] WhatsApp OTP")
		return nil
	}

	body, _ := json.Marshal(map[string]string{
		"to":      phone,
		"message": message,
	})

	req, err := http.NewRequest(http.MethodPost, w.serviceURL+"/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build whatsapp request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if w.token != "" {
		req.Header.Set("Authorization", "Bearer "+w.token)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp otp send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("whatsapp service returned %d", resp.StatusCode)
	}
	return nil
}
