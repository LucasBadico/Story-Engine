package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrArchetypeNameRequired = errors.New("archetype name is required")
)

// Archetype represents an archetype entity
type Archetype struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewArchetype creates a new archetype
func NewArchetype(tenantID uuid.UUID, name string) (*Archetype, error) {
	if name == "" {
		return nil, ErrArchetypeNameRequired
	}

	now := time.Now()
	return &Archetype{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the archetype entity
func (a *Archetype) Validate() error {
	if a.Name == "" {
		return ErrArchetypeNameRequired
	}
	return nil
}

// UpdateName updates the archetype name
func (a *Archetype) UpdateName(name string) error {
	if name == "" {
		return ErrArchetypeNameRequired
	}
	a.Name = name
	a.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the archetype description
func (a *Archetype) UpdateDescription(description string) {
	a.Description = description
	a.UpdatedAt = time.Now()
}

