//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

func TestConceptRepository_CreateAndGetByID(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewConceptRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()
	concept := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)
	concept.Metadata = map[string]interface{}{"age": float64(30), "role": "protagonist"}

	err := repo.Create(ctx, concept)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	retrieved, err := repo.GetByID(ctx, concept.ID)
	if err != nil {
		t.Fatalf("Expected no error retrieving concept, got %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected concept, got nil")
	}
	if retrieved.ID != concept.ID {
		t.Errorf("Expected ID %v, got %v", concept.ID, retrieved.ID)
	}
	if retrieved.TenantID != tenantID {
		t.Errorf("Expected TenantID %v, got %v", tenantID, retrieved.TenantID)
	}
	if retrieved.StoryID != storyID {
		t.Errorf("Expected StoryID %v, got %v", storyID, retrieved.StoryID)
	}
	if retrieved.Name != "Alice" {
		t.Errorf("Expected Name %q, got %q", "Alice", retrieved.Name)
	}
	if retrieved.Type != memory.ConceptTypeCharacter {
		t.Errorf("Expected Type %q, got %q", memory.ConceptTypeCharacter, retrieved.Type)
	}
	if retrieved.Metadata["role"] != "protagonist" {
		t.Errorf("Expected Metadata role %q, got %v", "protagonist", retrieved.Metadata["role"])
	}
}

func TestConceptRepository_FindByName(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewConceptRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()
	concept := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)

	err := repo.Create(ctx, concept)
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	// Find by name with different case
	found, err := repo.FindByName(ctx, storyID, "alice", memory.ConceptTypeCharacter)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if found == nil {
		t.Fatal("Expected concept, got nil")
	}
	if found.ID != concept.ID {
		t.Errorf("Expected ID %v, got %v", concept.ID, found.ID)
	}

	// Find nonexistent returns nil
	notFound, err := repo.FindByName(ctx, storyID, "Bob", memory.ConceptTypeCharacter)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if notFound != nil {
		t.Errorf("Expected nil for nonexistent concept, got %v", notFound)
	}
}

func TestConceptRepository_List(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewConceptRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()

	// Create 2 characters and 1 location
	char1 := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)
	char2 := memory.NewConcept(tenantID, storyID, "Bob", memory.ConceptTypeCharacter)
	loc := memory.NewConcept(tenantID, storyID, "Castle", memory.ConceptTypeLocation)

	for _, c := range []*memory.Concept{char1, char2, loc} {
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("Failed to create concept: %v", err)
		}
	}

	// List characters only
	characters, err := repo.List(ctx, storyID, memory.ConceptTypeCharacter)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(characters) != 2 {
		t.Errorf("Expected 2 characters, got %d", len(characters))
	}
}

func TestConceptRepository_Update(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewConceptRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()
	concept := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)
	concept.Metadata = map[string]interface{}{"age": float64(30)}

	err := repo.Create(ctx, concept)
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	// Update name and metadata
	concept.Name = "Alice Updated"
	concept.Metadata = map[string]interface{}{"age": float64(31), "title": "Queen"}
	concept.UpdatedAt = time.Now()

	err = repo.Update(ctx, concept)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	retrieved, err := repo.GetByID(ctx, concept.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved.Name != "Alice Updated" {
		t.Errorf("Expected Name %q, got %q", "Alice Updated", retrieved.Name)
	}
	if retrieved.Metadata["title"] != "Queen" {
		t.Errorf("Expected Metadata title %q, got %v", "Queen", retrieved.Metadata["title"])
	}
}

func TestConceptRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewConceptRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()
	concept := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)

	err := repo.Create(ctx, concept)
	if err != nil {
		t.Fatalf("Failed to create concept: %v", err)
	}

	err = repo.Delete(ctx, concept.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	retrieved, err := repo.GetByID(ctx, concept.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved != nil {
		t.Errorf("Expected nil after delete, got %v", retrieved)
	}
}

func TestConceptRepository_UniqueConstraint(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewConceptRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	storyID := uuid.New()

	concept1 := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)
	err := repo.Create(ctx, concept1)
	if err != nil {
		t.Fatalf("Failed to create first concept: %v", err)
	}

	// Create second concept with same (tenant, story, type, name)
	concept2 := memory.NewConcept(tenantID, storyID, "Alice", memory.ConceptTypeCharacter)
	err = repo.Create(ctx, concept2)
	if err == nil {
		t.Error("Expected error when creating duplicate concept")
	}
}
