package service

import (
	"context"

	"github.com/google/uuid"

	"karvon/internal/dto"
	"karvon/internal/model"
	"karvon/internal/repository"
)

type UserService struct {
	userRepo    *repository.UserRepo
	companyRepo *repository.CompanyRepo
}

func NewUserService(userRepo *repository.UserRepo, companyRepo *repository.CompanyRepo) *UserService {
	return &UserService{userRepo: userRepo, companyRepo: companyRepo}
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
	return &dto.UserStatsResponse{
		TokenBalance:      u.TokenBalance,
		CompaniesCount:    int(total),
		VerifiedCompanies: int(verified),
	}, nil
}
