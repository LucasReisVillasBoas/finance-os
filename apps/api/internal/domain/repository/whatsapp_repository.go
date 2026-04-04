package repository

import (
	"context"

	"github.com/financeos/api/internal/domain/entity"
)

// WhatsAppRepository defines data access operations for WhatsApp sessions.
type WhatsAppRepository interface {
	FindSessionByPhone(ctx context.Context, phone string) (*entity.WhatsAppSession, error)
	CreateSession(ctx context.Context, s *entity.WhatsAppSession) error
	UpdateSession(ctx context.Context, s *entity.WhatsAppSession) error
	FindUserByPhone(ctx context.Context, phone string) (*entity.User, error)
}
