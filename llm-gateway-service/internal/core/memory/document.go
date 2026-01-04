package memory

import (
	"time"

	"github.com/google/uuid"
)

// SourceType represents the type of source entity
type SourceType string

const (
	SourceTypeStory        SourceType = "story"
	SourceTypeChapter      SourceType = "chapter"
	SourceTypeScene        SourceType = "scene"
	SourceTypeBeat         SourceType = "beat"
	SourceTypeContentBlock SourceType = "content_block"
	SourceTypeWorld        SourceType = "world"
	SourceTypeCharacter    SourceType = "character"
	SourceTypeLocation     SourceType = "location"
	SourceTypeEvent        SourceType = "event"
	SourceTypeArtifact     SourceType = "artifact"
	SourceTypeFaction      SourceType = "faction"
	SourceTypeLore         SourceType = "lore"
)

// Document represents an embedding document
type Document struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	SourceType SourceType
	SourceID   uuid.UUID
	Title      string
	Content    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewDocument creates a new document
func NewDocument(tenantID uuid.UUID, sourceType SourceType, sourceID uuid.UUID, title, content string) *Document {
	now := time.Now()
	return &Document{
		ID:         uuid.New(),
		TenantID:   tenantID,
		SourceType: sourceType,
		SourceID:   sourceID,
		Title:      title,
		Content:    content,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Validate validates the document
func (d *Document) Validate() error {
	if d.TenantID == uuid.Nil {
		return ErrTenantIDRequired
	}
	if d.SourceID == uuid.Nil {
		return ErrSourceIDRequired
	}
	if d.Content == "" {
		return ErrContentRequired
	}
	return nil
}

