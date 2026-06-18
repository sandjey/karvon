// Package rahmat вЂ” РєР»РёРµРЅС‚ РїР»Р°С‚С‘Р¶РЅРѕР№ СЃРёСЃС‚РµРјС‹ Multicard (Rahmat).
// Р РµР°Р»РёР·СѓРµС‚ Р’Р°СЂРёР°РЅС‚ Рђ: СЃРѕР·РґР°РЅРёРµ РёРЅРІРѕР№СЃР° в†’ СЂРµРґРёСЂРµРєС‚ РЅР° checkout_url.
// РўРѕРєРµРЅ Р°РІС‚РѕСЂРёР·Р°С†РёРё РєРµС€РёСЂСѓРµС‚СЃСЏ Рё РѕР±РЅРѕРІР»СЏРµС‚СЃСЏ Р°РІС‚РѕРјР°С‚РёС‡РµСЃРєРё РїСЂРё РёСЃС‚РµС‡РµРЅРёРё.
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

// Config СЃРѕРґРµСЂР¶РёС‚ СЂРµРєРІРёР·РёС‚С‹ Multicard.
type Config struct {
	BaseURL     string // https://dev-mesh.multicard.uz (sandbox) | https://mesh.multicard.uz (prod)
	AppID       string // application_id, РІС‹РґР°РЅРЅС‹Р№ Multicard
	Secret      string // secret (РїР°СЂРѕР»СЊ РїСЂРёР»РѕР¶РµРЅРёСЏ)
	StoreID     int    // store_id РєР°СЃСЃС‹
	MXIK        string // РРљРџРЈ РёР· tasnif.soliq.uz
	PackageCode string // РєРѕРґ СѓРїР°РєРѕРІРєРё РёР· tasnif.soliq.uz
	CallbackURL string // РєСѓРґР° Multicard С€Р»С‘С‚ callback РїСЂРё РѕРїР»Р°С‚Рµ
	ReturnURL   string // РєСѓРґР° РІРµСЂРЅСѓС‚СЊ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ РїРѕСЃР»Рµ РѕРїР»Р°С‚С‹
}

// Client вЂ” РїРѕС‚РѕРєРѕР±РµР·РѕРїР°СЃРЅС‹Р№ РєР»РёРµРЅС‚ Multicard API.
type Client struct {
	cfg    Config
	hc     *http.Client
	mu     sync.Mutex
	token  string
	expiry time.Time
}

// NewClient СЃРѕР·РґР°С‘С‚ РєР»РёРµРЅС‚. Р Р°Р±РѕС‚Р°РµС‚ РІ СЂРµР¶РёРјРµ Р·Р°РіР»СѓС€РєРё, РµСЃР»Рё AppID РёР»Рё Secret РїСѓСЃС‚С‹.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		hc:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Configured СЃРѕРѕР±С‰Р°РµС‚, Р·Р°РґР°РЅС‹ Р»Рё СЂРµРєРІРёР·РёС‚С‹ (РёРЅР°С‡Рµ вЂ” СЂРµР¶РёРј Р·Р°РіР»СѓС€РєРё).
func (c *Client) Configured() bool {
	return c.cfg.AppID != "" && c.cfg.Secret != ""
}

// в”Ђв”Ђв”Ђ Auth в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ

type authReq struct {
	ApplicationID string `json:"application_id"`
	Secret        string `json:"secret"`
}

type authResp struct {
	Token  string `json:"token"`
	Expiry string `json:"expiry"` // "2023-03-18 16:40:31" GMT+5
}

// getToken РІРѕР·РІСЂР°С‰Р°РµС‚ РґРµР№СЃС‚РІСѓСЋС‰РёР№ JWT, РѕР±РЅРѕРІР»СЏСЏ РµРіРѕ РїСЂРё РЅРµРѕР±С…РѕРґРёРјРѕСЃС‚Рё.
func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 2-РјРёРЅСѓС‚РЅС‹Р№ Р·Р°РїР°СЃ РїРµСЂРµРґ РёСЃС‚РµС‡РµРЅРёРµРј
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

	// РџР°СЂСЃРёРј РІСЂРµРјСЏ РёСЃС‚РµС‡РµРЅРёСЏ; С„РѕСЂРјР°С‚: "2006-01-02 15:04:05", С‡Р°СЃРѕРІРѕР№ РїРѕСЏСЃ GMT+5
	loc := time.FixedZone("UZT", 5*3600)
	exp, err := time.ParseInLocation("2006-01-02 15:04:05", ar.Expiry, loc)
	if err != nil {
		exp = time.Now().Add(23 * time.Hour)
	}
	c.token, c.expiry = ar.Token, exp
	return c.token, nil
}

// в”Ђв”Ђв”Ђ Invoice (Р’Р°СЂРёР°РЅС‚ Рђ) в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ

type ofdItem struct {
	Qty         int    `json:"qty"`
	Price       int64  `json:"price"`
	MXIK        string `json:"mxik"`
	Total       int64  `json:"total"`
	PackageCode string `json:"package_code"`
	Name        string `json:"name"`
	VAT         int    `json:"vat,omitempty"`  // РќР”РЎ РІ %
	TIN         string `json:"tin,omitempty"`  // РРќРќ РїСЂРѕРґР°РІС†Р°
	Mark        string `json:"mark,omitempty"` // РјР°СЂРєРёСЂРѕРІРєР° С‚РѕРІР°СЂР°
}

