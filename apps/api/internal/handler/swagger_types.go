package handler

import "github.com/financeos/api/internal/domain/entity"

// ─── Generic wrappers ────────────────────────────────────────────────────────

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the error code and human-readable message.
type ErrorDetail struct {
	Code    string `json:"code"    example:"INVALID_INPUT"`
	Message string `json:"message" example:"name is required"`
}

// MessageResponse is returned for operations that only communicate success.
type MessageResponse struct {
	Data MessageData `json:"data"`
}

// MessageData holds the confirmation message.
type MessageData struct {
	Message string `json:"message" example:"logged out successfully"`
}

// ─── Auth ────────────────────────────────────────────────────────────────────

// AuthDataResponse wraps AuthResponse in the standard data envelope.
type AuthDataResponse struct {
	Data AuthResponseBody `json:"data"`
}

// AuthResponseBody is the payload returned on login, register and refresh.
type AuthResponseBody struct {
	AccessToken  string      `json:"access_token"  example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string      `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User         entity.User `json:"user"`
}

// ─── Accounts ────────────────────────────────────────────────────────────────

// AccountResponse wraps a single account.
type AccountResponse struct {
	Data entity.Account `json:"data"`
}

// AccountListResponse wraps a list of accounts.
type AccountListResponse struct {
	Data []*entity.Account `json:"data"`
}

// AccountSummaryResponse wraps the balance summary.
type AccountSummaryResponse struct {
	Data AccountSummaryData `json:"data"`
}

// AccountSummaryData contains aggregated balance information.
type AccountSummaryData struct {
	TotalBalance    float64            `json:"total_balance"    example:"15000"`
	AccountBalances map[string]float64 `json:"account_balances"`
}

// ─── Categories ──────────────────────────────────────────────────────────────

// CategoryResponse wraps a single category.
type CategoryResponse struct {
	Data entity.Category `json:"data"`
}

// CategoryListResponse wraps a list of categories.
type CategoryListResponse struct {
	Data []*entity.Category `json:"data"`
}

// ─── Transactions ─────────────────────────────────────────────────────────────

// TransactionResponse wraps a single transaction.
type TransactionResponse struct {
	Data entity.Transaction `json:"data"`
}

// TransactionListResponse wraps a paginated list of transactions.
type TransactionListResponse struct {
	Data []*entity.Transaction `json:"data"`
	Meta PaginationMeta        `json:"meta"`
}

// TransactionSummaryResponse wraps the transaction summary.
type TransactionSummaryResponse struct {
	Data TransactionSummaryData `json:"data"`
}

// TransactionSummaryData contains income/expense breakdown.
type TransactionSummaryData struct {
	TotalIncome   float64                    `json:"total_income"   example:"8000"`
	TotalExpense  float64                    `json:"total_expense"  example:"1500"`
	Balance       float64                    `json:"balance"        example:"6500"`
	ByCategory    map[string]float64         `json:"by_category"`
	TopCategories []CategoryExpense          `json:"top_categories"`
}

// CategoryExpense represents spending for a single category.
type CategoryExpense struct {
	CategoryID   string  `json:"category_id"   example:"uuid"`
	CategoryName string  `json:"category_name" example:"Alimentação"`
	Amount       float64 `json:"amount"        example:"350.00"`
}

// TransferResponse wraps the pair of transactions created by a transfer.
type TransferResponse struct {
	Data []*entity.Transaction `json:"data"`
}

// ─── Recurrences ─────────────────────────────────────────────────────────────

// RecurrenceResponse wraps a single recurrence.
type RecurrenceResponse struct {
	Data entity.Recurrence `json:"data"`
}

// RecurrenceListResponse wraps a list of recurrences.
type RecurrenceListResponse struct {
	Data []*entity.Recurrence `json:"data"`
	Meta PaginationMeta       `json:"meta"`
}

// ─── Budgets ──────────────────────────────────────────────────────────────────

// BudgetResponse wraps a single budget.
type BudgetResponse struct {
	Data entity.Budget `json:"data"`
}

// BudgetListResponse wraps a list of budgets.
type BudgetListResponse struct {
	Data []*entity.Budget `json:"data"`
	Meta PaginationMeta   `json:"meta"`
}

// BudgetProgressResponse wraps budget progress items.
type BudgetProgressResponse struct {
	Data []BudgetProgressItem `json:"data"`
	Meta PaginationMeta       `json:"meta"`
}

