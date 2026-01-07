package queue

import (
	"context"

	"github.com/google/uuid"
)

// IngestionQueue defines the interface for LLM ingestion notifications.
type IngestionQueue interface {
	// Push adds/updates an item with the current timestamp.
	Push(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error
}
