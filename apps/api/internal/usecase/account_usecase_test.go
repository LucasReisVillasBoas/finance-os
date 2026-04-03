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

// --- Mock account repository ---

type mockAccountRepo struct {
	accounts  map[uuid.UUID]*entity.Account
	createErr error
	findErr   error
	updateErr error
	deleteErr error
}

func newMockAccountRepo() *mockAccountRepo {
	return &mockAccountRepo{
		accounts: make(map[uuid.UUID]*entity.Account),
	}
}

func (m *mockAccountRepo) Create(ctx context.Context, account *entity.Account) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.accounts[account.ID] = account
	return nil
}

func (m *mockAccountRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Account, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	a, ok := m.accounts[id]
	if !ok || !a.IsActive || a.UserID != userID {
		return nil, nil
	}
	return a, nil
}

func (m *mockAccountRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Account, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []*entity.Account
	for _, a := range m.accounts {
		if a.UserID == userID && a.IsActive {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAccountRepo) Update(ctx context.Context, account *entity.Account) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.accounts[account.ID] = account
	return nil
}

func (m *mockAccountRepo) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if a, ok := m.accounts[id]; ok && a.UserID == userID {
		a.IsActive = false
	}
	return nil
}

func (m *mockAccountRepo) GetSummary(ctx context.Context, userID uuid.UUID) (*domainrepo.AccountSummary, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	summary := &domainrepo.AccountSummary{
		AccountBalances: []domainrepo.AccountBalance{},
	}
	for _, a := range m.accounts {
		if a.UserID == userID && a.IsActive {
			summary.TotalBalance += a.Balance
			if a.Type != "credit_card" {
				summary.NetBalance += a.Balance
			}
			summary.AccountBalances = append(summary.AccountBalances, domainrepo.AccountBalance{
				AccountID:   a.ID,
				AccountName: a.Name,
				Balance:     a.Balance,
			})
		}
	}
	return summary, nil
}

// --- Helpers ---

func newAccountUseCase(repo *mockAccountRepo) usecase.AccountUseCase {
	return usecase.NewAccountUseCase(repo)
}

func seedAccount(repo *mockAccountRepo, userID uuid.UUID) *entity.Account {
	a := &entity.Account{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      "Conta Corrente",
		Type:      "checking",
		Balance:   1000.0,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.accounts[a.ID] = a
	return a
}

// --- Tests ---

func TestCreateAccount_Success(t *testing.T) {
	repo := newMockAccountRepo()
	uc := newAccountUseCase(repo)
	userID := uuid.New()

	inst := "Banco do Brasil"
	req := usecase.CreateAccountRequest{
		Name:        "Poupança BB",
		Type:        "savings",
		Institution: &inst,
		Balance:     500.0,
	}

	account, err := uc.Create(context.Background(), userID, req)
	require.NoError(t, err)
	require.NotNil(t, account)

	assert.Equal(t, "Poupança BB", account.Name)
	assert.Equal(t, "savings", account.Type)
	assert.Equal(t, &inst, account.Institution)
	assert.Equal(t, 500.0, account.Balance)
	assert.True(t, account.IsActive)
	assert.Equal(t, userID, account.UserID)
	assert.NotEqual(t, uuid.Nil, account.ID)
}

func TestCreateAccount_RepoError(t *testing.T) {
	repo := newMockAccountRepo()
	repo.createErr = errors.New("db error")
	uc := newAccountUseCase(repo)

	_, err := uc.Create(context.Background(), uuid.New(), usecase.CreateAccountRequest{
		Name: "Test", Type: "checking",
	})
	require.Error(t, err)
}

func TestGetAll_ReturnsOnlyUserAccounts(t *testing.T) {
	repo := newMockAccountRepo()
	uc := newAccountUseCase(repo)

	user1 := uuid.New()
	user2 := uuid.New()

	seedAccount(repo, user1)
	seedAccount(repo, user1)
	seedAccount(repo, user2)

	accounts, err := uc.GetAll(context.Background(), user1)
	require.NoError(t, err)
	assert.Len(t, accounts, 2)
	for _, a := range accounts {
		assert.Equal(t, user1, a.UserID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newMockAccountRepo()
	uc := newAccountUseCase(repo)

	_, err := uc.GetByID(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrAccountNotFound))
}

func TestGetByID_WrongUser(t *testing.T) {
	repo := newMockAccountRepo()
	uc := newAccountUseCase(repo)

	user1 := uuid.New()
	a := seedAccount(repo, user1)

	_, err := uc.GetByID(context.Background(), a.ID, uuid.New())
	require.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrAccountNotFound))
}

func TestUpdateAccount_Success(t *testing.T) {
	repo := newMockAccountRepo()
	uc := newAccountUseCase(repo)
	userID := uuid.New()
	a := seedAccount(repo, userID)

	newName := "Conta Atualizada"
	req := usecase.UpdateAccountRequest{Name: &newName}

	updated, err := uc.Update(context.Background(), a.ID, userID, req)
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
}

func TestDeleteAccount_SoftDelete(t *testing.T) {
	repo := newMockAccountRepo()
	uc := newAccountUseCase(repo)
	userID := uuid.New()
	a := seedAccount(repo, userID)

	err := uc.Delete(context.Background(), a.ID, userID)
	require.NoError(t, err)

	// After soft delete, the account should not be findable
	_, err = uc.GetByID(context.Background(), a.ID, userID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrAccountNotFound))

	// Verify it's still in the store but inactive
	assert.False(t, repo.accounts[a.ID].IsActive)
}

func TestDeleteAccount_NotFound(t *testing.T) {
	repo := newMockAccountRepo()
	uc := newAccountUseCase(repo)

	err := uc.Delete(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrAccountNotFound))
}

func TestGetSummary_CalculatesCorrectly(t *testing.T) {
	repo := newMockAccountRepo()
	uc := newAccountUseCase(repo)
	userID := uuid.New()

	// Add a checking account with balance 1000
	a1 := seedAccount(repo, userID)
	a1.Balance = 1000.0
	a1.Type = "checking"

	// Add a credit card with negative balance (debt)
	a2 := &entity.Account{
		ID:       uuid.New(),
		UserID:   userID,
		Name:     "Cartão",
		Type:     "credit_card",
		Balance:  -200.0,
		IsActive: true,
	}
	repo.accounts[a2.ID] = a2

	summary, err := uc.GetSummary(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 800.0, summary.TotalBalance)
	assert.Equal(t, 1000.0, summary.NetBalance) // credit card excluded
	assert.Len(t, summary.AccountBalances, 2)
}
