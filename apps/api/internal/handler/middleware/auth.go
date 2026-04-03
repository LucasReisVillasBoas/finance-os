package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// AuthMiddleware validates JWT Bearer tokens and injects user claims into the Gin context.
func AuthMiddleware(jwtSecret string, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "authorization header missing"},
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid authorization header format"},
			})
			return
		}

		rawToken := parts[1]

		// Check blacklist
		h := sha256.Sum256([]byte(rawToken))
		tokenHash := hex.EncodeToString(h[:])
		blacklistKey := "blacklist:" + tokenHash

		exists, err := rdb.Exists(c.Request.Context(), blacklistKey).Result()
		if err == nil && exists > 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "TOKEN_REVOKED", "message": "token has been revoked"},
			})
			return
		}

		// Parse and validate JWT
		claims := &usecase.Claims{}
		token, err := jwt.ParseWithClaims(rawToken, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "INVALID_TOKEN", "message": "invalid or expired token"},
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_plan", claims.Plan)
		c.Next()
	}
}
