package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"karvon/internal/dto"
	"karvon/internal/model"
	"karvon/internal/repository"
	"karvon/pkg/rahmat"
)

type PaymentService struct {
	payments  *repository.PaymentRepo
	pricing   *repository.PricingRepo
	users     *repository.UserRepo
	cargo     *repository.CargoRepo
	warehouse *repository.WarehouseRepo
	rahmat    *rahmat.Client
}

func NewPaymentService(
	payments *repository.PaymentRepo,
	pricing *repository.PricingRepo,
	users *repository.UserRepo,
	cargo *repository.CargoRepo,
	warehouse *repository.WarehouseRepo,
	rahmatClient *rahmat.Client,
) *PaymentService {
	return &PaymentService{payments: payments, pricing: pricing, users: users, cargo: cargo, warehouse: warehouse, rahmat: rahmatClient}
}

type CreatePaymentInput struct {
	UserID      uuid.UUID
	PaymentType string // tokens | subscription | listing | boost
	PricingKey  string // tokens_basic | sub_month | boost_3day | listing_paid
	Currency    string // UZS | USD
	ListingType string // для listing/boost: cargo|warehouse
	ListingID   string // для listing/boost
	Lang        string // ru|uz|en — язык страницы оплаты
	ReturnURL   string // куда вернуть после оплаты (опционально, переопределяет конфиг)
}

// Create создаёт платёжный заказ (pending) и возвращает order + checkout_url.
func (s *PaymentService) Create(ctx context.Context, in CreatePaymentInput) (*model.PaymentOrder, string, error) {
	p, err := s.pricing.FindByKey(ctx, in.PricingKey)
	if err != nil {
		return nil, "", err
	}
	if p == nil {
		return nil, "", ErrNotFound
	}
	currency := in.Currency
	if currency == "" {
		currency = "UZS"
	}
	amount := p.PriceUZS
	if currency == "USD" {
		amount = p.PriceUSD
	}

	// item_id кодирует тариф и (для listing/boost) ссылку на объявление.
	itemID := in.PricingKey
	if in.PaymentType == "listing" || in.PaymentType == "boost" {
		itemID = in.ListingType + "|" + in.ListingID + "|" + in.PricingKey
	}

	order := &model.PaymentOrder{
		ID:          uuid.New(),
		UserID:      in.UserID,
		PaymentType: in.PaymentType,
		ItemID:      &itemID,
		Amount:      amount,
		Currency:    currency,
		Status:      "pending",
	}
	if err := s.payments.Create(ctx, order); err != nil {
		return nil, "", err
	}

	invoiceUUID, checkoutURL, err := s.rahmat.CreatePayment(ctx, order.ID.String(), amount, in.Lang, p.Label, in.ReturnURL)
	if err != nil {
		return nil, "", err
	}

	// Сохраняем UUID инвойса Multicard для последующей сверки
	if invoiceUUID != "" {
		_ = s.payments.SaveInvoiceUUID(ctx, order.ID, invoiceUUID)
	}

	return order, checkoutURL, nil
}

// Webhook обрабатывает полный callback (PaymentModel) от Multicard (идемпотентно).
// Обрабатывает статусы "success" и "revert"; остальные игнорируются на уровне хэндлера.
func (s *PaymentService) Webhook(ctx context.Context, cb dto.MulticardCallback) error {
	orderID, err := uuid.Parse(cb.StoreInvoiceID)
	if err != nil {
		return nil // неизвестный store_invoice_id — игнорируем
	}

	order, err := s.payments.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrNotFound
	}

	// Идемпотентность: уже обработанные заказы не трогаем
	if order.Status == "paid" || order.Status == "reverted" {
		return nil
	}

	// Верификация суммы: предупреждение в лог (не блокируем — комиссия может расходиться)
	if cb.PaymentAmount > 0 {
		expectedTiyin := int64(math.Round(order.Amount * 100))
		if cb.PaymentAmount != expectedTiyin {
			log.Warn().
				Int64("expected_tiyin", expectedTiyin).
				Int64("got_tiyin", cb.PaymentAmount).
				Str("order_id", order.ID.String()).
				Msg("payment amount mismatch in webhook")
		}
	}

	if cb.Status == "revert" {
		if err := s.payments.MarkReverted(ctx, order.ID); err != nil {
			return err
		}
		log.Info().Str("order_id", order.ID.String()).Str("type", order.PaymentType).Msg("payment reverted")
		return s.revert(ctx, order)
	}

	// success
	if err := s.payments.MarkPaid(ctx, order.ID, cb.PS); err != nil {
		return err
	}
	_ = s.payments.SavePaymentDetails(ctx, order.ID, cb.UUID, cb.Phone, cb.ReceiptURL, cb.TotalAmount, cb.CommissionAmount)
	return s.apply(ctx, order)
}

// GetByID возвращает платёжный заказ по ID (для клиентского поллинга статуса).
func (s *PaymentService) GetByID(ctx context.Context, orderID, userID uuid.UUID) (*model.PaymentOrder, error) {
	order, err := s.payments.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrNotFound
	}
	if order.UserID != userID {
		return nil, ErrNotOwner
	}
	return order, nil
}

