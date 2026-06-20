package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/repository"
)

type CargoService struct {
	repo          *repository.CargoRepo
	company       *repository.CompanyRepo
	routes        *repository.RouteRepo
	notif         *repository.NotificationRepo
	favorites     *repository.FavoriteRepo
	media         *repository.MediaRepo
	warehouseRepo *repository.WarehouseRepo
	pricingRepo   *repository.PricingRepo
	contacts      *repository.ContactRepo
}

func NewCargoService(
	repo *repository.CargoRepo,
	company *repository.CompanyRepo,
	routes *repository.RouteRepo,
	notif *repository.NotificationRepo,
	favorites *repository.FavoriteRepo,
	media *repository.MediaRepo,
	warehouseRepo *repository.WarehouseRepo,
	pricingRepo *repository.PricingRepo,
	contactRepo *repository.ContactRepo,
) *CargoService {
	return &CargoService{repo: repo, company: company, routes: routes, notif: notif, favorites: favorites, media: media, warehouseRepo: warehouseRepo, pricingRepo: pricingRepo, contacts: contactRepo}
}

func (s *CargoService) resolveCompany(ctx context.Context, userID uuid.UUID, explicit *uuid.UUID) (uuid.UUID, error) {
	companies, err := s.company.FindByUserID(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}
	if explicit != nil {
		for _, c := range companies {
			if c.ID == *explicit && c.Status == "approved" {
				return c.ID, nil
			}
		}
		return uuid.Nil, ErrCompanyNotOwned
	}
	for _, c := range companies {
		if c.Status == "approved" {
			return c.ID, nil
		}
	}
	return uuid.Nil, ErrCompanyNotFound
}

func countPhotos(media []dto.MediaItem) int {
	n := 0
	for _, m := range media {
		if m.FileType == "photo" {
			n++
		}
	}
	return n
}

func (s *CargoService) Create(ctx context.Context, userID uuid.UUID, req dto.CargoUpsertRequest) (*model.CargoListing, error) {
	if req.CargoName == nil || *req.CargoName == "" {
		return nil, ErrValidation
	}
	if req.Category == nil || *req.Category == "" {
		return nil, ErrValidation
	}
	if req.CompanyID == nil {
		return nil, ErrValidation
	}
	if countPhotos(req.Media) > 10 {
		return nil, ErrPhotoLimitExceeded
	}
	companyID, err := s.resolveCompany(ctx, userID, req.CompanyID)
	if err != nil {
		return nil, err
	}

	// Проверяем: есть ли уже бесплатное активное объявление (груз или склад).
	freeCargo, err := s.repo.CountFreeActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	freeWarehouse, err := s.warehouseRepo.CountFreeActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	freeQuota := s.pricingRepo.TokensAmount(ctx, "free_listing_quota", 1)
	if int(freeCargo+freeWarehouse) >= freeQuota {
		return nil, ErrFreeListingUsed
	}

	m := &model.CargoListing{
		ID:        uuid.New(),
		CompanyID: companyID,
		UserID:    userID,
		Status:    "active",
		IsPaid:    false,
		InStock:   true,
		Type:      "domestic", // legacy NOT NULL
	}
	applyCargo(&req, m)
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	s.saveMedia(ctx, m.ID, req.Media)
	go s.notifyMatchingRoutes(m)
	return s.loadFull(ctx, m.ID)
}

func (s *CargoService) GetByID(ctx context.Context, id, viewerID uuid.UUID) (*model.CargoListing, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrListingNotFound
	}
	if c.IsAdminBlocked && c.UserID != viewerID {
		return nil, ErrListingNotFound
	}
	if c.UserID != viewerID {
		_ = s.repo.IncrementViews(ctx, id)
	}
	c.Media, _ = s.media.ListByEntity(ctx, "cargo", id)

	// Populyatsiya contact maydoni
	if viewerID != uuid.Nil {
		contact := &model.CargoContact{}
		if viewerID == c.UserID {
			contact.IsUnlocked = true
		} else {
			view, _ := s.contacts.RecentView(ctx, viewerID, "cargo", id, 30*24*time.Hour)
			contact.IsUnlocked = view != nil
		}
		if contact.IsUnlocked {
			if c.User != nil {
				contact.Phone = &c.User.Phone
				contact.WhatsApp = c.User.WhatsApp
				contact.Telegram = c.User.Telegram
				contact.Email = c.User.Email
				contact.City = c.User.City
				contact.Country = c.User.Country
			}
			if c.Company != nil {
				contact.CompanyName = &c.Company.Name
				if contact.City == nil {
					city := c.Company.City
					contact.City = &city
				}
			}
		}
		c.Contact = contact
	}

	return c, nil
}

func (s *CargoService) loadFull(ctx context.Context, id uuid.UUID) (*model.CargoListing, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil || c == nil {
		return c, err
	}
	c.Media, _ = s.media.ListByEntity(ctx, "cargo", id)
	return c, nil
}

