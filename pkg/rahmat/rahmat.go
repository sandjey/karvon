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
	"math"
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
	VAT         int    `json:"vat,omitempty"`  // НДС в %
	TIN         string `json:"tin,omitempty"`  // ИНН продавца
	Mark        string `json:"mark,omitempty"` // маркировка товара
}

type invoiceReq struct {
	StoreID     int       `json:"store_id"`
	Amount      int64     `json:"amount"`               // в тийинах (1 UZS = 100 тийин)
	InvoiceID   string    `json:"invoice_id"`           // наш UUID заказа
	Lang        string    `json:"lang"`
	TTL         int       `json:"ttl,omitempty"`        // время жизни инвойса в секундах
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
// returnURL переопределяет дефолтный ReturnURL из конфига; передайте "", чтобы использовать конфиг.
// Если клиент не сконфигурирован — возвращает stub-данные для локального тестирования.
func (c *Client) CreatePayment(ctx context.Context, orderID string, amountUZS float64, lang, itemName, returnURL string) (invoiceUUID, checkoutURL string, err error) {
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
	if returnURL == "" {
		returnURL = c.cfg.ReturnURL
	}
	// Конвертируем UZS → тийины с округлением (исключаем float-артефакты)
	amountTiyin := int64(math.Round(amountUZS * 100))

	body, _ := json.Marshal(invoiceReq{
		StoreID:     c.cfg.StoreID,
		Amount:      amountTiyin,
		InvoiceID:   orderID,
		Lang:        lang,
		TTL:         3600,
		ReturnURL:   returnURL,
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

// CallbackPayload — полное тело callback-запроса (PaymentModel) от Multicard.
// Статусы: draft | progress | billing | success | error | revert.
type CallbackPayload struct {
	UUID             string `json:"uuid"`              // UUID транзакции в Multicard
	StoreInvoiceID   string `json:"store_invoice_id"`  // наш UUID заказа
	Status           string `json:"status"`            // draft|progress|billing|success|error|revert
	PS               string `json:"ps"`                // uzcard|humo|visa|mastercard|...
	PaymentAmount    int64  `json:"payment_amount"`    // сумма оплаты в тийинах
	TotalAmount      int64  `json:"total_amount"`      // итого с комиссией в тийинах
	CommissionAmount int64  `json:"commission_amount"` // комиссия в тийинах
	CardPan          string `json:"card_pan"`          // маскированный номер карты
	Phone            string `json:"phone"`             // телефон плательщика
	ReceiptURL       string `json:"receipt_url"`       // URL ОФД-чека
	OtpHash          string `json:"otp_hash"`          // хэш OTP-подтверждения
	PaymentTime      string `json:"payment_time"`      // время оплаты ("2006-01-02 15:04:05")
}