func (s *PaymentService) apply(ctx context.Context, o *model.PaymentOrder) error {
	key := ""
	if o.ItemID != nil {
		key = *o.ItemID
	}
	switch o.PaymentType {
	case "tokens":
		amount := s.pricing.TokensAmount(ctx, key, 0)
		if amount > 0 {
			return s.users.CreditTokens(ctx, o.UserID, amount, "purchase")
		}
	case "subscription":
		p, _ := s.pricing.FindByKey(ctx, key)
		days := 30
		plan := "month"
		if p != nil && p.DurationDays > 0 {
			days = p.DurationDays
		}
		switch key {
		case "sub_week":
			plan = "week"
		case "sub_year":
			plan = "year"
		}
		now := time.Now()
		return s.payments.CreateSubscription(ctx, &model.Subscription{
			UserID:         o.UserID,
			Plan:           plan,
			StartsAt:       now,
			ExpiresAt:      now.Add(time.Duration(days) * 24 * time.Hour),
			IsActive:       true,
			PaymentOrderID: &o.ID,
		})
	case "listing":
		lt, id, _ := parseListingRef(key)
		return s.markListingPaid(ctx, lt, id)
	case "boost":
		lt, id, bkey := parseListingRef(key)
		p, _ := s.pricing.FindByKey(ctx, bkey)
		days := 1
		if p != nil && p.DurationDays > 0 {
			days = p.DurationDays
		}
		until := time.Now().Add(time.Duration(days) * 24 * time.Hour)
		if lt == "warehouse" {
			return s.warehouse.SetBoost(ctx, id, until)
		}
		return s.cargo.SetBoost(ctx, id, until)
	}
	return nil
}

func (s *PaymentService) markListingPaid(ctx context.Context, lt string, id uuid.UUID) error {
	if lt == "warehouse" {
		return s.warehouse.MarkPaid(ctx, id)
	}
	return s.cargo.UpdateStatus(ctx, id, map[string]interface{}{"is_paid": true, "status": "active"})
}

// revert откатывает эффект оплаченного заказа после revert-статуса от Multicard.
func (s *PaymentService) revert(ctx context.Context, o *model.PaymentOrder) error {
	key := ""
	if o.ItemID != nil {
		key = *o.ItemID
	}
	switch o.PaymentType {
	case "tokens":
		amount := s.pricing.TokensAmount(ctx, key, 0)
		if amount > 0 {
			if err := s.users.DebitTokens(ctx, o.UserID, amount, "revert"); err != nil {
				log.Error().Err(err).Str("order_id", o.ID.String()).Msg("revert: failed to debit tokens")
			}
		}
	case "subscription":
		if err := s.payments.DeactivateSubscriptionByOrder(ctx, o.ID); err != nil {
			log.Error().Err(err).Str("order_id", o.ID.String()).Msg("revert: failed to deactivate subscription")
		}
	case "listing":
		lt, id, _ := parseListingRef(key)
		if err := s.revertListing(ctx, lt, id); err != nil {
			log.Error().Err(err).Str("order_id", o.ID.String()).Msg("revert: failed to revert listing")
		}
	case "boost":
		lt, id, _ := parseListingRef(key)
		if err := s.removeBoost(ctx, lt, id); err != nil {
			log.Error().Err(err).Str("order_id", o.ID.String()).Msg("revert: failed to remove boost")
		}
	}
	return nil
}

func (s *PaymentService) revertListing(ctx context.Context, lt string, id uuid.UUID) error {
	if lt == "warehouse" {
		return s.warehouse.RevertPaid(ctx, id)
	}
	return s.cargo.UpdateStatus(ctx, id, map[string]interface{}{"is_paid": false, "status": "archived"})
}

func (s *PaymentService) removeBoost(ctx context.Context, lt string, id uuid.UUID) error {
	if lt == "warehouse" {
		return s.warehouse.RemoveBoost(ctx, id)
	}
	return s.cargo.RemoveBoost(ctx, id)
}

func (s *PaymentService) History(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.PaymentOrder, int64, error) {
	return s.payments.History(ctx, userID, offset, limit)
}

func (s *PaymentService) ActiveSubscription(ctx context.Context, userID uuid.UUID) (*model.Subscription, error) {
	return s.payments.ActiveSubscription(ctx, userID)
}

// Boost создаёт платёжный заказ на продвижение объявления.
func (s *PaymentService) Boost(ctx context.Context, userID uuid.UUID, listingType, listingID, pricingKey, currency, lang, returnURL string) (*model.PaymentOrder, string, error) {
	if err := s.assertOwner(ctx, userID, listingType, listingID); err != nil {
		return nil, "", err
	}
	return s.Create(ctx, CreatePaymentInput{
		UserID:      userID,
		PaymentType: "boost",
		PricingKey:  pricingKey,
		Currency:    currency,
		ListingType: listingType,
		ListingID:   listingID,
		Lang:        lang,
		ReturnURL:   returnURL,
	})
}

func (s *PaymentService) assertOwner(ctx context.Context, userID uuid.UUID, lt, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ErrValidation
	}
	if lt == "warehouse" {
		w, err := s.warehouse.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if w == nil {
			return ErrListingNotFound
		}
		if w.UserID != userID {
			return ErrNotOwner
		}
		return nil
	}
	c, err := s.cargo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrListingNotFound
	}
	if c.UserID != userID {
		return ErrNotOwner
	}
	return nil
}

func parseListingRef(s string) (listingType string, id uuid.UUID, extra string) {
	parts := strings.Split(s, "|")
	if len(parts) >= 2 {
		listingType = parts[0]
		id, _ = uuid.Parse(parts[1])
	}
	if len(parts) >= 3 {
		extra = parts[2]
	}
	return
}