func (s *CargoService) List(ctx context.Context, f repository.CargoFilter) ([]model.CargoListing, int64, error) {
	list, total, err := s.repo.List(ctx, f)
	if err == nil {
		s.attachMedia(ctx, list)
	}
	return list, total, err
}

func (s *CargoService) ListMine(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.CargoListing, int64, error) {
	list, total, err := s.repo.ListByUser(ctx, userID, offset, limit)
	if err == nil {
		s.attachMedia(ctx, list)
	}
	return list, total, err
}

func (s *CargoService) attachMedia(ctx context.Context, list []model.CargoListing) {
	if len(list) == 0 {
		return
	}
	ids := make([]uuid.UUID, len(list))
	for i := range list {
		ids[i] = list[i].ID
	}
	mediaMap, _ := s.media.ListByEntityIDs(ctx, "cargo", ids)
	for i := range list {
		list[i].Media = mediaMap[list[i].ID]
	}
}

func (s *CargoService) ownedListing(ctx context.Context, id, userID uuid.UUID) (*model.CargoListing, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrListingNotFound
	}
	if c.UserID != userID {
		return nil, ErrNotOwner
	}
	return c, nil
}

func (s *CargoService) Similar(ctx context.Context, id uuid.UUID, category string) ([]model.CargoListing, error) {
	return s.repo.Similar(ctx, id, category, 5)
}

