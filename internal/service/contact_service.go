package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/repository"
)

const contactViewWindow = 30 * 24 * time.Hour

type ContactService struct {
	cargo     *repository.CargoRepo
	warehouse *repository.WarehouseRepo
	contacts  *repository.ContactRepo
	users     *repository.UserRepo
	notif     *repository.NotificationRepo
}

func NewContactService(
	cargo *repository.CargoRepo,
	warehouse *repository.WarehouseRepo,
	contacts *repository.ContactRepo,
	users *repository.UserRepo,
	notif *repository.NotificationRepo,
) *ContactService {
	return &ContactService{cargo: cargo, warehouse: warehouse, contacts: contacts, users: users, notif: notif}
}

// resolveContact достаёт владельца и контактные данные объявления.
func (s *ContactService) resolveContact(ctx context.Context, listingType string, listingID uuid.UUID) (ownerID uuid.UUID, info *dto.ContactInfo, err error) {
	switch listingType {
	case "cargo":
		c, e := s.cargo.FindByID(ctx, listingID)
		if e != nil {
			return uuid.Nil, nil, e
		}
		if c == nil {
			return uuid.Nil, nil, ErrListingNotFound
		}
		info = &dto.ContactInfo{}
		if c.User != nil {
			info.Phone = &c.User.Phone
			info.WhatsApp = c.User.WhatsApp
			info.Telegram = c.User.Telegram
			info.Email = c.User.Email
			info.City = c.User.City
			info.Country = c.User.Country
		}
		if c.Company != nil {
			info.CompanyName = &c.Company.Name
			if info.City == nil {
				info.City = &c.Company.City
			}
		}
		return c.UserID, info, nil
	case "warehouse":
		w, e := s.warehouse.FindByID(ctx, listingID)
		if e != nil {
			return uuid.Nil, nil, e
		}
		if w == nil {
			return uuid.Nil, nil, ErrListingNotFound
		}
		info = &dto.ContactInfo{
			Phone:         w.PhoneMain,
			PhoneExtra:    w.PhoneExtra,
			Email:         w.Email,
			ContactPerson: w.ContactPerson,
		}
		if w.Company != nil {
			info.CompanyName = &w.Company.Name
			info.City = &w.Company.City
		}
		// WhatsApp / Telegram берём из профиля владельца склада
		if owner, err := s.users.FindByID(ctx, w.UserID); err == nil && owner != nil {
			info.WhatsApp = owner.WhatsApp
			info.Telegram = owner.Telegram
		}
		return w.UserID, info, nil
	default:
		return uuid.Nil, nil, ErrListingNotFound
	}
}

func (s *ContactService) Open(ctx context.Context, viewerID uuid.UUID, req dto.OpenContactRequest) (*dto.ContactInfo, error) {
	ownerID, info, err := s.resolveContact(ctx, req.ListingType, req.ListingID)
	if err != nil {
		return nil, err
	}

	viewer, err := s.users.FindByID(ctx, viewerID)
	if err != nil || viewer == nil {
		return nil, ErrNotFound
	}
	info.TokenBalance = viewer.TokenBalance

	// своё объявление — без списания
	if ownerID == viewerID {
		info.TokensSpent = 0
		return info, nil
	}

	// уже смотрел за последние 30 дней
	if recent, err := s.contacts.RecentView(ctx, viewerID, req.ListingType, req.ListingID, contactViewWindow); err == nil && recent != nil {
		info.TokensSpent = 0
		return info, nil
	}

	// активная подписка — без списания, но фиксируем
	if active, _ := s.contacts.HasActiveSubscription(ctx, viewerID); active {
		_ = s.contacts.RecordFree(ctx, viewerID, req.ListingType, req.ListingID)
		info.TokensSpent = 0
		return info, nil
	}

	// списать токен
	newBalance, err := s.contacts.DebitAndRecord(ctx, viewerID, req.ListingType, req.ListingID)
	if err == repository.ErrNoTokens {
		return nil, ErrInsufficientTokens
	}
	if err != nil {
		return nil, err
	}
	info.TokensSpent = 1
	info.TokenBalance = newBalance

	// счётчик купленных контактов
	switch req.ListingType {
	case "cargo":
		_ = s.cargo.IncrementContactsBought(ctx, req.ListingID)
	case "warehouse":
		_ = s.warehouse.IncrementContactsBought(ctx, req.ListingID)
	}

	// уведомить автора
	meta, _ := json.Marshal(map[string]string{
		"listing_type": req.ListingType,
		"listing_id":   req.ListingID.String(),
		"buyer_id":     viewerID.String(),
	})
	buyerName := "пользователь"
	if viewer.Name != nil && *viewer.Name != "" {
		buyerName = *viewer.Name
	}
	body := "Ваш контакт открыл " + buyerName
	_ = s.notif.Create(ctx, &model.Notification{
		UserID: ownerID,
		Type:   "contact_purchased",
		Title:  "Контакт открыт",
		Body:   &body,
		Meta:   meta,
	})

	return info, nil
}

