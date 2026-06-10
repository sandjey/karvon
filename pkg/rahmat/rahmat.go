// Package rahmat — клиент платёжной системы Rahmat (Visa/Mir/UzCard/Humo/Payme/Click).
//
// СТАТУС: scaffold/заглушка. Бизнес-логика платежей (заказ → webhook → активация)
// реализована и работает; остаётся подключить реальный HTTP-вызов к Rahmat API
// в методе CreatePayment и проверку подписи в VerifySignature (TODO ниже).
package rahmat

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type Client struct {
	merchantID string
	secretKey  string
	baseURL    string
}

func NewClient(merchantID, secretKey string) *Client {
	return &Client{
		merchantID: merchantID,
		secretKey:  secretKey,
		baseURL:    "https://api.rahmat.uz", // TODO: уточнить по договору с Rahmat
	}
}

// Configured сообщает, заданы ли мерчант-креды (иначе работаем в режиме заглушки).
func (c *Client) Configured() bool {
	return c.merchantID != "" && c.secretKey != ""
}

// CreatePayment создаёт платёж на стороне Rahmat и возвращает URL платёжной страницы.
//
// TODO (подключение): отправить POST в Rahmat API c {amount, currency, merchant_id,
// order_id, callback_url} и вернуть payment_url из ответа. Сейчас возвращается
// детерминированная заглушка-ссылка, чтобы фронт/флоу можно было собрать заранее.
func (c *Client) CreatePayment(orderID string, amount float64, currency, callbackURL string) (string, error) {
	return c.baseURL + "/pay/" + orderID, nil
}

// VerifySignature проверяет подпись webhook-а от Rahmat (HMAC-SHA256).
//
// TODO (подключение): сверить с алгоритмом из документации Rahmat. Если креды не
// заданы (режим заглушки) — пропускаем проверку, чтобы можно было тестировать webhook.
func (c *Client) VerifySignature(payload []byte, signature string) bool {
	if !c.Configured() {
		return true
	}
	mac := hmac.New(sha256.New, []byte(c.secretKey))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
