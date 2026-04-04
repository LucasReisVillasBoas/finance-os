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

// Sentinel errors
var (
	ErrRecurrenceNotFound = errors.New("recurrence not found")
)

// CreateRecurrenceRequest holds data needed to create a recurrence.
type CreateRecurrenceRequest struct {
	AccountID   uuid.UUID  `json:"account_id" binding:"required"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Type        string     `json:"type" binding:"required,oneof=income expense"`
	Amount      float64    `json:"amount" binding:"required,gt=0"`
	Description *string    `json:"description"`
	Frequency   string     `json:"frequency" binding:"required,oneof=daily weekly biweekly monthly yearly"`
	StartDate   time.Time  `json:"start_date" binding:"required"`
	EndDate     *time.Time `json:"end_date"`
	AutoLaunch  bool       `json:"auto_launch"`
}

// UpdateRecurrenceRequest holds fields that can be updated on a recurrence.
type UpdateRecurrenceRequest struct {
	AccountID   *uuid.UUID `json:"account_id"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Type        *string    `json:"type" binding:"omitempty,oneof=income expense"`
	Amount      *float64   `json:"amount" binding:"omitempty,gt=0"`
	Description *string    `json:"description"`
	Frequency   *string    `json:"frequency" binding:"omitempty,oneof=daily weekly biweekly monthly yearly"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	AutoLaunch  *bool      `json:"auto_launch"`
	IsActive    *bool      `json:"is_active"`
}

// RecurrenceUseCase defines business logic for recurrences.
type RecurrenceUseCase interface {
	Create(ctx context.Context, userID uuid.UUID, req CreateRecurrenceRequest) (*entity.Recurrence, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Recurrence, error)
	List(ctx context.Context, userID uuid.UUID) ([]*entity.Recurrence, error)
	Update(ctx context.Context, id, userID uuid.UUID, req UpdateRecurrenceRequest) (*entity.Recurrence, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type recurrenceUseCase struct {
	repo domainrepo.RecurrenceRepository
}

// NewRecurrenceUseCase creates a new RecurrenceUseCase implementation.
func NewRecurrenceUseCase(repo domainrepo.RecurrenceRepository) RecurrenceUseCase {
	return &recurrenceUseCase{repo: repo}
}

func (uc *recurrenceUseCase) Create(ctx context.Context, userID uuid.UUID, req CreateRecurrenceRequest) (*entity.Recurrence, error) {
	now := time.Now()
	rec := &entity.Recurrence{
		ID:          uuid.New(),
		UserID:      userID,
		AccountID:   req.AccountID,
		CategoryID:  req.CategoryID,
		Type:        req.Type,
		Amount:      req.Amount,
		Description: req.Description,
		Frequency:   req.Frequency,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		NextDueDate: req.StartDate,
		AutoLaunch:  req.AutoLaunch,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.repo.Create(ctx, rec); err != nil {
		return nil, fmt.Errorf("recurrenceUseCase.Create: %w", err)
	}
	return rec, nil
}

func (uc *recurrenceUseCase) GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Recurrence, error) {
	rec, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("recurrenceUseCase.GetByID: %w", err)
	}
	if rec == nil {
		return nil, ErrRecurrenceNotFound
	}
	return rec, nil
}

func (uc *recurrenceUseCase) List(ctx context.Context, userID uuid.UUID) ([]*entity.Recurrence, error) {
	recs, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("recurrenceUseCase.List: %w", err)
	}
	return recs, nil
}

func (uc *recurrenceUseCase) Update(ctx context.Context, id, userID uuid.UUID, req UpdateRecurrenceRequest) (*entity.Recurrence, error) {
	rec, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("recurrenceUseCase.Update find: %w", err)
	}
	if rec == nil {
		return nil, ErrRecurrenceNotFound
	}

	if req.AccountID != nil {
		rec.AccountID = *req.AccountID
	}
	if req.CategoryID != nil {
		rec.CategoryID = req.CategoryID
	}
	if req.Type != nil {
		rec.Type = *req.Type
	}
	if req.Amount != nil {
		rec.Amount = *req.Amount
	}
	if req.Description != nil {
		rec.Description = req.Description
	}
	if req.Frequency != nil {
		rec.Frequency = *req.Frequency
	}
	if req.StartDate != nil {
		rec.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		rec.EndDate = req.EndDate
	}
	if req.AutoLaunch != nil {
		rec.AutoLaunch = *req.AutoLaunch
	}
	if req.IsActive != nil {
		rec.IsActive = *req.IsActive
	}

	if err := uc.repo.Update(ctx, rec); err != nil {
		return nil, fmt.Errorf("recurrenceUseCase.Update save: %w", err)
	}
	return rec, nil
}

func (uc *recurrenceUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	rec, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("recurrenceUseCase.Delete find: %w", err)
	}
	if rec == nil {
		return ErrRecurrenceNotFound
	}
	if err := uc.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("recurrenceUseCase.Delete: %w", err)
	}
	return nil
}

// CalculateNextDueDate calculates the next due date from current based on frequency.
// This is an exported alias used by the worker package.
func CalculateNextDueDate(current time.Time, frequency string) time.Time {
	return NextDueDate(current, frequency)
}

// NextDueDate calculates the next due date from current based on frequency.
func NextDueDate(current time.Time, frequency string) time.Time {
	switch frequency {
	case "daily":
		return current.AddDate(0, 0, 1)
	case "weekly":
		return current.AddDate(0, 0, 7)
	case "biweekly":
		return current.AddDate(0, 0, 14)
	case "monthly":
		return current.AddDate(0, 1, 0)
	case "yearly":
		return current.AddDate(1, 0, 0)
	default:
		return current.AddDate(0, 1, 0)
	}
}
