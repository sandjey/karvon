package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"karvon/internal/dto"
	"karvon/internal/model"
	"karvon/internal/repository"
	jwtpkg "karvon/pkg/jwt"
)

var adminPhoneRe = regexp.MustCompile(`[^0-9+]`)

type AdminService struct {
	admin     *repository.AdminRepo
	users     *repository.UserRepo
	tokens    *repository.TokenRepo
	cargo     *repository.CargoRepo
	warehouse *repository.WarehouseRepo
	pricing   *PricingService
	jwtMgr    *jwtpkg.Manager

	adminLogin    string
	adminPassword string
}

func NewAdminService(
	admin *repository.AdminRepo,
	users *repository.UserRepo,
	tokens *repository.TokenRepo,
	cargo *repository.CargoRepo,
	warehouse *repository.WarehouseRepo,
	pricing *PricingService,
	jwtMgr *jwtpkg.Manager,
	adminLogin, adminPassword string,
) *AdminService {
	return &AdminService{
		admin: admin, users: users, tokens: tokens,
		cargo: cargo, warehouse: warehouse, pricing: pricing, jwtMgr: jwtMgr,
		adminLogin: adminLogin, adminPassword: adminPassword,
	}
}

// SeedSuperAdmin создаёт скрытого статик-админа при старте, если его ещё нет.
func (s *AdminService) SeedSuperAdmin(ctx context.Context) error {
	phone := s.adminPhone()
	u, err := s.users.FindByPhone(ctx, phone)
	if err != nil {
		return err
	}
	if u != nil {
		return nil
	}
	name := "Super Admin"
	return s.users.Create(ctx, &model.User{
		ID:    uuid.New(),
		Phone: phone,
		Name:  &name,
		Role:  "super_admin",
	})
}

func (s *AdminService) adminPhone() string {
	// логин используется как уникальный «телефон» скрытого админа
	return s.adminLogin
}

// Login проверяет статик-креды и выдаёт JWT супер-админу (без OTP).
func (s *AdminService) Login(ctx context.Context, login, password string) (*dto.TokenPair, error) {
	if login != s.adminLogin || password != s.adminPassword {
		return nil, ErrInvalidCredentials
	}
	u, err := s.users.FindByPhone(ctx, s.adminPhone())
	if err != nil {
		return nil, err
	}
	if u == nil {
		// на случай если seed не отработал
		if err := s.SeedSuperAdmin(ctx); err != nil {
			return nil, err
		}
		u, err = s.users.FindByPhone(ctx, s.adminPhone())
		if err != nil || u == nil {
			return nil, ErrNotFound
		}
	}
	return s.issueTokens(ctx, u)
}

func (s *AdminService) issueTokens(ctx context.Context, u *model.User) (*dto.TokenPair, error) {
	access, err := s.jwtMgr.GenerateAccess(u.ID, u.Role)
	if err != nil {
		return nil, err
	}
	refresh, expiresAt, err := s.jwtMgr.GenerateRefresh(u.ID)
	if err != nil {
		return nil, err
	}
	if err := s.tokens.Save(ctx, u.ID, refresh, expiresAt); err != nil {
		return nil, err
	}
	return &dto.TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

// CreateModerator создаёт нового модератора или повышает существующего пользователя.
func (s *AdminService) CreateModerator(ctx context.Context, phone string, name *string) (*model.User, error) {
	phone = normalizeAdminPhone(phone)
	existing, err := s.users.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if err := s.users.UpdateProfile(ctx, existing.ID, map[string]interface{}{"role": "moderator"}); err != nil {
			return nil, err
		}
		existing.Role = "moderator"
		return existing, nil
	}
	u := &model.User{ID: uuid.New(), Phone: phone, Name: name, Role: "moderator"}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// DeleteModerator понижает модератора обратно до обычного пользователя.
func (s *AdminService) DeleteModerator(ctx context.Context, id uuid.UUID) error {
	u, err := s.users.FindByID(ctx, id)
	if err != nil || u == nil {
		return ErrNotFound
	}
	return s.users.UpdateProfile(ctx, id, map[string]interface{}{"role": "user"})
}

func (s *AdminService) Dashboard(ctx context.Context, period string) (*repository.DashboardMetrics, error) {
	since := time.Now().Add(-720 * time.Hour) // 30d
	switch period {
	case "7d":
		since = time.Now().Add(-7 * 24 * time.Hour)
	case "90d":
		since = time.Now().Add(-90 * 24 * time.Hour)
	}
	return s.admin.Dashboard(ctx, since)
}

func (s *AdminService) Users(ctx context.Context, search, role string, offset, limit int) ([]model.User, int64, error) {
	return s.admin.Users(ctx, search, role, offset, limit)
}

func (s *AdminService) User(ctx context.Context, id uuid.UUID) (*model.User, error) {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *AdminService) SetBlocked(ctx context.Context, id uuid.UUID, blocked bool) error {
	return s.admin.SetBlocked(ctx, id, blocked)
}

func (s *AdminService) TopupTokens(ctx context.Context, userID uuid.UUID, amount int) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil || u == nil {
		return ErrNotFound
	}
	return s.users.CreditTokens(ctx, userID, amount, "manual")
}

func (s *AdminService) Companies(ctx context.Context, status string, offset, limit int) ([]model.Company, int64, error) {
	return s.admin.AllCompanies(ctx, status, offset, limit)
}

func (s *AdminService) Payments(ctx context.Context, offset, limit int) ([]model.PaymentOrder, int64, error) {
	return s.admin.AllPayments(ctx, offset, limit)
}

func (s *AdminService) DeleteListing(ctx context.Context, listingType string, id uuid.UUID) error {
	switch listingType {
	case "cargo":
		return s.cargo.Delete(ctx, id)
	case "warehouse":
		return s.warehouse.Delete(ctx, id)
	}
	return ErrListingNotFound
}

func (s *AdminService) BlockListing(ctx context.Context, listingType string, id uuid.UUID) error {
	switch listingType {
	case "cargo":
		return s.cargo.UpdateStatus(ctx, id, "archived")
	case "warehouse":
		return s.warehouse.UpdateStatus(ctx, id, "archived")
	}
	return ErrListingNotFound
}

func (s *AdminService) Cargo(ctx context.Context, offset, limit int) ([]model.CargoListing, int64, error) {
	return s.admin.AllCargo(ctx, offset, limit)
}

func (s *AdminService) Warehouses(ctx context.Context, offset, limit int) ([]model.WarehouseListing, int64, error) {
	return s.admin.AllWarehouses(ctx, offset, limit)
}

func (s *AdminService) Pricing(ctx context.Context) ([]model.PricingConfig, error) {
	return s.pricing.ListAll(ctx)
}

func (s *AdminService) UpdatePricing(ctx context.Context, key string, fields map[string]interface{}, by uuid.UUID) error {
	return s.pricing.Update(ctx, key, fields, by)
}

func normalizeAdminPhone(phone string) string {
	phone = adminPhoneRe.ReplaceAllString(phone, "")
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}
	return phone
}
