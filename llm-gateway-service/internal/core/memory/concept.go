package memory

import (
	"time"

	"github.com/google/uuid"
)

// ConceptType represents the type of concept entity
type ConceptType string

const (
	ConceptTypeCharacter ConceptType = "character"
	ConceptTypeLocation  ConceptType = "location"
	ConceptTypeArtefact  ConceptType = "artefact"
	ConceptTypeFaction   ConceptType = "faction"
	ConceptTypeEvent     ConceptType = "event"
)

// Concept represents an entity identity in the story world
type Concept struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	StoryID   uuid.UUID
	Name      string
	Type      ConceptType
	Metadata  map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewConcept creates a new concept
func NewConcept(tenantID, storyID uuid.UUID, name string, conceptType ConceptType) *Concept {
	now := time.Now()
	return &Concept{
		ID:        uuid.New(),
		TenantID:  tenantID,
		StoryID:   storyID,
		Name:      name,
		Type:      conceptType,
		Metadata:  map[string]interface{}{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates the concept
func (c *Concept) Validate() error {
	if c.TenantID == uuid.Nil {
		return ErrTenantIDRequired
	}
	if c.StoryID == uuid.Nil {
		return ErrStoryIDRequired
	}
	if c.Name == "" {
		return ErrNameRequired
	}
	return nil
}
