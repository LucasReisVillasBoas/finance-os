package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
)

// FamilyUseCase defines business logic for family groups.
type FamilyUseCase interface {
	CreateGroup(ctx context.Context, ownerID uuid.UUID, name string) (*entity.FamilyGroup, error)
	GetGroup(ctx context.Context, userID uuid.UUID) (*entity.FamilyGroup, error)
	GenerateInvite(ctx context.Context, ownerID uuid.UUID) (string, error)
	JoinGroup(ctx context.Context, userID uuid.UUID, inviteCode string) (*entity.FamilyGroup, error)
	RemoveMember(ctx context.Context, memberID, ownerID uuid.UUID) error
	GetMembers(ctx context.Context, userID uuid.UUID) ([]*entity.FamilyMember, error)
	GetDashboard(ctx context.Context, userID uuid.UUID) (*domainrepo.FamilyDashboard, error)
}

type familyUseCase struct {
	repo domainrepo.FamilyRepository
}

// NewFamilyUseCase creates a new FamilyUseCase.
func NewFamilyUseCase(repo domainrepo.FamilyRepository) FamilyUseCase {
	return &familyUseCase{repo: repo}
}

const inviteCodeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateInviteCode() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	b := make([]byte, 8)
	for i := range b {
		b[i] = inviteCodeChars[rng.Intn(len(inviteCodeChars))]
	}
	return string(b)
}

func (uc *familyUseCase) CreateGroup(ctx context.Context, ownerID uuid.UUID, name string) (*entity.FamilyGroup, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("familyUseCase.CreateGroup: name is required")
	}

	// Check if user already has a group
	existing, err := uc.repo.FindGroupByUserID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("familyUseCase.CreateGroup check existing: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("familyUseCase.CreateGroup: user already belongs to a group")
	}

	g := &entity.FamilyGroup{
		ID:         uuid.New(),
		Name:       name,
		OwnerID:    ownerID,
		InviteCode: generateInviteCode(),
		CreatedAt:  time.Now(),
	}
	if err := uc.repo.CreateGroup(ctx, g); err != nil {
		return nil, fmt.Errorf("familyUseCase.CreateGroup: %w", err)
	}
	return g, nil
}

func (uc *familyUseCase) GetGroup(ctx context.Context, userID uuid.UUID) (*entity.FamilyGroup, error) {
	g, err := uc.repo.FindGroupByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("familyUseCase.GetGroup: %w", err)
	}
	return g, nil
}

func (uc *familyUseCase) GenerateInvite(ctx context.Context, ownerID uuid.UUID) (string, error) {
	g, err := uc.repo.FindGroupByUserID(ctx, ownerID)
	if err != nil {
		return "", fmt.Errorf("familyUseCase.GenerateInvite: %w", err)
	}
	if g == nil {
		return "", fmt.Errorf("familyUseCase.GenerateInvite: no group found")
	}
	if g.OwnerID != ownerID {
		return "", fmt.Errorf("familyUseCase.GenerateInvite: only the owner can regenerate the invite code")
	}
	return g.InviteCode, nil
}

func (uc *familyUseCase) JoinGroup(ctx context.Context, userID uuid.UUID, inviteCode string) (*entity.FamilyGroup, error) {
	inviteCode = strings.TrimSpace(strings.ToUpper(inviteCode))
	if inviteCode == "" {
		return nil, fmt.Errorf("familyUseCase.JoinGroup: invite_code is required")
	}

	// Check if user already has a group
	existing, err := uc.repo.FindGroupByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("familyUseCase.JoinGroup check existing: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("familyUseCase.JoinGroup: user already belongs to a group")
	}

	g, err := uc.repo.FindGroupByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, fmt.Errorf("familyUseCase.JoinGroup: %w", err)
	}
	if g == nil {
		return nil, fmt.Errorf("familyUseCase.JoinGroup: invalid invite code")
	}

	m := &entity.FamilyMember{
		ID:          uuid.New(),
		GroupID:     g.ID,
		UserID:      userID,
		Permissions: map[string]interface{}{"view": true},
		JoinedAt:    time.Now(),
	}
	if err := uc.repo.AddMember(ctx, m); err != nil {
		return nil, fmt.Errorf("familyUseCase.JoinGroup add member: %w", err)
	}
	return g, nil
}

func (uc *familyUseCase) RemoveMember(ctx context.Context, memberID, ownerID uuid.UUID) error {
	if err := uc.repo.RemoveMember(ctx, memberID, ownerID); err != nil {
		return fmt.Errorf("familyUseCase.RemoveMember: %w", err)
	}
	return nil
}

func (uc *familyUseCase) GetMembers(ctx context.Context, userID uuid.UUID) ([]*entity.FamilyMember, error) {
	g, err := uc.repo.FindGroupByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("familyUseCase.GetMembers: %w", err)
	}
	if g == nil {
		return []*entity.FamilyMember{}, nil
	}
	members, err := uc.repo.GetMembers(ctx, g.ID)
	if err != nil {
		return nil, fmt.Errorf("familyUseCase.GetMembers: %w", err)
	}
	return members, nil
}

func (uc *familyUseCase) GetDashboard(ctx context.Context, userID uuid.UUID) (*domainrepo.FamilyDashboard, error) {
	g, err := uc.repo.FindGroupByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("familyUseCase.GetDashboard: %w", err)
	}
	if g == nil {
		return nil, fmt.Errorf("familyUseCase.GetDashboard: user does not belong to any group")
	}

	now := time.Now()
	dashboard, err := uc.repo.GetDashboard(ctx, g.ID, int(now.Month()), now.Year())
	if err != nil {
		return nil, fmt.Errorf("familyUseCase.GetDashboard: %w", err)
	}
	return dashboard, nil
}
