package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/story-engine/llm-gateway-service/internal/platform/config"
)

// NewClient creates a new Redis client
func NewClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

