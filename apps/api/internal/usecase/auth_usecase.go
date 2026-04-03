package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/financeos/api/pkg/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors
var (
	ErrEmailAlreadyExists  = errors.New("email already in use")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrTokenRevoked        = errors.New("token has been revoked")
)

// RegisterRequest holds the data needed to register a new user.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=255"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest holds the data needed to authenticate a user.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is the response returned after successful authentication.
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         entity.User `json:"user"`
}

// Claims are the JWT claims used in access tokens.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Plan   string `json:"plan"`
	jwt.RegisteredClaims
}

// AuthUseCase defines the authentication business logic interface.
type AuthUseCase interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, rawRefreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, rawAccessToken string, rawRefreshToken string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token string, newPassword string) error
}

type authUseCase struct {
	repo   domainrepo.UserRepository
	rdb    *redis.Client
	cfg    *config.Config
	logger *zap.Logger
}

// NewAuthUseCase creates a new AuthUseCase implementation.
func NewAuthUseCase(repo domainrepo.UserRepository, rdb *redis.Client, cfg *config.Config, logger *zap.Logger) AuthUseCase {
	return &authUseCase{
		repo:   repo,
		rdb:    rdb,
		cfg:    cfg,
		logger: logger,
	}
}

func (uc *authUseCase) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	existing, err := uc.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("authUseCase.Register find email: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("authUseCase.Register bcrypt: %w", err)
	}

	verificationToken, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("authUseCase.Register generate token: %w", err)
	}

	now := time.Now()
	user := &entity.User{
		ID:                     uuid.New(),
		Email:                  req.Email,
		PasswordHash:           string(hash),
		Name:                   req.Name,
		Plan:                   "free",
		EmailVerified:          false,
		EmailVerificationToken: &verificationToken,
		Timezone:               "UTC",
		Currency:               "BRL",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("authUseCase.Register create user: %w", err)
	}

	accessToken, refreshToken, rt, err := uc.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("authUseCase.Register generate tokens: %w", err)
	}

	if err := uc.repo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, fmt.Errorf("authUseCase.Register save refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (uc *authUseCase) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := uc.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("authUseCase.Login find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, refreshToken, rt, err := uc.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("authUseCase.Login generate tokens: %w", err)
	}

	if err := uc.repo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, fmt.Errorf("authUseCase.Login save refresh token: %w", err)
	}

	if err := uc.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		uc.logger.Warn("failed to update last login", zap.Error(err))
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (uc *authUseCase) RefreshToken(ctx context.Context, rawRefreshToken string) (*AuthResponse, error) {
	tokenHash := hashToken(rawRefreshToken)

	rt, err := uc.repo.FindRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("authUseCase.RefreshToken find token: %w", err)
	}
	if rt == nil || rt.Revoked || time.Now().After(rt.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	if err := uc.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("authUseCase.RefreshToken revoke token: %w", err)
	}

	user, err := uc.repo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("authUseCase.RefreshToken find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	accessToken, refreshToken, newRT, err := uc.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("authUseCase.RefreshToken generate tokens: %w", err)
	}

	if err := uc.repo.CreateRefreshToken(ctx, newRT); err != nil {
		return nil, fmt.Errorf("authUseCase.RefreshToken save refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (uc *authUseCase) Logout(ctx context.Context, rawAccessToken string, rawRefreshToken string) error {
	// Parse the access token to get its remaining TTL
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(rawAccessToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(uc.cfg.JWT.Secret), nil
	})

	if err == nil && token.Valid {
		remaining := time.Until(claims.ExpiresAt.Time)
		if remaining > 0 {
			accessHash := hashToken(rawAccessToken)
			key := "blacklist:" + accessHash
			if setErr := uc.rdb.Set(ctx, key, "1", remaining).Err(); setErr != nil {
				uc.logger.Warn("failed to blacklist access token", zap.Error(setErr))
			}
		}
	}

	// Revoke the refresh token
	refreshHash := hashToken(rawRefreshToken)
	if err := uc.repo.RevokeRefreshToken(ctx, refreshHash); err != nil {
		return fmt.Errorf("authUseCase.Logout revoke refresh token: %w", err)
	}

	return nil
}

func (uc *authUseCase) ForgotPassword(ctx context.Context, email string) error {
	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("authUseCase.ForgotPassword find user: %w", err)
	}
	// Don't reveal whether the email exists
	if user == nil {
		return nil
	}

	rawToken, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("authUseCase.ForgotPassword generate token: %w", err)
	}

	tokenHash := hashToken(rawToken)
	expiry := time.Now().Add(1 * time.Hour)
	user.PasswordResetToken = &tokenHash
	user.PasswordResetExpiresAt = &expiry

	if err := uc.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("authUseCase.ForgotPassword update user: %w", err)
	}

	// In a real app, send an email. For now, log it.
	uc.logger.Info("mock email sent",
		zap.String("to", email),
		zap.String("token", rawToken),
	)

	return nil
}

func (uc *authUseCase) ResetPassword(ctx context.Context, rawToken string, newPassword string) error {
	tokenHash := hashToken(rawToken)

	user, err := uc.repo.FindByResetToken(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("authUseCase.ResetPassword find user: %w", err)
	}
	if user == nil {
		return ErrInvalidToken
	}
	if user.PasswordResetExpiresAt == nil || time.Now().After(*user.PasswordResetExpiresAt) {
		return ErrInvalidToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("authUseCase.ResetPassword bcrypt: %w", err)
	}

	user.PasswordHash = string(hash)
	user.PasswordResetToken = nil
	user.PasswordResetExpiresAt = nil

	if err := uc.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("authUseCase.ResetPassword update user: %w", err)
	}

	if err := uc.repo.RevokeAllUserRefreshTokens(ctx, user.ID); err != nil {
		return fmt.Errorf("authUseCase.ResetPassword revoke tokens: %w", err)
	}

	return nil
}

// generateTokenPair creates an access JWT, a raw refresh token string, and the RefreshToken entity.
func (uc *authUseCase) generateTokenPair(user *entity.User) (accessToken string, rawRefreshToken string, rt *entity.RefreshToken, err error) {
	now := time.Now()

	claims := Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Plan:   user.Plan,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(uc.cfg.JWT.AccessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString([]byte(uc.cfg.JWT.Secret))
	if err != nil {
		return "", "", nil, fmt.Errorf("sign access token: %w", err)
	}

	rawBytes := make([]byte, 32)
	if _, err = rand.Read(rawBytes); err != nil {
		return "", "", nil, fmt.Errorf("generate refresh token bytes: %w", err)
	}
	rawRefreshToken = hex.EncodeToString(rawBytes)
	tokenHash := hashToken(rawRefreshToken)

	rt = &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(uc.cfg.JWT.RefreshTTL),
		Revoked:   false,
		CreatedAt: now,
	}

	return accessToken, rawRefreshToken, rt, nil
}

// generateSecureToken creates a cryptographically secure random hex token.
func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken returns the SHA256 hex digest of a token string.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
