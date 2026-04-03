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

type accountRepository struct {
	db *pgxpool.Pool
}

// NewAccountRepository creates a new PostgreSQL-backed AccountRepository.
func NewAccountRepository(db *pgxpool.Pool) domainrepo.AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(ctx context.Context, account *entity.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, name, type, institution, balance, credit_limit, color, icon, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Exec(ctx, query,
		account.ID, account.UserID, account.Name, account.Type, account.Institution,
		account.Balance, account.CreditLimit, account.Color, account.Icon,
		account.IsActive, account.CreatedAt, account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("accountRepository.Create: %w", err)
	}
	return nil
}

func (r *accountRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Account, error) {
	query := `
		SELECT id, user_id, name, type, institution, balance, credit_limit, color, icon, is_active, created_at, updated_at
		FROM accounts
		WHERE id = $1 AND user_id = $2 AND is_active = true`
	row := r.db.QueryRow(ctx, query, id, userID)
	account, err := scanAccount(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("accountRepository.FindByID: %w", err)
	}
	return account, nil
}

func (r *accountRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Account, error) {
	query := `
		SELECT id, user_id, name, type, institution, balance, credit_limit, color, icon, is_active, created_at, updated_at
		FROM accounts
		WHERE user_id = $1 AND is_active = true
		ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("accountRepository.FindByUserID: %w", err)
	}
	defer rows.Close()

	var accounts []*entity.Account
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			return nil, fmt.Errorf("accountRepository.FindByUserID scan: %w", err)
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("accountRepository.FindByUserID rows: %w", err)
	}
	return accounts, nil
}

func (r *accountRepository) Update(ctx context.Context, account *entity.Account) error {
	query := `
		UPDATE accounts SET
			name = $3, institution = $4, credit_limit = $5, color = $6, icon = $7, updated_at = $8
		WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query,
		account.ID, account.UserID, account.Name, account.Institution,
		account.CreditLimit, account.Color, account.Icon, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("accountRepository.Update: %w", err)
	}
	return nil
}

func (r *accountRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE accounts SET is_active = false, updated_at = $3 WHERE id = $1 AND user_id = $2`,
		id, userID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("accountRepository.SoftDelete: %w", err)
	}
	return nil
}

func (r *accountRepository) GetSummary(ctx context.Context, userID uuid.UUID) (*domainrepo.AccountSummary, error) {
	query := `
		SELECT id, name, balance, type
		FROM accounts
		WHERE user_id = $1 AND is_active = true`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("accountRepository.GetSummary: %w", err)
	}
	defer rows.Close()

	summary := &domainrepo.AccountSummary{}
	for rows.Next() {
		var id uuid.UUID
		var name string
		var balance float64
		var accountType string
		if err := rows.Scan(&id, &name, &balance, &accountType); err != nil {
			return nil, fmt.Errorf("accountRepository.GetSummary scan: %w", err)
		}
		summary.TotalBalance += balance
		if accountType != "credit_card" {
			summary.NetBalance += balance
		}
		if accountType == "investment" || accountType == "savings" || accountType == "checking" || accountType == "wallet" {
			summary.TotalPatrimony += balance
		}
		summary.AccountBalances = append(summary.AccountBalances, domainrepo.AccountBalance{
			AccountID:   id,
			AccountName: name,
			Balance:     balance,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("accountRepository.GetSummary rows: %w", err)
	}
	if summary.AccountBalances == nil {
		summary.AccountBalances = []domainrepo.AccountBalance{}
	}
	return summary, nil
}

// scanAccount scans a pgx.Row or pgx.Rows into an Account entity.
func scanAccount(row pgx.Row) (*entity.Account, error) {
	a := &entity.Account{}
	err := row.Scan(
		&a.ID, &a.UserID, &a.Name, &a.Type, &a.Institution,
		&a.Balance, &a.CreditLimit, &a.Color, &a.Icon,
		&a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}
