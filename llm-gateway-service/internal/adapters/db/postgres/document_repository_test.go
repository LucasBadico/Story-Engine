//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

func TestDocumentRepository_Create(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewDocumentRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	sourceID := uuid.New()
	doc := memory.NewDocument(tenantID, memory.SourceTypeStory, sourceID, "Test Story", "Test content")

	err := repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if doc.ID == uuid.Nil {
		t.Error("Expected document ID to be set")
	}
}

func TestDocumentRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewDocumentRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	sourceID := uuid.New()
	doc := memory.NewDocument(tenantID, memory.SourceTypeStory, sourceID, "Test Story", "Test content")

	err := repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.ID != doc.ID {
		t.Errorf("Expected ID %v, got %v", doc.ID, retrieved.ID)
	}
	if retrieved.Title != doc.Title {
		t.Errorf("Expected Title %q, got %q", doc.Title, retrieved.Title)
	}
	if retrieved.Content != doc.Content {
		t.Errorf("Expected Content %q, got %q", doc.Content, retrieved.Content)
	}
}

func TestDocumentRepository_GetBySource(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewDocumentRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	sourceID := uuid.New()
	doc := memory.NewDocument(tenantID, memory.SourceTypeStory, sourceID, "Test Story", "Test content")

	err := repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	retrieved, err := repo.GetBySource(ctx, tenantID, memory.SourceTypeStory, sourceID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.ID != doc.ID {
		t.Errorf("Expected ID %v, got %v", doc.ID, retrieved.ID)
	}
}

func TestDocumentRepository_Update(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewDocumentRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	sourceID := uuid.New()
	doc := memory.NewDocument(tenantID, memory.SourceTypeStory, sourceID, "Test Story", "Test content")

	err := repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Update document
	doc.Title = "Updated Story"
	doc.Content = "Updated content"
	err = repo.Update(ctx, doc)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.Title != "Updated Story" {
		t.Errorf("Expected Title %q, got %q", "Updated Story", retrieved.Title)
	}
	if retrieved.Content != "Updated content" {
		t.Errorf("Expected Content %q, got %q", "Updated content", retrieved.Content)
	}
}

func TestDocumentRepository_UniqueConstraint(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewDocumentRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	sourceID := uuid.New()
	doc1 := memory.NewDocument(tenantID, memory.SourceTypeStory, sourceID, "Test Story", "Test content")

	err := repo.Create(ctx, doc1)
	if err != nil {
		t.Fatalf("Failed to create first document: %v", err)
	}

	// Try to create another document with same tenant, source type, and source ID
	doc2 := memory.NewDocument(tenantID, memory.SourceTypeStory, sourceID, "Another Story", "Another content")
	err = repo.Create(ctx, doc2)
	if err == nil {
		t.Error("Expected error when creating duplicate document")
	}
}

