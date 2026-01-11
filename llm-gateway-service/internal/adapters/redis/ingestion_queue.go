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
	processingQueue *ProcessingQueue
}

// NewIngestionQueue creates a new Redis-based ingestion queue
func NewIngestionQueue(client *redis.Client) *IngestionQueue {
	return &IngestionQueue{
		processingQueue: NewProcessingQueue(client, "ingestion:queue", "ingestion:processing"),
	}
}

// Push adds/updates item with current timestamp (debounce reset)
func (q *IngestionQueue) Push(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	member := fmt.Sprintf("%s:%s", sourceType, sourceID.String())
	return q.processingQueue.Push(ctx, tenantID, member, time.Now())
}

// PopStable returns items not updated since stableAt and moves them to processing
func (q *IngestionQueue) PopStable(ctx context.Context, tenantID uuid.UUID, stableAt time.Time, limit int) ([]*queue.QueueItem, error) {
	members, err := q.processingQueue.PopStableToProcessing(ctx, tenantID, stableAt, limit)
	if err != nil {
		return nil, err
	}

	return parseQueueItems(tenantID, members, stableAt), nil
}

// PopStableBySourceType returns items for a given source type not updated since stableAt and moves them to processing
func (q *IngestionQueue) PopStableBySourceType(ctx context.Context, tenantID uuid.UUID, sourceType string, stableAt time.Time, limit int) ([]*queue.QueueItem, error) {
	prefix := fmt.Sprintf("%s:", sourceType)
	members, err := q.processingQueue.PopStableToProcessingByPrefix(ctx, tenantID, stableAt, limit, prefix)
	if err != nil {
		return nil, err
	}

	return parseQueueItems(tenantID, members, stableAt), nil
}

// Ack removes an item from processing after successful handling
func (q *IngestionQueue) Ack(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	member := fmt.Sprintf("%s:%s", sourceType, sourceID.String())
	return q.processingQueue.Ack(ctx, tenantID, member)
}

// Release moves an item from processing back to the queue
func (q *IngestionQueue) Release(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	member := fmt.Sprintf("%s:%s", sourceType, sourceID.String())
	return q.processingQueue.Release(ctx, tenantID, member, time.Now())
}

// RequeueExpiredProcessing moves expired processing items back to the queue
func (q *IngestionQueue) RequeueExpiredProcessing(ctx context.Context, tenantID uuid.UUID, expiredBefore time.Time, limit int) (int, error) {
	members, err := q.processingQueue.RequeueExpiredProcessing(ctx, tenantID, expiredBefore, limit)
	if err != nil {
		return 0, err
	}
	return len(members), nil
}

// Remove removes an item from both queue and processing
func (q *IngestionQueue) Remove(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error {
	member := fmt.Sprintf("%s:%s", sourceType, sourceID.String())
	return q.processingQueue.Remove(ctx, tenantID, member)
}

// ListTenantsWithItems returns tenant IDs that have items in queue
func (q *IngestionQueue) ListTenantsWithItems(ctx context.Context) ([]uuid.UUID, error) {
	return q.processingQueue.ListTenantsWithQueueItems(ctx)
}

// ListTenantsWithProcessingItems returns tenant IDs that have items in processing
func (q *IngestionQueue) ListTenantsWithProcessingItems(ctx context.Context) ([]uuid.UUID, error) {
	return q.processingQueue.ListTenantsWithProcessingItems(ctx)
}

// queueKey returns the Redis key for a tenant's queue
func (q *IngestionQueue) queueKey(tenantID uuid.UUID) string {
	return q.processingQueue.queueKey(tenantID)
}

// processingKey returns the Redis key for a tenant's processing set
func (q *IngestionQueue) processingKey(tenantID uuid.UUID) string {
	return q.processingQueue.processingKey(tenantID)
}

func parseQueueItems(tenantID uuid.UUID, members []string, stableAt time.Time) []*queue.QueueItem {
	queueItems := make([]*queue.QueueItem, 0, len(members))
	for _, member := range members {
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
			Timestamp:  stableAt,
		})
	}

	return queueItems
}
