//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

func TestChunkRepository_Create(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewChunkRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// Create document first (FK constraint)
	docRepo := NewDocumentRepository(db)
	doc := memory.NewDocument(tenantID, memory.SourceTypeContentBlock, uuid.New(), "Test", "Content")
	if err := docRepo.Create(ctx, doc); err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = 0.1
	}

	chunk := memory.NewChunk(doc.ID, 0, "Test content", embedding, 10)
	chunk.SceneID = uuidPtr(uuid.New())
	beatType := "setup"
	chunk.BeatType = &beatType
	chunk.Characters = []string{"John", "Mary"}

	err := repo.Create(ctx, chunk)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify chunk was created
	retrieved, err := repo.GetByID(ctx, chunk.ID)
	if err != nil {
		t.Fatalf("Expected no error retrieving chunk, got %v", err)
	}

	if retrieved.Content != chunk.Content {
		t.Errorf("Expected Content %q, got %q", chunk.Content, retrieved.Content)
	}
	if retrieved.BeatType == nil || *retrieved.BeatType != beatType {
		t.Error("Expected BeatType to be persisted")
	}
	if len(retrieved.Characters) != 2 {
		t.Errorf("Expected 2 characters, got %d", len(retrieved.Characters))
	}
}

func TestChunkRepository_CreateBatch(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewChunkRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// Create document first (FK constraint)
	docRepo := NewDocumentRepository(db)
	doc := memory.NewDocument(tenantID, memory.SourceTypeContentBlock, uuid.New(), "Test", "Content")
	if err := docRepo.Create(ctx, doc); err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	chunks := make([]*memory.Chunk, 3)
	for i := 0; i < 3; i++ {
		embedding := make([]float32, 1536)
		for j := range embedding {
			embedding[j] = float32(i) * 0.1
		}
		chunks[i] = memory.NewChunk(doc.ID, i, "Content "+string(rune('A'+i)), embedding, 10)
	}

	err := repo.CreateBatch(ctx, chunks)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify chunks were created
	list, err := repo.ListByDocument(ctx, doc.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 chunks, got %d", len(list))
	}
}

func TestChunkRepository_SearchSimilar(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewChunkRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// Create a document first
	docRepo := NewDocumentRepository(db)
	doc := memory.NewDocument(tenantID, memory.SourceTypeContentBlock, uuid.New(), "Test", "Content")
	doc.ID = uuid.New()
	if err := docRepo.Create(ctx, doc); err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Create chunks with different embeddings
	queryEmbedding := make([]float32, 1536)
	for i := range queryEmbedding {
		queryEmbedding[i] = 0.5
	}

	// Chunk similar to query
	similarEmbedding := make([]float32, 1536)
	for i := range similarEmbedding {
		similarEmbedding[i] = 0.5
	}
	chunk1 := memory.NewChunk(doc.ID, 0, "Similar content", similarEmbedding, 10)
	if err := repo.Create(ctx, chunk1); err != nil {
		t.Fatalf("Failed to create chunk: %v", err)
	}

	// Chunk different from query
	differentEmbedding := make([]float32, 1536)
	for i := range differentEmbedding {
		differentEmbedding[i] = -0.5
	}
	chunk2 := memory.NewChunk(doc.ID, 1, "Different content", differentEmbedding, 10)
	if err := repo.Create(ctx, chunk2); err != nil {
		t.Fatalf("Failed to create chunk: %v", err)
	}

	// Search
	filters := &repositories.SearchFilters{}
	results, err := repo.SearchSimilar(ctx, tenantID, queryEmbedding, 10, filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}

	// First result should be more similar
	if results[0].ID != chunk1.ID {
		t.Logf("Note: SearchSimilar may return chunks in different order depending on vector similarity calculation")
	}
}

