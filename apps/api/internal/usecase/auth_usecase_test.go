package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/financeos/api/internal/usecase"
	"github.com/financeos/api/pkg/config"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// --- Mock repository ---

type mockUserRepo struct {
	users         map[string]*entity.User   // keyed by email
	usersByID     map[uuid.UUID]*entity.User
	refreshTokens map[string]*entity.RefreshToken

	createErr         error
	findByEmailErr    error
	findByIDErr       error
	updateErr         error
	createRTErr       error
	findRTErr         error
	revokeRTErr       error
	revokeAllRTErr    error
	updateLastLoginErr error
}

func newMockRepo() *mockUserRepo {
	return &mockUserRepo{
		users:         make(map[string]*entity.User),
		usersByID:     make(map[uuid.UUID]*entity.User),
		refreshTokens: make(map[string]*entity.RefreshToken),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *entity.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.users[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	if m.findByEmailErr != nil {
		return nil, m.findByEmailErr
	}
	u, ok := m.users[email]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	u, ok := m.usersByID[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *entity.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.users[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *mockUserRepo) FindByVerificationToken(ctx context.Context, token string) (*entity.User, error) {
	for _, u := range m.users {
		if u.EmailVerificationToken != nil && *u.EmailVerificationToken == token {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) FindByResetToken(ctx context.Context, token string) (*entity.User, error) {
	for _, u := range m.users {
		if u.PasswordResetToken != nil && *u.PasswordResetToken == token {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return m.updateLastLoginErr
}

func (m *mockUserRepo) CreateRefreshToken(ctx context.Context, rt *entity.RefreshToken) error {
	if m.createRTErr != nil {
		return m.createRTErr
	}
	m.refreshTokens[rt.TokenHash] = rt
	return nil
}

func (m *mockUserRepo) FindRefreshToken(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	if m.findRTErr != nil {
		return nil, m.findRTErr
	}
	rt, ok := m.refreshTokens[tokenHash]
	if !ok {
		return nil, nil
	}
	return rt, nil
}

func (m *mockUserRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	if m.revokeRTErr != nil {
		return m.revokeRTErr
	}
	if rt, ok := m.refreshTokens[tokenHash]; ok {
		rt.Revoked = true
	}
	return nil
}

func (m *mockUserRepo) RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	if m.revokeAllRTErr != nil {
		return m.revokeAllRTErr
	}
	for _, rt := range m.refreshTokens {
		if rt.UserID == userID {
			rt.Revoked = true
		}
	}
	return nil
}

// --- Test helpers ---

func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key-that-is-long-enough",
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 720 * time.Hour,
		},
	}
}

func newMiniRedis() *redis.Client {
	// Use a real Redis client pointed at a non-existent server.
	// Operations will fail gracefully (blacklist checks are non-fatal).
	return redis.NewClient(&redis.Options{Addr: "localhost:6399"})
}

func newAuthUseCase(repo *mockUserRepo) usecase.AuthUseCase {
	return usecase.NewAuthUseCase(repo, newMiniRedis(), testConfig(), zap.NewNop())
}

// --- Tests ---

func TestRegister_Success(t *testing.T) {
	repo := newMockRepo()
	uc := newAuthUseCase(repo)

	req := usecase.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "securepassword",
	}

	resp, err := uc.Register(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, req.Email, resp.User.Email)
	assert.Equal(t, req.Name, resp.User.Name)
	assert.Equal(t, "free", resp.User.Plan)
	assert.False(t, resp.User.EmailVerified)
	assert.NotEqual(t, req.Password, resp.User.PasswordHash) // must be hashed
}

func TestRegister_EmailDuplicate(t *testing.T) {
	repo := newMockRepo()
	uc := newAuthUseCase(repo)

	req := usecase.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "securepassword",
	}

	_, err := uc.Register(context.Background(), req)
	require.NoError(t, err)

	_, err = uc.Register(context.Background(), req)
	require.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrEmailAlreadyExists))
}

func TestLogin_Success(t *testing.T) {
	repo := newMockRepo()
	uc := newAuthUseCase(repo)

	password := "correctpassword"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	userID := uuid.New()
	user := &entity.User{
		ID:           userID,
		Email:        "bob@example.com",
		PasswordHash: string(hash),
		Name:         "Bob",
		Plan:         "free",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.users[user.Email] = user
	repo.usersByID[userID] = user

	req := usecase.LoginRequest{
		Email:    "bob@example.com",
		Password: password,
	}

	resp, err := uc.Login(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, user.Email, resp.User.Email)
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockRepo()
	uc := newAuthUseCase(repo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), 12)
	userID := uuid.New()
	user := &entity.User{
		ID:           userID,
		Email:        "carol@example.com",
		PasswordHash: string(hash),
		Name:         "Carol",
		Plan:         "free",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.users[user.Email] = user
	repo.usersByID[userID] = user

	req := usecase.LoginRequest{
		Email:    "carol@example.com",
		Password: "wrongpassword",
	}

	_, err := uc.Login(context.Background(), req)
	require.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidCredentials))
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockRepo()
	uc := newAuthUseCase(repo)

	req := usecase.LoginRequest{
		Email:    "nobody@example.com",
		Password: "somepassword",
	}

	_, err := uc.Login(context.Background(), req)
	require.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidCredentials))
}
