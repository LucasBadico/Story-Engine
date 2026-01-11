package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ProcessingQueue provides a reusable Redis sorted-set queue with processing/ack semantics.
type ProcessingQueue struct {
	client           *redis.Client
	queuePrefix      string
	processingPrefix string
}

// NewProcessingQueue creates a new processing queue with the given key prefixes.
func NewProcessingQueue(client *redis.Client, queuePrefix, processingPrefix string) *ProcessingQueue {
	return &ProcessingQueue{
		client:           client,
		queuePrefix:      strings.TrimRight(queuePrefix, ":"),
		processingPrefix: strings.TrimRight(processingPrefix, ":"),
	}
}

// Push adds/updates a member in the queue with the provided timestamp.
func (q *ProcessingQueue) Push(ctx context.Context, tenantID uuid.UUID, member string, at time.Time) error {
	return q.client.ZAdd(ctx, q.queueKey(tenantID), redis.Z{
		Score:  float64(at.Unix()),
		Member: member,
	}).Err()
}

// PopStableToProcessing moves stable items from the queue to processing and returns members.
func (q *ProcessingQueue) PopStableToProcessing(ctx context.Context, tenantID uuid.UUID, stableAt time.Time, limit int) ([]string, error) {
	queueKey := q.queueKey(tenantID)
	processingKey := q.processingKey(tenantID)
	maxScore := float64(stableAt.Unix())
	processingScore := float64(time.Now().Unix())

	luaScript := `
		local queue_key = KEYS[1]
		local processing_key = KEYS[2]
		local max_score = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		local processing_score = tonumber(ARGV[3])

		local items = redis.call('ZRANGEBYSCORE', queue_key, '0', max_score, 'LIMIT', '0', limit)
		if #items > 0 then
			for _, item in ipairs(items) do
				redis.call('ZREM', queue_key, item)
				redis.call('ZADD', processing_key, processing_score, item)
			end
		end
		return items
	`

	result, err := q.client.Eval(ctx, luaScript, []string{queueKey, processingKey}, maxScore, limit, processingScore).Result()
	if err != nil {
		if err == redis.Nil {
			return []string{}, nil
		}
		return nil, err
	}

	return toStringSlice(result), nil
}

// PopStableToProcessingByPrefix moves stable items with a prefix to processing and returns members.
func (q *ProcessingQueue) PopStableToProcessingByPrefix(ctx context.Context, tenantID uuid.UUID, stableAt time.Time, limit int, prefix string) ([]string, error) {
	queueKey := q.queueKey(tenantID)
	processingKey := q.processingKey(tenantID)
	maxScore := float64(stableAt.Unix())
	processingScore := float64(time.Now().Unix())

	luaScript := `
		local queue_key = KEYS[1]
		local processing_key = KEYS[2]
		local max_score = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		local processing_score = tonumber(ARGV[3])
		local prefix = ARGV[4]

		local items = redis.call('ZRANGEBYSCORE', queue_key, '0', max_score)
		local results = {}
		if #items > 0 then
			for _, item in ipairs(items) do
				if string.sub(item, 1, string.len(prefix)) == prefix then
					redis.call('ZREM', queue_key, item)
					redis.call('ZADD', processing_key, processing_score, item)
					table.insert(results, item)
					if #results >= limit then
						break
					end
				end
			end
		end
		return results
	`

	result, err := q.client.Eval(ctx, luaScript, []string{queueKey, processingKey}, maxScore, limit, processingScore, prefix).Result()
	if err != nil {
		if err == redis.Nil {
			return []string{}, nil
		}
		return nil, err
	}

	return toStringSlice(result), nil
}

// Ack removes a member from the processing set.
func (q *ProcessingQueue) Ack(ctx context.Context, tenantID uuid.UUID, member string) error {
	return q.client.ZRem(ctx, q.processingKey(tenantID), member).Err()
}

