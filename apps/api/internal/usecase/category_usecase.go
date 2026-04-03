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

// Sentinel errors for categories
var (
	ErrCategoryNotFound         = errors.New("category not found")
	ErrCannotModifySystemCategory = errors.New("cannot modify a system category")
)

// CreateCategoryRequest holds data needed to create a user category.
type CreateCategoryRequest struct {
	Name     string     `json:"name" binding:"required,min=2,max=255"`
	Type     string     `json:"type" binding:"required,oneof=income expense transfer"`
	Icon     *string    `json:"icon"`
	Color    *string    `json:"color"`
	ParentID *uuid.UUID `json:"parent_id"`
}

// UpdateCategoryRequest holds fields that can be updated on a category.
type UpdateCategoryRequest struct {
	Name  *string `json:"name"`
	Icon  *string `json:"icon"`
	Color *string `json:"color"`
}

// CategoryUseCase defines the business logic interface for categories.
type CategoryUseCase interface {
	Create(ctx context.Context, userID uuid.UUID, req CreateCategoryRequest) (*entity.Category, error)
	GetAll(ctx context.Context, userID uuid.UUID) ([]*entity.Category, error)
	Update(ctx context.Context, id, userID uuid.UUID, req UpdateCategoryRequest) (*entity.Category, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type categoryUseCase struct {
	repo domainrepo.CategoryRepository
}

// NewCategoryUseCase creates a new CategoryUseCase implementation.
func NewCategoryUseCase(repo domainrepo.CategoryRepository) CategoryUseCase {
	return &categoryUseCase{repo: repo}
}

func (uc *categoryUseCase) Create(ctx context.Context, userID uuid.UUID, req CreateCategoryRequest) (*entity.Category, error) {
	category := &entity.Category{
		ID:        uuid.New(),
		UserID:    &userID,
		Name:      req.Name,
		Type:      req.Type,
		Icon:      req.Icon,
		Color:     req.Color,
		IsSystem:  false,
		ParentID:  req.ParentID,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err := uc.repo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("categoryUseCase.Create: %w", err)
	}
	return category, nil
}

func (uc *categoryUseCase) GetAll(ctx context.Context, userID uuid.UUID) ([]*entity.Category, error) {
	categories, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("categoryUseCase.GetAll: %w", err)
	}
	return categories, nil
}

func (uc *categoryUseCase) Update(ctx context.Context, id, userID uuid.UUID, req UpdateCategoryRequest) (*entity.Category, error) {
	category, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("categoryUseCase.Update find: %w", err)
	}
	if category == nil {
		return nil, ErrCategoryNotFound
	}
	if category.IsSystem {
		return nil, ErrCannotModifySystemCategory
	}
	// Ensure the category belongs to this user
	if category.UserID == nil || *category.UserID != userID {
		return nil, ErrCategoryNotFound
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Icon != nil {
		category.Icon = req.Icon
	}
	if req.Color != nil {
		category.Color = req.Color
	}

	if err := uc.repo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("categoryUseCase.Update save: %w", err)
	}
	return category, nil
}

func (uc *categoryUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	category, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("categoryUseCase.Delete find: %w", err)
	}
	if category == nil {
		return ErrCategoryNotFound
	}
	if category.IsSystem {
		return ErrCannotModifySystemCategory
	}
	if category.UserID == nil || *category.UserID != userID {
		return ErrCategoryNotFound
	}

	if err := uc.repo.SoftDelete(ctx, id, userID); err != nil {
		return fmt.Errorf("categoryUseCase.Delete soft delete: %w", err)
	}
	return nil
}
