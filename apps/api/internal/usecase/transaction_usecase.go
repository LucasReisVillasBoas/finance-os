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

// Sentinel errors for transactions
var (
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrSameAccountTransfer    = errors.New("transfer source and destination must be different accounts")
)

// CreateTransactionRequest holds the data needed to create a transaction.
type CreateTransactionRequest struct {
	AccountID   uuid.UUID  `json:"account_id" binding:"required"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Type        string     `json:"type" binding:"required,oneof=income expense"`
	Amount      float64    `json:"amount" binding:"required,gt=0"`
	Description *string    `json:"description"`
	Notes       *string    `json:"notes"`
	Date        time.Time  `json:"date" binding:"required"`
	Tags        []string   `json:"tags"`
}

// CreateTransferRequest holds the data needed to create a transfer between accounts.
type CreateTransferRequest struct {
	FromAccountID uuid.UUID `json:"from_account_id" binding:"required"`
	ToAccountID   uuid.UUID `json:"to_account_id" binding:"required"`
	Amount        float64   `json:"amount" binding:"required,gt=0"`
	Description   *string   `json:"description"`
	Date          time.Time `json:"date" binding:"required"`
}

// UpdateTransactionRequest holds the fields that can be updated on a transaction.
type UpdateTransactionRequest struct {
	CategoryID  *uuid.UUID `json:"category_id"`
	Description *string    `json:"description"`
	Notes       *string    `json:"notes"`
	Date        *time.Time `json:"date"`
	Tags        []string   `json:"tags"`
}

// ListTransactionsRequest holds filter/pagination params for listing transactions.
type ListTransactionsRequest struct {
	StartDate  *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate    *time.Time `form:"end_date" time_format:"2006-01-02"`
	CategoryID *uuid.UUID `form:"category_id"`
	AccountID  *uuid.UUID `form:"account_id"`
	Type       *string    `form:"type"`
	Search     *string    `form:"search"`
	Page       int        `form:"page,default=1"`
	PageSize   int        `form:"page_size,default=20"`
}

// TransactionUseCase defines the business logic interface for transactions.
type TransactionUseCase interface {
	Create(ctx context.Context, userID uuid.UUID, req CreateTransactionRequest) (*entity.Transaction, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Transaction, error)
	List(ctx context.Context, userID uuid.UUID, req ListTransactionsRequest) ([]*entity.Transaction, int, error)
	Update(ctx context.Context, id, userID uuid.UUID, req UpdateTransactionRequest) (*entity.Transaction, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	GetSummary(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*domainrepo.TransactionSummary, error)
	CreateTransfer(ctx context.Context, userID uuid.UUID, req CreateTransferRequest) ([]*entity.Transaction, error)
}

type transactionUseCase struct {
	repo domainrepo.TransactionRepository
}

// NewTransactionUseCase creates a new TransactionUseCase implementation.
func NewTransactionUseCase(repo domainrepo.TransactionRepository) TransactionUseCase {
	return &transactionUseCase{repo: repo}
}

func (uc *transactionUseCase) Create(ctx context.Context, userID uuid.UUID, req CreateTransactionRequest) (*entity.Transaction, error) {
	now := time.Now()
	tx := &entity.Transaction{
		ID:            uuid.New(),
		UserID:        userID,
		AccountID:     req.AccountID,
		CategoryID:    req.CategoryID,
		Type:          req.Type,
		Amount:        req.Amount,
		Description:   req.Description,
		Notes:         req.Notes,
		Date:          req.Date,
		Tags:          req.Tags,
		AICategorized: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if tx.Tags == nil {
		tx.Tags = []string{}
	}

	if err := uc.repo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("transactionUseCase.Create: %w", err)
	}
	return tx, nil
}

func (uc *transactionUseCase) GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Transaction, error) {
	tx, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("transactionUseCase.GetByID: %w", err)
	}
	if tx == nil {
		return nil, ErrTransactionNotFound
	}
	return tx, nil
}

func (uc *transactionUseCase) List(ctx context.Context, userID uuid.UUID, req ListTransactionsRequest) ([]*entity.Transaction, int, error) {
	filter := domainrepo.TransactionFilter{
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		CategoryID: req.CategoryID,
		AccountID:  req.AccountID,
		Type:       req.Type,
		Search:     req.Search,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	txs, total, err := uc.repo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("transactionUseCase.List: %w", err)
	}
	return txs, total, nil
}

func (uc *transactionUseCase) Update(ctx context.Context, id, userID uuid.UUID, req UpdateTransactionRequest) (*entity.Transaction, error) {
	tx, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("transactionUseCase.Update find: %w", err)
	}
	if tx == nil {
		return nil, ErrTransactionNotFound
	}

	if req.CategoryID != nil {
		tx.CategoryID = req.CategoryID
	}
	if req.Description != nil {
		tx.Description = req.Description
	}
	if req.Notes != nil {
		tx.Notes = req.Notes
	}
	if req.Date != nil {
		tx.Date = *req.Date
	}
	if req.Tags != nil {
		tx.Tags = req.Tags
	}

	if err := uc.repo.Update(ctx, tx); err != nil {
		return nil, fmt.Errorf("transactionUseCase.Update save: %w", err)
	}
	return tx, nil
}

func (uc *transactionUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tx, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("transactionUseCase.Delete find: %w", err)
	}
	if tx == nil {
		return ErrTransactionNotFound
	}

	if err := uc.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("transactionUseCase.Delete: %w", err)
	}
	return nil
}

func (uc *transactionUseCase) GetSummary(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*domainrepo.TransactionSummary, error) {
	summary, err := uc.repo.GetSummary(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("transactionUseCase.GetSummary: %w", err)
	}
	return summary, nil
}

func (uc *transactionUseCase) CreateTransfer(ctx context.Context, userID uuid.UUID, req CreateTransferRequest) ([]*entity.Transaction, error) {
	if req.FromAccountID == req.ToAccountID {
		return nil, ErrSameAccountTransfer
	}

	now := time.Now()
	pairID := uuid.New()

	debit := &entity.Transaction{
		ID:             uuid.New(),
		UserID:         userID,
		AccountID:      req.FromAccountID,
		Type:           "transfer",
		Amount:         req.Amount,
		Description:    req.Description,
		Date:           req.Date,
		TransferPairID: &pairID,
		Tags:           []string{},
		AICategorized:  false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	credit := &entity.Transaction{
		ID:             uuid.New(),
		UserID:         userID,
		AccountID:      req.ToAccountID,
		Type:           "transfer",
		Amount:         req.Amount,
		Description:    req.Description,
		Date:           req.Date,
		TransferPairID: &pairID,
		Tags:           []string{},
		AICategorized:  false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := uc.repo.CreateTransfer(ctx, debit, credit); err != nil {
		return nil, fmt.Errorf("transactionUseCase.CreateTransfer: %w", err)
	}
	return []*entity.Transaction{debit, credit}, nil
}
