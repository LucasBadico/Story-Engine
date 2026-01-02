package rpg

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrClassNameRequired = errors.New("class name is required")
	ErrClassSystemRequired = errors.New("RPG system ID is required")
	ErrInvalidClassJSON = errors.New("invalid class JSON")
	ErrInvalidTier = errors.New("tier must be between 1 and 3")
)

// RPGClass represents an RPG class/job definition
type RPGClass struct {
	ID            uuid.UUID       `json:"id"`
	RPGSystemID   uuid.UUID       `json:"rpg_system_id"`
	ParentClassID *uuid.UUID      `json:"parent_class_id,omitempty"`
	Name          string          `json:"name"`
	Tier          int             `json:"tier"`
	Description   *string         `json:"description,omitempty"`
	Requirements  *json.RawMessage `json:"requirements,omitempty"` // JSONB
	StatBonuses   *json.RawMessage `json:"stat_bonuses,omitempty"` // JSONB
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// NewRPGClass creates a new RPG class
func NewRPGClass(rpgSystemID uuid.UUID, name string) (*RPGClass, error) {
	if name == "" {
		return nil, ErrClassNameRequired
	}
	if rpgSystemID == uuid.Nil {
		return nil, ErrClassSystemRequired
	}

	now := time.Now()
	return &RPGClass{
		ID:          uuid.New(),
		RPGSystemID: rpgSystemID,
		Name:        name,
		Tier:        1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Validate validates the RPG class entity
func (c *RPGClass) Validate() error {
	if c.Name == "" {
		return ErrClassNameRequired
	}
	if c.RPGSystemID == uuid.Nil {
		return ErrClassSystemRequired
	}
	if c.Tier < 1 || c.Tier > 3 {
		return ErrInvalidTier
	}

	// Validate JSON fields
	if c.Requirements != nil && len(*c.Requirements) > 0 {
		var req map[string]interface{}
		if err := json.Unmarshal(*c.Requirements, &req); err != nil {
			return ErrInvalidClassJSON
		}
	}

	if c.StatBonuses != nil && len(*c.StatBonuses) > 0 {
		var bonuses map[string]interface{}
		if err := json.Unmarshal(*c.StatBonuses, &bonuses); err != nil {
			return ErrInvalidClassJSON
		}
	}

	return nil
}

// UpdateName updates the class name
func (c *RPGClass) UpdateName(name string) error {
	if name == "" {
		return ErrClassNameRequired
	}
	c.Name = name
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateTier updates the class tier
func (c *RPGClass) UpdateTier(tier int) error {
	if tier < 1 || tier > 3 {
		return ErrInvalidTier
	}
	c.Tier = tier
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the description
func (c *RPGClass) UpdateDescription(description *string) {
	c.Description = description
	c.UpdatedAt = time.Now()
}

// SetParentClass sets the parent class (evolution)
func (c *RPGClass) SetParentClass(parentClassID *uuid.UUID) {
	c.ParentClassID = parentClassID
	c.UpdatedAt = time.Now()
}

// UpdateRequirements updates the requirements
func (c *RPGClass) UpdateRequirements(requirements *json.RawMessage) error {
	if requirements != nil && len(*requirements) > 0 {
		var req map[string]interface{}
		if err := json.Unmarshal(*requirements, &req); err != nil {
			return ErrInvalidClassJSON
		}
	}
	c.Requirements = requirements
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateStatBonuses updates the stat bonuses
func (c *RPGClass) UpdateStatBonuses(statBonuses *json.RawMessage) error {
	if statBonuses != nil && len(*statBonuses) > 0 {
		var bonuses map[string]interface{}
		if err := json.Unmarshal(*statBonuses, &bonuses); err != nil {
			return ErrInvalidClassJSON
		}
	}
	c.StatBonuses = statBonuses
	c.UpdatedAt = time.Now()
	return nil
}


