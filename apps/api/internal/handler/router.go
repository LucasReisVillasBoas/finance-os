package handler

import (
	"net/http"
	"time"

	"github.com/financeos/api/internal/handler/middleware"
	"github.com/financeos/api/internal/repository"
	"github.com/financeos/api/internal/usecase"
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

	// Protected auth routes
	authProtected := v1.Group("/auth")
	authProtected.Use(middleware.AuthMiddleware(cfg.JWT.Secret, rdb))
	{
		authProtected.POST("/logout", authH.Logout)
	}

	return router
}
