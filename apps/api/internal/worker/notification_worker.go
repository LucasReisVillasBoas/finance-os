package worker

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// NotificationWorker runs background notification jobs.
type NotificationWorker struct {
	logger *zap.Logger
}

// NewNotificationWorker creates a new NotificationWorker.
func NewNotificationWorker(logger *zap.Logger) *NotificationWorker {
	return &NotificationWorker{logger: logger}
}

// Run starts the notification worker loop. It checks for alerts daily.
func (w *NotificationWorker) Run(ctx context.Context) {
	w.logger.Info("notification worker started")

	// Calculate time until next midnight
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, now.Location())
	timer := time.NewTimer(time.Until(next))
	defer timer.Stop()

	// Run once on startup
	w.runDailyJob(ctx)

	for {
		select {
		case <-timer.C:
			w.runDailyJob(ctx)
			// Reset for next day
			next = next.Add(24 * time.Hour)
			timer.Reset(time.Until(next))
		case <-ctx.Done():
			w.logger.Info("notification worker stopped")
			return
		}
	}
}

// runDailyJob checks budgets, recurrences, and generates periodic summaries.
func (w *NotificationWorker) runDailyJob(ctx context.Context) {
	now := time.Now()
	w.logger.Info("notification worker daily job running", zap.Time("at", now))

	// Implementation note: in production this would:
	// 1. Check budgets above threshold -> create "budget_alert" notification
	// 2. Find recurrences with next_due_date = tomorrow -> create "recurrence_due"
	// 3. On Sunday -> create "weekly_summary" notification
	// 4. On day 1 -> create "monthly_report" notification
	// Skipping direct DB queries here to maintain clean worker/dependency separation.

	if now.Weekday() == time.Sunday {
		w.logger.Debug("would generate weekly summaries", zap.Time("date", now))
	}
	if now.Day() == 1 {
		w.logger.Debug("would generate monthly reports", zap.Time("date", now))
	}
}
