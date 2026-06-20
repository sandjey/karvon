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

func (s *ContactService) TokenInfo(ctx context.Context, userID uuid.UUID) (int, []model.TokenTransaction, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil || u == nil {
		return 0, nil, ErrNotFound
	}
	txs, err := s.contacts.TokenTransactions(ctx, userID, 100)
	return u.TokenBalance, txs, err
}
