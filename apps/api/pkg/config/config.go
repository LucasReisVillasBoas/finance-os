package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Claude    ClaudeConfig
	Evolution EvolutionConfig
	Brapi     BrapiConfig
	CORS      CORSConfig
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Env      string
	Port     int
	LogLevel string
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// JWTConfig holds JWT signing settings.
type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// ClaudeConfig holds Anthropic Claude API settings.
type ClaudeConfig struct {
	APIKey string
	Model  string
}

// EvolutionConfig holds Evolution API (WhatsApp) settings.
type EvolutionConfig struct {
	APIURL string
	APIKey string
}

// BrapiConfig holds BRAPI (Brazilian stock quotes) settings.
// Token is optional: brapi.dev grants a free token that unlocks quote
// endpoints. Without it the project still runs in best-effort mode.
type BrapiConfig struct {
	Token string
}

// CORSConfig holds CORS settings.
type CORSConfig struct {
	Origins []string
}

// Load reads configuration from environment variables and returns a Config.
func Load() (*Config, error) {
	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// App defaults
	v.SetDefault("app.env", "development")
	v.SetDefault("app.port", 8000)
	v.SetDefault("app.log_level", "info")

	// Database defaults
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "5m")

	// Redis defaults
	v.SetDefault("redis.db", 0)

	// JWT defaults
	v.SetDefault("jwt.access_ttl", "15m")
	v.SetDefault("jwt.refresh_ttl", "720h")

	// Claude defaults
	v.SetDefault("claude.model", "claude-sonnet-4-20250514")

	// Bind environment variables explicitly
	envBindings := map[string]string{
		"app.env":       "APP_ENV",
		"app.port":      "APP_PORT",
		"app.log_level": "LOG_LEVEL",

		"database.url": "DATABASE_URL",

		"redis.url":      "REDIS_URL",
		"redis.password": "REDIS_PASSWORD",
		"redis.db":       "REDIS_DB",

		"jwt.secret":      "JWT_SECRET",
		"jwt.access_ttl":  "JWT_ACCESS_TTL",
		"jwt.refresh_ttl": "JWT_REFRESH_TTL",

		"claude.api_key": "ANTHROPIC_API_KEY",
		"claude.model":   "CLAUDE_MODEL",

		"evolution.api_url": "EVOLUTION_API_URL",
		"evolution.api_key": "EVOLUTION_API_KEY",

		"brapi.token": "BRAPI_TOKEN",

		"cors.origins": "CORS_ORIGINS",
	}

	for key, env := range envBindings {
		if err := v.BindEnv(key, env); err != nil {
			return nil, fmt.Errorf("binding env var %s: %w", env, err)
		}
	}

	accessTTL, err := time.ParseDuration(v.GetString("jwt.access_ttl"))
	if err != nil {
		return nil, fmt.Errorf("parsing JWT_ACCESS_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(v.GetString("jwt.refresh_ttl"))
	if err != nil {
		return nil, fmt.Errorf("parsing JWT_REFRESH_TTL: %w", err)
	}

	connMaxLifetime, err := time.ParseDuration(v.GetString("database.conn_max_lifetime"))
	if err != nil {
		connMaxLifetime = 5 * time.Minute
	}

	originsRaw := v.GetString("cors.origins")
	var origins []string
	if originsRaw != "" {
		for _, o := range strings.Split(originsRaw, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}
	if len(origins) == 0 {
		origins = []string{"http://localhost:3000"}
	}

	cfg := &Config{
		App: AppConfig{
			Env:      v.GetString("app.env"),
			Port:     v.GetInt("app.port"),
			LogLevel: v.GetString("app.log_level"),
		},
		Database: DatabaseConfig{
			URL:             v.GetString("database.url"),
			MaxOpenConns:    v.GetInt("database.max_open_conns"),
			MaxIdleConns:    v.GetInt("database.max_idle_conns"),
			ConnMaxLifetime: connMaxLifetime,
		},
		Redis: RedisConfig{
			URL:      v.GetString("redis.url"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
		},
		JWT: JWTConfig{
			Secret:     v.GetString("jwt.secret"),
			AccessTTL:  accessTTL,
			RefreshTTL: refreshTTL,
		},
		Claude: ClaudeConfig{
			APIKey: v.GetString("claude.api_key"),
			Model:  v.GetString("claude.model"),
		},
		Evolution: EvolutionConfig{
			APIURL: v.GetString("evolution.api_url"),
			APIKey: v.GetString("evolution.api_key"),
		},
		Brapi: BrapiConfig{
			Token: v.GetString("brapi.token"),
		},
		CORS: CORSConfig{
			Origins: origins,
		},
	}

	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWT.Secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	if cfg.JWT.Secret == "change-me-in-production-use-random-32-chars" {
		return nil, fmt.Errorf("JWT_SECRET is set to the example value — generate a secure random secret before running in production")
	}

	return cfg, nil
}
