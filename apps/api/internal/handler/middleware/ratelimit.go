package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig defines the parameters for a rate limit rule.
type RateLimitConfig struct {
	// MaxAttempts is the number of requests allowed within Window.
	MaxAttempts int
	// Window is the rolling time period for the counter.
	Window time.Duration
	// KeyPrefix distinguishes rules for different endpoints.
	KeyPrefix string
}

// RateLimit returns a Gin middleware that enforces per-IP rate limits using Redis.
// If Redis is unavailable the middleware fails open (lets the request through).
func RateLimit(rdb *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s:%s", cfg.KeyPrefix, ip)
		ctx := c.Request.Context()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Redis unavailable — fail open to avoid blocking legitimate users.
			c.Next()
			return
		}

		// Set TTL only on the first increment so the window resets naturally.
		if count == 1 {
			rdb.Expire(ctx, key, cfg.Window)
		}

		if count > int64(cfg.MaxAttempts) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Muitas tentativas. Tente novamente mais tarde.",
				},
			})
			return
		}

		c.Next()
	}
}
