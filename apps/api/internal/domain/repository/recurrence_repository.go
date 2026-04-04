package repository

import (
	"context"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

// RecurrenceRepository defines data access operations for recurrences.
type RecurrenceRepository interface {
	Create(ctx context.Context, r *entity.Recurrence) error
	FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Recurrence, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Recurrence, error)
	FindDue(ctx context.Context, before time.Time) ([]*entity.Recurrence, error)
	Update(ctx context.Context, r *entity.Recurrence) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	UpdateNextDueDate(ctx context.Context, id uuid.UUID, nextDate time.Time) error
}
