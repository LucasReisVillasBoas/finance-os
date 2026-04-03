package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/financeos/api/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock ---

type mockTransactionRepo struct {
	transactions   map[uuid.UUID]*entity.Transaction
	balanceUpdates map[uuid.UUID]float64
	createErr      error
	findErr        error
	updateErr      error
	deleteErr      error
	summaryErr     error
}

func newMockTransactionRepo() *mockTransactionRepo {
	return &mockTransactionRepo{
		transactions:   make(map[uuid.UUID]*entity.Transaction),
		balanceUpdates: make(map[uuid.UUID]float64),
	}
}

func (m *mockTransactionRepo) Create(ctx context.Context, tx *entity.Transaction) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.transactions[tx.ID] = tx
	// simulate balance update
	if tx.Type == "income" {
		m.balanceUpdates[tx.AccountID] += tx.Amount
	} else if tx.Type == "expense" {
		m.balanceUpdates[tx.AccountID] -= tx.Amount
	}
	return nil
}

func (m *mockTransactionRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Transaction, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	tx, ok := m.transactions[id]
	if !ok {
		return nil, nil
	}
	if tx.UserID != userID {
		return nil, nil
	}
	return tx, nil
}

func (m *mockTransactionRepo) FindByUserID(ctx context.Context, userID uuid.UUID, filter domainrepo.TransactionFilter) ([]*entity.Transaction, int, error) {
	if m.findErr != nil {
		return nil, 0, m.findErr
	}
	var result []*entity.Transaction
	for _, tx := range m.transactions {
		if tx.UserID != userID {
			continue
		}
		if filter.Type != nil && tx.Type != *filter.Type {
			continue
		}
		if filter.AccountID != nil && tx.AccountID != *filter.AccountID {
			continue
		}
		result = append(result, tx)
	}
	return result, len(result), nil
}

func (m *mockTransactionRepo) Update(ctx context.Context, tx *entity.Transaction) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.transactions[tx.ID] = tx
	return nil
}

func (m *mockTransactionRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	tx, ok := m.transactions[id]
	if !ok {
		return nil
	}
	// reverse balance
	if tx.Type == "income" {
		m.balanceUpdates[tx.AccountID] -= tx.Amount
	} else if tx.Type == "expense" {
		m.balanceUpdates[tx.AccountID] += tx.Amount
	}
	delete(m.transactions, id)
	return nil
}

func (m *mockTransactionRepo) GetSummary(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*domainrepo.TransactionSummary, error) {
	if m.summaryErr != nil {
		return nil, m.summaryErr
	}
	summary := &domainrepo.TransactionSummary{
		ByCategory: []domainrepo.CategorySummary{},
	}
	for _, tx := range m.transactions {
		if tx.UserID != userID {
			continue
		}
		if tx.Type == "income" {
			summary.TotalIncome += tx.Amount
		} else if tx.Type == "expense" {
			summary.TotalExpense += tx.Amount
		}
	}
	summary.Balance = summary.TotalIncome - summary.TotalExpense
	return summary, nil
}

func (m *mockTransactionRepo) CreateTransfer(ctx context.Context, debit, credit *entity.Transaction) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.transactions[debit.ID] = debit
	m.transactions[credit.ID] = credit
	m.balanceUpdates[debit.AccountID] -= debit.Amount
	m.balanceUpdates[credit.AccountID] += credit.Amount
	return nil
}

func (m *mockTransactionRepo) UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, delta float64) error {
	m.balanceUpdates[accountID] += delta
	return nil
}

// --- Tests ---

func TestCreateTransaction_Income_UpdatesBalance(t *testing.T) {
	repo := newMockTransactionRepo()
	uc := usecase.NewTransactionUseCase(repo)

	userID := uuid.New()
	accountID := uuid.New()
	desc := "Salary"
	req := usecase.CreateTransactionRequest{
		AccountID:   accountID,
		Type:        "income",
		Amount:      5000.00,
		Description: &desc,
		Date:        time.Now(),
	}

	tx, err := uc.Create(context.Background(), userID, req)
	require.NoError(t, err)
	assert.Equal(t, "income", tx.Type)
	assert.Equal(t, 5000.00, tx.Amount)
	assert.Equal(t, userID, tx.UserID)
	assert.Equal(t, accountID, tx.AccountID)
	assert.Equal(t, 5000.00, repo.balanceUpdates[accountID])
}

func TestCreateTransaction_Expense_UpdatesBalance(t *testing.T) {
	repo := newMockTransactionRepo()
	uc := usecase.NewTransactionUseCase(repo)

	userID := uuid.New()
	accountID := uuid.New()
	desc := "Grocery"
	req := usecase.CreateTransactionRequest{
		AccountID:   accountID,
		Type:        "expense",
		Amount:      150.00,
		Description: &desc,
		Date:        time.Now(),
	}

	tx, err := uc.Create(context.Background(), userID, req)
	require.NoError(t, err)
	assert.Equal(t, "expense", tx.Type)
	assert.Equal(t, 150.00, tx.Amount)
	assert.Equal(t, -150.00, repo.balanceUpdates[accountID])
}

