package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/financeos/api/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestPlanMiddleware_FreeBlocksPro(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/pro-only", func(c *gin.Context) {
		c.Set("user_plan", "free")
		c.Next()
	}, middleware.PlanMiddleware("pro"), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/pro-only", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusPaymentRequired, w.Code)
}

func TestPlanMiddleware_ProAllowsPro(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/pro-only",
		func(c *gin.Context) {
			c.Set("user_plan", "pro")
			c.Next()
		},
		middleware.PlanMiddleware("pro"),
		func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true})
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/pro-only", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestPlanMiddleware_PremiumAllowsPro(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/pro-only",
		func(c *gin.Context) {
			c.Set("user_plan", "premium")
			c.Next()
		},
		middleware.PlanMiddleware("pro"),
		func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true})
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/pro-only", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestPlanMiddleware_EmptyPlanDefaultsFree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// No user_plan set — defaults to free
	r.GET("/pro-only",
		middleware.PlanMiddleware("pro"),
		func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true})
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/pro-only", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusPaymentRequired, w.Code)
}
