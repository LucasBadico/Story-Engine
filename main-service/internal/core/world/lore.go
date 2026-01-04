package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLoreNameRequired = errors.New("lore name is required")
)

// Lore represents a lore entity with hierarchical structure
type Lore struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	WorldID        uuid.UUID  `json:"world_id"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	Name           string     `json:"name"`
	Category       *string    `json:"category,omitempty"`
	Description    string     `json:"description"`
	Rules          string     `json:"rules"`
	Limitations    string     `json:"limitations"`
	Requirements   string     `json:"requirements"`
	HierarchyLevel int        `json:"hierarchy_level"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// NewLore creates a new lore
func NewLore(tenantID, worldID uuid.UUID, name string, parentID *uuid.UUID) (*Lore, error) {
	if name == "" {
		return nil, ErrLoreNameRequired
	}

	now := time.Now()
	level := 0
	if parentID != nil {
		level = 1 // Will be recalculated based on parent's level
	}

	return &Lore{
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

// Validate validates the lore entity
func (l *Lore) Validate() error {
	if l.Name == "" {
		return ErrLoreNameRequired
	}
	return nil
}

// UpdateName updates the lore name
func (l *Lore) UpdateName(name string) error {
	if name == "" {
		return ErrLoreNameRequired
	}
	l.Name = name
	l.UpdatedAt = time.Now()
	return nil
}

// UpdateCategory updates the lore category
func (l *Lore) UpdateCategory(category *string) {
	l.Category = category
	l.UpdatedAt = time.Now()
}

// UpdateDescription updates the lore description
func (l *Lore) UpdateDescription(description string) {
	l.Description = description
	l.UpdatedAt = time.Now()
}

// UpdateRules updates the lore rules
func (l *Lore) UpdateRules(rules string) {
	l.Rules = rules
	l.UpdatedAt = time.Now()
}

// UpdateLimitations updates the lore limitations
func (l *Lore) UpdateLimitations(limitations string) {
	l.Limitations = limitations
	l.UpdatedAt = time.Now()
}

// UpdateRequirements updates the lore requirements
func (l *Lore) UpdateRequirements(requirements string) {
	l.Requirements = requirements
	l.UpdatedAt = time.Now()
}

// SetParent updates the parent lore and recalculates hierarchy level
func (l *Lore) SetParent(parentID *uuid.UUID, parentLevel int) {
	l.ParentID = parentID
	if parentID != nil {
		l.HierarchyLevel = parentLevel + 1
	} else {
		l.HierarchyLevel = 0
	}
	l.UpdatedAt = time.Now()
}

// SetHierarchyLevel sets the hierarchy level directly
func (l *Lore) SetHierarchyLevel(level int) {
	l.HierarchyLevel = level
	l.UpdatedAt = time.Now()
}

