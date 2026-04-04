package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
)

// NotificationUseCase defines business logic for notifications.
type NotificationUseCase interface {
	Create(ctx context.Context, userID uuid.UUID, notifType, title, message string, data map[string]interface{}) (*entity.Notification, error)
	List(ctx context.Context, userID uuid.UUID, onlyUnread bool) ([]*entity.Notification, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	DeleteAll(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
}

type notificationUseCase struct {
	repo domainrepo.NotificationRepository
}

// NewNotificationUseCase creates a new NotificationUseCase.
func NewNotificationUseCase(repo domainrepo.NotificationRepository) NotificationUseCase {
	return &notificationUseCase{repo: repo}
}

func (uc *notificationUseCase) Create(ctx context.Context, userID uuid.UUID, notifType, title, message string, data map[string]interface{}) (*entity.Notification, error) {
	var msgPtr *string
	if message != "" {
		msgPtr = &message
	}
	n := &entity.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Message:   msgPtr,
		Data:      data,
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, n); err != nil {
		return nil, fmt.Errorf("notificationUseCase.Create: %w", err)
	}
	return n, nil
}

func (uc *notificationUseCase) List(ctx context.Context, userID uuid.UUID, onlyUnread bool) ([]*entity.Notification, error) {
	notifications, err := uc.repo.FindByUserID(ctx, userID, onlyUnread)
	if err != nil {
		return nil, fmt.Errorf("notificationUseCase.List: %w", err)
	}
	return notifications, nil
}

func (uc *notificationUseCase) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	if err := uc.repo.MarkAsRead(ctx, id, userID); err != nil {
		return fmt.Errorf("notificationUseCase.MarkAsRead: %w", err)
	}
	return nil
}

func (uc *notificationUseCase) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if err := uc.repo.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("notificationUseCase.MarkAllAsRead: %w", err)
	}
	return nil
}

func (uc *notificationUseCase) DeleteAll(ctx context.Context, userID uuid.UUID) error {
	if err := uc.repo.DeleteAll(ctx, userID); err != nil {
		return fmt.Errorf("notificationUseCase.DeleteAll: %w", err)
	}
	return nil
}

func (uc *notificationUseCase) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := uc.repo.CountUnread(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("notificationUseCase.CountUnread: %w", err)
	}
	return count, nil
}
