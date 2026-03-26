//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/core/tenant"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
)

// mustCreateTenantCS creates a tenant for character state tests
func mustCreateTenantCS(t *testing.T, ctx context.Context, db *DB) *tenant.Tenant {
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

// mustCreateWorldCS creates a world for character state tests
func mustCreateWorldCS(t *testing.T, ctx context.Context, db *DB, tenantID uuid.UUID) *world.World {
	t.Helper()
	repo := NewWorldRepository(db)
	w, err := world.NewWorld(tenantID, "Test World", false)
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}
	if err := repo.Create(ctx, w); err != nil {
		t.Fatalf("failed to persist world: %v", err)
	}
	return w
}

// mustCreateCharacterCS creates a character for character state tests
func mustCreateCharacterCS(t *testing.T, ctx context.Context, db *DB, tenantID, worldID uuid.UUID, name string) *world.Character {
	t.Helper()
	repo := NewCharacterRepository(db)
	c, err := world.NewCharacter(tenantID, worldID, name)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("failed to persist character: %v", err)
	}
	return c
}

// mustCreateStoryCS creates a story for character state tests
func mustCreateStoryCS(t *testing.T, ctx context.Context, db *DB, tenantID uuid.UUID) *story.Story {
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

// mustCreateChapterCS creates a chapter for character state tests
func mustCreateChapterCS(t *testing.T, ctx context.Context, db *DB, tenantID, storyID uuid.UUID, number int) *story.Chapter {
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

func TestCharacterStateRepository_CreateAndGetByID(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantCS(t, ctx, db)
	w := mustCreateWorldCS(t, ctx, db, tn.ID)
	ch := mustCreateCharacterCS(t, ctx, db, tn.ID, w.ID, "Hero")
	s := mustCreateStoryCS(t, ctx, db, tn.ID)
	chap := mustCreateChapterCS(t, ctx, db, tn.ID, s.ID, 1)

	repo := NewCharacterStateRepository(db)

	cs, err := world.NewCharacterState(tn.ID, ch.ID, 1)
	if err != nil {
		t.Fatalf("failed to create character state: %v", err)
	}
	cs.ChapterID = &chap.ID
	cs.SetEmotion("happy")
	cs.SetGoal("save the world")
	cs.SetSecrets([]string{"knows the truth", "hides a ring"})
	cs.SetClothes("armor")
	cs.SetPhysicalState("healthy")

	if err := repo.Create(ctx, cs); err != nil {
		t.Fatalf("failed to persist character state: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, tn.ID, cs.ID)
	if err != nil {
		t.Fatalf("failed to retrieve character state: %v", err)
	}

	if retrieved.ID != cs.ID {
		t.Errorf("expected ID %s, got %s", cs.ID, retrieved.ID)
	}
	if retrieved.TenantID != tn.ID {
		t.Errorf("expected TenantID %s, got %s", tn.ID, retrieved.TenantID)
	}
	if retrieved.CharacterID != ch.ID {
		t.Errorf("expected CharacterID %s, got %s", ch.ID, retrieved.CharacterID)
	}
	if retrieved.ChapterID == nil || *retrieved.ChapterID != chap.ID {
		t.Errorf("expected ChapterID %s, got %v", chap.ID, retrieved.ChapterID)
	}
	if retrieved.ChapterNumber != 1 {
		t.Errorf("expected ChapterNumber 1, got %d", retrieved.ChapterNumber)
	}
	if retrieved.Emotion != "happy" {
		t.Errorf("expected Emotion 'happy', got '%s'", retrieved.Emotion)
	}
	if retrieved.Goal != "save the world" {
		t.Errorf("expected Goal 'save the world', got '%s'", retrieved.Goal)
	}
	if len(retrieved.Secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(retrieved.Secrets))
	}
	if retrieved.Clothes != "armor" {
		t.Errorf("expected Clothes 'armor', got '%s'", retrieved.Clothes)
	}
	if retrieved.PhysicalState != "healthy" {
		t.Errorf("expected PhysicalState 'healthy', got '%s'", retrieved.PhysicalState)
	}

	// Non-existing ID
	_, err = repo.GetByID(ctx, tn.ID, uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existing character state")
	}
	if !platformerrors.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCharacterStateRepository_GetLatestByCharacter(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantCS(t, ctx, db)
	w := mustCreateWorldCS(t, ctx, db, tn.ID)
	ch := mustCreateCharacterCS(t, ctx, db, tn.ID, w.ID, "Hero")

	repo := NewCharacterStateRepository(db)

	// Create states at chapters 1, 3, 5
	cs1, _ := world.NewCharacterState(tn.ID, ch.ID, 1)
	cs1.SetEmotion("calm")
	repo.Create(ctx, cs1)

	cs3, _ := world.NewCharacterState(tn.ID, ch.ID, 3)
	cs3.SetEmotion("angry")
	repo.Create(ctx, cs3)

	cs5, _ := world.NewCharacterState(tn.ID, ch.ID, 5)
	cs5.SetEmotion("sad")
	repo.Create(ctx, cs5)

	// Query at chapter 4 should return chapter 3 state
	latest, err := repo.GetLatestByCharacter(ctx, tn.ID, ch.ID, 4)
	if err != nil {
		t.Fatalf("failed to get latest by character: %v", err)
	}
	if latest.ID != cs3.ID {
		t.Errorf("expected state from chapter 3 (ID=%s), got chapter %d (ID=%s)", cs3.ID, latest.ChapterNumber, latest.ID)
	}
	if latest.Emotion != "angry" {
		t.Errorf("expected Emotion 'angry', got '%s'", latest.Emotion)
	}

	// Query at chapter 5 should return chapter 5 state
	latest5, err := repo.GetLatestByCharacter(ctx, tn.ID, ch.ID, 5)
	if err != nil {
		t.Fatalf("failed to get latest by character at chapter 5: %v", err)
	}
	if latest5.ID != cs5.ID {
		t.Errorf("expected state from chapter 5, got chapter %d", latest5.ChapterNumber)
	}
}

func TestCharacterStateRepository_ListByChapter(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantCS(t, ctx, db)
	w := mustCreateWorldCS(t, ctx, db, tn.ID)
	ch1 := mustCreateCharacterCS(t, ctx, db, tn.ID, w.ID, "Hero")
	ch2 := mustCreateCharacterCS(t, ctx, db, tn.ID, w.ID, "Villain")
	s := mustCreateStoryCS(t, ctx, db, tn.ID)
	chap := mustCreateChapterCS(t, ctx, db, tn.ID, s.ID, 1)

	repo := NewCharacterStateRepository(db)

	cs1, _ := world.NewCharacterState(tn.ID, ch1.ID, 1)
	cs1.ChapterID = &chap.ID
	repo.Create(ctx, cs1)

	cs2, _ := world.NewCharacterState(tn.ID, ch2.ID, 1)
	cs2.ChapterID = &chap.ID
	repo.Create(ctx, cs2)

	states, err := repo.ListByChapter(ctx, tn.ID, chap.ID)
	if err != nil {
		t.Fatalf("failed to list by chapter: %v", err)
	}
	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
	}
}

func TestCharacterStateRepository_ListByCharacter(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantCS(t, ctx, db)
	w := mustCreateWorldCS(t, ctx, db, tn.ID)
	ch := mustCreateCharacterCS(t, ctx, db, tn.ID, w.ID, "Hero")

	repo := NewCharacterStateRepository(db)

	cs1, _ := world.NewCharacterState(tn.ID, ch.ID, 1)
	cs1.SetEmotion("calm")
	repo.Create(ctx, cs1)

	cs3, _ := world.NewCharacterState(tn.ID, ch.ID, 3)
	cs3.SetEmotion("angry")
	repo.Create(ctx, cs3)

	states, err := repo.ListByCharacter(ctx, tn.ID, ch.ID)
	if err != nil {
		t.Fatalf("failed to list by character: %v", err)
	}
	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
	}

	// Should be ordered by chapter_number
	if states[0].ChapterNumber > states[1].ChapterNumber {
		t.Errorf("expected states ordered by chapter_number, got %d before %d", states[0].ChapterNumber, states[1].ChapterNumber)
	}
}

func TestCharacterStateRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantCS(t, ctx, db)
	w := mustCreateWorldCS(t, ctx, db, tn.ID)
	ch := mustCreateCharacterCS(t, ctx, db, tn.ID, w.ID, "Hero")

	repo := NewCharacterStateRepository(db)

	cs, _ := world.NewCharacterState(tn.ID, ch.ID, 1)
	repo.Create(ctx, cs)

	err := repo.Delete(ctx, tn.ID, cs.ID)
	if err != nil {
		t.Fatalf("failed to delete character state: %v", err)
	}

	_, err = repo.GetByID(ctx, tn.ID, cs.ID)
	if err == nil {
		t.Fatal("expected error for deleted character state")
	}
	if !platformerrors.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCharacterStateRepository_TimelineFields(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantCS(t, ctx, db)
	w := mustCreateWorldCS(t, ctx, db, tn.ID)
	ch := mustCreateCharacterCS(t, ctx, db, tn.ID, w.ID, "Hero")

	repo := NewCharacterStateRepository(db)

	cs, _ := world.NewCharacterState(tn.ID, ch.ID, 1)
	cs.SetTimeline(-40, "explicit")

	if err := repo.Create(ctx, cs); err != nil {
		t.Fatalf("failed to persist character state: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, tn.ID, cs.ID)
	if err != nil {
		t.Fatalf("failed to retrieve character state: %v", err)
	}

	if retrieved.TimelineDay == nil {
		t.Fatal("expected TimelineDay to be set")
	}
	if *retrieved.TimelineDay != -40 {
		t.Errorf("expected TimelineDay -40, got %d", *retrieved.TimelineDay)
	}

	if retrieved.TimelineConfidence == nil {
		t.Fatal("expected TimelineConfidence to be set")
	}
	if *retrieved.TimelineConfidence != "explicit" {
		t.Errorf("expected TimelineConfidence 'explicit', got '%s'", *retrieved.TimelineConfidence)
	}
}
