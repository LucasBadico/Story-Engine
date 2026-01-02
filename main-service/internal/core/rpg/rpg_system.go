package rpg

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRPGSystemNameRequired     = errors.New("RPG system name is required")
	ErrRPGSystemSchemaRequired   = errors.New("base_stats_schema is required")
	ErrInvalidRPGSystemSchema    = errors.New("invalid RPG system schema")
)

// RPGSystem represents an RPG system definition
type RPGSystem struct {
	ID                uuid.UUID       `json:"id"`
	TenantID          *uuid.UUID      `json:"tenant_id,omitempty"` // null = builtin
	Name              string          `json:"name"`
	Description       *string         `json:"description,omitempty"`
	BaseStatsSchema   json.RawMessage `json:"base_stats_schema"`   // JSONB
	DerivedStatsSchema *json.RawMessage `json:"derived_stats_schema,omitempty"` // JSONB
	ProgressionSchema *json.RawMessage `json:"progression_schema,omitempty"` // JSONB
	IsBuiltin         bool            `json:"is_builtin"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// NewRPGSystem creates a new RPG system
func NewRPGSystem(tenantID *uuid.UUID, name string, baseStatsSchema json.RawMessage) (*RPGSystem, error) {
	if name == "" {
		return nil, ErrRPGSystemNameRequired
	}
	if len(baseStatsSchema) == 0 {
		return nil, ErrRPGSystemSchemaRequired
	}

	// Validate JSON
	var schema map[string]interface{}
	if err := json.Unmarshal(baseStatsSchema, &schema); err != nil {
		return nil, ErrInvalidRPGSystemSchema
	}

	now := time.Now()
	return &RPGSystem{
		ID:              uuid.New(),
		TenantID:        tenantID,
		Name:            name,
		BaseStatsSchema: baseStatsSchema,
		IsBuiltin:       tenantID == nil,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Validate validates the RPG system entity
func (r *RPGSystem) Validate() error {
	if r.Name == "" {
		return ErrRPGSystemNameRequired
	}
	if len(r.BaseStatsSchema) == 0 {
		return ErrRPGSystemSchemaRequired
	}

	// Validate JSON schemas
	var schema map[string]interface{}
	if err := json.Unmarshal(r.BaseStatsSchema, &schema); err != nil {
		return ErrInvalidRPGSystemSchema
	}

	if r.DerivedStatsSchema != nil && len(*r.DerivedStatsSchema) > 0 {
		var derived map[string]interface{}
		if err := json.Unmarshal(*r.DerivedStatsSchema, &derived); err != nil {
			return ErrInvalidRPGSystemSchema
		}
	}

	if r.ProgressionSchema != nil && len(*r.ProgressionSchema) > 0 {
		var progression map[string]interface{}
		if err := json.Unmarshal(*r.ProgressionSchema, &progression); err != nil {
			return ErrInvalidRPGSystemSchema
		}
	}

	return nil
}

// UpdateName updates the system name
func (r *RPGSystem) UpdateName(name string) error {
	if name == "" {
		return ErrRPGSystemNameRequired
	}
	r.Name = name
	r.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the description
func (r *RPGSystem) UpdateDescription(description *string) {
	r.Description = description
	r.UpdatedAt = time.Now()
}

// UpdateSchemas updates the schemas
func (r *RPGSystem) UpdateSchemas(baseStatsSchema json.RawMessage, derivedStatsSchema, progressionSchema *json.RawMessage) error {
	if len(baseStatsSchema) == 0 {
		return ErrRPGSystemSchemaRequired
	}

	// Validate JSON
	var schema map[string]interface{}
	if err := json.Unmarshal(baseStatsSchema, &schema); err != nil {
		return ErrInvalidRPGSystemSchema
	}

	r.BaseStatsSchema = baseStatsSchema
	r.DerivedStatsSchema = derivedStatsSchema
	r.ProgressionSchema = progressionSchema
	r.UpdatedAt = time.Now()
	return nil
}


