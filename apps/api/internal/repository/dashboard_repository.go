package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dashboardRepository struct {
	db *pgxpool.Pool
}

// NewDashboardRepository creates a new PostgreSQL-backed DashboardRepository.
func NewDashboardRepository(db *pgxpool.Pool) domainrepo.DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) GetOverview(ctx context.Context, userID uuid.UUID, month, year int) (*domainrepo.DashboardOverview, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	overview := &domainrepo.DashboardOverview{}

	// Net balance: SUM of all active non-credit-card accounts
	if err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(balance), 0)
		FROM accounts
		WHERE user_id = $1 AND is_active = true AND type != 'credit_card'
	`, userID).Scan(&overview.NetBalance); err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetOverview net_balance: %w", err)
	}

	// Total patrimony: SUM of ALL active account balances
	if err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(balance), 0)
		FROM accounts
		WHERE user_id = $1 AND is_active = true
	`, userID).Scan(&overview.TotalPatrimony); err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetOverview total_patrimony: %w", err)
	}

	// Investment value: current market value of all holdings across user's portfolios
	if err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(h.current_value), 0)
		FROM holdings h
		JOIN portfolios p ON p.id = h.portfolio_id
		WHERE p.user_id = $1
	`, userID).Scan(&overview.InvestmentValue); err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetOverview investment_value: %w", err)
	}

	// Custom asset value: sum of active custom assets (real estate, vehicles, etc.)
	if err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(current_value), 0)
		FROM custom_assets
		WHERE user_id = $1 AND is_active = true
	`, userID).Scan(&overview.CustomAssetValue); err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetOverview custom_asset_value: %w", err)
	}

	overview.TotalNetWorth = overview.NetBalance + overview.InvestmentValue + overview.CustomAssetValue

	// Total income and expense for the month
	if err := r.db.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS total_expense
		FROM transactions
		WHERE user_id = $1 AND date >= $2 AND date <= $3 AND type != 'transfer'
	`, userID, startDate, endDate).Scan(&overview.TotalIncome, &overview.TotalExpense); err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetOverview income_expense: %w", err)
	}

	// Investment capacity: how much is left to invest this month
	if overview.TotalIncome > overview.TotalExpense {
		overview.InvestmentCapacity = overview.TotalIncome - overview.TotalExpense
	}
	if overview.TotalIncome > 0 {
		overview.InvestmentCapacityPct = (overview.InvestmentCapacity / overview.TotalIncome) * 100
	}

	// Top 5 expense categories for the month
	topRows, err := r.db.Query(ctx, `
		SELECT
			t.category_id,
			COALESCE(c.name, 'Sem categoria') AS category_name,
			SUM(t.amount) AS total,
			COUNT(*) AS cnt,
			c.color
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1 AND t.type = 'expense' AND t.date >= $2 AND t.date <= $3
		GROUP BY t.category_id, c.name, c.color
		ORDER BY total DESC
		LIMIT 5
	`, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetOverview top_categories query: %w", err)
	}
	defer topRows.Close()

	overview.TopCategories = []domainrepo.CategorySummary{}
	for topRows.Next() {
		var cs domainrepo.CategorySummary
		if err := topRows.Scan(&cs.CategoryID, &cs.CategoryName, &cs.Total, &cs.Count, &cs.Color); err != nil {
			return nil, fmt.Errorf("dashboardRepository.GetOverview top_categories scan: %w", err)
		}
		overview.TopCategories = append(overview.TopCategories, cs)
	}
	if err := topRows.Err(); err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetOverview top_categories rows: %w", err)
	}

	// Alert budgets: budgets where actual/planned >= threshold_pct/100
	alertRows, err := r.db.Query(ctx, `
		SELECT
			b.id AS budget_id,
			b.category_id,
			COALESCE(c.name, 'Geral') AS category_name,
			c.color AS category_color,
			c.icon AS category_icon,
			b.amount AS planned,
			b.threshold_pct,
			COALESCE(SUM(t.amount), 0) AS actual
		FROM budgets b
		LEFT JOIN categories c ON c.id = b.category_id
		LEFT JOIN transactions t ON (
			t.user_id = b.user_id
			AND t.type = 'expense'
			AND t.date >= $3
			AND t.date <= $4
			AND (b.category_id IS NULL OR t.category_id = b.category_id)
		)
		WHERE b.user_id = $1
		  AND (b.month = $2 OR b.month IS NULL)
		  AND (b.year = $5 OR b.year IS NULL)
		GROUP BY b.id, b.category_id, b.amount, b.threshold_pct, c.name, c.color, c.icon
		HAVING b.amount > 0
		   AND (COALESCE(SUM(t.amount), 0) / b.amount) * 100 >= b.threshold_pct
		ORDER BY (COALESCE(SUM(t.amount), 0) / b.amount) DESC
	`, userID, month, startDate, endDate, year)
	if err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetOverview alert_budgets query: %w", err)
	}
	defer alertRows.Close()

	overview.AlertBudgets = []domainrepo.BudgetProgress{}
	for alertRows.Next() {
		var bp domainrepo.BudgetProgress
		var thresholdPct float64
		if err := alertRows.Scan(
			&bp.BudgetID, &bp.CategoryID, &bp.CategoryName,
			&bp.CategoryColor, &bp.CategoryIcon,
			&bp.Planned, &thresholdPct, &bp.Actual,
		); err != nil {
			return nil, fmt.Errorf("dashboardRepository.GetOverview alert_budgets scan: %w", err)
		}
		if bp.Planned > 0 {
			bp.Percentage = (bp.Actual / bp.Planned) * 100
		}
		bp.IsAlert = bp.Percentage >= thresholdPct
		overview.AlertBudgets = append(overview.AlertBudgets, bp)
	}
	if err := alertRows.Err(); err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetOverview alert_budgets rows: %w", err)
	}

	// Recent 5 transactions with JOIN on accounts and categories
	recentRows, err := r.db.Query(ctx, `
		SELECT
			t.id, t.user_id, t.account_id, t.category_id,
			t.type, t.amount, t.description, t.notes, t.date,
			t.transfer_pair_id, t.recurrence_id, t.import_id,
			t.tags, t.ai_categorized, t.ai_confidence,
			t.created_at, t.updated_at,
			a.name AS account_name,
			c.name AS category_name,
			c.color AS category_color,
			c.icon AS category_icon
		FROM transactions t
		LEFT JOIN accounts a ON a.id = t.account_id
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1
		ORDER BY t.date DESC, t.created_at DESC
		LIMIT 5
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetOverview recent_transactions query: %w", err)
	}
	defer recentRows.Close()

	overview.RecentTransactions = []*entity.Transaction{}
	for recentRows.Next() {
		tx := &entity.Transaction{}
		if err := recentRows.Scan(
			&tx.ID, &tx.UserID, &tx.AccountID, &tx.CategoryID,
			&tx.Type, &tx.Amount, &tx.Description, &tx.Notes, &tx.Date,
			&tx.TransferPairID, &tx.RecurrenceID, &tx.ImportID,
			&tx.Tags, &tx.AICategorized, &tx.AIConfidence,
			&tx.CreatedAt, &tx.UpdatedAt,
			&tx.AccountName, &tx.CategoryName, &tx.CategoryColor, &tx.CategoryIcon,
		); err != nil {
			return nil, fmt.Errorf("dashboardRepository.GetOverview recent_transactions scan: %w", err)
		}
		overview.RecentTransactions = append(overview.RecentTransactions, tx)
	}
	if err := recentRows.Err(); err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetOverview recent_transactions rows: %w", err)
	}

	return overview, nil
}

