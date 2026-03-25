//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

func TestStateSnapshotRepository_CreateAndGetLatest(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewStateSnapshotRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()
	entityID := uuid.New()

	// Create snapshots at chapters 1, 3, 5
	snap1 := memory.NewStateSnapshot(tenantID, storyID, entityID, "character", 1, map[string]interface{}{"mood": "happy"})
	snap3 := memory.NewStateSnapshot(tenantID, storyID, entityID, "character", 3, map[string]interface{}{"mood": "sad"})
	snap5 := memory.NewStateSnapshot(tenantID, storyID, entityID, "character", 5, map[string]interface{}{"mood": "angry"})

	for _, s := range []*memory.StateSnapshot{snap1, snap3, snap5} {
		if err := repo.Create(ctx, s); err != nil {
			t.Fatalf("Failed to create snapshot: %v", err)
		}
	}

	// GetLatestByEntity at chapter 4 should return chapter 3 snapshot
	retrieved, err := repo.GetLatestByEntity(ctx, storyID, entityID, 4)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected snapshot, got nil")
	}
	if retrieved.ChapterNumber != 3 {
		t.Errorf("Expected ChapterNumber 3, got %d", retrieved.ChapterNumber)
	}
	if retrieved.StateData["mood"] != "sad" {
		t.Errorf("Expected mood %q, got %v", "sad", retrieved.StateData["mood"])
	}
}

func TestStateSnapshotRepository_GetLatestByEntityAtDay(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewStateSnapshotRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()
	entityID := uuid.New()

	// Create snapshots at days -60, -40, -20
	dayMinus60 := intPtr(-60)
	dayMinus40 := intPtr(-40)
	dayMinus20 := intPtr(-20)

	snap1 := memory.NewStateSnapshot(tenantID, storyID, entityID, "character", 1, map[string]interface{}{"location": "village"})
	snap1.TimelineDay = dayMinus60

	snap2 := memory.NewStateSnapshot(tenantID, storyID, entityID, "character", 2, map[string]interface{}{"location": "forest"})
	snap2.TimelineDay = dayMinus40

	snap3 := memory.NewStateSnapshot(tenantID, storyID, entityID, "character", 3, map[string]interface{}{"location": "castle"})
	snap3.TimelineDay = dayMinus20

	for _, s := range []*memory.StateSnapshot{snap1, snap2, snap3} {
		if err := repo.Create(ctx, s); err != nil {
			t.Fatalf("Failed to create snapshot: %v", err)
		}
	}

	// GetLatestByEntityAtDay at day -30 should return day -40 snapshot
	retrieved, err := repo.GetLatestByEntityAtDay(ctx, storyID, entityID, -30)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected snapshot, got nil")
	}
	if retrieved.TimelineDay == nil || *retrieved.TimelineDay != -40 {
		t.Errorf("Expected TimelineDay -40, got %v", retrieved.TimelineDay)
	}
	if retrieved.StateData["location"] != "forest" {
		t.Errorf("Expected location %q, got %v", "forest", retrieved.StateData["location"])
	}
}

func TestStateSnapshotRepository_ListByStory(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewStateSnapshotRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()
	entityA := uuid.New()
	entityB := uuid.New()

	// Create snapshots for entity A at chapters 1, 3
	snapA1 := memory.NewStateSnapshot(tenantID, storyID, entityA, "character", 1, map[string]interface{}{"state": "A1"})
	snapA3 := memory.NewStateSnapshot(tenantID, storyID, entityA, "character", 3, map[string]interface{}{"state": "A3"})

	// Create snapshots for entity B at chapters 1, 3
	snapB1 := memory.NewStateSnapshot(tenantID, storyID, entityB, "location", 1, map[string]interface{}{"state": "B1"})
	snapB3 := memory.NewStateSnapshot(tenantID, storyID, entityB, "location", 3, map[string]interface{}{"state": "B3"})

	for _, s := range []*memory.StateSnapshot{snapA1, snapA3, snapB1, snapB3} {
		if err := repo.Create(ctx, s); err != nil {
			t.Fatalf("Failed to create snapshot: %v", err)
		}
	}

	// ListByStory at chapter 5 returns latest per entity (DISTINCT ON)
	results, err := repo.ListByStory(ctx, storyID, 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 snapshots (one per entity), got %d", len(results))
	}

	// Verify we got the latest snapshots (chapter 3 for both)
	for _, r := range results {
		if r.ChapterNumber != 3 {
			t.Errorf("Expected ChapterNumber 3 (latest), got %d for entity %v", r.ChapterNumber, r.EntityID)
		}
	}
}

func TestStateSnapshotRepository_TimelineConfidence(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewStateSnapshotRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()
	entityID := uuid.New()

	snap := memory.NewStateSnapshot(tenantID, storyID, entityID, "character", 1, map[string]interface{}{"mood": "calm"})
	confidence := "explicit"
	snap.TimelineConfidence = &confidence
	day := 10
	snap.TimelineDay = &day

	err := repo.Create(ctx, snap)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	retrieved, err := repo.GetLatestByEntity(ctx, storyID, entityID, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected snapshot, got nil")
	}
	if retrieved.TimelineConfidence == nil {
		t.Fatal("Expected TimelineConfidence to be set")
	}
	if *retrieved.TimelineConfidence != "explicit" {
		t.Errorf("Expected TimelineConfidence %q, got %q", "explicit", *retrieved.TimelineConfidence)
	}
}
