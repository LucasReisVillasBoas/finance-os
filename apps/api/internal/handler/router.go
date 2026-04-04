package handler

import (
	"net/http"
	"time"

	"github.com/financeos/api/internal/handler/middleware"
	"github.com/financeos/api/internal/repository"
	"github.com/financeos/api/internal/usecase"
	"github.com/financeos/api/pkg/claude"
	"github.com/financeos/api/pkg/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// SetupRouter configures and returns the Gin engine with all routes registered.
func SetupRouter(cfg *config.Config, db *pgxpool.Pool, rdb *redis.Client, logger *zap.Logger) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// CORS
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "financeos-api",
			"env":     cfg.App.Env,
		})
	})

	// Dependencies
	userRepo := repository.NewUserRepository(db)
	authUC := usecase.NewAuthUseCase(userRepo, rdb, cfg, logger)
	authH := NewAuthHandler(authUC, logger)

	accountRepo := repository.NewAccountRepository(db)
	accountUC := usecase.NewAccountUseCase(accountRepo)
	accountH := NewAccountHandler(accountUC, logger)

	categoryRepo := repository.NewCategoryRepository(db)
	categoryUC := usecase.NewCategoryUseCase(categoryRepo)
	categoryH := NewCategoryHandler(categoryUC, logger)

	transactionRepo := repository.NewTransactionRepository(db)
	transactionUC := usecase.NewTransactionUseCase(transactionRepo)
	transactionH := NewTransactionHandler(transactionUC, logger)

	recurrenceRepo := repository.NewRecurrenceRepository(db)
	recurrenceUC := usecase.NewRecurrenceUseCase(recurrenceRepo)
	recurrenceH := NewRecurrenceHandler(recurrenceUC, logger)

	budgetRepo := repository.NewBudgetRepository(db)
	budgetUC := usecase.NewBudgetUseCase(budgetRepo)
	budgetH := NewBudgetHandler(budgetUC, logger)

	dashboardRepo := repository.NewDashboardRepository(db)
	dashboardUC := usecase.NewDashboardUseCase(dashboardRepo)
	dashboardH := NewDashboardHandler(dashboardUC, logger)

	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	investTxRepo := repository.NewInvestmentTransactionRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	customAssetRepo := repository.NewCustomAssetRepository(db)
	investmentUC := usecase.NewInvestmentUseCase(portfolioRepo, holdingRepo, investTxRepo, assetRepo, customAssetRepo)
	investmentH := NewInvestmentHandler(investmentUC, logger)

	goalRepo := repository.NewGoalRepository(db)
	goalUC := usecase.NewGoalUseCase(goalRepo)
	goalH := NewGoalHandler(goalUC, logger)

	importUC := usecase.NewImportUseCase(transactionRepo)
	importH := NewImportHandler(importUC, logger)

	whatsappRepo := repository.NewWhatsAppRepository(db)
	whatsappUC := usecase.NewWhatsAppUseCase(whatsappRepo, transactionRepo, accountRepo)
	webhookH := NewWebhookHandler(whatsappUC, cfg.Evolution.APIURL, cfg.Evolution.APIKey, logger)

	claudeClient := claude.New(cfg.Claude.APIKey, cfg.Claude.Model)
	aiUC := usecase.NewAIUseCase(claudeClient, transactionRepo, rdb)
	aiH := NewAIHandler(aiUC, logger)

	notificationRepo := repository.NewNotificationRepository(db)
	notificationUC := usecase.NewNotificationUseCase(notificationRepo)
	notificationH := NewNotificationHandler(notificationUC, logger)

	familyRepo := repository.NewFamilyRepository(db)
	familyUC := usecase.NewFamilyUseCase(familyRepo)
	familyH := NewFamilyHandler(familyUC, logger)

	// API v1
	v1 := router.Group("/api/v1")

	// Public auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.POST("/refresh", authH.Refresh)
		auth.POST("/forgot-password", authH.ForgotPassword)
		auth.POST("/reset-password", authH.ResetPassword)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret, rdb))
	{
		// Auth (protected)
		protected.POST("/auth/logout", authH.Logout)

		// Accounts
		protected.GET("/accounts/summary", accountH.Summary)
		protected.GET("/accounts", accountH.List)
		protected.POST("/accounts", accountH.Create)
		protected.GET("/accounts/:id", accountH.GetByID)
		protected.PUT("/accounts/:id", accountH.Update)
		protected.DELETE("/accounts/:id", accountH.Delete)

		// Categories
		protected.GET("/categories", categoryH.List)
		protected.POST("/categories", categoryH.Create)
		protected.PUT("/categories/:id", categoryH.Update)
		protected.DELETE("/categories/:id", categoryH.Delete)

		// Transactions
		protected.GET("/transactions/summary", transactionH.GetSummary)
		protected.POST("/transactions/transfer", transactionH.CreateTransfer)
		protected.GET("/transactions", transactionH.List)
		protected.POST("/transactions", transactionH.Create)
		protected.GET("/transactions/:id", transactionH.GetByID)
		protected.PUT("/transactions/:id", transactionH.Update)
		protected.DELETE("/transactions/:id", transactionH.Delete)

		// Recurrences
		protected.GET("/recurrences", recurrenceH.List)
		protected.POST("/recurrences", recurrenceH.Create)
		protected.PUT("/recurrences/:id", recurrenceH.Update)
		protected.DELETE("/recurrences/:id", recurrenceH.Delete)

		// Budgets
		protected.GET("/budgets/progress", budgetH.GetProgress)
		protected.GET("/budgets", budgetH.List)
		protected.POST("/budgets", budgetH.Create)
		protected.PUT("/budgets/:id", budgetH.Update)
		protected.DELETE("/budgets/:id", budgetH.Delete)

		// Dashboard
		protected.GET("/dashboard/overview", dashboardH.GetOverview)
		protected.GET("/dashboard/cashflow", dashboardH.GetCashflow)

		// Investments — Portfolios
		protected.GET("/portfolios", investmentH.ListPortfolios)
		protected.POST("/portfolios", investmentH.CreatePortfolio)
		protected.PUT("/portfolios/:id", investmentH.UpdatePortfolio)
		protected.DELETE("/portfolios/:id", investmentH.DeletePortfolio)
		protected.GET("/portfolios/:id/holdings", investmentH.ListHoldings)
		protected.POST("/portfolios/:id/holdings", investmentH.CreateHolding)

		// Investments — Holdings
		protected.PUT("/holdings/:id", investmentH.UpdateHolding)
		protected.DELETE("/holdings/:id", investmentH.DeleteHolding)
		protected.GET("/holdings/:id/transactions", investmentH.ListInvestmentTransactions)
		protected.POST("/holdings/:id/transactions", investmentH.CreateInvestmentTransaction)

		// Investments — Transactions
		protected.DELETE("/investment-transactions/:id", investmentH.DeleteInvestmentTransaction)

		// Assets
		protected.GET("/assets/search", investmentH.SearchAssets)

		// Custom Assets
		protected.GET("/custom-assets", investmentH.ListCustomAssets)
		protected.POST("/custom-assets", investmentH.CreateCustomAsset)
		protected.PUT("/custom-assets/:id", investmentH.UpdateCustomAsset)
		protected.DELETE("/custom-assets/:id", investmentH.DeleteCustomAsset)

		// Goals
		protected.GET("/goals/projections", goalH.GetProjections)
		protected.GET("/goals", goalH.List)
		protected.POST("/goals", goalH.Create)
		protected.PUT("/goals/:id", goalH.Update)
		protected.DELETE("/goals/:id", goalH.Delete)
		protected.POST("/goals/:id/contribute", goalH.Contribute)

		// Imports (pro plan required)
		imports := protected.Group("/imports")
		imports.Use(middleware.PlanMiddleware("pro"))
		{
			imports.POST("/ofx", importH.ImportOFX)
			imports.POST("/csv", importH.ImportCSV)
			imports.POST("/csv/preview", importH.PreviewCSV)
		}

		// Notifications
		protected.GET("/notifications", notificationH.List)
		protected.PUT("/notifications/read-all", notificationH.MarkAllAsRead)
		protected.PUT("/notifications/:id/read", notificationH.MarkAsRead)
		protected.DELETE("/notifications", notificationH.DeleteAll)

		// Family
		protected.POST("/family", familyH.Create)
		protected.GET("/family", familyH.Get)
		protected.POST("/family/invite", familyH.GetInvite)
		protected.POST("/family/join", familyH.Join)
		protected.DELETE("/family/members/:id", familyH.RemoveMember)
		protected.GET("/family/dashboard", familyH.GetDashboard)

		// AI (pro plan required for forecast and portfolio)
		protected.POST("/ai/chat", aiH.Chat)
		aiProtected := protected.Group("/ai")
		aiProtected.Use(middleware.PlanMiddleware("pro"))
		{
			aiProtected.GET("/spending-forecast", aiH.GetSpendingForecast)
			aiProtected.GET("/portfolio-analysis", aiH.GetPortfolioAnalysis)
		}
	}

	// Public webhook routes (no auth)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/whatsapp", webhookH.WhatsApp)
	}

	return router
}
