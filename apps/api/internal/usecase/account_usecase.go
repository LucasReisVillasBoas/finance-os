package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
)

// Sentinel errors for accounts
var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrCannotDeleteSystem = errors.New("cannot delete a system resource")
)

// CreateAccountRequest holds the data needed to create an account.
type CreateAccountRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=255"`
	Type        string   `json:"type" binding:"required,oneof=checking savings credit_card investment wallet other"`
	Institution *string  `json:"institution"`
	Balance     float64  `json:"balance"`
	CreditLimit *float64 `json:"credit_limit"`
	Color       *string  `json:"color"`
	Icon        *string  `json:"icon"`
}

// UpdateAccountRequest holds the fields that can be updated on an account.
type UpdateAccountRequest struct {
	Name        *string  `json:"name"`
	Institution *string  `json:"institution"`
	CreditLimit *float64 `json:"credit_limit"`
	Color       *string  `json:"color"`
	Icon        *string  `json:"icon"`
}

// AccountUseCase defines the business logic interface for accounts.
type AccountUseCase interface {
	Create(ctx context.Context, userID uuid.UUID, req CreateAccountRequest) (*entity.Account, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Account, error)
	GetAll(ctx context.Context, userID uuid.UUID) ([]*entity.Account, error)
	Update(ctx context.Context, id, userID uuid.UUID, req UpdateAccountRequest) (*entity.Account, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	GetSummary(ctx context.Context, userID uuid.UUID) (*domainrepo.AccountSummary, error)
}

type accountUseCase struct {
	repo domainrepo.AccountRepository
}

// NewAccountUseCase creates a new AccountUseCase implementation.
func NewAccountUseCase(repo domainrepo.AccountRepository) AccountUseCase {
	return &accountUseCase{repo: repo}
}

func (uc *accountUseCase) Create(ctx context.Context, userID uuid.UUID, req CreateAccountRequest) (*entity.Account, error) {
	now := time.Now()
	account := &entity.Account{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Type:        req.Type,
		Institution: req.Institution,
		Balance:     req.Balance,
		CreditLimit: req.CreditLimit,
		Color:       req.Color,
		Icon:        req.Icon,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.repo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("accountUseCase.Create: %w", err)
	}
	return account, nil
}

func (uc *accountUseCase) GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Account, error) {
	account, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("accountUseCase.GetByID: %w", err)
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

func (uc *accountUseCase) GetAll(ctx context.Context, userID uuid.UUID) ([]*entity.Account, error) {
	accounts, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("accountUseCase.GetAll: %w", err)
	}
	return accounts, nil
}

func (uc *accountUseCase) Update(ctx context.Context, id, userID uuid.UUID, req UpdateAccountRequest) (*entity.Account, error) {
	account, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("accountUseCase.Update find: %w", err)
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	if req.Name != nil {
		account.Name = *req.Name
	}
	if req.Institution != nil {
		account.Institution = req.Institution
	}
	if req.CreditLimit != nil {
		account.CreditLimit = req.CreditLimit
	}
	if req.Color != nil {
		account.Color = req.Color
	}
	if req.Icon != nil {
		account.Icon = req.Icon
	}

	if err := uc.repo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("accountUseCase.Update save: %w", err)
	}
	return account, nil
}

func (uc *accountUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	account, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("accountUseCase.Delete find: %w", err)
	}
	if account == nil {
		return ErrAccountNotFound
	}

	if err := uc.repo.SoftDelete(ctx, id, userID); err != nil {
		return fmt.Errorf("accountUseCase.Delete soft delete: %w", err)
	}
	return nil
}

func (uc *accountUseCase) GetSummary(ctx context.Context, userID uuid.UUID) (*domainrepo.AccountSummary, error) {
	summary, err := uc.repo.GetSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("accountUseCase.GetSummary: %w", err)
	}
	return summary, nil
}
