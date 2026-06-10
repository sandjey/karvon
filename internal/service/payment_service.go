package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

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
	CallbackURL string
}

// Create создаёт платёжный заказ (pending) и возвращает order + payment_url.
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

	url, err := s.rahmat.CreatePayment(order.ID.String(), amount, currency, in.CallbackURL)
	if err != nil {
		return nil, "", err
	}
	return order, url, nil
}

// Webhook обрабатывает уведомление об оплате (идемпотентно).
func (s *PaymentService) Webhook(ctx context.Context, orderID uuid.UUID, method string, payload []byte, signature string) error {
	if !s.rahmat.VerifySignature(payload, signature) {
		return ErrInvalidCredentials
	}
	order, err := s.payments.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrNotFound
	}
	if order.Status == "paid" {
		return nil // идемпотентность
	}
	if err := s.payments.MarkPaid(ctx, order.ID, method); err != nil {
		return err
	}
	return s.apply(ctx, order)
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
		return s.warehouse.UpdateStatus(ctx, id, "active") // заглушка: помечаем активным
	}
	return s.cargo.UpdateStatus(ctx, id, map[string]interface{}{"is_paid": true})
}

func (s *PaymentService) History(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.PaymentOrder, int64, error) {
	return s.payments.History(ctx, userID, offset, limit)
}

func (s *PaymentService) ActiveSubscription(ctx context.Context, userID uuid.UUID) (*model.Subscription, error) {
	return s.payments.ActiveSubscription(ctx, userID)
}

// Boost создаёт платёжный заказ на продвижение объявления.
func (s *PaymentService) Boost(ctx context.Context, userID uuid.UUID, listingType, listingID, pricingKey, callbackURL string) (*model.PaymentOrder, string, error) {
	if err := s.assertOwner(ctx, userID, listingType, listingID); err != nil {
		return nil, "", err
	}
	return s.Create(ctx, CreatePaymentInput{
		UserID:      userID,
		PaymentType: "boost",
		PricingKey:  pricingKey,
		ListingType: listingType,
		ListingID:   listingID,
		CallbackURL: callbackURL,
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
