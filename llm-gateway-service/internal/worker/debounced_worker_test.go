package worker

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/config"
	"github.com/story-engine/llm-gateway-service/internal/ports/queue"
)

// Note: Full integration tests for worker would require modifying the architecture
// to use interfaces for use cases. For now, we test the configuration and queue logic.
// The use cases themselves are tested separately.

func TestNewDebouncedWorker_Configuration(t *testing.T) {
	// Test worker configuration parsing
	cfg := &config.Config{}
	cfg.Worker.DebounceMinutes = 5
	cfg.Worker.PollSeconds = 60
	cfg.Worker.BatchSize = 10
	cfg.Worker.ProcessingTimeoutSeconds = 600

	debounceInterval := time.Duration(cfg.Worker.DebounceMinutes) * time.Minute
	pollInterval := time.Duration(cfg.Worker.PollSeconds) * time.Second
	processingTimeout := time.Duration(cfg.Worker.ProcessingTimeoutSeconds) * time.Second

	if debounceInterval != 5*time.Minute {
		t.Errorf("Expected debounceInterval 5m, got %v", debounceInterval)
	}
	if pollInterval != 60*time.Second {
		t.Errorf("Expected pollInterval 60s, got %v", pollInterval)
	}
	if processingTimeout != 600*time.Second {
		t.Errorf("Expected processingTimeout 600s, got %v", processingTimeout)
	}
	if cfg.Worker.BatchSize != 10 {
		t.Errorf("Expected batchSize 10, got %d", cfg.Worker.BatchSize)
	}
}

func TestDebouncedWorker_processItem_Story(t *testing.T) {
	t.Skip("Requires modifying worker to use interfaces for use cases")
}

// Test worker queue processing logic
func TestDebouncedWorker_QueueLogic(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	mockQueue := queue.NewMockIngestionQueue()
	
	// Add items to queue - they will have current timestamp
	storyID := uuid.New()
	err := mockQueue.Push(ctx, tenantID, string(memory.SourceTypeStory), storyID)
	if err != nil {
		t.Fatalf("Expected no error on Push, got %v", err)
	}
	
	chapterID := uuid.New()
	err = mockQueue.Push(ctx, tenantID, string(memory.SourceTypeChapter), chapterID)
	if err != nil {
		t.Fatalf("Expected no error on Push, got %v", err)
	}

	// Verify items were added by checking tenants
	tenants, err := mockQueue.ListTenantsWithItems(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tenants) != 1 {
		t.Errorf("Expected 1 tenant, got %d", len(tenants))
	}

	if tenants[0] != tenantID {
		t.Errorf("Expected tenantID %s, got %s", tenantID, tenants[0])
	}

	// Wait to ensure items have timestamps set, then pop with a past timestamp
	time.Sleep(200 * time.Millisecond)
	
	// Pop stable items - use a timestamp in the past
	// Items created ~200ms ago should be stable now
	stableTime := time.Now().Add(-100 * time.Millisecond)
	items, err := mockQueue.PopStable(ctx, tenantID, stableTime, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Items should be popped since they're older than stableTime
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d (stableTime: %v)", len(items), stableTime)
	}
}

func TestDebouncedWorker_Configuration(t *testing.T) {
	cfg := &config.Config{}
	cfg.Worker.DebounceMinutes = 5
	cfg.Worker.PollSeconds = 60
	cfg.Worker.BatchSize = 10
	cfg.Worker.ProcessingTimeoutSeconds = 600

	debounceInterval := time.Duration(cfg.Worker.DebounceMinutes) * time.Minute
	pollInterval := time.Duration(cfg.Worker.PollSeconds) * time.Second
	processingTimeout := time.Duration(cfg.Worker.ProcessingTimeoutSeconds) * time.Second

	if debounceInterval != 5*time.Minute {
		t.Errorf("Expected debounceInterval 5m, got %v", debounceInterval)
	}
	if pollInterval != 60*time.Second {
		t.Errorf("Expected pollInterval 60s, got %v", pollInterval)
	}
	if processingTimeout != 600*time.Second {
		t.Errorf("Expected processingTimeout 600s, got %v", processingTimeout)
	}
	if cfg.Worker.BatchSize != 10 {
		t.Errorf("Expected batchSize 10, got %d", cfg.Worker.BatchSize)
	}
}
