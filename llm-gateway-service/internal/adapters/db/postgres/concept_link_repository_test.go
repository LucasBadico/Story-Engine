//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

func TestConceptLinkRepository_CreateAndGetByID(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	conceptRepo := NewConceptRepository(db)
	repo := NewConceptLinkRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()

	// Create two concepts
	conceptA := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)
	conceptB := memory.NewConcept(tenantID, storyID, "Bob", memory.ConceptTypeCharacter)
	if err := conceptRepo.Create(ctx, conceptA); err != nil {
		t.Fatalf("Failed to create concept A: %v", err)
	}
	if err := conceptRepo.Create(ctx, conceptB); err != nil {
		t.Fatalf("Failed to create concept B: %v", err)
	}

	// Create link
	link := memory.NewConceptLink(tenantID, conceptA.ID, conceptB.ID, "friend")
	link.Summary = "Childhood friends"
	err := repo.Create(ctx, link)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Get by ID
	retrieved, err := repo.GetByID(ctx, link.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected link, got nil")
	}
	if retrieved.ID != link.ID {
		t.Errorf("Expected ID %v, got %v", link.ID, retrieved.ID)
	}
	if retrieved.SourceConceptID != conceptA.ID {
		t.Errorf("Expected SourceConceptID %v, got %v", conceptA.ID, retrieved.SourceConceptID)
	}
	if retrieved.TargetConceptID != conceptB.ID {
		t.Errorf("Expected TargetConceptID %v, got %v", conceptB.ID, retrieved.TargetConceptID)
	}
	if retrieved.RelationType != "friend" {
		t.Errorf("Expected RelationType %q, got %q", "friend", retrieved.RelationType)
	}
	if retrieved.Summary != "Childhood friends" {
		t.Errorf("Expected Summary %q, got %q", "Childhood friends", retrieved.Summary)
	}
	if retrieved.ValidFromChapter != 0 {
		t.Errorf("Expected ValidFromChapter 0, got %d", retrieved.ValidFromChapter)
	}
	if retrieved.ValidToChapter != nil {
		t.Errorf("Expected ValidToChapter nil, got %v", retrieved.ValidToChapter)
	}
}

