package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transactionRepository struct {
	db *pgxpool.Pool
}

// NewTransactionRepository creates a new PostgreSQL-backed TransactionRepository.
func NewTransactionRepository(db *pgxpool.Pool) domainrepo.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx *entity.Transaction) error {
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transactionRepository.Create begin: %w", err)
	}
	defer dbTx.Rollback(ctx) //nolint:errcheck

	query := `
		INSERT INTO transactions (
			id, user_id, account_id, category_id, type, amount,
			description, notes, date, transfer_pair_id, recurrence_id,
			import_id, tags, ai_categorized, ai_confidence, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17
		)`
	_, err = dbTx.Exec(ctx, query,
		tx.ID, tx.UserID, tx.AccountID, tx.CategoryID, tx.Type, tx.Amount,
		tx.Description, tx.Notes, tx.Date, tx.TransferPairID, tx.RecurrenceID,
		tx.ImportID, tx.Tags, tx.AICategorized, tx.AIConfidence, tx.CreatedAt, tx.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("transactionRepository.Create insert: %w", err)
	}

	// Update account balance
	var delta float64
	if tx.Type == "income" {
		delta = tx.Amount
	} else if tx.Type == "expense" {
		delta = -tx.Amount
	}
	if delta != 0 {
		_, err = dbTx.Exec(ctx,
			`UPDATE accounts SET balance = balance + $1, updated_at = $2 WHERE id = $3`,
			delta, time.Now(), tx.AccountID,
		)
		if err != nil {
			return fmt.Errorf("transactionRepository.Create update balance: %w", err)
		}
	}

	if err := dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("transactionRepository.Create commit: %w", err)
	}
	return nil
}

func (r *transactionRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Transaction, error) {
	query := `
		SELECT t.id, t.user_id, t.account_id, t.category_id, t.type, t.amount,
			t.description, t.notes, t.date, t.transfer_pair_id, t.recurrence_id,
			t.import_id, t.tags, t.ai_categorized, t.ai_confidence, t.created_at, t.updated_at,
			a.name AS account_name, c.name AS category_name, c.color AS category_color, c.icon AS category_icon
		FROM transactions t
		LEFT JOIN accounts a ON a.id = t.account_id
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.id = $1 AND t.user_id = $2`
	row := r.db.QueryRow(ctx, query, id, userID)
	tx, err := scanTransaction(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("transactionRepository.FindByID: %w", err)
	}
	return tx, nil
}

func (r *transactionRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter domainrepo.TransactionFilter) ([]*entity.Transaction, int, error) {
	args := []interface{}{userID}
	argIdx := 2
	conditions := []string{"t.user_id = $1"}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("t.date >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("t.date <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}
	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("t.category_id = $%d", argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}
	if filter.AccountID != nil {
		conditions = append(conditions, fmt.Sprintf("t.account_id = $%d", argIdx))
		args = append(args, *filter.AccountID)
		argIdx++
	}
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("t.type = $%d", argIdx))
		args = append(args, *filter.Type)
		argIdx++
	}
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("t.description ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM transactions t WHERE %s`, where)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("transactionRepository.FindByUserID count: %w", err)
	}

	// Pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	dataArgs := append(args, pageSize, offset)
	dataQuery := fmt.Sprintf(`
		SELECT t.id, t.user_id, t.account_id, t.category_id, t.type, t.amount,
			t.description, t.notes, t.date, t.transfer_pair_id, t.recurrence_id,
			t.import_id, t.tags, t.ai_categorized, t.ai_confidence, t.created_at, t.updated_at,
			a.name AS account_name, c.name AS category_name, c.color AS category_color, c.icon AS category_icon
		FROM transactions t
		LEFT JOIN accounts a ON a.id = t.account_id
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE %s
		ORDER BY t.date DESC, t.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("transactionRepository.FindByUserID query: %w", err)
	}
	defer rows.Close()

	var transactions []*entity.Transaction
	for rows.Next() {
		tx, err := scanTransaction(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("transactionRepository.FindByUserID scan: %w", err)
		}
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("transactionRepository.FindByUserID rows: %w", err)
	}
	if transactions == nil {
		transactions = []*entity.Transaction{}
	}
	return transactions, total, nil
}

func (r *transactionRepository) Update(ctx context.Context, tx *entity.Transaction) error {
	query := `
		UPDATE transactions SET
			category_id = $3, description = $4, notes = $5,
			date = $6, tags = $7, updated_at = $8
		WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query,
		tx.ID, tx.UserID, tx.CategoryID, tx.Description,
		tx.Notes, tx.Date, tx.Tags, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("transactionRepository.Update: %w", err)
	}
	return nil
}