// BudgetProgressItem contains planned vs actual spending for a budget.
type BudgetProgressItem struct {
	BudgetID     string  `json:"budget_id"     example:"uuid"`
	CategoryName string  `json:"category_name" example:"Alimentação"`
	Planned      float64 `json:"planned"       example:"600"`
	Actual       float64 `json:"actual"        example:"150"`
	Percentage   float64 `json:"percentage"    example:"25"`
	IsAlert      bool    `json:"is_alert"      example:"false"`
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

// DashboardOverviewResponse wraps the monthly overview.
type DashboardOverviewResponse struct {
	Data DashboardOverviewData `json:"data"`
}

// DashboardOverviewData contains the aggregated monthly financial overview.
type DashboardOverviewData struct {
	NetBalance         float64              `json:"net_balance"          example:"22850"`
	MonthIncome        float64              `json:"month_income"         example:"8000"`
	MonthExpense       float64              `json:"month_expense"        example:"1500"`
	TopCategories      []CategoryExpense    `json:"top_categories"`
	RecentTransactions []entity.Transaction `json:"recent_transactions"`
}

// CashflowResponse wraps 12-month cashflow data.
type CashflowResponse struct {
	Data []MonthlyCashflowItem `json:"data"`
}

// MonthlyCashflowItem contains income/expense for a single month.
type MonthlyCashflowItem struct {
	Month   int     `json:"month"   example:"4"`
	Year    int     `json:"year"    example:"2026"`
	Income  float64 `json:"income"  example:"8000"`
	Expense float64 `json:"expense" example:"1500"`
	Balance float64 `json:"balance" example:"6500"`
}

// ─── Investments ──────────────────────────────────────────────────────────────

// PortfolioResponse wraps a single portfolio.
type PortfolioResponse struct {
	Data entity.Portfolio `json:"data"`
}

// PortfolioListResponse wraps a list of portfolios.
type PortfolioListResponse struct {
	Data []*entity.Portfolio `json:"data"`
}

// HoldingResponse wraps a single holding.
type HoldingResponse struct {
	Data entity.Holding `json:"data"`
}

// HoldingListResponse wraps a list of holdings.
type HoldingListResponse struct {
	Data []*entity.Holding `json:"data"`
}

// InvestmentTransactionResponse wraps a single investment transaction.
type InvestmentTransactionResponse struct {
	Data entity.InvestmentTransaction `json:"data"`
}

// InvestmentTransactionListResponse wraps a list of investment transactions.
type InvestmentTransactionListResponse struct {
	Data []*entity.InvestmentTransaction `json:"data"`
}

// AssetSearchResponse wraps asset search results.
type AssetSearchResponse struct {
	Data []*entity.Asset `json:"data"`
}

// CustomAssetResponse wraps a single custom asset.
type CustomAssetResponse struct {
	Data entity.CustomAsset `json:"data"`
}

// CustomAssetListResponse wraps a list of custom assets.
type CustomAssetListResponse struct {
	Data []*entity.CustomAsset `json:"data"`
}

// ─── Goals ────────────────────────────────────────────────────────────────────

// GoalResponse wraps a single goal.
type GoalResponse struct {
	Data entity.Goal `json:"data"`
}

// GoalListResponse wraps a list of goals.
type GoalListResponse struct {
	Data []*entity.Goal `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// GoalContributionResponse wraps a goal contribution.
type GoalContributionResponse struct {
	Data entity.GoalContribution `json:"data"`
}

// GoalProjectionsResponse wraps goal projection data.
type GoalProjectionsResponse struct {
	Data []GoalProjectionItem `json:"data"`
}

// GoalProjectionItem contains projected completion data for a goal.
type GoalProjectionItem struct {
	GoalID        string  `json:"goal_id"        example:"uuid"`
	GoalName      string  `json:"goal_name"      example:"Reserva de Emergência"`
	ProgressPct   float64 `json:"progress_pct"   example:"1.67"`
	MonthsToGoal  int     `json:"months_to_goal" example:"30"`
	EstimatedDate string  `json:"estimated_date" example:"2028-10-04"`
}

// ─── Notifications ────────────────────────────────────────────────────────────

// NotificationListResponse wraps a list of notifications.
type NotificationListResponse struct {
	Data []*entity.Notification `json:"data"`
	Meta NotificationMeta       `json:"meta"`
}

// NotificationMeta contains notification metadata.
type NotificationMeta struct {
	UnreadCount int `json:"unread_count" example:"3"`
}

// UpdatedResponse is returned when a record was updated.
type UpdatedResponse struct {
	Data UpdatedData `json:"data"`
}

// UpdatedData contains the update confirmation.
type UpdatedData struct {
	Updated bool `json:"updated" example:"true"`
}

// ─── Family ───────────────────────────────────────────────────────────────────

// FamilyGroupResponse wraps a family group.
type FamilyGroupResponse struct {
	Data entity.FamilyGroup `json:"data"`
}

// InviteCodeResponse wraps the invite code.
type InviteCodeResponse struct {
	Data InviteCodeData `json:"data"`
}

// InviteCodeData contains the invite code.
type InviteCodeData struct {
	InviteCode string `json:"invite_code" example:"ABCD1234"`
}

// FamilyDashboardResponse wraps the family dashboard.
type FamilyDashboardResponse struct {
	Data FamilyDashboardData `json:"data"`
}

// FamilyDashboardData contains aggregated family financial data.
type FamilyDashboardData struct {
	Members     []entity.FamilyMember `json:"members"`
	TotalIncome float64               `json:"total_income"  example:"16000"`
	TotalExpense float64              `json:"total_expense" example:"3000"`
}

// ─── AI ───────────────────────────────────────────────────────────────────────

// AIChatResponse wraps the AI assistant reply.
type AIChatResponse struct {
	Data AIChatData `json:"data"`
}

// AIChatData contains the AI response text.
type AIChatData struct {
	Response string `json:"response" example:"Você gastou R$1.500 em Alimentação este mês, 25% acima da média."`
}

// AIForecastResponse wraps spending forecast data.
type AIForecastResponse struct {
	Data map[string]interface{} `json:"data"`
}

// ─── Health ───────────────────────────────────────────────────────────────────

// HealthResponse is the API health status.
type HealthResponse struct {
	Status  string `json:"status"  example:"ok"`
	Service string `json:"service" example:"financeos-api"`
	Env     string `json:"env"     example:"development"`
}

// ─── Pagination ───────────────────────────────────────────────────────────────

// PaginationMeta contains pagination information.
type PaginationMeta struct {
	Page     int `json:"page"      example:"1"`
	PageSize int `json:"page_size" example:"20"`
	Total    int `json:"total"     example:"100"`
}