// Release moves a member from processing back to the queue.
func (q *ProcessingQueue) Release(ctx context.Context, tenantID uuid.UUID, member string, at time.Time) error {
	processingKey := q.processingKey(tenantID)
	queueKey := q.queueKey(tenantID)
	score := float64(at.Unix())

	luaScript := `
		local processing_key = KEYS[1]
		local queue_key = KEYS[2]
		local member = ARGV[1]
		local score = tonumber(ARGV[2])

		redis.call('ZREM', processing_key, member)
		return redis.call('ZADD', queue_key, score, member)
	`

	return q.client.Eval(ctx, luaScript, []string{processingKey, queueKey}, member, score).Err()
}

// RequeueExpiredProcessing moves expired processing items back to the queue.
func (q *ProcessingQueue) RequeueExpiredProcessing(ctx context.Context, tenantID uuid.UUID, expiredBefore time.Time, limit int) ([]string, error) {
	processingKey := q.processingKey(tenantID)
	queueKey := q.queueKey(tenantID)
	maxScore := float64(expiredBefore.Unix())
	newScore := float64(time.Now().Unix())

	luaScript := `
		local processing_key = KEYS[1]
		local queue_key = KEYS[2]
		local max_score = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		local new_score = tonumber(ARGV[3])

		local items = redis.call('ZRANGEBYSCORE', processing_key, '0', max_score, 'LIMIT', '0', limit)
		if #items > 0 then
			for _, item in ipairs(items) do
				redis.call('ZREM', processing_key, item)
				redis.call('ZADD', queue_key, new_score, item)
			end
		end
		return items
	`

	result, err := q.client.Eval(ctx, luaScript, []string{processingKey, queueKey}, maxScore, limit, newScore).Result()
	if err != nil {
		if err == redis.Nil {
			return []string{}, nil
		}
		return nil, err
	}

	return toStringSlice(result), nil
}

// Remove removes a member from both queue and processing sets.
func (q *ProcessingQueue) Remove(ctx context.Context, tenantID uuid.UUID, member string) error {
	pipe := q.client.Pipeline()
	pipe.ZRem(ctx, q.queueKey(tenantID), member)
	pipe.ZRem(ctx, q.processingKey(tenantID), member)
	_, err := pipe.Exec(ctx)
	return err
}

// ListTenantsWithQueueItems returns tenants with queued items.
func (q *ProcessingQueue) ListTenantsWithQueueItems(ctx context.Context) ([]uuid.UUID, error) {
	return q.listTenantsByPattern(ctx, fmt.Sprintf("%s:*", q.queuePrefix))
}

// ListTenantsWithProcessingItems returns tenants with processing items.
func (q *ProcessingQueue) ListTenantsWithProcessingItems(ctx context.Context) ([]uuid.UUID, error) {
	return q.listTenantsByPattern(ctx, fmt.Sprintf("%s:*", q.processingPrefix))
}

func (q *ProcessingQueue) queueKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("%s:%s", q.queuePrefix, tenantID.String())
}

func (q *ProcessingQueue) processingKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("%s:%s", q.processingPrefix, tenantID.String())
}

func (q *ProcessingQueue) listTenantsByPattern(ctx context.Context, pattern string) ([]uuid.UUID, error) {
	keys, err := q.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	tenantIDs := make([]uuid.UUID, 0, len(keys))
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) < 2 {
			continue
		}
		tenantID, err := uuid.Parse(parts[len(parts)-1])
		if err != nil {
			continue
		}
		tenantIDs = append(tenantIDs, tenantID)
	}

	return tenantIDs, nil
}

func toStringSlice(result interface{}) []string {
	items, ok := result.([]interface{})
	if !ok {
		return []string{}
	}

	out := make([]string, 0, len(items))
	for _, item := range items {
		member, ok := item.(string)
		if !ok {
			continue
		}
		out = append(out, member)
	}
	return out
}