func (r *transactionRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	// First find the transaction to reverse the balance
	tx, err := r.FindByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("transactionRepository.Delete find: %w", err)
	}
	if tx == nil {
		return nil
	}

	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transactionRepository.Delete begin: %w", err)
	}
	defer dbTx.Rollback(ctx) //nolint:errcheck

	_, err = dbTx.Exec(ctx, `DELETE FROM transactions WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("transactionRepository.Delete delete: %w", err)
	}

	// Reverse balance impact
	var delta float64
	if tx.Type == "income" {
		delta = -tx.Amount
	} else if tx.Type == "expense" {
		delta = tx.Amount
	}
	if delta != 0 {
		_, err = dbTx.Exec(ctx,
			`UPDATE accounts SET balance = balance + $1, updated_at = $2 WHERE id = $3`,
			delta, time.Now(), tx.AccountID,
		)
		if err != nil {
			return fmt.Errorf("transactionRepository.Delete reverse balance: %w", err)
		}
	}

	if err := dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("transactionRepository.Delete commit: %w", err)
	}
	return nil
}

func (r *transactionRepository) GetSummary(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*domainrepo.TransactionSummary, error) {
	summary := &domainrepo.TransactionSummary{}

	// Current period income/expense
	periodQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
		FROM transactions
		WHERE user_id = $1 AND date >= $2 AND date <= $3 AND type != 'transfer'`
	if err := r.db.QueryRow(ctx, periodQuery, userID, startDate, endDate).
		Scan(&summary.TotalIncome, &summary.TotalExpense); err != nil {
		return nil, fmt.Errorf("transactionRepository.GetSummary period: %w", err)
	}
	summary.Balance = summary.TotalIncome - summary.TotalExpense

	// Previous period (same duration before startDate)
	duration := endDate.Sub(startDate)
	prevEnd := startDate.Add(-time.Second)
	prevStart := prevEnd.Add(-duration)
	prevQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
		FROM transactions
		WHERE user_id = $1 AND date >= $2 AND date <= $3 AND type != 'transfer'`
	if err := r.db.QueryRow(ctx, prevQuery, userID, prevStart, prevEnd).
		Scan(&summary.PrevIncome, &summary.PrevExpense); err != nil {
		return nil, fmt.Errorf("transactionRepository.GetSummary prev: %w", err)
	}

	// By category (expenses only)
	catQuery := `
		SELECT c.id, COALESCE(c.name, 'Sem categoria') AS category_name,
			SUM(t.amount) AS total, COUNT(*) AS count, c.color
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1 AND t.date >= $2 AND t.date <= $3 AND t.type = 'expense'
		GROUP BY c.id, c.name, c.color
		ORDER BY total DESC`
	rows, err := r.db.Query(ctx, catQuery, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("transactionRepository.GetSummary by_category: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cs domainrepo.CategorySummary
		if err := rows.Scan(&cs.CategoryID, &cs.CategoryName, &cs.Total, &cs.Count, &cs.Color); err != nil {
			return nil, fmt.Errorf("transactionRepository.GetSummary scan cat: %w", err)
		}
		summary.ByCategory = append(summary.ByCategory, cs)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("transactionRepository.GetSummary rows: %w", err)
	}
	if summary.ByCategory == nil {
		summary.ByCategory = []domainrepo.CategorySummary{}
	}
	return summary, nil
}

func (r *transactionRepository) CreateTransfer(ctx context.Context, debit, credit *entity.Transaction) error {
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transactionRepository.CreateTransfer begin: %w", err)
	}
	defer dbTx.Rollback(ctx) //nolint:errcheck

	insertQuery := `
		INSERT INTO transactions (
			id, user_id, account_id, category_id, type, amount,
			description, notes, date, transfer_pair_id, recurrence_id,
			import_id, tags, ai_categorized, ai_confidence, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17
		)`

	for _, tx := range []*entity.Transaction{debit, credit} {
		_, err := dbTx.Exec(ctx, insertQuery,
			tx.ID, tx.UserID, tx.AccountID, tx.CategoryID, tx.Type, tx.Amount,
			tx.Description, tx.Notes, tx.Date, tx.TransferPairID, tx.RecurrenceID,
			tx.ImportID, tx.Tags, tx.AICategorized, tx.AIConfidence, tx.CreatedAt, tx.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("transactionRepository.CreateTransfer insert: %w", err)
		}
	}

	// Debit: subtract from source account, Credit: add to destination account
	_, err = dbTx.Exec(ctx,
		`UPDATE accounts SET balance = balance - $1, updated_at = $2 WHERE id = $3`,
		debit.Amount, time.Now(), debit.AccountID,
	)
	if err != nil {
		return fmt.Errorf("transactionRepository.CreateTransfer debit balance: %w", err)
	}
	_, err = dbTx.Exec(ctx,
		`UPDATE accounts SET balance = balance + $1, updated_at = $2 WHERE id = $3`,
		credit.Amount, time.Now(), credit.AccountID,
	)
	if err != nil {
		return fmt.Errorf("transactionRepository.CreateTransfer credit balance: %w", err)
	}

	if err := dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("transactionRepository.CreateTransfer commit: %w", err)
	}
	return nil
}

func (r *transactionRepository) UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, delta float64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE accounts SET balance = balance + $1, updated_at = $2 WHERE id = $3`,
		delta, time.Now(), accountID,
	)
	if err != nil {
		return fmt.Errorf("transactionRepository.UpdateAccountBalance: %w", err)
	}
	return nil
}

// scanTransaction scans a pgx.Row or pgx.Rows into a Transaction entity.
func scanTransaction(row pgx.Row) (*entity.Transaction, error) {
	t := &entity.Transaction{}
	err := row.Scan(
		&t.ID, &t.UserID, &t.AccountID, &t.CategoryID, &t.Type, &t.Amount,
		&t.Description, &t.Notes, &t.Date, &t.TransferPairID, &t.RecurrenceID,
		&t.ImportID, &t.Tags, &t.AICategorized, &t.AIConfidence, &t.CreatedAt, &t.UpdatedAt,
		&t.AccountName, &t.CategoryName, &t.CategoryColor, &t.CategoryIcon,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}