func TestConceptLinkRepository_FindByPair(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	conceptRepo := NewConceptRepository(db)
	repo := NewConceptLinkRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()

	conceptA := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)
	conceptB := memory.NewConcept(tenantID, storyID, "Bob", memory.ConceptTypeCharacter)
	if err := conceptRepo.Create(ctx, conceptA); err != nil {
		t.Fatalf("Failed to create concept A: %v", err)
	}
	if err := conceptRepo.Create(ctx, conceptB); err != nil {
		t.Fatalf("Failed to create concept B: %v", err)
	}

	link := memory.NewConceptLink(tenantID, conceptA.ID, conceptB.ID, "rival")
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Failed to create link: %v", err)
	}

	// Find by source+target
	found, err := repo.FindByPair(ctx, conceptA.ID, conceptB.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if found == nil {
		t.Fatal("Expected link, got nil")
	}
	if found.ID != link.ID {
		t.Errorf("Expected ID %v, got %v", link.ID, found.ID)
	}

	// Find nonexistent pair returns nil
	notFound, err := repo.FindByPair(ctx, conceptB.ID, conceptA.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if notFound != nil {
		t.Errorf("Expected nil for nonexistent pair, got %v", notFound)
	}
}

func TestConceptLinkRepository_CloseVersion(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	conceptRepo := NewConceptRepository(db)
	repo := NewConceptLinkRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()

	conceptA := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)
	conceptB := memory.NewConcept(tenantID, storyID, "Bob", memory.ConceptTypeCharacter)
	if err := conceptRepo.Create(ctx, conceptA); err != nil {
		t.Fatalf("Failed to create concept A: %v", err)
	}
	if err := conceptRepo.Create(ctx, conceptB); err != nil {
		t.Fatalf("Failed to create concept B: %v", err)
	}

	link := memory.NewConceptLink(tenantID, conceptA.ID, conceptB.ID, "ally")
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Failed to create link: %v", err)
	}

	// Close version at chapter 5
	err := repo.CloseVersion(ctx, link.ID, 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify valid_to_chapter is set
	retrieved, err := repo.GetByID(ctx, link.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved.ValidToChapter == nil {
		t.Fatal("Expected ValidToChapter to be set")
	}
	if *retrieved.ValidToChapter != 5 {
		t.Errorf("Expected ValidToChapter 5, got %d", *retrieved.ValidToChapter)
	}
}

func TestConceptLinkRepository_List_WithChapterFilter(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	conceptRepo := NewConceptRepository(db)
	repo := NewConceptLinkRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()

	conceptA := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)
	conceptB := memory.NewConcept(tenantID, storyID, "Bob", memory.ConceptTypeCharacter)
	if err := conceptRepo.Create(ctx, conceptA); err != nil {
		t.Fatalf("Failed to create concept A: %v", err)
	}
	if err := conceptRepo.Create(ctx, conceptB); err != nil {
		t.Fatalf("Failed to create concept B: %v", err)
	}

	// Link 1: valid_from_chapter=0, valid_to_chapter=5
	link1 := memory.NewConceptLink(tenantID, conceptA.ID, conceptB.ID, "friend")
	link1.ValidFromChapter = 0
	link1.ValidToChapter = intPtr(5)
	if err := repo.Create(ctx, link1); err != nil {
		t.Fatalf("Failed to create link1: %v", err)
	}

	// Link 2: valid_from_chapter=5, valid_to_chapter=nil
	link2 := memory.NewConceptLink(tenantID, conceptA.ID, conceptB.ID, "enemy")
	link2.ValidFromChapter = 5
	if err := repo.Create(ctx, link2); err != nil {
		t.Fatalf("Failed to create link2: %v", err)
	}

	// List at chapter 3 should return link1 only
	chapter3 := 3
	linksAt3, err := repo.List(ctx, storyID, &chapter3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(linksAt3) != 1 {
		t.Fatalf("Expected 1 link at chapter 3, got %d", len(linksAt3))
	}
	if linksAt3[0].RelationType != "friend" {
		t.Errorf("Expected relation_type %q at chapter 3, got %q", "friend", linksAt3[0].RelationType)
	}

	// List at chapter 7 should return link2 only
	chapter7 := 7
	linksAt7, err := repo.List(ctx, storyID, &chapter7)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(linksAt7) != 1 {
		t.Fatalf("Expected 1 link at chapter 7, got %d", len(linksAt7))
	}
	if linksAt7[0].RelationType != "enemy" {
		t.Errorf("Expected relation_type %q at chapter 7, got %q", "enemy", linksAt7[0].RelationType)
	}
}

func TestConceptLinkRepository_ListByConceptID(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	conceptRepo := NewConceptRepository(db)
	repo := NewConceptLinkRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()

	conceptA := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)
	conceptB := memory.NewConcept(tenantID, storyID, "Bob", memory.ConceptTypeCharacter)
	conceptC := memory.NewConcept(tenantID, storyID, "Charlie", memory.ConceptTypeCharacter)
	for _, c := range []*memory.Concept{conceptA, conceptB, conceptC} {
		if err := conceptRepo.Create(ctx, c); err != nil {
			t.Fatalf("Failed to create concept: %v", err)
		}
	}

	// Link A -> B
	linkAB := memory.NewConceptLink(tenantID, conceptA.ID, conceptB.ID, "friend")
	if err := repo.Create(ctx, linkAB); err != nil {
		t.Fatalf("Failed to create link A->B: %v", err)
	}

	// Link C -> A
	linkCA := memory.NewConceptLink(tenantID, conceptC.ID, conceptA.ID, "mentor")
	if err := repo.Create(ctx, linkCA); err != nil {
		t.Fatalf("Failed to create link C->A: %v", err)
	}

	// List by concept A should return both links
	links, err := repo.ListByConceptID(ctx, conceptA.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(links) != 2 {
		t.Errorf("Expected 2 links for concept A, got %d", len(links))
	}
}

func intPtr(i int) *int {
	return &i
}
