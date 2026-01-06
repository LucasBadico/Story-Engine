//go:build integration

package redis

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Cleanup function - flush all keys
	cleanup := func() {
		client.FlushDB(ctx)
		client.Close()
	}

	return client, cleanup
}

func TestIngestionQueue_Push(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	queue := NewIngestionQueue(client)
	ctx := context.Background()
	tenantID := uuid.New()
	sourceID := uuid.New()

	err := queue.Push(ctx, tenantID, string(memory.SourceTypeStory), sourceID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify item was added
	key := queue.queueKey(tenantID)
	members, err := client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		t.Fatalf("Failed to verify queue: %v", err)
	}

	if len(members) != 1 {
		t.Errorf("Expected 1 item in queue, got %d", len(members))
	}

	expectedMember := string(memory.SourceTypeStory) + ":" + sourceID.String()
	if members[0] != expectedMember {
		t.Errorf("Expected member %s, got %s", expectedMember, members[0])
	}
}

func TestIngestionQueue_Push_Debounce(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	queue := NewIngestionQueue(client)
	ctx := context.Background()
	tenantID := uuid.New()
	sourceID := uuid.New()

	// Push first time
	err := queue.Push(ctx, tenantID, string(memory.SourceTypeStory), sourceID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Push again (should update timestamp, not create duplicate)
	err = queue.Push(ctx, tenantID, string(memory.SourceTypeStory), sourceID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify only one item exists
	key := queue.queueKey(tenantID)
	count, err := client.ZCard(ctx, key).Result()
	if err != nil {
		t.Fatalf("Failed to verify queue: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 item after debounce, got %d", count)
	}
}

func TestIngestionQueue_PopStable(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	queue := NewIngestionQueue(client)
	ctx := context.Background()
	tenantID := uuid.New()
	sourceID1 := uuid.New()
	sourceID2 := uuid.New()

	// Push items
	queue.Push(ctx, tenantID, string(memory.SourceTypeStory), sourceID1)
	queue.Push(ctx, tenantID, string(memory.SourceTypeChapter), sourceID2)

	// Wait a bit to ensure timestamps are stable
	time.Sleep(100 * time.Millisecond)
	stableAt := time.Now()

	// Pop stable items
	items, err := queue.PopStable(ctx, tenantID, stableAt, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	// Verify items were removed from queue
	key := queue.queueKey(tenantID)
	count, err := client.ZCard(ctx, key).Result()
	if err != nil {
		t.Fatalf("Failed to verify queue: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected queue to be empty after PopStable, got %d items", count)
	}
}

func TestIngestionQueue_PopStable_WithLimit(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	queue := NewIngestionQueue(client)
	ctx := context.Background()
	tenantID := uuid.New()

	// Push multiple items
	for i := 0; i < 5; i++ {
		queue.Push(ctx, tenantID, string(memory.SourceTypeStory), uuid.New())
	}

	time.Sleep(100 * time.Millisecond)
	stableAt := time.Now()

	// Pop with limit
	items, err := queue.PopStable(ctx, tenantID, stableAt, 3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(items) != 3 {
		t.Errorf("Expected 3 items with limit, got %d", len(items))
	}

	// Verify remaining items
	key := queue.queueKey(tenantID)
	count, err := client.ZCard(ctx, key).Result()
	if err != nil {
		t.Fatalf("Failed to verify queue: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 remaining items, got %d", count)
	}
}

func TestIngestionQueue_PopStable_NotStable(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	queue := NewIngestionQueue(client)
	ctx := context.Background()
	tenantID := uuid.New()
	sourceID := uuid.New()

	// Push item
	queue.Push(ctx, tenantID, string(memory.SourceTypeStory), sourceID)

	// Try to pop with a timestamp in the past (before push)
	pastTime := time.Now().Add(-1 * time.Hour)
	items, err := queue.PopStable(ctx, tenantID, pastTime, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should return empty since item is not stable
	if len(items) != 0 {
		t.Errorf("Expected 0 items (not stable), got %d", len(items))
	}
}

func TestIngestionQueue_Remove(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	queue := NewIngestionQueue(client)
	ctx := context.Background()
	tenantID := uuid.New()
	sourceID1 := uuid.New()
	sourceID2 := uuid.New()

	// Push items
	queue.Push(ctx, tenantID, string(memory.SourceTypeStory), sourceID1)
	queue.Push(ctx, tenantID, string(memory.SourceTypeChapter), sourceID2)

	// Remove one item
	err := queue.Remove(ctx, tenantID, string(memory.SourceTypeStory), sourceID1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify item was removed
	key := queue.queueKey(tenantID)
	members, err := client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		t.Fatalf("Failed to verify queue: %v", err)
	}

	if len(members) != 1 {
		t.Errorf("Expected 1 item after removal, got %d", len(members))
	}

	expectedMember := string(memory.SourceTypeChapter) + ":" + sourceID2.String()
	if members[0] != expectedMember {
		t.Errorf("Expected remaining member %s, got %s", expectedMember, members[0])
	}
}

func TestIngestionQueue_ListTenantsWithItems(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	queue := NewIngestionQueue(client)
	ctx := context.Background()
	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	tenantID3 := uuid.New()

	// Push items for different tenants
	queue.Push(ctx, tenantID1, string(memory.SourceTypeStory), uuid.New())
	queue.Push(ctx, tenantID2, string(memory.SourceTypeChapter), uuid.New())
	queue.Push(ctx, tenantID2, string(memory.SourceTypeContentBlock), uuid.New())
	// tenantID3 has no items

	// List tenants with items
	tenants, err := queue.ListTenantsWithItems(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tenants) != 2 {
		t.Errorf("Expected 2 tenants with items, got %d", len(tenants))
	}

	// Verify tenant IDs
	tenantMap := make(map[uuid.UUID]bool)
	for _, tid := range tenants {
		tenantMap[tid] = true
	}

	if !tenantMap[tenantID1] {
		t.Error("Expected tenantID1 to be in list")
	}
	if !tenantMap[tenantID2] {
		t.Error("Expected tenantID2 to be in list")
	}
	if tenantMap[tenantID3] {
		t.Error("Expected tenantID3 NOT to be in list")
	}
}

func TestIngestionQueue_EmptyQueue(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	queue := NewIngestionQueue(client)
	ctx := context.Background()
	tenantID := uuid.New()

	// Pop from empty queue
	items, err := queue.PopStable(ctx, tenantID, time.Now(), 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 items from empty queue, got %d", len(items))
	}

	// List tenants from empty queue
	tenants, err := queue.ListTenantsWithItems(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tenants) != 0 {
		t.Errorf("Expected 0 tenants, got %d", len(tenants))
	}
}


