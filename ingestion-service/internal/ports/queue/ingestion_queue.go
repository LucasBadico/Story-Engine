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

	// PopStable returns items not updated since stableAt and removes them atomically
	PopStable(ctx context.Context, tenantID uuid.UUID, stableAt time.Time, limit int) ([]*QueueItem, error)

	// Remove removes an item from queue
	Remove(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error

	// ListTenantsWithItems returns tenant IDs that have items in queue
	ListTenantsWithItems(ctx context.Context) ([]uuid.UUID, error)
}

