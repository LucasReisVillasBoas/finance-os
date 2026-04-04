package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/financeos/api/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Fake RecurrenceRepository ---

type fakeRecurrenceRepo struct {
	items map[uuid.UUID]*entity.Recurrence
}

func newFakeRecurrenceRepo() *fakeRecurrenceRepo {
	return &fakeRecurrenceRepo{items: make(map[uuid.UUID]*entity.Recurrence)}
}

func (r *fakeRecurrenceRepo) Create(ctx context.Context, rec *entity.Recurrence) error {
	r.items[rec.ID] = rec
	return nil
}

func (r *fakeRecurrenceRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Recurrence, error) {
	rec, ok := r.items[id]
	if !ok || rec.UserID != userID {
		return nil, nil
	}
	return rec, nil
}

func (r *fakeRecurrenceRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Recurrence, error) {
	var result []*entity.Recurrence
	for _, rec := range r.items {
		if rec.UserID == userID {
			result = append(result, rec)
		}
	}
	return result, nil
}

func (r *fakeRecurrenceRepo) FindDue(ctx context.Context, before time.Time) ([]*entity.Recurrence, error) {
	var result []*entity.Recurrence
	for _, rec := range r.items {
		if rec.IsActive && !rec.NextDueDate.After(before) {
			result = append(result, rec)
		}
	}
	return result, nil
}

func (r *fakeRecurrenceRepo) Update(ctx context.Context, rec *entity.Recurrence) error {
	r.items[rec.ID] = rec
	return nil
}

func (r *fakeRecurrenceRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	delete(r.items, id)
	return nil
}

func (r *fakeRecurrenceRepo) UpdateNextDueDate(ctx context.Context, id uuid.UUID, nextDate time.Time) error {
	if rec, ok := r.items[id]; ok {
		rec.NextDueDate = nextDate
	}
	return nil
}

// --- Tests ---

func TestCalculateNextDueDate(t *testing.T) {
	base := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		frequency string
		expected  time.Time
	}{
		{"daily", "daily", time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)},
		{"weekly", "weekly", time.Date(2026, 1, 22, 0, 0, 0, 0, time.UTC)},
		{"biweekly", "biweekly", time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)},
		{"monthly", "monthly", time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)},
		{"yearly", "yearly", time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := usecase.CalculateNextDueDate(base, tt.frequency)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNextDueDate_DefaultFrequency(t *testing.T) {
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	result := usecase.CalculateNextDueDate(base, "unknown")
	expected := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, result)
}

func TestCreateRecurrence_Success(t *testing.T) {
	repo := newFakeRecurrenceRepo()
	uc := usecase.NewRecurrenceUseCase(repo)

	userID := uuid.New()
	accountID := uuid.New()
	desc := "Monthly rent"
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	req := usecase.CreateRecurrenceRequest{
		AccountID:   accountID,
		Type:        "expense",
		Amount:      1500.0,
		Description: &desc,
		Frequency:   "monthly",
		StartDate:   startDate,
		AutoLaunch:  true,
	}

	rec, err := uc.Create(context.Background(), userID, req)
	require.NoError(t, err)
	require.NotNil(t, rec)

	assert.Equal(t, userID, rec.UserID)
	assert.Equal(t, accountID, rec.AccountID)
	assert.Equal(t, "expense", rec.Type)
	assert.Equal(t, 1500.0, rec.Amount)
	assert.Equal(t, "monthly", rec.Frequency)
	assert.Equal(t, startDate, rec.StartDate)
	assert.Equal(t, startDate, rec.NextDueDate)
	assert.True(t, rec.AutoLaunch)
	assert.True(t, rec.IsActive)
	assert.NotEqual(t, uuid.Nil, rec.ID)
}

func TestCreateRecurrence_StoredInRepo(t *testing.T) {
	repo := newFakeRecurrenceRepo()
	uc := usecase.NewRecurrenceUseCase(repo)

	userID := uuid.New()
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	req := usecase.CreateRecurrenceRequest{
		AccountID: uuid.New(),
		Type:      "income",
		Amount:    500.0,
		Frequency: "weekly",
		StartDate: startDate,
	}

	rec, err := uc.Create(context.Background(), userID, req)
	require.NoError(t, err)

	stored, err := uc.GetByID(context.Background(), rec.ID, userID)
	require.NoError(t, err)
	require.NotNil(t, stored)
	assert.Equal(t, rec.ID, stored.ID)
}

func TestDeleteRecurrence_NotFound(t *testing.T) {
	repo := newFakeRecurrenceRepo()
	uc := usecase.NewRecurrenceUseCase(repo)

	err := uc.Delete(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, usecase.ErrRecurrenceNotFound)
}
