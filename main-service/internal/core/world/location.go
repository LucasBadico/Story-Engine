package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLocationNameRequired = errors.New("location name is required")
)

// Location represents a location entity with hierarchical structure
type Location struct {
	ID             uuid.UUID  `json:"id"`
	WorldID        uuid.UUID  `json:"world_id"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	Name           string     `json:"name"`
	Type           string     `json:"type"`
	Description    string     `json:"description"`
	HierarchyLevel int        `json:"hierarchy_level"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// NewLocation creates a new location
func NewLocation(worldID uuid.UUID, name string, parentID *uuid.UUID) (*Location, error) {
	if name == "" {
		return nil, ErrLocationNameRequired
	}

	now := time.Now()
	level := 0
	if parentID != nil {
		level = 1 // Will be recalculated based on parent's level
	}

	return &Location{
		ID:             uuid.New(),
		WorldID:        worldID,
		ParentID:       parentID,
		Name:           name,
		HierarchyLevel: level,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Validate validates the location entity
func (l *Location) Validate() error {
	if l.Name == "" {
		return ErrLocationNameRequired
	}
	return nil
}

// UpdateName updates the location name
func (l *Location) UpdateName(name string) error {
	if name == "" {
		return ErrLocationNameRequired
	}
	l.Name = name
	l.UpdatedAt = time.Now()
	return nil
}

// UpdateType updates the location type
func (l *Location) UpdateType(locationType string) {
	l.Type = locationType
	l.UpdatedAt = time.Now()
}

// UpdateDescription updates the location description
func (l *Location) UpdateDescription(description string) {
	l.Description = description
	l.UpdatedAt = time.Now()
}

// SetParent updates the parent location and recalculates hierarchy level
func (l *Location) SetParent(parentID *uuid.UUID, parentLevel int) {
	l.ParentID = parentID
	if parentID != nil {
		l.HierarchyLevel = parentLevel + 1
	} else {
		l.HierarchyLevel = 0
	}
	l.UpdatedAt = time.Now()
}

// SetHierarchyLevel sets the hierarchy level directly
func (l *Location) SetHierarchyLevel(level int) {
	l.HierarchyLevel = level
	l.UpdatedAt = time.Now()
}

