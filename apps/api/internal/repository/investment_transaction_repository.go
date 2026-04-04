package repository

import (
	"context"
	"fmt"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type investmentTransactionRepository struct {
	db *pgxpool.Pool
}

// NewInvestmentTransactionRepository creates a new PostgreSQL-backed InvestmentTransactionRepository.
func NewInvestmentTransactionRepository(db *pgxpool.Pool) domainrepo.InvestmentTransactionRepository {
	return &investmentTransactionRepository{db: db}
}

func (r *investmentTransactionRepository) Create(ctx context.Context, t *entity.InvestmentTransaction) error {
	query := `
		INSERT INTO investment_transactions (id, holding_id, type, quantity, price, fees, total, date, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(ctx, query,
		t.ID, t.HoldingID, t.Type, t.Quantity, t.Price, t.Fees, t.Total, t.Date, t.Notes, t.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("investmentTransactionRepository.Create: %w", err)
	}
	return nil
}

func (r *investmentTransactionRepository) FindByHoldingID(ctx context.Context, holdingID uuid.UUID) ([]*entity.InvestmentTransaction, error) {
	query := `
		SELECT id, holding_id, type, quantity, price, fees, total, date, notes, created_at
		FROM investment_transactions
		WHERE holding_id = $1
		ORDER BY date DESC`
	rows, err := r.db.Query(ctx, query, holdingID)
	if err != nil {
		return nil, fmt.Errorf("investmentTransactionRepository.FindByHoldingID: %w", err)
	}
	defer rows.Close()

	var txs []*entity.InvestmentTransaction
	for rows.Next() {
		t, err := scanInvestmentTransaction(rows)
		if err != nil {
			return nil, fmt.Errorf("investmentTransactionRepository.FindByHoldingID scan: %w", err)
		}
		txs = append(txs, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("investmentTransactionRepository.FindByHoldingID rows: %w", err)
	}
	return txs, nil
}

func (r *investmentTransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM investment_transactions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("investmentTransactionRepository.Delete: %w", err)
	}
	return nil
}

func scanInvestmentTransaction(row pgx.Row) (*entity.InvestmentTransaction, error) {
	t := &entity.InvestmentTransaction{}
	err := row.Scan(
		&t.ID, &t.HoldingID, &t.Type, &t.Quantity, &t.Price, &t.Fees, &t.Total, &t.Date, &t.Notes, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}