func (s *ContactService) History(ctx context.Context, userID uuid.UUID, offset, limit int) ([]dto.ContactHistoryItem, int64, error) {
	views, total, err := s.contacts.HistoryPaginated(ctx, userID, contactViewWindow, offset, limit)
	if err != nil || len(views) == 0 {
		return nil, total, err
	}

	// Listing IDlarini cargo va warehouse bo'yicha ajrat
	var cargoIDs, warehouseIDs, allListingIDs []uuid.UUID
	for _, v := range views {
		allListingIDs = append(allListingIDs, v.ListingID)
		if v.ListingType == "cargo" {
			cargoIDs = append(cargoIDs, v.ListingID)
		} else {
			warehouseIDs = append(warehouseIDs, v.ListingID)
		}
	}

	// Batch yuklash
	cargoList, _ := s.cargo.FindByIDs(ctx, cargoIDs)
	warehouseList, _ := s.warehouse.FindByIDs(ctx, warehouseIDs)

	// Map qilish
	cargoMap := make(map[uuid.UUID]*model.CargoListing, len(cargoList))
	for _, c := range cargoList {
		cargoMap[c.ID] = c
	}
	warehouseMap := make(map[uuid.UUID]*model.WarehouseListing, len(warehouseList))
	for _, w := range warehouseList {
		warehouseMap[w.ID] = w
	}

	// Birinchi viewed_at per listing (repeat_view aniqlash uchun)
	firstViews, _ := s.contacts.FirstViewedAtPerListing(ctx, userID, allListingIDs)

	items := make([]dto.ContactHistoryItem, 0, len(views))
	for _, v := range views {
		item := dto.ContactHistoryItem{
			ID:          v.ID,
			ListingType: v.ListingType,
			TokensSpent: v.TokensSpent,
			ViewedAt:    v.ViewedAt,
			FreeUntil:   v.ViewedAt.Add(30 * 24 * time.Hour),
		}

		// view_reason aniqlash
		if v.TokensSpent > 0 {
			item.ViewReason = "token"
		}

		if v.ListingType == "cargo" {
			if c, ok := cargoMap[v.ListingID]; ok {
				companyName := ""
				if c.Company != nil {
					companyName = c.Company.Name
				}
				item.Listing = dto.ContactHistoryListing{
					ID:          c.ID,
					Name:        c.CargoName,
					CompanyName: companyName,
					FromCity:    c.FromCity,
					ToCity:      c.ToCity,
					Status:      c.Status,
				}
				if c.User != nil {
					item.Contact = dto.ContactHistoryContact{
						Phone:    &c.User.Phone,
						WhatsApp: c.User.WhatsApp,
						Telegram: c.User.Telegram,
					}
				}
				if v.TokensSpent == 0 {
					if c.UserID == userID {
						item.ViewReason = "own_listing"
					} else if first, ok := firstViews[v.ListingID]; ok && v.ViewedAt.After(first) {
						item.ViewReason = "repeat_view"
					} else {
						item.ViewReason = "subscription"
					}
				}
			}
		} else {
			if w, ok := warehouseMap[v.ListingID]; ok {
				companyName := ""
				if w.Company != nil {
					companyName = w.Company.Name
				}
				item.Listing = dto.ContactHistoryListing{
					ID:          w.ID,
					Name:        w.Name,
					CompanyName: companyName,
					FromCity:    w.City,
					Status:      w.Status,
				}
				item.Contact = dto.ContactHistoryContact{
					Phone: w.PhoneMain,
				}
				if w.User != nil {
					item.Contact.WhatsApp = w.User.WhatsApp
					item.Contact.Telegram = w.User.Telegram
				}
				if v.TokensSpent == 0 {
					if w.UserID == userID {
						item.ViewReason = "own_listing"
					} else if first, ok := firstViews[v.ListingID]; ok && v.ViewedAt.After(first) {
						item.ViewReason = "repeat_view"
					} else {
						item.ViewReason = "subscription"
					}
				}
			}
		}

		items = append(items, item)
	}
	return items, total, nil
}

