package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTraitNameRequired = errors.New("trait name is required")
)

// Trait represents a trait entity
type Trait struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewTrait creates a new trait
func NewTrait(tenantID uuid.UUID, name string) (*Trait, error) {
	if name == "" {
		return nil, ErrTraitNameRequired
	}

	now := time.Now()
	return &Trait{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the trait entity
func (t *Trait) Validate() error {
	if t.Name == "" {
		return ErrTraitNameRequired
	}
	return nil
}

// UpdateName updates the trait name
func (t *Trait) UpdateName(name string) error {
	if name == "" {
		return ErrTraitNameRequired
	}
	t.Name = name
	t.UpdatedAt = time.Now()
	return nil
}

// UpdateCategory updates the trait category
func (t *Trait) UpdateCategory(category string) {
	t.Category = category
	t.UpdatedAt = time.Now()
}

// UpdateDescription updates the trait description
func (t *Trait) UpdateDescription(description string) {
	t.Description = description
	t.UpdatedAt = time.Now()
}

