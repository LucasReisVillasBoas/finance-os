package repository

import (
	"context"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

// NotificationRepository defines data access operations for notifications.
type NotificationRepository interface {
	Create(ctx context.Context, n *entity.Notification) error
	FindByUserID(ctx context.Context, userID uuid.UUID, onlyUnread bool) ([]*entity.Notification, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	DeleteAll(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
}
