package service

import (
	"context"

	"github.com/google/uuid"

	"ctm/internal/dto"
	"ctm/internal/model"
	"ctm/internal/repository"
)

type CompanyService struct {
	repo *repository.CompanyRepo
}

func NewCompanyService(repo *repository.CompanyRepo) *CompanyService {
	return &CompanyService{repo: repo}
}

func (s *CompanyService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateCompanyRequest) (*model.Company, error) {
	exists, err := s.repo.INNExistsApproved(ctx, req.INN, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	c := &model.Company{
		ID:         uuid.New(),
		UserID:     userID,
		Country:    req.Country,
		OrgType:    req.OrgType,
		Name:       req.Name,
		INN:        req.INN,
		Phone:      req.Phone,
		Email:      req.Email,
		City:       req.City,
		Region:     req.Region,
		Street:     req.Street,
		PostalCode: req.PostalCode,
		RegDocURL:  req.RegDocURL,
		Status:     "pending",
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CompanyService) GetAll(ctx context.Context, userID uuid.UUID) ([]model.Company, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *CompanyService) GetByID(ctx context.Context, userID, companyID uuid.UUID) (*model.Company, error) {
	c, err := s.repo.FindByID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCompanyNotFound
	}
	if c.UserID != userID {
		return nil, ErrCompanyNotOwned
	}
	return c, nil
}

func (s *CompanyService) Update(ctx context.Context, userID, companyID uuid.UUID, req dto.UpdateCompanyRequest) error {
	c, err := s.repo.FindByID(ctx, companyID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCompanyNotFound
	}
	if c.UserID != userID {
		return ErrCompanyNotOwned
	}
	// pending — нельзя редактировать ничего, кроме docs_requested
	if c.Status == "pending" {
		return ErrCompanyNotEditable
	}

	fields := map[string]interface{}{}
	if req.Country != nil {
		fields["country"] = *req.Country
	}
	if req.OrgType != nil {
		fields["org_type"] = *req.OrgType
	}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Phone != nil {
		fields["phone"] = *req.Phone
	}
	if req.Email != nil {
		fields["email"] = *req.Email
	}
	if req.City != nil {
		fields["city"] = *req.City
	}
	if req.Region != nil {
		fields["region"] = *req.Region
	}
	if req.Street != nil {
		fields["street"] = *req.Street
	}
	if req.PostalCode != nil {
		fields["postal_code"] = *req.PostalCode
	}
	// документ (свидетельство) можно обновить только при docs_requested
	if c.Status == "docs_requested" {
		if req.RegDocURL != nil {
			fields["reg_doc_url"] = *req.RegDocURL
		}
		// при обновлении после запроса документов — вернуть статус в pending
		if len(fields) > 0 {
			fields["status"] = "pending"
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(ctx, companyID, fields)
}