func (s *ContactService) TokenInfo(ctx context.Context, userID uuid.UUID, offset, limit int) (int, []dto.TokenTransactionResponse, int64, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil || u == nil {
		return 0, nil, 0, ErrNotFound
	}
	txs, total, err := s.contacts.TokenTransactions(ctx, userID, offset, limit)
	if err != nil {
		return u.TokenBalance, nil, 0, err
	}

	type listingRef struct {
		listingType string
		listingID   uuid.UUID
	}
	cargoIDs := make([]uuid.UUID, 0)
	warehouseIDs := make([]uuid.UUID, 0)
	txMetas := make([]listingRef, len(txs))

	for i, tx := range txs {
		if tx.Reason == "contact_view" && len(tx.Meta) > 0 {
			var m map[string]string
			if json.Unmarshal(tx.Meta, &m) == nil {
				lid, _ := uuid.Parse(m["listing_id"])
				txMetas[i] = listingRef{listingType: m["listing_type"], listingID: lid}
				if m["listing_type"] == "cargo" {
					cargoIDs = append(cargoIDs, lid)
				} else if m["listing_type"] == "warehouse" {
					warehouseIDs = append(warehouseIDs, lid)
				}
			}
		}
	}

	cargoList, _ := s.cargo.FindByIDs(ctx, cargoIDs)
	warehouseList, _ := s.warehouse.FindByIDs(ctx, warehouseIDs)

	cargoMap := make(map[uuid.UUID]*model.CargoListing, len(cargoList))
	for _, c := range cargoList {
		cargoMap[c.ID] = c
	}
	warehouseMap := make(map[uuid.UUID]*model.WarehouseListing, len(warehouseList))
	for _, w := range warehouseList {
		warehouseMap[w.ID] = w
	}

	result := make([]dto.TokenTransactionResponse, len(txs))
	for i, tx := range txs {
		var balanceBefore int
		if tx.Type == "credit" {
			balanceBefore = tx.BalanceAfter - tx.Amount
		} else {
			balanceBefore = tx.BalanceAfter + tx.Amount
		}

		resp := dto.TokenTransactionResponse{
			ID:            tx.ID,
			Type:          tx.Type,
			Amount:        tx.Amount,
			Reason:        tx.Reason,
			BalanceBefore: balanceBefore,
			BalanceAfter:  tx.BalanceAfter,
			Meta:          json.RawMessage(tx.Meta),
			CreatedAt:     tx.CreatedAt,
		}

		ref := txMetas[i]
		if ref.listingType == "cargo" {
			if c, ok := cargoMap[ref.listingID]; ok {
				resp.Listing = &dto.TokenTxListing{
					ID:       c.ID,
					Type:     "cargo",
					Name:     c.CargoName,
					FromCity: c.FromCity,
				}
				resp.Description = "Kontakt ko'rish — " + c.CargoName
			}
		} else if ref.listingType == "warehouse" {
			if w, ok := warehouseMap[ref.listingID]; ok {
				resp.Listing = &dto.TokenTxListing{
					ID:       w.ID,
					Type:     "warehouse",
					Name:     w.Name,
					FromCity: w.City,
				}
				resp.Description = "Kontakt ko'rish — " + w.Name
			}
		}

		if resp.Description == "" {
			switch tx.Reason {
			case "manual":
				resp.Description = "Administrator tomonidan qo'shildi"
			case "payment_completed", "token_purchase":
				resp.Description = "Token sotib olish"
			case "subscription_grant":
				resp.Description = "Obuna orqali tokenlar berildi"
			case "refund", "payment_revert":
				resp.Description = "To'lov qaytarildi"
			default:
				resp.Description = tx.Reason
			}
		}

		result[i] = resp
	}

	return u.TokenBalance, result, total, nil
}
