package worker

import (
	"context"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/financeos/api/internal/usecase"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RecurrenceWorker processes due recurrences on a schedule.
type RecurrenceWorker struct {
	recurrenceRepo  domainrepo.RecurrenceRepository
	transactionRepo domainrepo.TransactionRepository
	logger          *zap.Logger
}

// NewRecurrenceWorker creates a new RecurrenceWorker.
func NewRecurrenceWorker(rr domainrepo.RecurrenceRepository, tr domainrepo.TransactionRepository, l *zap.Logger) *RecurrenceWorker {
	return &RecurrenceWorker{recurrenceRepo: rr, transactionRepo: tr, logger: l}
}

// Run starts the worker loop, processing due recurrences every hour.
func (w *RecurrenceWorker) Run(ctx context.Context) {
	w.processRecurrences(ctx)
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.processRecurrences(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (w *RecurrenceWorker) processRecurrences(ctx context.Context) {
	recurrences, err := w.recurrenceRepo.FindDue(ctx, time.Now())
	if err != nil {
		w.logger.Error("find due recurrences", zap.Error(err))
		return
	}

	for _, r := range recurrences {
		if r.AutoLaunch {
			tx := &entity.Transaction{
				ID:           uuid.New(),
				UserID:       r.UserID,
				AccountID:    r.AccountID,
				CategoryID:   r.CategoryID,
				Type:         r.Type,
				Amount:       r.Amount,
				Description:  r.Description,
				Date:         r.NextDueDate,
				RecurrenceID: &r.ID,
				Tags:         []string{},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			if err := w.transactionRepo.Create(ctx, tx); err != nil {
				w.logger.Error("auto-launch transaction",
					zap.Error(err),
					zap.String("recurrence_id", r.ID.String()),
				)
				continue
			}
		} else {
			w.logger.Info("recurrence due (manual launch required)",
				zap.String("id", r.ID.String()),
				zap.String("description", func() string {
					if r.Description != nil {
						return *r.Description
					}
					return ""
				}()),
			)
		}

		next := usecase.CalculateNextDueDate(r.NextDueDate, r.Frequency)
		if err := w.recurrenceRepo.UpdateNextDueDate(ctx, r.ID, next); err != nil {
			w.logger.Error("update next due date", zap.Error(err))
		}
	}

	if len(recurrences) > 0 {
		w.logger.Info("processed recurrences", zap.Int("count", len(recurrences)))
	}
}