func (r *dashboardRepository) GetCashflow(ctx context.Context, userID uuid.UUID, months int) ([]domainrepo.MonthlyCashflow, error) {
	// Calculate start date: first day of (current month - (months-1))
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month()-time.Month(months-1), 1, 0, 0, 0, 0, time.UTC)

	rows, err := r.db.Query(ctx, `
		SELECT
			EXTRACT(MONTH FROM date)::int AS month,
			EXTRACT(YEAR FROM date)::int AS year,
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS expense
		FROM transactions
		WHERE user_id = $1 AND date >= $2 AND type != 'transfer'
		GROUP BY month, year
		ORDER BY year ASC, month ASC
	`, userID, startDate)
	if err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetCashflow query: %w", err)
	}
	defer rows.Close()

	monthNames := []string{"Jan", "Fev", "Mar", "Abr", "Mai", "Jun", "Jul", "Ago", "Set", "Out", "Nov", "Dez"}

	var result []domainrepo.MonthlyCashflow
	for rows.Next() {
		var cf domainrepo.MonthlyCashflow
		if err := rows.Scan(&cf.Month, &cf.Year, &cf.Income, &cf.Expense); err != nil {
			return nil, fmt.Errorf("dashboardRepository.GetCashflow scan: %w", err)
		}
		cf.Balance = cf.Income - cf.Expense
		yearShort := cf.Year % 100
		cf.Label = fmt.Sprintf("%s/%02d", monthNames[cf.Month-1], yearShort)
		result = append(result, cf)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetCashflow rows: %w", err)
	}
	if result == nil {
		result = []domainrepo.MonthlyCashflow{}
	}
	return result, nil
}

