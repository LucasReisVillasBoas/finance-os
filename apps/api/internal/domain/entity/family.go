package entity

import (
	"time"

	"github.com/google/uuid"
)

// FamilyGroup represents a shared financial group.
type FamilyGroup struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	OwnerID    uuid.UUID `json:"owner_id" db:"owner_id"`
	InviteCode string    `json:"invite_code" db:"invite_code"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// FamilyMember links a user to a family group.
type FamilyMember struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	GroupID     uuid.UUID              `json:"group_id" db:"group_id"`
	UserID      uuid.UUID              `json:"user_id" db:"user_id"`
	Permissions map[string]interface{} `json:"permissions" db:"permissions"`
	JoinedAt    time.Time              `json:"joined_at" db:"joined_at"`
	// Joined fields
	UserName  string `json:"user_name,omitempty" db:"user_name"`
	UserEmail string `json:"user_email,omitempty" db:"user_email"`
}
