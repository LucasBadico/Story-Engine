package world

import (
	"time"

	"github.com/google/uuid"
)

// FactionReference represents the relationship between a faction and another entity
type FactionReference struct {
	ID         uuid.UUID  `json:"id"`
	FactionID  uuid.UUID  `json:"faction_id"`
	EntityType string     `json:"entity_type"`
	EntityID   uuid.UUID  `json:"entity_id"`
	Role       *string    `json:"role,omitempty"`
	Notes      string     `json:"notes"`
	CreatedAt  time.Time  `json:"created_at"`
}

// NewFactionReference creates a new faction-reference relationship
func NewFactionReference(factionID uuid.UUID, entityType string, entityID uuid.UUID, role *string) *FactionReference {
	return &FactionReference{
		ID:         uuid.New(),
		FactionID:  factionID,
		EntityType: entityType,
		EntityID:   entityID,
		Role:       role,
		CreatedAt:  time.Now(),
	}
}

// UpdateRole updates the role
func (fr *FactionReference) UpdateRole(role *string) {
	fr.Role = role
}

// UpdateNotes updates the notes
func (fr *FactionReference) UpdateNotes(notes string) {
	fr.Notes = notes
}

