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

// TelegramGatewayNotifier отправляет OTP через Telegram Gateway API.
// Пользователю не нужно знать chat_id — достаточно номера телефона.
type TelegramGatewayNotifier struct {
	baseURL string
	token   string
	bypass  bool
	client  *http.Client
}

func NewTelegramGateway(baseURL, token string, bypass bool) *TelegramGatewayNotifier {
	return &TelegramGatewayNotifier{
		baseURL: baseURL,
		token:   token,
		bypass:  bypass,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

var otpCodeRegexp = regexp.MustCompile(`\*(\d{6})\*`)

func (t *TelegramGatewayNotifier) CheckNumber(ctx context.Context, phone string) error {
	if t.bypass {
		return nil
	}
	body, _ := json.Marshal(map[string]string{"phone_number": phone})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/checkSendAbility", bytes.NewReader(body))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.token)
	resp, err := t.client.Do(req)
	if err != nil {
		// сеть недоступна — не блокируем
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// не можем разобрать ответ — не блокируем
		return nil
	}
	// Блокируем только при явном "номер не найден в Telegram".
	// Ошибки баланса, rate-limit, 5xx — пропускаем (пусть Send сам вернёт ошибку).
	if !result.OK && (result.Error == "PHONE_NOT_FOUND" || result.Error == "PHONE_NUMBER_INVALID") {
		return ErrNumberNotRegistered
	}
	return nil
}

// Send — to = номер телефона (+998901234567), message = отформатированный текст с OTP.
// Код извлекается из message и передаётся в Telegram Gateway API.
func (t *TelegramGatewayNotifier) Send(_ context.Context, phone, message string) error {
	matches := otpCodeRegexp.FindStringSubmatch(message)
	if len(matches) < 2 {
		return fmt.Errorf("cannot extract OTP code from message")
	}
	code := matches[1]

	if t.bypass {
		log.Info().Str("phone", phone).Str("code", code).Msg("[BYPASS] Telegram Gateway OTP")
		return nil
	}

	body, _ := json.Marshal(map[string]interface{}{
		"phone_number": phone,
		"code":         code,
		"ttl":          300,
	})

	req, err := http.NewRequest(http.MethodPost, t.baseURL+"/sendVerificationMessage", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build telegram gateway request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.token)

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram gateway send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram gateway returned %d", resp.StatusCode)
	}
	return nil
}
