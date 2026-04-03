package config_test

import (
	"os"
	"testing"

	"github.com/financeos/api/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "secret")

	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")
	t.Setenv("JWT_SECRET", "")

	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")
	t.Setenv("JWT_SECRET", "mysecret")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "development", cfg.App.Env)
	assert.Equal(t, 8000, cfg.App.Port)
	assert.Equal(t, "info", cfg.App.LogLevel)
	assert.Equal(t, "claude-sonnet-4-20250514", cfg.Claude.Model)
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")
	t.Setenv("JWT_SECRET", "mysecret")
	t.Setenv("APP_PORT", "9000")
	t.Setenv("APP_ENV", "production")
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("JWT_ACCESS_TTL", "30m")
	t.Setenv("JWT_REFRESH_TTL", "48h")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "production", cfg.App.Env)
	assert.Equal(t, 9000, cfg.App.Port)
	assert.Equal(t, "warn", cfg.App.LogLevel)
	assert.Equal(t, "mysecret", cfg.JWT.Secret)
}

func TestLoad_InvalidJWTAccessTTL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")
	t.Setenv("JWT_SECRET", "mysecret")
	t.Setenv("JWT_ACCESS_TTL", "not-a-duration")

	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_ACCESS_TTL")
}

func TestLoad_CORSOrigins(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")
	t.Setenv("JWT_SECRET", "mysecret")
	t.Setenv("CORS_ORIGINS", "http://localhost:3000,http://localhost:8080")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Len(t, cfg.CORS.Origins, 2)
	assert.Equal(t, "http://localhost:3000", cfg.CORS.Origins[0])
	assert.Equal(t, "http://localhost:8080", cfg.CORS.Origins[1])
}
