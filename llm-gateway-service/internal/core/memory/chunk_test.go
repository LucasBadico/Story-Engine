package memory

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewChunk(t *testing.T) {
	documentID := uuid.New()
	content := "Test content"
	embedding := []float32{0.1, 0.2, 0.3}
	tokenCount := 10

	chunk := NewChunk(documentID, 0, content, embedding, tokenCount)

	if chunk.ID == uuid.Nil {
		t.Error("Expected chunk ID to be set")
	}
	if chunk.DocumentID != documentID {
		t.Errorf("Expected DocumentID %v, got %v", documentID, chunk.DocumentID)
	}
	if chunk.ChunkIndex != 0 {
		t.Errorf("Expected ChunkIndex 0, got %d", chunk.ChunkIndex)
	}
	if chunk.Content != content {
		t.Errorf("Expected Content %q, got %q", content, chunk.Content)
	}
	if len(chunk.Embedding) != len(embedding) {
		t.Errorf("Expected embedding length %d, got %d", len(embedding), len(chunk.Embedding))
	}
	if chunk.TokenCount != tokenCount {
		t.Errorf("Expected TokenCount %d, got %d", tokenCount, chunk.TokenCount)
	}
}

func TestChunk_Validate(t *testing.T) {
	documentID := uuid.New()
	chunk := NewChunk(documentID, 0, "Test content", []float32{0.1}, 10)

	if err := chunk.Validate(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestChunk_Validate_InvalidDocumentID(t *testing.T) {
	chunk := NewChunk(uuid.Nil, 0, "Test content", []float32{0.1}, 10)

	if err := chunk.Validate(); err != ErrDocumentIDRequired {
		t.Errorf("Expected ErrDocumentIDRequired, got %v", err)
	}
}

func TestChunk_Validate_InvalidChunkIndex(t *testing.T) {
	documentID := uuid.New()
	chunk := NewChunk(documentID, -1, "Test content", []float32{0.1}, 10)

	if err := chunk.Validate(); err != ErrInvalidChunkIndex {
		t.Errorf("Expected ErrInvalidChunkIndex, got %v", err)
	}
}

func TestChunk_Validate_EmptyContent(t *testing.T) {
	documentID := uuid.New()
	chunk := NewChunk(documentID, 0, "", []float32{0.1}, 10)

	if err := chunk.Validate(); err != ErrContentRequired {
		t.Errorf("Expected ErrContentRequired, got %v", err)
	}
}

func TestChunk_WithMetadata(t *testing.T) {
	documentID := uuid.New()
	sceneID := uuid.New()
	beatID := uuid.New()
	beatType := "setup"
	beatIntent := "Introduce character"
	characters := []string{"John", "Mary"}
	locationID := uuid.New()
	locationName := "Library"
	timeline := "Morning"
	povCharacter := "John"
	proseKind := "final"

	chunk := NewChunk(documentID, 0, "Test content", []float32{0.1}, 10)
	chunk.SceneID = &sceneID
	chunk.BeatID = &beatID
	chunk.BeatType = &beatType
	chunk.BeatIntent = &beatIntent
	chunk.Characters = characters
	chunk.LocationID = &locationID
	chunk.LocationName = &locationName
	chunk.Timeline = &timeline
	chunk.POVCharacter = &povCharacter
	chunk.ProseKind = &proseKind

	if chunk.SceneID == nil || *chunk.SceneID != sceneID {
		t.Error("SceneID not set correctly")
	}
	if chunk.BeatID == nil || *chunk.BeatID != beatID {
		t.Error("BeatID not set correctly")
	}
	if chunk.BeatType == nil || *chunk.BeatType != beatType {
		t.Error("BeatType not set correctly")
	}
	if chunk.BeatIntent == nil || *chunk.BeatIntent != beatIntent {
		t.Error("BeatIntent not set correctly")
	}
	if len(chunk.Characters) != len(characters) {
		t.Error("Characters not set correctly")
	}
	if chunk.LocationID == nil || *chunk.LocationID != locationID {
		t.Error("LocationID not set correctly")
	}
	if chunk.LocationName == nil || *chunk.LocationName != locationName {
		t.Error("LocationName not set correctly")
	}
	if chunk.Timeline == nil || *chunk.Timeline != timeline {
		t.Error("Timeline not set correctly")
	}
	if chunk.POVCharacter == nil || *chunk.POVCharacter != povCharacter {
		t.Error("POVCharacter not set correctly")
	}
	if chunk.ProseKind == nil || *chunk.ProseKind != proseKind {
		t.Error("ProseKind not set correctly")
	}
}

