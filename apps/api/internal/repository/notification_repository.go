package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type notificationRepository struct {
	db *pgxpool.Pool
}

// NewNotificationRepository creates a new PostgreSQL-backed NotificationRepository.
func NewNotificationRepository(db *pgxpool.Pool) domainrepo.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *entity.Notification) error {
	dataJSON, err := json.Marshal(n.Data)
	if err != nil {
		dataJSON = []byte("{}")
	}
	query := `
		INSERT INTO notifications (id, user_id, type, title, message, data, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = r.db.Exec(ctx, query,
		n.ID, n.UserID, n.Type, n.Title, n.Message, dataJSON, n.IsRead, n.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("notificationRepository.Create: %w", err)
	}
	return nil
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, onlyUnread bool) ([]*entity.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, is_read, created_at
		FROM notifications
		WHERE user_id = $1`
	args := []interface{}{userID}
	if onlyUnread {
		query += " AND is_read = FALSE"
	}
	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("notificationRepository.FindByUserID: %w", err)
	}
	defer rows.Close()

	var notifications []*entity.Notification
	for rows.Next() {
		n := &entity.Notification{}
		var dataJSON []byte
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &dataJSON, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("notificationRepository.FindByUserID scan: %w", err)
		}
		if len(dataJSON) > 0 {
			_ = json.Unmarshal(dataJSON, &n.Data)
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("notificationRepository.FindByUserID rows: %w", err)
	}
	if notifications == nil {
		notifications = []*entity.Notification{}
	}
	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("notificationRepository.MarkAsRead: %w", err)
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("notificationRepository.MarkAllAsRead: %w", err)
	}
	return nil
}

func (r *notificationRepository) DeleteAll(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM notifications WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("notificationRepository.DeleteAll: %w", err)
	}
	return nil
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("notificationRepository.CountUnread: %w", err)
	}
	return count, nil
}
