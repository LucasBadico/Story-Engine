package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrFactionNameRequired = errors.New("faction name is required")
)

// Faction represents a faction entity with hierarchical structure
type Faction struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	WorldID        uuid.UUID  `json:"world_id"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	Name           string     `json:"name"`
	Type           *string    `json:"type,omitempty"`
	Description    string     `json:"description"`
	Beliefs        string     `json:"beliefs"`
	Structure      string     `json:"structure"`
	Symbols        string     `json:"symbols"`
	HierarchyLevel int        `json:"hierarchy_level"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// NewFaction creates a new faction
func NewFaction(tenantID, worldID uuid.UUID, name string, parentID *uuid.UUID) (*Faction, error) {
	if name == "" {
		return nil, ErrFactionNameRequired
	}

	now := time.Now()
	level := 0
	if parentID != nil {
		level = 1 // Will be recalculated based on parent's level
	}

	return &Faction{
		ID:             uuid.New(),
		TenantID:       tenantID,
		WorldID:        worldID,
		ParentID:       parentID,
		Name:           name,
		HierarchyLevel: level,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Validate validates the faction entity
func (f *Faction) Validate() error {
	if f.Name == "" {
		return ErrFactionNameRequired
	}
	return nil
}

// UpdateName updates the faction name
func (f *Faction) UpdateName(name string) error {
	if name == "" {
		return ErrFactionNameRequired
	}
	f.Name = name
	f.UpdatedAt = time.Now()
	return nil
}

// UpdateType updates the faction type
func (f *Faction) UpdateType(factionType *string) {
	f.Type = factionType
	f.UpdatedAt = time.Now()
}

// UpdateDescription updates the faction description
func (f *Faction) UpdateDescription(description string) {
	f.Description = description
	f.UpdatedAt = time.Now()
}

// UpdateBeliefs updates the faction beliefs
func (f *Faction) UpdateBeliefs(beliefs string) {
	f.Beliefs = beliefs
	f.UpdatedAt = time.Now()
}

// UpdateStructure updates the faction structure
func (f *Faction) UpdateStructure(structure string) {
	f.Structure = structure
	f.UpdatedAt = time.Now()
}

// UpdateSymbols updates the faction symbols
func (f *Faction) UpdateSymbols(symbols string) {
	f.Symbols = symbols
	f.UpdatedAt = time.Now()
}

// SetParent updates the parent faction and recalculates hierarchy level
func (f *Faction) SetParent(parentID *uuid.UUID, parentLevel int) {
	f.ParentID = parentID
	if parentID != nil {
		f.HierarchyLevel = parentLevel + 1
	} else {
		f.HierarchyLevel = 0
	}
	f.UpdatedAt = time.Now()
}

// SetHierarchyLevel sets the hierarchy level directly
func (f *Faction) SetHierarchyLevel(level int) {
	f.HierarchyLevel = level
	f.UpdatedAt = time.Now()
}

