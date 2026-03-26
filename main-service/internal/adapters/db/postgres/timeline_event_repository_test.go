//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/core/tenant"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
)

// mustCreateTenantTE creates a tenant for timeline event tests
func mustCreateTenantTE(t *testing.T, ctx context.Context, db *DB) *tenant.Tenant {
	t.Helper()
	repo := NewTenantRepository(db)
	tn, err := tenant.NewTenant("Test Tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	if err := repo.Create(ctx, tn); err != nil {
		t.Fatalf("failed to persist tenant: %v", err)
	}
	return tn
}

// mustCreateStoryTE creates a story for timeline event tests
func mustCreateStoryTE(t *testing.T, ctx context.Context, db *DB, tenantID uuid.UUID) *story.Story {
	t.Helper()
	repo := NewStoryRepository(db)
	s, err := story.NewStory(tenantID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	if err := repo.Create(ctx, s); err != nil {
		t.Fatalf("failed to persist story: %v", err)
	}
	return s
}

// mustCreateChapterTE creates a chapter for timeline event tests
func mustCreateChapterTE(t *testing.T, ctx context.Context, db *DB, tenantID, storyID uuid.UUID, number int) *story.Chapter {
	t.Helper()
	repo := NewChapterRepository(db)
	ch, err := story.NewChapter(tenantID, storyID, number, "Chapter")
	if err != nil {
		t.Fatalf("failed to create chapter: %v", err)
	}
	if err := repo.Create(ctx, ch); err != nil {
		t.Fatalf("failed to persist chapter: %v", err)
	}
	return ch
}

func TestTimelineEventRepository_CreateAndGetByID(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantTE(t, ctx, db)
	s := mustCreateStoryTE(t, ctx, db, tn.ID)
	chap := mustCreateChapterTE(t, ctx, db, tn.ID, s.ID, 1)

	repo := NewTimelineEventRepository(db)

	te, err := story.NewTimelineEvent(tn.ID, s.ID, 1, "The hero arrived at the castle")
	if err != nil {
		t.Fatalf("failed to create timeline event: %v", err)
	}
	te.SetChapterID(chap.ID)
	te.SetTimeline(5, "explicit")
	te.AddEntity("participant", "character", uuid.New())
	te.AddEntity("location", "location", uuid.New())

	if err := repo.Create(ctx, te); err != nil {
		t.Fatalf("failed to persist timeline event: %v", err)
	}

	// Create entity links
	for _, e := range te.Entities {
		if err := repo.AddEntity(ctx, &e); err != nil {
			t.Fatalf("failed to add entity: %v", err)
		}
	}

	retrieved, err := repo.GetByID(ctx, tn.ID, te.ID)
	if err != nil {
		t.Fatalf("failed to retrieve timeline event: %v", err)
	}

	if retrieved.ID != te.ID {
		t.Errorf("expected ID %s, got %s", te.ID, retrieved.ID)
	}
	if retrieved.TenantID != tn.ID {
		t.Errorf("expected TenantID %s, got %s", tn.ID, retrieved.TenantID)
	}
	if retrieved.StoryID != s.ID {
		t.Errorf("expected StoryID %s, got %s", s.ID, retrieved.StoryID)
	}
	if retrieved.ChapterID != chap.ID {
		t.Errorf("expected ChapterID %s, got %s", chap.ID, retrieved.ChapterID)
	}
	if retrieved.ChapterNumber != 1 {
		t.Errorf("expected ChapterNumber 1, got %d", retrieved.ChapterNumber)
	}
	if retrieved.Description != "The hero arrived at the castle" {
		t.Errorf("expected Description 'The hero arrived at the castle', got '%s'", retrieved.Description)
	}
	if retrieved.TimelineDay == nil || *retrieved.TimelineDay != 5 {
		t.Errorf("expected TimelineDay 5, got %v", retrieved.TimelineDay)
	}
	if retrieved.TimelineConfidence == nil || *retrieved.TimelineConfidence != "explicit" {
		t.Errorf("expected TimelineConfidence 'explicit', got %v", retrieved.TimelineConfidence)
	}
	if len(retrieved.Entities) != 2 {
		t.Errorf("expected 2 entities, got %d", len(retrieved.Entities))
	}

	// Non-existing ID
	_, err = repo.GetByID(ctx, tn.ID, uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existing timeline event")
	}
	if !platformerrors.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestTimelineEventRepository_ListByStory(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantTE(t, ctx, db)
	s := mustCreateStoryTE(t, ctx, db, tn.ID)
	chap1 := mustCreateChapterTE(t, ctx, db, tn.ID, s.ID, 1)
	chap3 := mustCreateChapterTE(t, ctx, db, tn.ID, s.ID, 3)
	chap5 := mustCreateChapterTE(t, ctx, db, tn.ID, s.ID, 5)

	repo := NewTimelineEventRepository(db)

	// Create events at chapters 1, 3, 5
	te1, _ := story.NewTimelineEvent(tn.ID, s.ID, 1, "Event at chapter 1")
	te1.SetChapterID(chap1.ID)
	repo.Create(ctx, te1)

	te3, _ := story.NewTimelineEvent(tn.ID, s.ID, 3, "Event at chapter 3")
	te3.SetChapterID(chap3.ID)
	repo.Create(ctx, te3)

	te5, _ := story.NewTimelineEvent(tn.ID, s.ID, 5, "Event at chapter 5")
	te5.SetChapterID(chap5.ID)
	repo.Create(ctx, te5)

	// Query at chapter 4 with limit 10 should return 2 events (chapters 1 and 3)
	events, err := repo.ListByStory(ctx, tn.ID, s.ID, 4, 10)
	if err != nil {
		t.Fatalf("failed to list by story: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	// Query at chapter 5 should return all 3
	events5, err := repo.ListByStory(ctx, tn.ID, s.ID, 5, 10)
	if err != nil {
		t.Fatalf("failed to list by story at chapter 5: %v", err)
	}
	if len(events5) != 3 {
		t.Errorf("expected 3 events, got %d", len(events5))
	}
}

func TestTimelineEventRepository_ListByStory_WithLimit(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantTE(t, ctx, db)
	s := mustCreateStoryTE(t, ctx, db, tn.ID)

	repo := NewTimelineEventRepository(db)

	// Create 5 events with chapters
	for i := 1; i <= 5; i++ {
		chap := mustCreateChapterTE(t, ctx, db, tn.ID, s.ID, i)
		te, _ := story.NewTimelineEvent(tn.ID, s.ID, i, "Event")
		te.SetChapterID(chap.ID)
		repo.Create(ctx, te)
	}

	// Query with limit 2 should return exactly 2
	events, err := repo.ListByStory(ctx, tn.ID, s.ID, 10, 2)
	if err != nil {
		t.Fatalf("failed to list by story with limit: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestTimelineEventRepository_ListByChapter(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantTE(t, ctx, db)
	s := mustCreateStoryTE(t, ctx, db, tn.ID)
	chap := mustCreateChapterTE(t, ctx, db, tn.ID, s.ID, 1)

	repo := NewTimelineEventRepository(db)

	te1, _ := story.NewTimelineEvent(tn.ID, s.ID, 1, "Event A")
	te1.SetChapterID(chap.ID)
	repo.Create(ctx, te1)

	te2, _ := story.NewTimelineEvent(tn.ID, s.ID, 1, "Event B")
	te2.SetChapterID(chap.ID)
	repo.Create(ctx, te2)

	events, err := repo.ListByChapter(ctx, tn.ID, chap.ID)
	if err != nil {
		t.Fatalf("failed to list by chapter: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestTimelineEventRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantTE(t, ctx, db)
	s := mustCreateStoryTE(t, ctx, db, tn.ID)
	chap := mustCreateChapterTE(t, ctx, db, tn.ID, s.ID, 1)

	repo := NewTimelineEventRepository(db)

	te, _ := story.NewTimelineEvent(tn.ID, s.ID, 1, "To be deleted")
	te.SetChapterID(chap.ID)
	repo.Create(ctx, te)

	err := repo.Delete(ctx, tn.ID, te.ID)
	if err != nil {
		t.Fatalf("failed to delete timeline event: %v", err)
	}

	_, err = repo.GetByID(ctx, tn.ID, te.ID)
	if err == nil {
		t.Fatal("expected error for deleted timeline event")
	}
	if !platformerrors.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}