func (s *CargoService) Update(ctx context.Context, id, userID uuid.UUID, req dto.CargoUpsertRequest) (*model.CargoListing, error) {
	if countPhotos(req.Media) > 10 {
		return nil, ErrPhotoLimitExceeded
	}
	c, err := s.ownedListing(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	applyCargo(&req, c)
	if err := s.repo.Save(ctx, c); err != nil {
		return nil, err
	}
	if req.Media != nil {
		s.saveMedia(ctx, id, req.Media)
	}
	return s.loadFull(ctx, id)
}

func (s *CargoService) SetStatus(ctx context.Context, id, userID uuid.UUID, status *string, inStock *bool) error {
	c, err := s.ownedListing(ctx, id, userID)
	if err != nil {
		return err
	}
	fields := map[string]interface{}{}
	if status != nil {
		if c.IsAdminBlocked && *status == "active" {
			return ErrListingNotFound
		}
		fields["status"] = *status
	}
	if inStock != nil {
		fields["in_stock"] = *inStock
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateStatus(ctx, id, fields)
}

func (s *CargoService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.ownedListing(ctx, id, userID); err != nil {
		return err
	}
	_ = s.media.DeleteByEntity(ctx, "cargo", id)
	return s.repo.Delete(ctx, id)
}

func (s *CargoService) Duplicate(ctx context.Context, id, userID uuid.UUID) (*model.CargoListing, error) {
	src, err := s.ownedListing(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	freeCargo, err := s.repo.CountFreeActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	freeWarehouse, err := s.warehouseRepo.CountFreeActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	initialStatus := "active"
	if freeCargo+freeWarehouse > 0 {
		initialStatus = "archived"
	}
	cp := *src
	cp.ID = uuid.New()
	cp.Status = initialStatus
	cp.IsPaid = false
	cp.InStock = true
	cp.IsTemplate = false
	cp.TemplateName = nil
	cp.ViewsCount = 0
	cp.ContactsBoughtCount = 0
	cp.IsBoosted = false
	cp.BoostExpiresAt = nil
	cp.Company = nil
	cp.User = nil
	cp.Media = nil
	if err := s.repo.Create(ctx, &cp); err != nil {
		return nil, err
	}
	s.copyMedia(ctx, id, cp.ID)
	return s.loadFull(ctx, cp.ID)
}

func (s *CargoService) SaveAsTemplate(ctx context.Context, id, userID uuid.UUID, name string) (*model.CargoListing, error) {
	src, err := s.ownedListing(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	cp := *src
	cp.ID = uuid.New()
	cp.IsTemplate = true
	cp.TemplateName = &name
	cp.Status = "archived"
	cp.Company = nil
	cp.User = nil
	cp.Media = nil
	if err := s.repo.Create(ctx, &cp); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, cp.ID)
}

func (s *CargoService) Templates(ctx context.Context, userID uuid.UUID) ([]model.CargoListing, error) {
	return s.repo.ListTemplates(ctx, userID)
}

func (s *CargoService) FromTemplate(ctx context.Context, templateID, userID uuid.UUID) (*model.CargoListing, error) {
	src, err := s.repo.FindByID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	if src == nil || !src.IsTemplate {
		return nil, ErrListingNotFound
	}
	if src.UserID != userID {
		return nil, ErrNotOwner
	}
	freeCargo, err := s.repo.CountFreeActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	freeWarehouse, err := s.warehouseRepo.CountFreeActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	initialStatus := "active"
	if freeCargo+freeWarehouse > 0 {
		initialStatus = "archived"
	}
	cp := *src
	cp.ID = uuid.New()
	cp.IsTemplate = false
	cp.TemplateName = nil
	cp.Status = initialStatus
	cp.IsPaid = false
	cp.InStock = true
	cp.Company = nil
	cp.User = nil
	cp.Media = nil
	if err := s.repo.Create(ctx, &cp); err != nil {
		return nil, err
	}
	return s.loadFull(ctx, cp.ID)
}

func (s *CargoService) Stats(ctx context.Context, id, userID uuid.UUID) (*dto.CargoStatsResponse, error) {
	c, err := s.ownedListing(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.favorites.UsersByListing(ctx, "cargo", id)
	if err != nil {
		return nil, err
	}
	favs := make([]dto.FavoriteUser, len(rows))
	for i, r := range rows {
		favs[i] = dto.FavoriteUser{UserName: r.UserName, CompanyName: r.CompanyName, AddedAt: r.CreatedAt}
	}
	return &dto.CargoStatsResponse{
		ViewsCount:          c.ViewsCount,
		ContactsBoughtCount: c.ContactsBoughtCount,
		Favorites:           favs,
	}, nil
}

func (s *CargoService) saveMedia(ctx context.Context, cargoID uuid.UUID, items []dto.MediaItem) {
	if items == nil {
		return
	}
	media := make([]model.ListingMedia, len(items))
	for i, it := range items {
		media[i] = model.ListingMedia{
			FileURL:      it.FileURL,
			FileType:     it.FileType,
			OriginalName: it.OriginalName,
			SortOrder:    it.SortOrder,
		}
	}
	_ = s.media.Replace(ctx, "cargo", cargoID, media)
}

func (s *CargoService) copyMedia(ctx context.Context, srcID, dstID uuid.UUID) {
	items, err := s.media.ListByEntity(ctx, "cargo", srcID)
	if err != nil || len(items) == 0 {
		return
	}
	for i := range items {
		items[i].ID = uuid.Nil
	}
	_ = s.media.Replace(ctx, "cargo", dstID, items)
}

func (s *CargoService) notifyMatchingRoutes(c *model.CargoListing) {
	from := ""
	if c.FromCity != nil {
		from = *c.FromCity
	}
	if from == "" {
		return
	}
	ctx := context.Background()
	routes, err := s.routes.FindMatching(ctx, from, "")
	if err != nil {
		return
	}
	meta, _ := json.Marshal(map[string]string{"listing_type": "cargo", "listing_id": c.ID.String()})
	for _, rt := range routes {
		if rt.UserID == c.UserID {
			continue
		}
		body := fmt.Sprintf("Новый товар из %s по вашему маршруту", from)
		_ = s.notif.Create(ctx, &model.Notification{
			UserID: rt.UserID,
			Type:   "new_cargo_on_route",
			Title:  "Новый товар по вашему маршруту",
			Body:   &body,
			Meta:   meta,
		})
	}
}

// applyCargo переносит поля запроса в модель карточки товара.
func applyCargo(req *dto.CargoUpsertRequest, m *model.CargoListing) {
	if req.CargoName != nil {
		m.CargoName = *req.CargoName
	}
	if req.Category != nil {
		m.Category = *req.Category
	}
	m.Description = req.Description
	m.QuantityAvailable = req.QuantityAvailable
	m.QuantityUnit = req.QuantityUnit
	m.MinOrder = req.MinOrder
	m.MinOrderUnit = req.MinOrderUnit
	m.Divisibility = req.Divisibility
	m.Packaging = req.Packaging
	m.WeightPerPlace = req.WeightPerPlace
	m.LengthM = req.LengthM
	m.WidthM = req.WidthM
	m.HeightM = req.HeightM
	m.PriceMode = req.PriceMode
	m.PricePerUnit = req.PricePerUnit
	m.Currency = req.Currency
	m.VatMode = req.VatMode
	m.PaymentTerms = req.PaymentTerms
	m.PaymentMethod = req.PaymentMethod
	m.FromCountry = req.FromCountry
	m.FromCity = req.FromCity
	m.ToCountry = req.ToCountry
	m.ToCity = req.ToCity
	m.PickupAddress = req.PickupAddress
	m.Lat = req.Lat
	m.Lng = req.Lng
	if req.LoadingMethods != nil {
		m.LoadingMethods = pq.StringArray(req.LoadingMethods)
	}
	if req.WorkingHours != nil {
		m.WorkingHours = req.WorkingHours
	}
	if req.IsADR != nil {
		m.IsADR = *req.IsADR
	}
	m.ADRClass = req.ADRClass
	if req.HasTempRegime != nil {
		m.HasTempRegime = *req.HasTempRegime
	}
	m.TempMin = req.TempMin
	m.TempMax = req.TempMax
	if req.RequiredBodyTypes != nil {
		m.RequiredBodyTypes = pq.StringArray(req.RequiredBodyTypes)
	}
	m.Incoterms = req.Incoterms
	if req.DeliveryGeography != nil {
		m.DeliveryGeography = pq.StringArray(req.DeliveryGeography)
	}
	if req.Permits != nil {
		m.Permits = pq.StringArray(req.Permits)
	}
	m.CommentForCarrier = req.CommentForCarrier
}
