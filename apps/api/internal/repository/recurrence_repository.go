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

type recurrenceRepository struct {
	db *pgxpool.Pool
}

// NewRecurrenceRepository creates a new PostgreSQL-backed RecurrenceRepository.
func NewRecurrenceRepository(db *pgxpool.Pool) domainrepo.RecurrenceRepository {
	return &recurrenceRepository{db: db}
}

func (r *recurrenceRepository) Create(ctx context.Context, rec *entity.Recurrence) error {
	query := `
		INSERT INTO recurrences (
			id, user_id, account_id, category_id, type, amount,
			description, frequency, start_date, end_date, next_due_date,
			auto_launch, is_active, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
		)`
	_, err := r.db.Exec(ctx, query,
		rec.ID, rec.UserID, rec.AccountID, rec.CategoryID, rec.Type, rec.Amount,
		rec.Description, rec.Frequency, rec.StartDate, rec.EndDate, rec.NextDueDate,
		rec.AutoLaunch, rec.IsActive, rec.CreatedAt, rec.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("recurrenceRepository.Create: %w", err)
	}
	return nil
}

func (r *recurrenceRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Recurrence, error) {
	query := `
		SELECT r.id, r.user_id, r.account_id, r.category_id, r.type, r.amount,
			r.description, r.frequency, r.start_date, r.end_date, r.next_due_date,
			r.auto_launch, r.is_active, r.created_at, r.updated_at,
			a.name AS account_name, c.name AS category_name
		FROM recurrences r
		LEFT JOIN accounts a ON a.id = r.account_id
		LEFT JOIN categories c ON c.id = r.category_id
		WHERE r.id = $1 AND r.user_id = $2`
	row := r.db.QueryRow(ctx, query, id, userID)
	rec, err := scanRecurrence(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("recurrenceRepository.FindByID: %w", err)
	}
	return rec, nil
}

func (r *recurrenceRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Recurrence, error) {
	query := `
		SELECT r.id, r.user_id, r.account_id, r.category_id, r.type, r.amount,
			r.description, r.frequency, r.start_date, r.end_date, r.next_due_date,
			r.auto_launch, r.is_active, r.created_at, r.updated_at,
			a.name AS account_name, c.name AS category_name
		FROM recurrences r
		LEFT JOIN accounts a ON a.id = r.account_id
		LEFT JOIN categories c ON c.id = r.category_id
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("recurrenceRepository.FindByUserID query: %w", err)
	}
	defer rows.Close()

	var result []*entity.Recurrence
	for rows.Next() {
		rec, err := scanRecurrence(rows)
		if err != nil {
			return nil, fmt.Errorf("recurrenceRepository.FindByUserID scan: %w", err)
		}
		result = append(result, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recurrenceRepository.FindByUserID rows: %w", err)
	}
	if result == nil {
		result = []*entity.Recurrence{}
	}
	return result, nil
}

func (r *recurrenceRepository) FindDue(ctx context.Context, before time.Time) ([]*entity.Recurrence, error) {
	query := `
		SELECT r.id, r.user_id, r.account_id, r.category_id, r.type, r.amount,
			r.description, r.frequency, r.start_date, r.end_date, r.next_due_date,
			r.auto_launch, r.is_active, r.created_at, r.updated_at,
			a.name AS account_name, c.name AS category_name
		FROM recurrences r
		LEFT JOIN accounts a ON a.id = r.account_id
		LEFT JOIN categories c ON c.id = r.category_id
		WHERE r.is_active = true AND r.next_due_date <= $1
		ORDER BY r.next_due_date ASC`
	rows, err := r.db.Query(ctx, query, before)
	if err != nil {
		return nil, fmt.Errorf("recurrenceRepository.FindDue query: %w", err)
	}
	defer rows.Close()

	var result []*entity.Recurrence
	for rows.Next() {
		rec, err := scanRecurrence(rows)
		if err != nil {
			return nil, fmt.Errorf("recurrenceRepository.FindDue scan: %w", err)
		}
		result = append(result, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recurrenceRepository.FindDue rows: %w", err)
	}
	if result == nil {
		result = []*entity.Recurrence{}
	}
	return result, nil
}

func (r *recurrenceRepository) Update(ctx context.Context, rec *entity.Recurrence) error {
	query := `
		UPDATE recurrences SET
			account_id = $3, category_id = $4, type = $5, amount = $6,
			description = $7, frequency = $8, start_date = $9, end_date = $10,
			next_due_date = $11, auto_launch = $12, is_active = $13, updated_at = $14
		WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query,
		rec.ID, rec.UserID, rec.AccountID, rec.CategoryID, rec.Type, rec.Amount,
		rec.Description, rec.Frequency, rec.StartDate, rec.EndDate,
		rec.NextDueDate, rec.AutoLaunch, rec.IsActive, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("recurrenceRepository.Update: %w", err)
	}
	return nil
}

func (r *recurrenceRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM recurrences WHERE id = $1 AND user_id = $2`, id, userID,
	)
	if err != nil {
		return fmt.Errorf("recurrenceRepository.Delete: %w", err)
	}
	return nil
}

func (r *recurrenceRepository) UpdateNextDueDate(ctx context.Context, id uuid.UUID, nextDate time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE recurrences SET next_due_date = $2, updated_at = $3 WHERE id = $1`,
		id, nextDate, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("recurrenceRepository.UpdateNextDueDate: %w", err)
	}
	return nil
}

func scanRecurrence(row pgx.Row) (*entity.Recurrence, error) {
	r := &entity.Recurrence{}
	err := row.Scan(
		&r.ID, &r.UserID, &r.AccountID, &r.CategoryID, &r.Type, &r.Amount,
		&r.Description, &r.Frequency, &r.StartDate, &r.EndDate, &r.NextDueDate,
		&r.AutoLaunch, &r.IsActive, &r.CreatedAt, &r.UpdatedAt,
		&r.AccountName, &r.CategoryName,
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}
