package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis connection parameters.
type Config struct {
	URL      string
	Password string
	DB       int
}

// Connect creates a Redis client, parses the URL, and verifies connectivity
// with a PING. It returns the client or an error.
func Connect(ctx context.Context, cfg Config) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing Redis URL: %w", err)
	}

	// Allow override of password and DB from explicit config.
	if cfg.Password != "" {
		opts.Password = cfg.Password
	}
	if cfg.DB != 0 {
		opts.DB = cfg.DB
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("pinging Redis: %w", err)
	}

	return client, nil
}
