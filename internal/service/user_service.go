package service

import (
	"context"

	"github.com/google/uuid"

	"karvon/internal/dto"
	"karvon/internal/model"
	"karvon/internal/repository"
)

type UserService struct {
	userRepo      *repository.UserRepo
	companyRepo   *repository.CompanyRepo
	cargoRepo     *repository.CargoRepo
	warehouseRepo *repository.WarehouseRepo
}

func NewUserService(
	userRepo *repository.UserRepo,
	companyRepo *repository.CompanyRepo,
	cargoRepo *repository.CargoRepo,
	warehouseRepo *repository.WarehouseRepo,
) *UserService {
	return &UserService{
		userRepo:      userRepo,
		companyRepo:   companyRepo,
		cargoRepo:     cargoRepo,
		warehouseRepo: warehouseRepo,
	}
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateProfileRequest) error {
	fields := map[string]interface{}{}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Email != nil {
		fields["email"] = *req.Email
	}
	if req.ExtraPhone != nil {
		fields["extra_phone"] = *req.ExtraPhone
	}
	if req.WhatsApp != nil {
		fields["whatsapp"] = *req.WhatsApp
	}
	if req.Telegram != nil {
		fields["telegram"] = *req.Telegram
	}
	if req.City != nil {
		fields["city"] = *req.City
	}
	if req.Country != nil {
		fields["country"] = *req.Country
	}
	if len(fields) == 0 {
		return nil
	}
	return s.userRepo.UpdateProfile(ctx, userID, fields)
}

func (s *UserService) GetEvents(ctx context.Context, userID uuid.UUID) ([]repository.DashboardEvent, error) {
	return s.userRepo.DashboardEvents(ctx, userID, 50)
}

func (s *UserService) GetStats(ctx context.Context, userID uuid.UUID) (*dto.UserStatsResponse, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u == nil {
		return nil, ErrNotFound
	}
	total, verified, err := s.userRepo.CountCompanies(ctx, userID)
	if err != nil {
		return nil, err
	}
	activeCargo, err := s.cargoRepo.CountActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	activeWarehouses, err := s.warehouseRepo.CountActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	contacts, err := s.userRepo.CountContactsPurchasedByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &dto.UserStatsResponse{
		TokenBalance:      u.TokenBalance,
		CompaniesCount:    int(total),
		VerifiedCompanies: int(verified),
		ActiveCargo:       int(activeCargo),
		ActiveWarehouses:  int(activeWarehouses),
		ContactsPurchased: int(contacts),
	}, nil
}

// GetMyCargo возвращает все объявления-грузы пользователя (не шаблоны).
func (s *UserService) GetMyCargo(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.CargoListing, int64, error) {
	return s.cargoRepo.ListByUser(ctx, userID, offset, limit)
}

// GetMyWarehouses возвращает все склады пользователя.
func (s *UserService) GetMyWarehouses(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.WarehouseListing, int64, error) {
	return s.warehouseRepo.ListByUser(ctx, userID, offset, limit)
}

// GetListingQuota возвращает информацию о бесплатном слоте объявлений.
func (s *UserService) GetListingQuota(ctx context.Context, userID uuid.UUID) (*dto.ListingQuotaResponse, error) {
	freeCargo, err := s.cargoRepo.CountFreeActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	freeWarehouse, err := s.warehouseRepo.CountFreeActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	freeUsed := int(freeCargo + freeWarehouse)
	const freeTotal = 1
	return &dto.ListingQuotaResponse{
		FreeUsed:    freeUsed,
		FreeTotal:   freeTotal,
		CanPostFree: freeUsed < freeTotal,
		PricingKey:  "listing_paid",
	}, nil
}
