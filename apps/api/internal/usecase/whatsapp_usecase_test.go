package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/financeos/api/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Fake WhatsApp Repository ---

type fakeWhatsAppRepo struct {
	sessions map[string]*entity.WhatsAppSession
	users    map[string]*entity.User
}

func newFakeWhatsAppRepo() *fakeWhatsAppRepo {
	return &fakeWhatsAppRepo{
		sessions: make(map[string]*entity.WhatsAppSession),
		users:    make(map[string]*entity.User),
	}
}

func (r *fakeWhatsAppRepo) FindSessionByPhone(ctx context.Context, phone string) (*entity.WhatsAppSession, error) {
	s, ok := r.sessions[phone]
	if !ok {
		return nil, nil
	}
	// Return a copy
	copy := *s
	copyData := make(map[string]interface{})
	for k, v := range s.SessionData {
		copyData[k] = v
	}
	copy.SessionData = copyData
	return &copy, nil
}

func (r *fakeWhatsAppRepo) CreateSession(ctx context.Context, s *entity.WhatsAppSession) error {
	r.sessions[s.PhoneNumber] = s
	return nil
}

func (r *fakeWhatsAppRepo) UpdateSession(ctx context.Context, s *entity.WhatsAppSession) error {
	if existing, ok := r.sessions[s.PhoneNumber]; ok {
		existing.State = s.State
		existing.SessionData = s.SessionData
		existing.LastActivity = s.LastActivity
		existing.IsActive = s.IsActive
	}
	return nil
}

func (r *fakeWhatsAppRepo) FindUserByPhone(ctx context.Context, phone string) (*entity.User, error) {
	u, ok := r.users[phone]
	if !ok {
		return nil, nil
	}
	return u, nil
}

// --- Fake Account Repository for WhatsApp ---

type fakeWhatsAppAccountRepo struct {
	accounts []*entity.Account
}

func newFakeWhatsAppAccountRepo(userID uuid.UUID) *fakeWhatsAppAccountRepo {
	return &fakeWhatsAppAccountRepo{
		accounts: []*entity.Account{
			{
				ID:      uuid.New(),
				UserID:  userID,
				Name:    "Conta Corrente",
				Type:    "checking",
				Balance: 5000.0,
			},
		},
	}
}

func (r *fakeWhatsAppAccountRepo) Create(ctx context.Context, account *entity.Account) error { return nil }
func (r *fakeWhatsAppAccountRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Account, error) {
	return nil, nil
}
func (r *fakeWhatsAppAccountRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Account, error) {
	return r.accounts, nil
}
func (r *fakeWhatsAppAccountRepo) Update(ctx context.Context, account *entity.Account) error {
	return nil
}
func (r *fakeWhatsAppAccountRepo) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}
func (r *fakeWhatsAppAccountRepo) GetSummary(ctx context.Context, userID uuid.UUID) (*domainrepo.AccountSummary, error) {
	return &domainrepo.AccountSummary{}, nil
}

// --- Fake Transaction Repository for WhatsApp ---

type fakeWhatsAppTxRepo struct {
	created []*entity.Transaction
}

func newFakeWhatsAppTxRepo() *fakeWhatsAppTxRepo {
	return &fakeWhatsAppTxRepo{created: []*entity.Transaction{}}
}

func (r *fakeWhatsAppTxRepo) Create(ctx context.Context, tx *entity.Transaction) error {
	r.created = append(r.created, tx)
	return nil
}
func (r *fakeWhatsAppTxRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Transaction, error) {
	return nil, nil
}
func (r *fakeWhatsAppTxRepo) FindByUserID(ctx context.Context, userID uuid.UUID, filter domainrepo.TransactionFilter) ([]*entity.Transaction, int, error) {
	return r.created, len(r.created), nil
}
func (r *fakeWhatsAppTxRepo) Update(ctx context.Context, tx *entity.Transaction) error { return nil }
func (r *fakeWhatsAppTxRepo) Delete(ctx context.Context, id, userID uuid.UUID) error   { return nil }
func (r *fakeWhatsAppTxRepo) GetSummary(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*domainrepo.TransactionSummary, error) {
	return &domainrepo.TransactionSummary{
		TotalIncome:  5000.0,
		TotalExpense: 1500.0,
		Balance:      3500.0,
		ByCategory: []domainrepo.CategorySummary{
			{CategoryName: "Supermercado", Total: 600.0, Count: 3},
			{CategoryName: "Restaurante", Total: 300.0, Count: 5},
		},
	}, nil
}
func (r *fakeWhatsAppTxRepo) CreateTransfer(ctx context.Context, debit, credit *entity.Transaction) error {
	return nil
}
func (r *fakeWhatsAppTxRepo) UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, delta float64) error {
	return nil
}

// --- Helper ---

