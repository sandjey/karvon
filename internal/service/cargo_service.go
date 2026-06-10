package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"karvon/internal/dto"
	"karvon/internal/model"
	"karvon/internal/repository"
)

type CargoService struct {
	repo      *repository.CargoRepo
	company   *repository.CompanyRepo
	routes    *repository.RouteRepo
	notif     *repository.NotificationRepo
	favorites *repository.FavoriteRepo
}

func NewCargoService(
	repo *repository.CargoRepo,
	company *repository.CompanyRepo,
	routes *repository.RouteRepo,
	notif *repository.NotificationRepo,
	favorites *repository.FavoriteRepo,
) *CargoService {
	return &CargoService{repo: repo, company: company, routes: routes, notif: notif, favorites: favorites}
}

// resolveCompany выбирает компанию для публикации: явно указанную (своя+approved) или первую approved.
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

func (s *CargoService) Create(ctx context.Context, userID uuid.UUID, req dto.CargoUpsertRequest) (*model.CargoListing, error) {
	companyID, err := s.resolveCompany(ctx, userID, req.CompanyID)
	if err != nil {
		return nil, err
	}

	m := &model.CargoListing{
		ID:        uuid.New(),
		CompanyID: companyID,
		UserID:    userID,
		Status:    "active",
	}
	applyCargo(&req, m)

	if req.DurationHours != nil && *req.DurationHours > 0 {
		m.DurationHours = *req.DurationHours
	} else {
		m.DurationHours = 24
	}
	exp := time.Now().Add(time.Duration(m.DurationHours) * time.Hour)
	m.ExpiresAt = &exp

	for i, w := range req.Waypoints {
		m.Waypoints = append(m.Waypoints, model.CargoWaypoint{
			SortOrder:    pickSort(w.SortOrder, i),
			WaypointType: w.WaypointType,
			City:         w.City,
			Country:      w.Country,
		})
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}

	s.notifyMatchingRoutes(ctx, m)
	return s.repo.FindByID(ctx, m.ID)
}

func (s *CargoService) GetByID(ctx context.Context, id, viewerID uuid.UUID) (*model.CargoListing, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrListingNotFound
	}
	if c.UserID != viewerID {
		_ = s.repo.IncrementViews(ctx, id)
	}
	return c, nil
}

func (s *CargoService) List(ctx context.Context, f repository.CargoFilter) ([]model.CargoListing, int64, error) {
	return s.repo.List(ctx, f)
}

func (s *CargoService) ListMine(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.CargoListing, int64, error) {
	return s.repo.ListByUser(ctx, userID, offset, limit)
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

func (s *CargoService) Update(ctx context.Context, id, userID uuid.UUID, req dto.CargoUpsertRequest) (*model.CargoListing, error) {
	c, err := s.ownedListing(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	c.Waypoints = nil
	applyCargo(&req, c)
	if err := s.repo.Save(ctx, c); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *CargoService) SetStatus(ctx context.Context, id, userID uuid.UUID, status string) error {
	if _, err := s.ownedListing(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *CargoService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.ownedListing(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *CargoService) Duplicate(ctx context.Context, id, userID uuid.UUID) (*model.CargoListing, error) {
	src, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, ErrListingNotFound
	}
	if src.UserID != userID {
		return nil, ErrNotOwner
	}
	cp := *src
	cp.ID = uuid.New()
	cp.Status = "active"
	cp.IsTemplate = false
	cp.TemplateName = nil
	cp.FromDate = nil
	cp.ToDate = nil
	cp.ViewsCount = 0
	cp.ContactsBoughtCount = 0
	cp.IsBoosted = false
	cp.BoostExpiresAt = nil
	cp.Company = nil
	cp.User = nil
	cp.Waypoints = nil
	exp := time.Now().Add(time.Duration(cp.DurationHours) * time.Hour)
	cp.ExpiresAt = &exp
	for _, w := range src.Waypoints {
		cp.Waypoints = append(cp.Waypoints, model.CargoWaypoint{
			SortOrder:    w.SortOrder,
			WaypointType: w.WaypointType,
			City:         w.City,
			Country:      w.Country,
		})
	}
	if err := s.repo.Create(ctx, &cp); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, cp.ID)
}

func (s *CargoService) SaveAsTemplate(ctx context.Context, id, userID uuid.UUID, name string) (*model.CargoListing, error) {
	src, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, ErrListingNotFound
	}
	if src.UserID != userID {
		return nil, ErrNotOwner
	}
	cp := *src
	cp.ID = uuid.New()
	cp.IsTemplate = true
	cp.TemplateName = &name
	cp.Status = "archived"
	cp.ExpiresAt = nil
	cp.Company = nil
	cp.User = nil
	cp.Waypoints = nil
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
	cp := *src
	cp.ID = uuid.New()
	cp.IsTemplate = false
	cp.TemplateName = nil
	cp.Status = "active"
	cp.Company = nil
	cp.User = nil
	cp.Waypoints = nil
	exp := time.Now().Add(time.Duration(cp.DurationHours) * time.Hour)
	cp.ExpiresAt = &exp
	if err := s.repo.Create(ctx, &cp); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, cp.ID)
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
		favs[i] = dto.FavoriteUser{UserName: r.UserName, AddedAt: r.CreatedAt}
	}
	return &dto.CargoStatsResponse{
		ViewsCount:          c.ViewsCount,
		ContactsBoughtCount: c.ContactsBoughtCount,
		Favorites:           favs,
	}, nil
}

