// Package rahmat — клиент платёжной системы Multicard (Rahmat).
// Реализует Вариант А: создание инвойса → редирект на checkout_url.
// Токен авторизации кешируется и обновляется автоматически при истечении.
package rahmat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Config содержит реквизиты Multicard.
type Config struct {
	BaseURL     string // https://dev-mesh.multicard.uz (sandbox) | https://mesh.multicard.uz (prod)
	AppID       string // application_id, выданный Multicard
	Secret      string // secret (пароль приложения)
	StoreID     int    // store_id кассы
	MXIK        string // ИКПУ из tasnif.soliq.uz
	PackageCode string // код упаковки из tasnif.soliq.uz
	CallbackURL string // куда Multicard шлёт callback при оплате
	ReturnURL   string // куда вернуть пользователя после оплаты
}

// Client — потокобезопасный клиент Multicard API.
type Client struct {
	cfg    Config
	hc     *http.Client
	mu     sync.Mutex
	token  string
	expiry time.Time
}

// NewClient создаёт клиент. Работает в режиме заглушки, если AppID или Secret пусты.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		hc:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Configured сообщает, заданы ли реквизиты (иначе — режим заглушки).
func (c *Client) Configured() bool {
	return c.cfg.AppID != "" && c.cfg.Secret != ""
}

// ─── Auth ────────────────────────────────────────────────────────────────────

type authReq struct {
	ApplicationID string `json:"application_id"`
	Secret        string `json:"secret"`
}

type authResp struct {
	Token  string `json:"token"`
	Expiry string `json:"expiry"` // "2023-03-18 16:40:31" GMT+5
}

// getToken возвращает действующий JWT, обновляя его при необходимости.
func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 2-минутный запас перед истечением
	if c.token != "" && time.Now().Before(c.expiry.Add(-2*time.Minute)) {
		return c.token, nil
	}

	body, _ := json.Marshal(authReq{ApplicationID: c.cfg.AppID, Secret: c.cfg.Secret})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/auth", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("multicard auth: %w", err)
	}
	defer resp.Body.Close()

	var ar authResp
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil || ar.Token == "" {
		return "", fmt.Errorf("multicard auth: invalid response (status %d)", resp.StatusCode)
	}

	// Парсим время истечения; формат: "2006-01-02 15:04:05", часовой пояс GMT+5
	loc := time.FixedZone("UZT", 5*3600)
	exp, err := time.ParseInLocation("2006-01-02 15:04:05", ar.Expiry, loc)
	if err != nil {
		exp = time.Now().Add(23 * time.Hour)
	}
	c.token, c.expiry = ar.Token, exp
	return c.token, nil
}

// ─── Invoice (Вариант А) ─────────────────────────────────────────────────────

type ofdItem struct {
	Qty         int    `json:"qty"`
	Price       int64  `json:"price"`
	MXIK        string `json:"mxik"`
	Total       int64  `json:"total"`
	PackageCode string `json:"package_code"`
	Name        string `json:"name"`
}

type invoiceReq struct {
	StoreID     int       `json:"store_id"`
	Amount      int64     `json:"amount"`      // в тийинах (1 UZS = 100 тийин)
	InvoiceID   string    `json:"invoice_id"`  // наш UUID заказа
	Lang        string    `json:"lang"`
	ReturnURL   string    `json:"return_url,omitempty"`
	CallbackURL string    `json:"callback_url"`
	OFD         []ofdItem `json:"ofd"`
}

type invoiceResp struct {
	Success bool `json:"success"`
	Data    struct {
		UUID        string `json:"uuid"`
		CheckoutURL string `json:"checkout_url"`
	} `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Details string `json:"details"`
	} `json:"error"`
}

// CreatePayment создаёт инвойс в Multicard и возвращает (invoiceUUID, checkoutURL).
// amountUZS — сумма в UZS; itemName — название услуги для ОФД.
// Если клиент не сконфигурирован — возвращает stub-данные для локального тестирования.
func (c *Client) CreatePayment(ctx context.Context, orderID string, amountUZS float64, lang, itemName string) (invoiceUUID, checkoutURL string, err error) {
	if !c.Configured() {
		return "stub-" + orderID, "https://pay.karvon.uz/stub/" + orderID, nil
	}

	token, err := c.getToken(ctx)
	if err != nil {
		return "", "", err
	}

	if lang == "" {
		lang = "ru"
	}
	if itemName == "" {
		itemName = "Сервис платформы Karvon"
	}
	amountTiyin := int64(amountUZS * 100)

	body, _ := json.Marshal(invoiceReq{
		StoreID:     c.cfg.StoreID,
		Amount:      amountTiyin,
		InvoiceID:   orderID,
		Lang:        lang,
		ReturnURL:   c.cfg.ReturnURL,
		CallbackURL: c.cfg.CallbackURL,
		OFD: []ofdItem{{
			Qty:         1,
			Price:       amountTiyin,
			MXIK:        c.cfg.MXIK,
			Total:       amountTiyin,
			PackageCode: c.cfg.PackageCode,
			Name:        itemName,
		}},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/payment/invoice", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("multicard create invoice: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var ir invoiceResp
	if err := json.Unmarshal(respBytes, &ir); err != nil {
		return "", "", fmt.Errorf("multicard invoice decode: %w", err)
	}
	if !ir.Success {
		if ir.Error != nil {
			return "", "", fmt.Errorf("multicard: %s: %s", ir.Error.Code, ir.Error.Details)
		}
		return "", "", fmt.Errorf("multicard: invoice failed (body: %.200s)", string(respBytes))
	}
	return ir.Data.UUID, ir.Data.CheckoutURL, nil
}

// ─── Callback ────────────────────────────────────────────────────────────────

// CallbackPayload — тело callback-запроса, который Multicard отправляет нам при оплате инвойса.
// Ключевые поля: StoreInvoiceID = наш order UUID; Status = "success"|"error"|"revert".
type CallbackPayload struct {
	StoreInvoiceID string `json:"store_invoice_id"`
	Status         string `json:"status"`
	PS             string `json:"ps"`  // uzcard|humo|visa|mastercard|...
	UUID           string `json:"uuid"` // UUID транзакции в Multicard
}
