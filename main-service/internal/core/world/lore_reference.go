package world

import (
	"time"

	"github.com/google/uuid"
)

// LoreReference represents the relationship between a lore and another entity
type LoreReference struct {
	ID               uuid.UUID  `json:"id"`
	LoreID           uuid.UUID  `json:"lore_id"`
	EntityType       string     `json:"entity_type"`
	EntityID         uuid.UUID  `json:"entity_id"`
	RelationshipType *string    `json:"relationship_type,omitempty"`
	Notes            string     `json:"notes"`
	CreatedAt        time.Time  `json:"created_at"`
}

// NewLoreReference creates a new lore-reference relationship
func NewLoreReference(loreID uuid.UUID, entityType string, entityID uuid.UUID, relationshipType *string) *LoreReference {
	return &LoreReference{
		ID:               uuid.New(),
		LoreID:           loreID,
		EntityType:       entityType,
		EntityID:         entityID,
		RelationshipType: relationshipType,
		CreatedAt:        time.Now(),
	}
}

// UpdateRelationshipType updates the relationship type
func (lr *LoreReference) UpdateRelationshipType(relationshipType *string) {
	lr.RelationshipType = relationshipType
}

// UpdateNotes updates the notes
func (lr *LoreReference) UpdateNotes(notes string) {
	lr.Notes = notes
}

