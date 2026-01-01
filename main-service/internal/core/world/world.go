package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired = errors.New("world name is required")
)

// World represents a world entity
type World struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Genre       string    `json:"genre"`
	IsImplicit  bool      `json:"is_implicit"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewWorld creates a new world
func NewWorld(tenantID uuid.UUID, name string, isImplicit bool) (*World, error) {
	if name == "" {
		return nil, ErrNameRequired
	}

	now := time.Now()
	return &World{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		IsImplicit:  isImplicit,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Validate validates the world entity
func (w *World) Validate() error {
	if w.Name == "" {
		return ErrNameRequired
	}
	return nil
}

// UpdateName updates the world name
func (w *World) UpdateName(name string) error {
	if name == "" {
		return ErrNameRequired
	}
	w.Name = name
	w.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the world description
func (w *World) UpdateDescription(description string) {
	w.Description = description
	w.UpdatedAt = time.Now()
}

// UpdateGenre updates the world genre
func (w *World) UpdateGenre(genre string) {
	w.Genre = genre
	w.UpdatedAt = time.Now()
}

// SetImplicit updates the implicit flag
func (w *World) SetImplicit(isImplicit bool) {
	w.IsImplicit = isImplicit
	w.UpdatedAt = time.Now()
}

