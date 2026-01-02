package world

import (
	"time"

	"github.com/google/uuid"
)

// ArchetypeTrait represents the junction between archetype and trait
type ArchetypeTrait struct {
	ID           uuid.UUID `json:"id"`
	ArchetypeID  uuid.UUID `json:"archetype_id"`
	TraitID      uuid.UUID `json:"trait_id"`
	DefaultValue string    `json:"default_value"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewArchetypeTrait creates a new archetype-trait relationship
func NewArchetypeTrait(archetypeID, traitID uuid.UUID, defaultValue string) *ArchetypeTrait {
	return &ArchetypeTrait{
		ID:           uuid.New(),
		ArchetypeID:  archetypeID,
		TraitID:      traitID,
		DefaultValue: defaultValue,
		CreatedAt:    time.Now(),
	}
}


