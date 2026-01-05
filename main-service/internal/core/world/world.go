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
	ID          uuid.UUID   `json:"id"`
	TenantID    uuid.UUID   `json:"tenant_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Genre       string      `json:"genre"`
	IsImplicit  bool        `json:"is_implicit"`
	RPGSystemID *uuid.UUID  `json:"rpg_system_id,omitempty"`
	TimeConfig  *TimeConfig  `json:"time_config,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
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

// SetRPGSystem sets the RPG system for this world
func (w *World) SetRPGSystem(rpgSystemID *uuid.UUID) {
	w.RPGSystemID = rpgSystemID
	w.UpdatedAt = time.Now()
}

// SetTimeConfig sets the time configuration for this world
func (w *World) SetTimeConfig(timeConfig *TimeConfig) {
	w.TimeConfig = timeConfig
	w.UpdatedAt = time.Now()
}

// GetTimeConfig returns the time configuration, or default if nil
func (w *World) GetTimeConfig() *TimeConfig {
	if w.TimeConfig == nil {
		return DefaultTimeConfig()
	}
	return w.TimeConfig
}

