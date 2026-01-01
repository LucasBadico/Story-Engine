package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCharacterNameRequired = errors.New("character name is required")
)

// Character represents a character entity
type Character struct {
	ID          uuid.UUID  `json:"id"`
	WorldID     uuid.UUID  `json:"world_id"`
	ArchetypeID *uuid.UUID `json:"archetype_id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewCharacter creates a new character
func NewCharacter(worldID uuid.UUID, name string) (*Character, error) {
	if name == "" {
		return nil, ErrCharacterNameRequired
	}

	now := time.Now()
	return &Character{
		ID:        uuid.New(),
		WorldID:   worldID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the character entity
func (c *Character) Validate() error {
	if c.Name == "" {
		return ErrCharacterNameRequired
	}
	return nil
}

// UpdateName updates the character name
func (c *Character) UpdateName(name string) error {
	if name == "" {
		return ErrCharacterNameRequired
	}
	c.Name = name
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the character description
func (c *Character) UpdateDescription(description string) {
	c.Description = description
	c.UpdatedAt = time.Now()
}

// SetArchetype sets the character's archetype
func (c *Character) SetArchetype(archetypeID *uuid.UUID) {
	c.ArchetypeID = archetypeID
	c.UpdatedAt = time.Now()
}

