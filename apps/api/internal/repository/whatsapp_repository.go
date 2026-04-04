package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type whatsAppRepository struct {
	db *pgxpool.Pool
}

// NewWhatsAppRepository creates a new PostgreSQL-backed WhatsAppRepository.
func NewWhatsAppRepository(db *pgxpool.Pool) domainrepo.WhatsAppRepository {
	return &whatsAppRepository{db: db}
}

func (r *whatsAppRepository) FindSessionByPhone(ctx context.Context, phone string) (*entity.WhatsAppSession, error) {
	query := `
		SELECT id, user_id, phone_number, state, session_data, last_activity, is_active, created_at
		FROM whatsapp_sessions
		WHERE phone_number = $1 AND is_active = true
		ORDER BY last_activity DESC
		LIMIT 1`
	row := r.db.QueryRow(ctx, query, phone)

	s := &entity.WhatsAppSession{}
	var sessionDataJSON []byte
	err := row.Scan(
		&s.ID, &s.UserID, &s.PhoneNumber, &s.State, &sessionDataJSON,
		&s.LastActivity, &s.IsActive, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("whatsAppRepository.FindSessionByPhone: %w", err)
	}

	if sessionDataJSON != nil {
		if err := json.Unmarshal(sessionDataJSON, &s.SessionData); err != nil {
			s.SessionData = make(map[string]interface{})
		}
	} else {
		s.SessionData = make(map[string]interface{})
	}

	return s, nil
}

func (r *whatsAppRepository) CreateSession(ctx context.Context, s *entity.WhatsAppSession) error {
	sessionDataJSON, err := json.Marshal(s.SessionData)
	if err != nil {
		return fmt.Errorf("whatsAppRepository.CreateSession marshal: %w", err)
	}

	query := `
		INSERT INTO whatsapp_sessions (id, user_id, phone_number, state, session_data, last_activity, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = r.db.Exec(ctx, query,
		s.ID, s.UserID, s.PhoneNumber, s.State, sessionDataJSON,
		s.LastActivity, s.IsActive, s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("whatsAppRepository.CreateSession: %w", err)
	}
	return nil
}

func (r *whatsAppRepository) UpdateSession(ctx context.Context, s *entity.WhatsAppSession) error {
	sessionDataJSON, err := json.Marshal(s.SessionData)
	if err != nil {
		return fmt.Errorf("whatsAppRepository.UpdateSession marshal: %w", err)
	}

	query := `
		UPDATE whatsapp_sessions SET
			state = $2, session_data = $3, last_activity = $4, is_active = $5
		WHERE id = $1`
	_, err = r.db.Exec(ctx, query,
		s.ID, s.State, sessionDataJSON, s.LastActivity, s.IsActive,
	)
	if err != nil {
		return fmt.Errorf("whatsAppRepository.UpdateSession: %w", err)
	}
	return nil
}

func (r *whatsAppRepository) FindUserByPhone(ctx context.Context, phone string) (*entity.User, error) {
	query := `
		SELECT u.id, u.email, u.name, u.password_hash, u.created_at, u.updated_at
		FROM users u
		JOIN whatsapp_sessions ws ON ws.user_id = u.id
		WHERE ws.phone_number = $1 AND ws.is_active = true
		LIMIT 1`
	row := r.db.QueryRow(ctx, query, phone)

	u := &entity.User{}
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("whatsAppRepository.FindUserByPhone: %w", err)
	}
	return u, nil
}