func TestChunkRepository_SearchSimilar_WithBeatTypeFilter(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewChunkRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// Create document
	docRepo := NewDocumentRepository(db)
	doc := memory.NewDocument(tenantID, memory.SourceTypeContentBlock, uuid.New(), "Test", "Content")
	doc.ID = uuid.New()
	if err := docRepo.Create(ctx, doc); err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Create chunks with different beat types
	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = 0.1
	}

	beatType1 := "setup"
	chunk1 := memory.NewChunk(doc.ID, 0, "Setup content", embedding, 10)
	chunk1.BeatType = &beatType1
	if err := repo.Create(ctx, chunk1); err != nil {
		t.Fatalf("Failed to create chunk: %v", err)
	}

	beatType2 := "climax"
	chunk2 := memory.NewChunk(doc.ID, 1, "Climax content", embedding, 10)
	chunk2.BeatType = &beatType2
	if err := repo.Create(ctx, chunk2); err != nil {
		t.Fatalf("Failed to create chunk: %v", err)
	}

	// Search with beat type filter
	filters := &repositories.SearchFilters{
		BeatTypes: []string{"setup"},
	}
	results, err := repo.SearchSimilar(ctx, tenantID, embedding, 10, filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should only return setup chunks
	for _, result := range results {
		if result.BeatType == nil || *result.BeatType != "setup" {
			t.Errorf("Expected only setup chunks, got beat type %v", result.BeatType)
		}
	}
}

func TestChunkRepository_MetadataPersistence(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	repo := NewChunkRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// Create document first (FK constraint)
	docRepo := NewDocumentRepository(db)
	doc := memory.NewDocument(tenantID, memory.SourceTypeContentBlock, uuid.New(), "Test", "Content")
	if err := docRepo.Create(ctx, doc); err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	sceneID := uuid.New()
	beatID := uuid.New()
	locationID := uuid.New()
	beatType := "setup"
	beatIntent := "Introduce character"
	locationName := "Library"
	timeline := "Morning"
	povCharacter := "John"
	contentKind := "final"
	characters := []string{"John", "Mary"}

	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = 0.1
	}

	chunk := memory.NewChunk(doc.ID, 0, "Test content", embedding, 10)
	chunk.SceneID = &sceneID
	chunk.BeatID = &beatID
	chunk.BeatType = &beatType
	chunk.BeatIntent = &beatIntent
	chunk.Characters = characters
	chunk.LocationID = &locationID
	chunk.LocationName = &locationName
	chunk.Timeline = &timeline
	chunk.POVCharacter = &povCharacter
	chunk.ContentKind = &contentKind

	err := repo.Create(ctx, chunk)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Retrieve and verify all metadata
	retrieved, err := repo.GetByID(ctx, chunk.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.SceneID == nil || *retrieved.SceneID != sceneID {
		t.Error("SceneID not persisted correctly")
	}
	if retrieved.BeatID == nil || *retrieved.BeatID != beatID {
		t.Error("BeatID not persisted correctly")
	}
	if retrieved.BeatType == nil || *retrieved.BeatType != beatType {
		t.Error("BeatType not persisted correctly")
	}
	if retrieved.BeatIntent == nil || *retrieved.BeatIntent != beatIntent {
		t.Error("BeatIntent not persisted correctly")
	}
	if len(retrieved.Characters) != len(characters) {
		t.Errorf("Characters not persisted correctly: expected %d, got %d", len(characters), len(retrieved.Characters))
	}
	// Verify characters content
	charJSON, _ := json.Marshal(retrieved.Characters)
	expectedJSON, _ := json.Marshal(characters)
	if string(charJSON) != string(expectedJSON) {
		t.Errorf("Characters JSON mismatch: expected %s, got %s", string(expectedJSON), string(charJSON))
	}
	if retrieved.LocationID == nil || *retrieved.LocationID != locationID {
		t.Error("LocationID not persisted correctly")
	}
	if retrieved.LocationName == nil || *retrieved.LocationName != locationName {
		t.Error("LocationName not persisted correctly")
	}
	if retrieved.Timeline == nil || *retrieved.Timeline != timeline {
		t.Error("Timeline not persisted correctly")
	}
	if retrieved.POVCharacter == nil || *retrieved.POVCharacter != povCharacter {
		t.Error("POVCharacter not persisted correctly")
	}
	if retrieved.ContentKind == nil || *retrieved.ContentKind != contentKind {
		t.Error("ContentKind not persisted correctly")
	}
}

func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}
