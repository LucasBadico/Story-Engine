package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/story-engine/main-service/internal/ports/queue"
)

var _ queue.IngestionQueue = (*IngestionQueue)(nil)

// IngestionQueue implements the queue interface using Redis Sorted Sets.
type IngestionQueue struct {
	client *redis.Client
}

// NewIngestionQueue creates a new Redis-based ingestion queue.
func NewIngestionQueue(client *redis.Client) *IngestionQueue {
	return &IngestionQueue{
		client: client,
	}
}

// Push adds/updates item with current timestamp (debounce reset).
func (q *IngestionQueue) Push(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	key := q.queueKey(tenantID)
	member := fmt.Sprintf("%s:%s", sourceType, sourceID.String())
	score := float64(time.Now().Unix())

	return q.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

func (q *IngestionQueue) queueKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("ingestion:queue:%s", tenantID.String())
}
