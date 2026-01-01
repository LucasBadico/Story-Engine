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

