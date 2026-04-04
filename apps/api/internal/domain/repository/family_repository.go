package repository

import (
	"context"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

// FamilyDashboard holds aggregated family financial summary.
type FamilyDashboard struct {
	GroupID      uuid.UUID     `json:"group_id"`
	GroupName    string        `json:"group_name"`
	TotalIncome  float64       `json:"total_income"`
	TotalExpense float64       `json:"total_expense"`
	Balance      float64       `json:"balance"`
	Members      []MemberStats `json:"members"`
}

// MemberStats holds per-member aggregate data.
type MemberStats struct {
	UserID       uuid.UUID `json:"user_id"`
	UserName     string    `json:"user_name"`
	TotalIncome  float64   `json:"total_income"`
	TotalExpense float64   `json:"total_expense"`
}

// FamilyRepository defines data access for family groups and members.
type FamilyRepository interface {
	CreateGroup(ctx context.Context, g *entity.FamilyGroup) error
	FindGroupByUserID(ctx context.Context, userID uuid.UUID) (*entity.FamilyGroup, error)
	FindGroupByInviteCode(ctx context.Context, code string) (*entity.FamilyGroup, error)
	AddMember(ctx context.Context, m *entity.FamilyMember) error
	RemoveMember(ctx context.Context, memberID, ownerID uuid.UUID) error
	GetMembers(ctx context.Context, groupID uuid.UUID) ([]*entity.FamilyMember, error)
	GetDashboard(ctx context.Context, groupID uuid.UUID, month, year int) (*FamilyDashboard, error)
}
