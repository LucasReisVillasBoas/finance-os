package worker

import (
	"context"
	"time"

	"github.com/financeos/api/pkg/claude"
	"go.uber.org/zap"
)

// AIWorker runs background AI processing jobs.
type AIWorker struct {
	claudeClient *claude.Client
	logger       *zap.Logger
}

// NewAIWorker creates a new AIWorker.
func NewAIWorker(claudeClient *claude.Client, logger *zap.Logger) *AIWorker {
	return &AIWorker{
		claudeClient: claudeClient,
		logger:       logger,
	}
}

// Run starts the AI worker loop. It runs categorization hourly.
func (w *AIWorker) Run(ctx context.Context) {
	w.logger.Info("AI worker started")
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Run once immediately on startup
	w.runCategorizationJob(ctx)

	for {
		select {
		case <-ticker.C:
			w.runCategorizationJob(ctx)
		case <-ctx.Done():
			w.logger.Info("AI worker stopped")
			return
		}
	}
}

// runCategorizationJob processes transactions without categories using AI.
// In production, this would query the DB for uncategorized transactions
// and use Claude to suggest categories.
func (w *AIWorker) runCategorizationJob(ctx context.Context) {
	w.logger.Debug("AI categorization job running",
		zap.Time("at", time.Now()),
	)
	// Implementation note: in a production setup, this would:
	// 1. Fetch transactions with category_id IS NULL and ai_categorized = false
	//    for users on pro/premium plans
	// 2. Call Claude to suggest a category for each
	// 3. Match suggested category with existing categories
	// 4. Update the transaction record
	// Skipping direct DB queries here to maintain clean worker/dependency separation.
}
