package memory

import (
	"time"

	"github.com/google/uuid"
)

// Chunk represents an embedding chunk
type Chunk struct {
	ID          uuid.UUID
	DocumentID  uuid.UUID
	ChunkIndex  int
	Content     string
	Embedding   []float32 // Vector embedding
	TokenCount  int
	CreatedAt   time.Time

	// Metadados estruturais
	SceneID    *uuid.UUID `json:"scene_id,omitempty"`
	BeatID     *uuid.UUID `json:"beat_id,omitempty"`
	BeatType   *string    `json:"beat_type,omitempty"`
	BeatIntent *string    `json:"beat_intent,omitempty"`

	// Contexto narrativo
	Characters   []string   `json:"characters,omitempty"`
	LocationID   *uuid.UUID `json:"location_id,omitempty"`
	LocationName *string    `json:"location_name,omitempty"`
	Timeline     *string    `json:"timeline,omitempty"`
	POVCharacter *string    `json:"pov_character,omitempty"`

	// Tipo do bloco
	ProseKind *string `json:"prose_kind,omitempty"`
}

// NewChunk creates a new chunk
func NewChunk(documentID uuid.UUID, chunkIndex int, content string, embedding []float32, tokenCount int) *Chunk {
	return &Chunk{
		ID:         uuid.New(),
		DocumentID: documentID,
		ChunkIndex: chunkIndex,
		Content:    content,
		Embedding:  embedding,
		TokenCount: tokenCount,
		CreatedAt:  time.Now(),
	}
}

// Validate validates the chunk
func (c *Chunk) Validate() error {
	if c.DocumentID == uuid.Nil {
		return ErrDocumentIDRequired
	}
	if c.ChunkIndex < 0 {
		return ErrInvalidChunkIndex
	}
	if c.Content == "" {
		return ErrContentRequired
	}
	return nil
}

