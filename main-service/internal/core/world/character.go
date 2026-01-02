package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCharacterNameRequired = errors.New("character name is required")
	ErrInvalidClassLevel = errors.New("class_level must be at least 1")
)

// Character represents a character entity
type Character struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	WorldID       uuid.UUID  `json:"world_id"`
	ArchetypeID   *uuid.UUID `json:"archetype_id,omitempty"`
	CurrentClassID *uuid.UUID `json:"current_class_id,omitempty"`
	ClassLevel    int        `json:"class_level"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// NewCharacter creates a new character
func NewCharacter(tenantID, worldID uuid.UUID, name string) (*Character, error) {
	if name == "" {
		return nil, ErrCharacterNameRequired
	}

	now := time.Now()
	return &Character{
		ID:         uuid.New(),
		TenantID:   tenantID,
		WorldID:    worldID,
		Name:       name,
		ClassLevel: 1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// Validate validates the character entity
func (c *Character) Validate() error {
	if c.Name == "" {
		return ErrCharacterNameRequired
	}
	if c.ClassLevel < 1 {
		return ErrInvalidClassLevel
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

// SetArchetype sets the archetype for this character
func (c *Character) SetArchetype(archetypeID *uuid.UUID) {
	c.ArchetypeID = archetypeID
	c.UpdatedAt = time.Now()
}

// SetClass sets the current class for this character
func (c *Character) SetClass(classID *uuid.UUID) {
	c.CurrentClassID = classID
	c.UpdatedAt = time.Now()
}

// SetClassLevel sets the class level
func (c *Character) SetClassLevel(level int) error {
	if level < 1 {
		return ErrInvalidClassLevel
	}
	c.ClassLevel = level
	c.UpdatedAt = time.Now()
	return nil
}
