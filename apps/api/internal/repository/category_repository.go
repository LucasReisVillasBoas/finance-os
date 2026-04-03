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

type categoryRepository struct {
	db *pgxpool.Pool
}

// NewCategoryRepository creates a new PostgreSQL-backed CategoryRepository.
func NewCategoryRepository(db *pgxpool.Pool) domainrepo.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	query := `
		INSERT INTO categories (id, user_id, name, type, icon, color, is_system, parent_id, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(ctx, query,
		category.ID, category.UserID, category.Name, category.Type,
		category.Icon, category.Color, category.IsSystem, category.ParentID,
		category.IsActive, category.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("categoryRepository.Create: %w", err)
	}
	return nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	query := `
		SELECT id, user_id, name, type, icon, color, is_system, parent_id, is_active, created_at
		FROM categories
		WHERE id = $1 AND is_active = true`
	row := r.db.QueryRow(ctx, query, id)
	category, err := scanCategory(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("categoryRepository.FindByID: %w", err)
	}
	return category, nil
}

// FindByUserID returns system categories (user_id IS NULL) plus categories belonging to userID.
func (r *categoryRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Category, error) {
	query := `
		SELECT id, user_id, name, type, icon, color, is_system, parent_id, is_active, created_at
		FROM categories
		WHERE (user_id IS NULL OR user_id = $1) AND is_active = true
		ORDER BY is_system DESC, name ASC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("categoryRepository.FindByUserID: %w", err)
	}
	defer rows.Close()

	var categories []*entity.Category
	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, fmt.Errorf("categoryRepository.FindByUserID scan: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("categoryRepository.FindByUserID rows: %w", err)
	}
	if categories == nil {
		categories = []*entity.Category{}
	}
	return categories, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *entity.Category) error {
	query := `
		UPDATE categories SET name = $2, icon = $3, color = $4
		WHERE id = $1 AND user_id IS NOT NULL`
	_, err := r.db.Exec(ctx, query, category.ID, category.Name, category.Icon, category.Color)
	if err != nil {
		return fmt.Errorf("categoryRepository.Update: %w", err)
	}
	return nil
}

func (r *categoryRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE categories SET is_active = false WHERE id = $1 AND user_id = $2 AND is_system = false`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("categoryRepository.SoftDelete: %w", err)
	}

	// Update the timestamp if the table has an updated_at field — for now categories only have created_at
	_ = time.Now() // keep import used if needed in future
	return nil
}

// scanCategory scans a pgx.Row or pgx.Rows into a Category entity.
func scanCategory(row pgx.Row) (*entity.Category, error) {
	c := &entity.Category{}
	err := row.Scan(
		&c.ID, &c.UserID, &c.Name, &c.Type,
		&c.Icon, &c.Color, &c.IsSystem, &c.ParentID,
		&c.IsActive, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}
