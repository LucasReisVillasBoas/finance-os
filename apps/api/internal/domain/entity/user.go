package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	Email                  string     `json:"email" db:"email"`
	PasswordHash           string     `json:"-" db:"password_hash"`
	Name                   string     `json:"name" db:"name"`
	AvatarURL              *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Plan                   string     `json:"plan" db:"plan"`
	PlanExpiresAt          *time.Time `json:"plan_expires_at,omitempty" db:"plan_expires_at"`
	EmailVerified          bool       `json:"email_verified" db:"email_verified"`
	EmailVerificationToken *string    `json:"-" db:"email_verification_token"`
	PasswordResetToken     *string    `json:"-" db:"password_reset_token"`
	PasswordResetExpiresAt *time.Time `json:"-" db:"password_reset_expires_at"`
	Timezone               string     `json:"timezone" db:"timezone"`
	Currency               string     `json:"currency" db:"currency"`
	LastLoginAt            *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

type RefreshToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	Revoked   bool      `db:"revoked"`
	CreatedAt time.Time `db:"created_at"`
}
