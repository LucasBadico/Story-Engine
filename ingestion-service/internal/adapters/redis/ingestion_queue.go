package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/story-engine/llm-gateway-service/internal/ports/queue"
)

var _ queue.IngestionQueue = (*IngestionQueue)(nil)

// IngestionQueue implements the queue interface using Redis Sorted Sets
type IngestionQueue struct {
	client *redis.Client
}

// NewIngestionQueue creates a new Redis-based ingestion queue
func NewIngestionQueue(client *redis.Client) *IngestionQueue {
	return &IngestionQueue{
		client: client,
	}
}

// Push adds/updates item with current timestamp (debounce reset)
func (q *IngestionQueue) Push(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	key := q.queueKey(tenantID)
	member := fmt.Sprintf("%s:%s", sourceType, sourceID.String())
	score := float64(time.Now().Unix())

	// ZADD updates the score if member exists, or adds if it doesn't
	return q.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// PopStable returns items not updated since stableAt and removes them atomically
func (q *IngestionQueue) PopStable(ctx context.Context, tenantID uuid.UUID, stableAt time.Time, limit int) ([]*QueueItem, error) {
	key := q.queueKey(tenantID)
	maxScore := float64(stableAt.Unix())

	// Lua script to atomically get and remove items
	// Returns items with score <= maxScore, limited by limit
	luaScript := `
		local key = KEYS[1]
		local max_score = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		
		local items = redis.call('ZRANGEBYSCORE', key, '0', max_score, 'LIMIT', '0', limit)
		if #items > 0 then
			redis.call('ZREM', key, unpack(items))
		end
		return items
	`

	result, err := q.client.Eval(ctx, luaScript, []string{key}, maxScore, limit).Result()
	if err != nil {
		if err == redis.Nil {
			return []*QueueItem{}, nil
		}
		return nil, err
	}

	// Parse result
	items, ok := result.([]interface{})
	if !ok {
		return []*QueueItem{}, nil
	}

	queueItems := make([]*QueueItem, 0, len(items))
	for _, item := range items {
		member, ok := item.(string)
		if !ok {
			continue
		}

		// Parse member format: "source_type:source_id"
		parts := strings.Split(member, ":")
		if len(parts) != 2 {
			continue
		}

		sourceType := parts[0]
		sourceID, err := uuid.Parse(parts[1])
		if err != nil {
			continue
		}

		queueItems = append(queueItems, &queue.QueueItem{
			TenantID:   tenantID,
			SourceType: sourceType,
			SourceID:   sourceID,
			Timestamp:  stableAt, // Approximate timestamp
		})
	}

	return queueItems, nil
}

// Remove removes an item from queue
func (q *IngestionQueue) Remove(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	key := q.queueKey(tenantID)
	member := fmt.Sprintf("%s:%s", sourceType, sourceID.String())
	return q.client.ZRem(ctx, key, member).Err()
}

// ListTenantsWithItems returns tenant IDs that have items in queue
func (q *IngestionQueue) ListTenantsWithItems(ctx context.Context) ([]uuid.UUID, error) {
	pattern := "ingestion:queue:*"
	keys, err := q.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	tenantIDs := make([]uuid.UUID, 0, len(keys))
	for _, key := range keys {
		// Extract tenant ID from key: "ingestion:queue:{tenant_id}"
		parts := strings.Split(key, ":")
		if len(parts) != 3 {
			continue
		}
		tenantID, err := uuid.Parse(parts[2])
		if err != nil {
			continue
		}
		tenantIDs = append(tenantIDs, tenantID)
	}

	return tenantIDs, nil
}

// queueKey returns the Redis key for a tenant's queue
func (q *IngestionQueue) queueKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("ingestion:queue:%s", tenantID.String())
}

