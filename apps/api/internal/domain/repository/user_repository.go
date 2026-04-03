package repository

import (
	"context"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

// UserRepository defines persistence operations for users and refresh tokens.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	FindByVerificationToken(ctx context.Context, token string) (*entity.User, error)
	FindByResetToken(ctx context.Context, token string) (*entity.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	CreateRefreshToken(ctx context.Context, rt *entity.RefreshToken) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
}