type invoiceReq struct {
	StoreID     int       `json:"store_id"`
	Amount      int64     `json:"amount"`               // РІ С‚РёР№РёРЅР°С… (1 UZS = 100 С‚РёР№РёРЅ)
	InvoiceID   string    `json:"invoice_id"`           // РЅР°С€ UUID Р·Р°РєР°Р·Р°
	Lang        string    `json:"lang"`
	TTL         int       `json:"ttl,omitempty"`        // РІСЂРµРјСЏ Р¶РёР·РЅРё РёРЅРІРѕР№СЃР° РІ СЃРµРєСѓРЅРґР°С…
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

// CreatePayment СЃРѕР·РґР°С‘С‚ РёРЅРІРѕР№СЃ РІ Multicard Рё РІРѕР·РІСЂР°С‰Р°РµС‚ (invoiceUUID, checkoutURL).
// amountUZS вЂ” СЃСѓРјРјР° РІ UZS; itemName вЂ” РЅР°Р·РІР°РЅРёРµ СѓСЃР»СѓРіРё РґР»СЏ РћР¤Р”.
// returnURL РїРµСЂРµРѕРїСЂРµРґРµР»СЏРµС‚ РґРµС„РѕР»С‚РЅС‹Р№ ReturnURL РёР· РєРѕРЅС„РёРіР°; РїРµСЂРµРґР°Р№С‚Рµ "", С‡С‚РѕР±С‹ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊ РєРѕРЅС„РёРі.
// Р•СЃР»Рё РєР»РёРµРЅС‚ РЅРµ СЃРєРѕРЅС„РёРіСѓСЂРёСЂРѕРІР°РЅ вЂ” РІРѕР·РІСЂР°С‰Р°РµС‚ stub-РґР°РЅРЅС‹Рµ РґР»СЏ Р»РѕРєР°Р»СЊРЅРѕРіРѕ С‚РµСЃС‚РёСЂРѕРІР°РЅРёСЏ.
func (c *Client) CreatePayment(ctx context.Context, orderID string, amountUZS float64, lang, itemName, returnURL string) (invoiceUUID, checkoutURL string, err error) {
	if !c.Configured() {
		return "stub-" + orderID, "https://pay.centraltrademarket.com/stub/" + orderID, nil
	}

	token, err := c.getToken(ctx)
	if err != nil {
		return "", "", err
	}

	if lang == "" {
		lang = "ru"
	}
	if itemName == "" {
		itemName = "РЎРµСЂРІРёСЃ РїР»Р°С‚С„РѕСЂРјС‹ Karvon"
	}
	if returnURL == "" {
		returnURL = c.cfg.ReturnURL
	}
	// РљРѕРЅРІРµСЂС‚РёСЂСѓРµРј UZS в†’ С‚РёР№РёРЅС‹ СЃ РѕРєСЂСѓРіР»РµРЅРёРµРј (РёСЃРєР»СЋС‡Р°РµРј float-Р°СЂС‚РµС„Р°РєС‚С‹)
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

// в”Ђв”Ђв”Ђ Callback в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ

// CallbackPayload вЂ” РїРѕР»РЅРѕРµ С‚РµР»Рѕ callback-Р·Р°РїСЂРѕСЃР° (PaymentModel) РѕС‚ Multicard.
// РЎС‚Р°С‚СѓСЃС‹: draft | progress | billing | success | error | revert.
type CallbackPayload struct {
	UUID             string `json:"uuid"`              // UUID С‚СЂР°РЅР·Р°РєС†РёРё РІ Multicard
	StoreInvoiceID   string `json:"store_invoice_id"`  // РЅР°С€ UUID Р·Р°РєР°Р·Р°
	Status           string `json:"status"`            // draft|progress|billing|success|error|revert
	PS               string `json:"ps"`                // uzcard|humo|visa|mastercard|...
	PaymentAmount    int64  `json:"payment_amount"`    // СЃСѓРјРјР° РѕРїР»Р°С‚С‹ РІ С‚РёР№РёРЅР°С…
	TotalAmount      int64  `json:"total_amount"`      // РёС‚РѕРіРѕ СЃ РєРѕРјРёСЃСЃРёРµР№ РІ С‚РёР№РёРЅР°С…
	CommissionAmount int64  `json:"commission_amount"` // РєРѕРјРёСЃСЃРёСЏ РІ С‚РёР№РёРЅР°С…
	CardPan          string `json:"card_pan"`          // РјР°СЃРєРёСЂРѕРІР°РЅРЅС‹Р№ РЅРѕРјРµСЂ РєР°СЂС‚С‹
	Phone            string `json:"phone"`             // С‚РµР»РµС„РѕРЅ РїР»Р°С‚РµР»СЊС‰РёРєР°
	ReceiptURL       string `json:"receipt_url"`       // URL РћР¤Р”-С‡РµРєР°
	OtpHash          string `json:"otp_hash"`          // С…СЌС€ OTP-РїРѕРґС‚РІРµСЂР¶РґРµРЅРёСЏ
	PaymentTime      string `json:"payment_time"`      // РІСЂРµРјСЏ РѕРїР»Р°С‚С‹ ("2006-01-02 15:04:05")
}