func (r *dashboardRepository) GetPatrimonyHistory(ctx context.Context, userID uuid.UUID, months int) ([]domainrepo.PatrimonySnapshot, error) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month()-time.Month(months-1), 1, 0, 0, 0, 0, time.UTC)
	monthNames := []string{"Jan", "Fev", "Mar", "Abr", "Mai", "Jun", "Jul", "Ago", "Set", "Out", "Nov", "Dez"}

	// Monthly net savings (income - expense) per month
	savingsRows, err := r.db.Query(ctx, `
		SELECT
			EXTRACT(YEAR FROM date)::int  AS year,
			EXTRACT(MONTH FROM date)::int AS month,
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0) AS net_savings
		FROM transactions
		WHERE user_id = $1 AND date >= $2 AND type != 'transfer'
		GROUP BY year, month
		ORDER BY year ASC, month ASC
	`, userID, startDate)
	if err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetPatrimonyHistory savings query: %w", err)
	}
	defer savingsRows.Close()

	type monthKey struct{ Year, Month int }
	savings := map[monthKey]float64{}
	for savingsRows.Next() {
		var yr, mo int
		var net float64
		if err := savingsRows.Scan(&yr, &mo, &net); err != nil {
			return nil, fmt.Errorf("dashboardRepository.GetPatrimonyHistory savings scan: %w", err)
		}
		savings[monthKey{yr, mo}] = net
	}
	if err := savingsRows.Err(); err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetPatrimonyHistory savings rows: %w", err)
	}

	// Monthly invested capital (buys - sells) per month
	investRows, err := r.db.Query(ctx, `
		SELECT
			EXTRACT(YEAR FROM it.date)::int  AS year,
			EXTRACT(MONTH FROM it.date)::int AS month,
			COALESCE(SUM(CASE WHEN it.type = 'buy' THEN it.total ELSE -it.total END), 0) AS invested
		FROM investment_transactions it
		JOIN holdings h ON h.id = it.holding_id
		JOIN portfolios p ON p.id = h.portfolio_id
		WHERE p.user_id = $1 AND it.date >= $2 AND it.type IN ('buy','sell')
		GROUP BY year, month
		ORDER BY year ASC, month ASC
	`, userID, startDate)
	if err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetPatrimonyHistory invest query: %w", err)
	}
	defer investRows.Close()

	invested := map[monthKey]float64{}
	for investRows.Next() {
		var yr, mo int
		var inv float64
		if err := investRows.Scan(&yr, &mo, &inv); err != nil {
			return nil, fmt.Errorf("dashboardRepository.GetPatrimonyHistory invest scan: %w", err)
		}
		invested[monthKey{yr, mo}] = inv
	}
	if err := investRows.Err(); err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetPatrimonyHistory invest rows: %w", err)
	}

	// Build cumulative snapshots for each of the past `months` months
	result := make([]domainrepo.PatrimonySnapshot, 0, months)
	var cumSavings, cumInvested float64
	for i := months - 1; i >= 0; i-- {
		t := time.Date(now.Year(), now.Month()-time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		key := monthKey{t.Year(), int(t.Month())}
		cumSavings += savings[key]
		cumInvested += invested[key]
		result = append(result, domainrepo.PatrimonySnapshot{
			Month:         int(t.Month()),
			Year:          t.Year(),
			Label:         fmt.Sprintf("%s/%02d", monthNames[int(t.Month())-1], t.Year()%100),
			BankSavings:   cumSavings,
			InvestedTotal: cumInvested,
			TotalNetWorth: cumSavings + cumInvested,
		})
	}
	return result, nil
}