func (s *CargoService) notifyMatchingRoutes(ctx context.Context, c *model.CargoListing) {
	from, to := "", ""
	if c.FromCity != nil {
		from = *c.FromCity
	}
	if c.ToCity != nil {
		to = *c.ToCity
	}
	if from == "" && to == "" {
		return
	}
	routes, err := s.routes.FindMatching(ctx, from, to)
	if err != nil {
		return
	}
	meta, _ := json.Marshal(map[string]string{"listing_type": "cargo", "listing_id": c.ID.String()})
	for _, rt := range routes {
		if rt.UserID == c.UserID {
			continue
		}
		body := fmt.Sprintf("Новый груз по маршруту %s → %s", rt.FromCity, rt.ToCity)
		_ = s.notif.Create(ctx, &model.Notification{
			UserID: rt.UserID,
			Type:   "new_cargo_on_route",
			Title:  "Новый груз по вашему маршруту",
			Body:   &body,
			Meta:   meta,
		})
	}
}

func pickSort(v, fallback int) int {
	if v != 0 {
		return v
	}
	return fallback
}

// applyCargo переносит непустые поля запроса в модель.
func applyCargo(req *dto.CargoUpsertRequest, m *model.CargoListing) {
	m.Type = req.Type
	m.CargoName = req.CargoName
	m.CargoType = req.CargoType
	if req.IsADR != nil {
		m.IsADR = *req.IsADR
	}
	if req.HasTempRegime != nil {
		m.HasTempRegime = *req.HasTempRegime
	}
	m.TempMin = req.TempMin
	m.TempMax = req.TempMax
	m.WeightTon = req.WeightTon
	m.VolumeM3 = req.VolumeM3
	m.PlacesCount = req.PlacesCount
	m.LengthM = req.LengthM
	m.WidthM = req.WidthM
	m.HeightM = req.HeightM
	m.LoadingType = req.LoadingType
	if req.BodyTypes != nil {
		m.BodyTypes = pq.StringArray(req.BodyTypes)
	}
	m.VehicleType = req.VehicleType
	m.TractorAxles = req.TractorAxles
	m.TrailerAxles = req.TrailerAxles
	if req.OnlyRecoupling != nil {
		m.OnlyRecoupling = *req.OnlyRecoupling
	}
	m.FromCity = req.FromCity
	m.FromCountry = req.FromCountry
	m.FromDate = req.FromDate
	m.LoadingMethod = req.LoadingMethod
	m.ToCity = req.ToCity
	m.ToCountry = req.ToCountry
	m.ToDate = req.ToDate
	m.UnloadingMethod = req.UnloadingMethod
	if req.TransitCountries != nil {
		m.TransitCountries = pq.StringArray(req.TransitCountries)
	}
	m.BorderCrossing = req.BorderCrossing
	m.CustomsType = req.CustomsType
	m.Incoterms = req.Incoterms
	if req.Permits != nil {
		m.Permits = pq.StringArray(req.Permits)
	}
	m.RateMode = req.RateMode
	m.RateAmount = req.RateAmount
	m.RateCurrency = req.RateCurrency
	if req.RateVAT != nil {
		m.RateVAT = *req.RateVAT
	}
	m.PaymentTerms = req.PaymentTerms
	m.PaymentMethod = req.PaymentMethod
	m.PrepaymentPercent = req.PrepaymentPercent
	m.Notes = req.Notes
}
