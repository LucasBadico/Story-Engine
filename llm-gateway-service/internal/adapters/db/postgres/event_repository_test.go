//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

func TestEventRepository_CreateAndList(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewEventRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()

	// Create events at chapters 1, 3, 5
	event1 := memory.NewEvent(tenantID, storyID, 1, "Hero arrives at the village")
	event3 := memory.NewEvent(tenantID, storyID, 3, "Hero meets the mentor")
	event5 := memory.NewEvent(tenantID, storyID, 5, "Hero defeats the dragon")

	for _, e := range []*memory.Event{event1, event3, event5} {
		if err := repo.Create(ctx, e); err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	// ListByStory at chapter 4 should return 2 events (chapters 1, 3)
	results, err := repo.ListByStory(ctx, storyID, 4, 100)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 events at chapter 4, got %d", len(results))
	}
}

func TestEventRepository_ListByStoryAtDay(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewEventRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()

	// Create events at days -60, -40, -20
	dayMinus60 := intPtr(-60)
	dayMinus40 := intPtr(-40)
	dayMinus20 := intPtr(-20)

	event1 := memory.NewEvent(tenantID, storyID, 1, "Ancient ritual performed")
	event1.TimelineDay = dayMinus60

	event2 := memory.NewEvent(tenantID, storyID, 2, "Travelers arrive")
	event2.TimelineDay = dayMinus40

	event3 := memory.NewEvent(tenantID, storyID, 3, "Battle begins")
	event3.TimelineDay = dayMinus20

	for _, e := range []*memory.Event{event1, event2, event3} {
		if err := repo.Create(ctx, e); err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	// ListByStoryAtDay at day -30 should return 2 events (day -60, -40)
	results, err := repo.ListByStoryAtDay(ctx, storyID, -30, 100)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 events at day -30, got %d", len(results))
	}
}

func TestEventRepository_WithLimit(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewEventRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()

	// Create 5 events
	for i := 1; i <= 5; i++ {
		event := memory.NewEvent(tenantID, storyID, i, "Event description")
		if err := repo.Create(ctx, event); err != nil {
			t.Fatalf("Failed to create event %d: %v", i, err)
		}
	}

	// List with limit 2 should return 2
	results, err := repo.ListByStory(ctx, storyID, 10, 2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 events with limit 2, got %d", len(results))
	}
}

func TestEventRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewEventRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()

	event := memory.NewEvent(tenantID, storyID, 1, "Hero arrives")
	if err := repo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	err := repo.Delete(ctx, event.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify list is empty
	results, err := repo.ListByStory(ctx, storyID, 10, 100)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 events after delete, got %d", len(results))
	}
}
