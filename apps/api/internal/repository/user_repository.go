package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL-backed UserRepository.
func NewUserRepository(db *pgxpool.Pool) domainrepo.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, name, avatar_url, plan, plan_expires_at,
			email_verified, email_verification_token, timezone, currency, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`
	_, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name, user.AvatarURL,
		user.Plan, user.PlanExpiresAt, user.EmailVerified, user.EmailVerificationToken,
		user.Timezone, user.Currency, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("userRepository.Create: %w", err)
	}
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, plan, plan_expires_at,
		       email_verified, email_verification_token, password_reset_token,
		       password_reset_expires_at, timezone, currency, last_login_at,
		       created_at, updated_at
		FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.FindByEmail: %w", err)
	}
	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, plan, plan_expires_at,
		       email_verified, email_verification_token, password_reset_token,
		       password_reset_expires_at, timezone, currency, last_login_at,
		       created_at, updated_at
		FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.FindByID: %w", err)
	}
	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users SET
			email = $2, password_hash = $3, name = $4, avatar_url = $5,
			plan = $6, plan_expires_at = $7, email_verified = $8,
			email_verification_token = $9, password_reset_token = $10,
			password_reset_expires_at = $11, timezone = $12, currency = $13,
			updated_at = $14
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name, user.AvatarURL,
		user.Plan, user.PlanExpiresAt, user.EmailVerified, user.EmailVerificationToken,
		user.PasswordResetToken, user.PasswordResetExpiresAt, user.Timezone, user.Currency,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("userRepository.Update: %w", err)
	}
	return nil
}

func (r *userRepository) FindByVerificationToken(ctx context.Context, token string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, plan, plan_expires_at,
		       email_verified, email_verification_token, password_reset_token,
		       password_reset_expires_at, timezone, currency, last_login_at,
		       created_at, updated_at
		FROM users WHERE email_verification_token = $1`
	row := r.db.QueryRow(ctx, query, token)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.FindByVerificationToken: %w", err)
	}
	return user, nil
}

func (r *userRepository) FindByResetToken(ctx context.Context, token string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, plan, plan_expires_at,
		       email_verified, email_verification_token, password_reset_token,
		       password_reset_expires_at, timezone, currency, last_login_at,
		       created_at, updated_at
		FROM users WHERE password_reset_token = $1`
	row := r.db.QueryRow(ctx, query, token)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.FindByResetToken: %w", err)
	}
	return user, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(ctx, `UPDATE users SET last_login_at = $2, updated_at = $2 WHERE id = $1`, id, now)
	if err != nil {
		return fmt.Errorf("userRepository.UpdateLastLogin: %w", err)
	}
	return nil
}

func (r *userRepository) CreateRefreshToken(ctx context.Context, rt *entity.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query,
		rt.ID, rt.UserID, rt.TokenHash, rt.ExpiresAt, rt.Revoked, rt.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("userRepository.CreateRefreshToken: %w", err)
	}
	return nil
}

func (r *userRepository) FindRefreshToken(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	query := `SELECT id, user_id, token_hash, expires_at, revoked, created_at FROM refresh_tokens WHERE token_hash = $1`
	row := r.db.QueryRow(ctx, query, tokenHash)
	rt := &entity.RefreshToken{}
	err := row.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.FindRefreshToken: %w", err)
	}
	return rt, nil
}

func (r *userRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("userRepository.RevokeRefreshToken: %w", err)
	}
	return nil
}

func (r *userRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("userRepository.RevokeAllUserRefreshTokens: %w", err)
	}
	return nil
}

// scanUser scans a pgx.Row into a User entity.
func scanUser(row pgx.Row) (*entity.User, error) {
	u := &entity.User{}
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.AvatarURL,
		&u.Plan, &u.PlanExpiresAt, &u.EmailVerified, &u.EmailVerificationToken,
		&u.PasswordResetToken, &u.PasswordResetExpiresAt, &u.Timezone, &u.Currency,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}
