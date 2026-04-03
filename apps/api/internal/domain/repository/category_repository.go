package repository

import (
	"context"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

// CategoryRepository defines data access operations for categories.
type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Category, error) // system + user's
	Update(ctx context.Context, category *entity.Category) error
	SoftDelete(ctx context.Context, id, userID uuid.UUID) error
}
