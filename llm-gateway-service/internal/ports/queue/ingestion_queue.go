package queue

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// QueueItem represents an item in the ingestion queue
type QueueItem struct {
	TenantID   uuid.UUID
	SourceType string
	SourceID   uuid.UUID
	Timestamp  time.Time
}

// IngestionQueue defines the interface for the ingestion queue
type IngestionQueue interface {
	// Push adds/updates item with current timestamp (debounce reset)
	Push(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error

	// PopStable returns items not updated since stableAt and moves them to processing
	PopStable(ctx context.Context, tenantID uuid.UUID, stableAt time.Time, limit int) ([]*QueueItem, error)

	// Ack removes an item from processing after successful handling
	Ack(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error

	// Release moves an item from processing back to the queue (debounce reset)
	Release(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error

	// RequeueExpiredProcessing moves expired processing items back to the queue
	RequeueExpiredProcessing(ctx context.Context, tenantID uuid.UUID, expiredBefore time.Time, limit int) (int, error)

	// Remove removes an item from both queue and processing
	Remove(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error

	// ListTenantsWithItems returns tenant IDs that have items in queue
	ListTenantsWithItems(ctx context.Context) ([]uuid.UUID, error)

	// ListTenantsWithProcessingItems returns tenant IDs that have items in processing
	ListTenantsWithProcessingItems(ctx context.Context) ([]uuid.UUID, error)
}