func TestCreateTransaction_InvalidType(t *testing.T) {
	repo := newMockTransactionRepo()
	repo.createErr = errors.New("invalid type")
	uc := usecase.NewTransactionUseCase(repo)

	userID := uuid.New()
	req := usecase.CreateTransactionRequest{
		AccountID: uuid.New(),
		Type:      "expense",
		Amount:    100.00,
		Date:      time.Now(),
	}

	_, err := uc.Create(context.Background(), userID, req)
	assert.Error(t, err)
}

func TestCreateTransfer_CreatesDebitAndCredit(t *testing.T) {
	repo := newMockTransactionRepo()
	uc := usecase.NewTransactionUseCase(repo)

	userID := uuid.New()
	fromAccountID := uuid.New()
	toAccountID := uuid.New()
	desc := "Transfer"
	req := usecase.CreateTransferRequest{
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        1000.00,
		Description:   &desc,
		Date:          time.Now(),
	}

	txs, err := uc.CreateTransfer(context.Background(), userID, req)
	require.NoError(t, err)
	require.Len(t, txs, 2)

	// Both should be transfer type
	assert.Equal(t, "transfer", txs[0].Type)
	assert.Equal(t, "transfer", txs[1].Type)

	// From account decremented
	assert.Equal(t, -1000.00, repo.balanceUpdates[fromAccountID])
	// To account incremented
	assert.Equal(t, 1000.00, repo.balanceUpdates[toAccountID])

	// Both share the same transfer_pair_id
	require.NotNil(t, txs[0].TransferPairID)
	require.NotNil(t, txs[1].TransferPairID)
	assert.Equal(t, *txs[0].TransferPairID, *txs[1].TransferPairID)
}

func TestCreateTransfer_SameAccount_ReturnsError(t *testing.T) {
	repo := newMockTransactionRepo()
	uc := usecase.NewTransactionUseCase(repo)

	accountID := uuid.New()
	req := usecase.CreateTransferRequest{
		FromAccountID: accountID,
		ToAccountID:   accountID,
		Amount:        500.00,
		Date:          time.Now(),
	}

	_, err := uc.CreateTransfer(context.Background(), uuid.New(), req)
	assert.ErrorIs(t, err, usecase.ErrSameAccountTransfer)
}

func TestListTransactions_WithFilters(t *testing.T) {
	repo := newMockTransactionRepo()
	uc := usecase.NewTransactionUseCase(repo)

	userID := uuid.New()
	accountID := uuid.New()

	// Create two transactions
	incomeType := "income"
	expenseType := "expense"

	incomeDesc := "Salary"
	expenseDesc := "Rent"

	repo.transactions[uuid.New()] = &entity.Transaction{
		ID: uuid.New(), UserID: userID, AccountID: accountID,
		Type: "income", Amount: 3000, Description: &incomeDesc,
		Date: time.Now(), Tags: []string{},
	}
	repo.transactions[uuid.New()] = &entity.Transaction{
		ID: uuid.New(), UserID: userID, AccountID: accountID,
		Type: "expense", Amount: 800, Description: &expenseDesc,
		Date: time.Now(), Tags: []string{},
	}

	// List only income
	txs, total, err := uc.List(context.Background(), userID, usecase.ListTransactionsRequest{
		Type:     &incomeType,
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, txs, 1)
	assert.Equal(t, "income", txs[0].Type)

	// List only expense
	txs, total, err = uc.List(context.Background(), userID, usecase.ListTransactionsRequest{
		Type:     &expenseType,
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, txs, 1)
	assert.Equal(t, "expense", txs[0].Type)
}

func TestGetTransactionByID_NotFound(t *testing.T) {
	repo := newMockTransactionRepo()
	uc := usecase.NewTransactionUseCase(repo)

	_, err := uc.GetByID(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, usecase.ErrTransactionNotFound)
}

func TestDeleteTransaction(t *testing.T) {
	repo := newMockTransactionRepo()
	uc := usecase.NewTransactionUseCase(repo)

	userID := uuid.New()
	accountID := uuid.New()
	txID := uuid.New()
	desc := "Test"
	repo.transactions[txID] = &entity.Transaction{
		ID: txID, UserID: userID, AccountID: accountID,
		Type: "expense", Amount: 200, Description: &desc,
		Date: time.Now(), Tags: []string{},
	}

	err := uc.Delete(context.Background(), txID, userID)
	require.NoError(t, err)

	// Should be removed
	_, exists := repo.transactions[txID]
	assert.False(t, exists)
	// Balance should be reversed
	assert.Equal(t, 200.00, repo.balanceUpdates[accountID])
}

func TestDeleteTransaction_NotFound(t *testing.T) {
	repo := newMockTransactionRepo()
	uc := usecase.NewTransactionUseCase(repo)

	err := uc.Delete(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, usecase.ErrTransactionNotFound)
}
