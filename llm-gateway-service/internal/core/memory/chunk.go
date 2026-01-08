package memory

import (
	"time"

	"github.com/google/uuid"
)

// Chunk represents an embedding chunk
type Chunk struct {
	ID         uuid.UUID
	DocumentID uuid.UUID
	ChunkIndex int
	Content    string
	Embedding  []float32 // Vector embedding
	TokenCount int
	CreatedAt  time.Time

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

	// Content Block metadata
	ContentType *string `json:"content_type,omitempty"` // text, image, video, etc.
	ContentKind *string `json:"content_kind,omitempty"` // final, alt_a, draft, etc.

	// Summary metadata
	ChunkType *string `json:"chunk_type,omitempty"` // raw, summary, relations
	EmbedText *string `json:"embed_text,omitempty"`

	// World entity metadata
	WorldID       *uuid.UUID `json:"world_id,omitempty"`
	WorldName     *string    `json:"world_name,omitempty"`
	WorldGenre    *string    `json:"world_genre,omitempty"`
	EntityName    *string    `json:"entity_name,omitempty"` // nome da entidade (character, location, etc)
	EventTimeline *string    `json:"event_timeline,omitempty"`
	Importance    *int       `json:"importance,omitempty"` // para events

	// Relacionamentos World
	RelatedCharacters []string `json:"related_characters,omitempty"`
	RelatedLocations  []string `json:"related_locations,omitempty"`
	RelatedArtifacts  []string `json:"related_artifacts,omitempty"`
	RelatedEvents     []string `json:"related_events,omitempty"`
}

// NewChunk creates a new chunk
func NewChunk(documentID uuid.UUID, chunkIndex int, content string, embedding []float32, tokenCount int) *Chunk {
	chunkType := "raw"
	return &Chunk{
		ID:         uuid.New(),
		DocumentID: documentID,
		ChunkIndex: chunkIndex,
		Content:    content,
		Embedding:  embedding,
		TokenCount: tokenCount,
		CreatedAt:  time.Now(),
		ChunkType:  &chunkType,
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