func buildWhatsAppUC(userID uuid.UUID) (usecase.WhatsAppUseCase, *fakeWhatsAppRepo, *fakeWhatsAppTxRepo) {
	whatsappRepo := newFakeWhatsAppRepo()
	txRepo := newFakeWhatsAppTxRepo()
	accountRepo := newFakeWhatsAppAccountRepo(userID)

	// Pre-link phone to user
	phone := "5511999999999"
	whatsappRepo.users[phone] = &entity.User{ID: userID, Name: "Test User", Email: "test@example.com"}

	uc := usecase.NewWhatsAppUseCase(whatsappRepo, txRepo, accountRepo)
	return uc, whatsappRepo, txRepo
}

// --- Tests ---

func TestHandleMessage_Resumo(t *testing.T) {
	userID := uuid.New()
	uc, _, _ := buildWhatsAppUC(userID)

	response, err := uc.HandleMessage(context.Background(), "5511999999999", "resumo")
	require.NoError(t, err)
	assert.NotEmpty(t, response)
	assert.Contains(t, response, "Resumo")
	assert.Contains(t, response, "R$")
}

func TestHandleMessage_GasteiX(t *testing.T) {
	userID := uuid.New()
	uc, whatsappRepo, _ := buildWhatsAppUC(userID)

	response, err := uc.HandleMessage(context.Background(), "5511999999999", "gastei 150 reais em supermercado")
	require.NoError(t, err)
	assert.NotEmpty(t, response)
	assert.Contains(t, response, "150.00")
	assert.Contains(t, response, "Confirmar")

	// Verify state changed to awaiting_confirmation
	session := whatsappRepo.sessions["5511999999999"]
	require.NotNil(t, session)
	assert.Equal(t, "awaiting_confirmation", session.State)
	assert.Equal(t, 150.0, session.SessionData["pending_amount"])
	assert.Equal(t, "expense", session.SessionData["pending_type"])
}

func TestHandleMessage_ConfirmTransaction(t *testing.T) {
	userID := uuid.New()
	uc, _, txRepo := buildWhatsAppUC(userID)

	// First: register a gasto
	_, err := uc.HandleMessage(context.Background(), "5511999999999", "gastei 200 em restaurante")
	require.NoError(t, err)

	// Then: confirm it
	response, err := uc.HandleMessage(context.Background(), "5511999999999", "sim")
	require.NoError(t, err)
	assert.Contains(t, response, "registrado com sucesso")

	// Verify transaction was created
	assert.Len(t, txRepo.created, 1)
	assert.Equal(t, "expense", txRepo.created[0].Type)
	assert.Equal(t, 200.0, txRepo.created[0].Amount)
}

func TestHandleMessage_CancelTransaction(t *testing.T) {
	userID := uuid.New()
	uc, whatsappRepo, txRepo := buildWhatsAppUC(userID)

	_, err := uc.HandleMessage(context.Background(), "5511999999999", "gastei 100")
	require.NoError(t, err)

	response, err := uc.HandleMessage(context.Background(), "5511999999999", "não")
	require.NoError(t, err)
	assert.Contains(t, response, "cancelada")

	// No transaction created
	assert.Empty(t, txRepo.created)

	// Session reset to idle
	session := whatsappRepo.sessions["5511999999999"]
	assert.Equal(t, "idle", session.State)
}

func TestHandleMessage_Timeout(t *testing.T) {
	userID := uuid.New()
	uc, whatsappRepo, _ := buildWhatsAppUC(userID)

	// Manually create a session with old activity
	oldTime := time.Now().Add(-20 * time.Minute)
	whatsappRepo.sessions["5511999999999"] = &entity.WhatsAppSession{
		ID:           uuid.New(),
		UserID:       userID,
		PhoneNumber:  "5511999999999",
		State:        "awaiting_confirmation",
		SessionData:  map[string]interface{}{"pending_amount": 100.0, "pending_type": "expense"},
		LastActivity: oldTime,
		IsActive:     true,
		CreatedAt:    oldTime,
	}

	// Any message after timeout should reset to idle and treat as new command
	response, err := uc.HandleMessage(context.Background(), "5511999999999", "resumo")
	require.NoError(t, err)

	// Should have processed as idle (resumo), not as confirmation
	assert.Contains(t, response, "Resumo")
}

func TestHandleMessage_UnknownCommand(t *testing.T) {
	userID := uuid.New()
	uc, _, _ := buildWhatsAppUC(userID)

	response, err := uc.HandleMessage(context.Background(), "5511999999999", "qualquer coisa aleatória")
	require.NoError(t, err)
	assert.Contains(t, response, "Não entendi")
	assert.Contains(t, response, "resumo")
}

func TestHandleMessage_Carteira(t *testing.T) {
	userID := uuid.New()
	uc, _, _ := buildWhatsAppUC(userID)

	response, err := uc.HandleMessage(context.Background(), "5511999999999", "carteira")
	require.NoError(t, err)
	assert.Contains(t, response, "Carteira")
	assert.Contains(t, response, "R$")
}

func TestHandleMessage_RecebiX(t *testing.T) {
	userID := uuid.New()
	uc, _, _ := buildWhatsAppUC(userID)

	response, err := uc.HandleMessage(context.Background(), "5511999999999", "recebi 5000")
	require.NoError(t, err)
	assert.Contains(t, response, "5000.00")
	assert.Contains(t, response, "Confirmar")
}
