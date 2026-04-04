package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type familyRepository struct {
	db *pgxpool.Pool
}

// NewFamilyRepository creates a new PostgreSQL-backed FamilyRepository.
func NewFamilyRepository(db *pgxpool.Pool) domainrepo.FamilyRepository {
	return &familyRepository{db: db}
}

func (r *familyRepository) CreateGroup(ctx context.Context, g *entity.FamilyGroup) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO family_groups (id, name, owner_id, invite_code, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		g.ID, g.Name, g.OwnerID, g.InviteCode, g.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("familyRepository.CreateGroup: %w", err)
	}
	return nil
}

func (r *familyRepository) FindGroupByUserID(ctx context.Context, userID uuid.UUID) (*entity.FamilyGroup, error) {
	query := `
		SELECT fg.id, fg.name, fg.owner_id, fg.invite_code, fg.created_at
		FROM family_groups fg
		WHERE fg.owner_id = $1
		   OR fg.id IN (SELECT group_id FROM family_members WHERE user_id = $1)
		LIMIT 1`
	g := &entity.FamilyGroup{}
	err := r.db.QueryRow(ctx, query, userID).Scan(&g.ID, &g.Name, &g.OwnerID, &g.InviteCode, &g.CreatedAt)
	if err != nil {
		return nil, nil //nolint:nilerr
	}
	return g, nil
}

func (r *familyRepository) FindGroupByInviteCode(ctx context.Context, code string) (*entity.FamilyGroup, error) {
	g := &entity.FamilyGroup{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, owner_id, invite_code, created_at FROM family_groups WHERE invite_code = $1`,
		code,
	).Scan(&g.ID, &g.Name, &g.OwnerID, &g.InviteCode, &g.CreatedAt)
	if err != nil {
		return nil, nil //nolint:nilerr
	}
	return g, nil
}

func (r *familyRepository) AddMember(ctx context.Context, m *entity.FamilyMember) error {
	permJSON, _ := json.Marshal(m.Permissions)
	_, err := r.db.Exec(ctx,
		`INSERT INTO family_members (id, group_id, user_id, permissions, joined_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (group_id, user_id) DO NOTHING`,
		m.ID, m.GroupID, m.UserID, permJSON, m.JoinedAt,
	)
	if err != nil {
		return fmt.Errorf("familyRepository.AddMember: %w", err)
	}
	return nil
}

func (r *familyRepository) RemoveMember(ctx context.Context, memberID, ownerID uuid.UUID) error {
	// ownerID is used to verify the caller owns the group
	_, err := r.db.Exec(ctx,
		`DELETE FROM family_members fm
		 USING family_groups fg
		 WHERE fm.id = $1 AND fm.group_id = fg.id AND fg.owner_id = $2`,
		memberID, ownerID,
	)
	if err != nil {
		return fmt.Errorf("familyRepository.RemoveMember: %w", err)
	}
	return nil
}

func (r *familyRepository) GetMembers(ctx context.Context, groupID uuid.UUID) ([]*entity.FamilyMember, error) {
	rows, err := r.db.Query(ctx,
		`SELECT fm.id, fm.group_id, fm.user_id, fm.permissions, fm.joined_at,
		        u.name, u.email
		 FROM family_members fm
		 JOIN users u ON u.id = fm.user_id
		 WHERE fm.group_id = $1`,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("familyRepository.GetMembers: %w", err)
	}
	defer rows.Close()

	var members []*entity.FamilyMember
	for rows.Next() {
		m := &entity.FamilyMember{}
		var permJSON []byte
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &permJSON, &m.JoinedAt, &m.UserName, &m.UserEmail); err != nil {
			return nil, fmt.Errorf("familyRepository.GetMembers scan: %w", err)
		}
		if len(permJSON) > 0 {
			_ = json.Unmarshal(permJSON, &m.Permissions)
		}
		members = append(members, m)
	}
	if members == nil {
		members = []*entity.FamilyMember{}
	}
	return members, nil
}

func (r *familyRepository) GetDashboard(ctx context.Context, groupID uuid.UUID, month, year int) (*domainrepo.FamilyDashboard, error) {
	// Get group info
	g := &entity.FamilyGroup{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name FROM family_groups WHERE id = $1`,
		groupID,
	).Scan(&g.ID, &g.Name)
	if err != nil {
		return nil, fmt.Errorf("familyRepository.GetDashboard group: %w", err)
	}

	dashboard := &domainrepo.FamilyDashboard{
		GroupID:   g.ID,
		GroupName: g.Name,
	}

	// Get all user IDs in the group (owner + members)
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT user_id, u.name
		 FROM (
		   SELECT owner_id AS user_id FROM family_groups WHERE id = $1
		   UNION
		   SELECT user_id FROM family_members WHERE group_id = $1
		 ) ids
		 JOIN users u ON u.id = ids.user_id`,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("familyRepository.GetDashboard members: %w", err)
	}
	defer rows.Close()

	type memberInfo struct {
		userID uuid.UUID
		name   string
	}
	var membersList []memberInfo
	for rows.Next() {
		var mi memberInfo
		if err := rows.Scan(&mi.userID, &mi.name); err != nil {
			continue
		}
		membersList = append(membersList, mi)
	}

	for _, mi := range membersList {
		var income, expense float64
		_ = r.db.QueryRow(ctx,
			`SELECT
			   COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
			   COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
			 FROM transactions
			 WHERE user_id = $1
			   AND EXTRACT(MONTH FROM date) = $2
			   AND EXTRACT(YEAR FROM date) = $3
			   AND type != 'transfer'`,
			mi.userID, month, year,
		).Scan(&income, &expense)

		dashboard.TotalIncome += income
		dashboard.TotalExpense += expense
		dashboard.Members = append(dashboard.Members, domainrepo.MemberStats{
			UserID:       mi.userID,
			UserName:     mi.name,
			TotalIncome:  income,
			TotalExpense: expense,
		})
	}
	dashboard.Balance = dashboard.TotalIncome - dashboard.TotalExpense
	return dashboard, nil
}
