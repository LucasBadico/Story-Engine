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

// mustCreateTenantRS creates a tenant for relationship state tests
func mustCreateTenantRS(t *testing.T, ctx context.Context, db *DB) *tenant.Tenant {
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

// mustCreateWorldRS creates a world for relationship state tests
func mustCreateWorldRS(t *testing.T, ctx context.Context, db *DB, tenantID uuid.UUID) *world.World {
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

// mustCreateCharacterRS creates a character for relationship state tests
func mustCreateCharacterRS(t *testing.T, ctx context.Context, db *DB, tenantID, worldID uuid.UUID, name string) *world.Character {
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

// mustCreateStoryRS creates a story for relationship state tests
func mustCreateStoryRS(t *testing.T, ctx context.Context, db *DB, tenantID uuid.UUID) *story.Story {
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

// mustCreateChapterRS creates a chapter for relationship state tests
func mustCreateChapterRS(t *testing.T, ctx context.Context, db *DB, tenantID, storyID uuid.UUID, number int) *story.Chapter {
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

func TestRelationshipStateRepository_CreateAndGetByID(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantRS(t, ctx, db)
	w := mustCreateWorldRS(t, ctx, db, tn.ID)
	src := mustCreateCharacterRS(t, ctx, db, tn.ID, w.ID, "Hero")
	tgt := mustCreateCharacterRS(t, ctx, db, tn.ID, w.ID, "Villain")
	s := mustCreateStoryRS(t, ctx, db, tn.ID)
	chap := mustCreateChapterRS(t, ctx, db, tn.ID, s.ID, 1)

	repo := NewRelationshipStateRepository(db)

	rs, err := world.NewRelationshipState(tn.ID, chap.ID, 1, "mutual distrust")
	if err != nil {
		t.Fatalf("failed to create relationship state: %v", err)
	}
	rs.AddEntity("source", "character", src.ID)
	rs.AddEntity("target", "character", tgt.ID)

	if err := repo.Create(ctx, rs); err != nil {
		t.Fatalf("failed to persist relationship state: %v", err)
	}

	// Create entity links
	for _, e := range rs.Entities {
		if err := repo.AddEntity(ctx, &e); err != nil {
			t.Fatalf("failed to add entity: %v", err)
		}
	}

	retrieved, err := repo.GetByID(ctx, tn.ID, rs.ID)
	if err != nil {
		t.Fatalf("failed to retrieve relationship state: %v", err)
	}

	if retrieved.ID != rs.ID {
		t.Errorf("expected ID %s, got %s", rs.ID, retrieved.ID)
	}
	if retrieved.TenantID != tn.ID {
		t.Errorf("expected TenantID %s, got %s", tn.ID, retrieved.TenantID)
	}
	if retrieved.ChapterID != chap.ID {
		t.Errorf("expected ChapterID %s, got %s", chap.ID, retrieved.ChapterID)
	}
	if retrieved.ChapterNumber != 1 {
		t.Errorf("expected ChapterNumber 1, got %d", retrieved.ChapterNumber)
	}
	if retrieved.TensionDescription != "mutual distrust" {
		t.Errorf("expected TensionDescription 'mutual distrust', got '%s'", retrieved.TensionDescription)
	}
	if len(retrieved.Entities) != 2 {
		t.Errorf("expected 2 entities, got %d", len(retrieved.Entities))
	}

	// Non-existing ID
	_, err = repo.GetByID(ctx, tn.ID, uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existing relationship state")
	}
	if !platformerrors.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestRelationshipStateRepository_GetLatestByEntityPair(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantRS(t, ctx, db)
	w := mustCreateWorldRS(t, ctx, db, tn.ID)
	src := mustCreateCharacterRS(t, ctx, db, tn.ID, w.ID, "Hero")
	tgt := mustCreateCharacterRS(t, ctx, db, tn.ID, w.ID, "Villain")
	s := mustCreateStoryRS(t, ctx, db, tn.ID)
	chap1 := mustCreateChapterRS(t, ctx, db, tn.ID, s.ID, 1)
	chap3 := mustCreateChapterRS(t, ctx, db, tn.ID, s.ID, 3)
	chap5 := mustCreateChapterRS(t, ctx, db, tn.ID, s.ID, 5)

	repo := NewRelationshipStateRepository(db)

	// Create states at chapters 1, 3, 5
	rs1, _ := world.NewRelationshipState(tn.ID, chap1.ID, 1, "strangers")
	rs1.AddEntity("source", "character", src.ID)
	rs1.AddEntity("target", "character", tgt.ID)
	repo.Create(ctx, rs1)
	for _, e := range rs1.Entities {
		repo.AddEntity(ctx, &e)
	}

	rs3, _ := world.NewRelationshipState(tn.ID, chap3.ID, 3, "rivals")
	rs3.AddEntity("source", "character", src.ID)
	rs3.AddEntity("target", "character", tgt.ID)
	repo.Create(ctx, rs3)
	for _, e := range rs3.Entities {
		repo.AddEntity(ctx, &e)
	}

	rs5, _ := world.NewRelationshipState(tn.ID, chap5.ID, 5, "enemies")
	rs5.AddEntity("source", "character", src.ID)
	rs5.AddEntity("target", "character", tgt.ID)
	repo.Create(ctx, rs5)
	for _, e := range rs5.Entities {
		repo.AddEntity(ctx, &e)
	}

	// Query at chapter 4 should return chapter 3 state
	latest, err := repo.GetLatestByEntityPair(ctx, tn.ID, "character", src.ID, "character", tgt.ID, 4)
	if err != nil {
		t.Fatalf("failed to get latest by entity pair: %v", err)
	}
	if latest.ID != rs3.ID {
		t.Errorf("expected state from chapter 3 (ID=%s), got chapter %d (ID=%s)", rs3.ID, latest.ChapterNumber, latest.ID)
	}
	if latest.TensionDescription != "rivals" {
		t.Errorf("expected TensionDescription 'rivals', got '%s'", latest.TensionDescription)
	}

	// Query at chapter 5 should return chapter 5 state
	latest5, err := repo.GetLatestByEntityPair(ctx, tn.ID, "character", src.ID, "character", tgt.ID, 5)
	if err != nil {
		t.Fatalf("failed to get latest by entity pair at chapter 5: %v", err)
	}
	if latest5.ID != rs5.ID {
		t.Errorf("expected state from chapter 5, got chapter %d", latest5.ChapterNumber)
	}
}

func TestRelationshipStateRepository_ListByChapter(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantRS(t, ctx, db)
	w := mustCreateWorldRS(t, ctx, db, tn.ID)
	ch1 := mustCreateCharacterRS(t, ctx, db, tn.ID, w.ID, "Hero")
	ch2 := mustCreateCharacterRS(t, ctx, db, tn.ID, w.ID, "Villain")
	ch3 := mustCreateCharacterRS(t, ctx, db, tn.ID, w.ID, "Sidekick")
	s := mustCreateStoryRS(t, ctx, db, tn.ID)
	chap := mustCreateChapterRS(t, ctx, db, tn.ID, s.ID, 1)

	repo := NewRelationshipStateRepository(db)

	rs1, _ := world.NewRelationshipState(tn.ID, chap.ID, 1, "rivalry")
	rs1.AddEntity("source", "character", ch1.ID)
	rs1.AddEntity("target", "character", ch2.ID)
	repo.Create(ctx, rs1)
	for _, e := range rs1.Entities {
		repo.AddEntity(ctx, &e)
	}

	rs2, _ := world.NewRelationshipState(tn.ID, chap.ID, 1, "friendship")
	rs2.AddEntity("source", "character", ch1.ID)
	rs2.AddEntity("target", "character", ch3.ID)
	repo.Create(ctx, rs2)
	for _, e := range rs2.Entities {
		repo.AddEntity(ctx, &e)
	}

	states, err := repo.ListByChapter(ctx, tn.ID, chap.ID)
	if err != nil {
		t.Fatalf("failed to list by chapter: %v", err)
	}
	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
	}
}

func TestRelationshipStateRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tn := mustCreateTenantRS(t, ctx, db)
	s := mustCreateStoryRS(t, ctx, db, tn.ID)
	chap := mustCreateChapterRS(t, ctx, db, tn.ID, s.ID, 1)

	repo := NewRelationshipStateRepository(db)

	rs, _ := world.NewRelationshipState(tn.ID, chap.ID, 1, "tension")
	repo.Create(ctx, rs)

	err := repo.Delete(ctx, tn.ID, rs.ID)
	if err != nil {
		t.Fatalf("failed to delete relationship state: %v", err)
	}

	_, err = repo.GetByID(ctx, tn.ID, rs.ID)
	if err == nil {
		t.Fatal("expected error for deleted relationship state")
	}
	if !platformerrors.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}
