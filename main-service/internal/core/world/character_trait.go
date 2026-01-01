package world

import (
	"time"

	"github.com/google/uuid"
)

// CharacterTrait represents a trait assigned to a character (as a copy/snapshot)
type CharacterTrait struct {
	ID              uuid.UUID `json:"id"`
	CharacterID     uuid.UUID `json:"character_id"`
	TraitID         uuid.UUID `json:"trait_id"`
	// Copied trait data (snapshot)
	TraitName        string `json:"trait_name"`
	TraitCategory    string `json:"trait_category"`
	TraitDescription string `json:"trait_description"`
	// Character-specific customization
	Value string `json:"value"`
	Notes string `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewCharacterTrait creates a new character-trait relationship with copied trait data
func NewCharacterTrait(characterID, traitID uuid.UUID, trait *Trait, defaultValue string) *CharacterTrait {
	now := time.Now()
	return &CharacterTrait{
		ID:               uuid.New(),
		CharacterID:      characterID,
		TraitID:          traitID,
		TraitName:        trait.Name,
		TraitCategory:    trait.Category,
		TraitDescription: trait.Description,
		Value:            defaultValue,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// UpdateValue updates the trait value for this character
func (ct *CharacterTrait) UpdateValue(value string) {
	ct.Value = value
	ct.UpdatedAt = time.Now()
}

// UpdateNotes updates the trait notes for this character
func (ct *CharacterTrait) UpdateNotes(notes string) {
	ct.Notes = notes
	ct.UpdatedAt = time.Now()
}

// UpdateTraitSnapshot updates the copied trait data (if trait template changed)
func (ct *CharacterTrait) UpdateTraitSnapshot(trait *Trait) {
	ct.TraitName = trait.Name
	ct.TraitCategory = trait.Category
	ct.TraitDescription = trait.Description
	ct.UpdatedAt = time.Now()
}

