package relation

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRelationTypeRequired  = errors.New("relation type is required")
	ErrSourceTargetDifferent = errors.New("source and target must be different")
	ErrSourceIDRequired      = errors.New("source_id is required")
	ErrTargetIDRequired      = errors.New("target_id is required")
)

// EntityRelation represents a relation between two entities
// Both source and target entities must exist - no temporary IDs allowed
type EntityRelation struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	WorldID         uuid.UUID
	SourceType      string
	SourceID        uuid.UUID // Required - entity must exist
	TargetType      string
	TargetID        uuid.UUID // Required - entity must exist
	RelationType    string
	ContextType     *string
	ContextID       *uuid.UUID
	Attributes      map[string]interface{} // Flexible metadata (role, notes, since, etc.)
	Summary         string                 // Human-readable description
	MirrorID        *uuid.UUID             // Points to auto-created inverse relation
	CreatedByUserID *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Inverse relation type mapping
var inverseRelations = map[string]string{
	"parent_of":    "child_of",
	"child_of":     "parent_of",
	"sibling_of":   "sibling_of",
	"spouse_of":    "spouse_of",
	"ally_of":      "ally_of",
	"enemy_of":     "enemy_of",
	"member_of":    "has_member",
	"has_member":   "member_of",
	"leader_of":    "led_by",
	"led_by":       "leader_of",
	"located_in":   "contains",
	"contains":     "located_in",
	"owns":         "owned_by",
	"owned_by":     "owns",
	"mentor_of":    "mentored_by",
	"mentored_by":  "mentor_of",
}

// GetInverseRelationType returns the inverse relation type
func GetInverseRelationType(relationType string) string {
	if inverse, ok := inverseRelations[relationType]; ok {
		return inverse
	}
	return relationType // fallback: symmetric
}

// NewEntityRelation creates a new entity relation
// Both source and target entities must exist (IDs are required)
func NewEntityRelation(
	tenantID, worldID uuid.UUID,
	sourceType string, sourceID uuid.UUID,
	targetType string, targetID uuid.UUID,
	relationType string,
) (*EntityRelation, error) {
	if relationType == "" {
		return nil, ErrRelationTypeRequired
	}
	if sourceID == uuid.Nil {
		return nil, ErrSourceIDRequired
	}
	if targetID == uuid.Nil {
		return nil, ErrTargetIDRequired
	}
	if sourceID == targetID {
		return nil, ErrSourceTargetDifferent
	}

	now := time.Now()
	return &EntityRelation{
		ID:           uuid.New(),
		TenantID:     tenantID,
		WorldID:      worldID,
		SourceType:   sourceType,
		SourceID:     sourceID,
		TargetType:   targetType,
		TargetID:     targetID,
		RelationType: relationType,
		Attributes:   make(map[string]interface{}),
		Summary:      "",
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Validate validates the entity relation
func (r *EntityRelation) Validate() error {
	if r.RelationType == "" {
		return ErrRelationTypeRequired
	}
	if r.SourceID == uuid.Nil {
		return ErrSourceIDRequired
	}
	if r.TargetID == uuid.Nil {
		return ErrTargetIDRequired
	}
	if r.SourceID == r.TargetID {
		return ErrSourceTargetDifferent
	}
	return nil
}

// CreateMirrorRelation creates the inverse relation
func (r *EntityRelation) CreateMirrorRelation() *EntityRelation {
	mirror := &EntityRelation{
		ID:              uuid.New(),
		TenantID:        r.TenantID,
		WorldID:         r.WorldID,
		SourceType:      r.TargetType,
		SourceID:        r.TargetID,
		TargetType:      r.SourceType,
		TargetID:        r.SourceID,
		RelationType:    GetInverseRelationType(r.RelationType),
		ContextType:     r.ContextType,
		ContextID:       r.ContextID,
		Attributes:      copyAttributes(r.Attributes),
		Summary:         r.Summary,
		MirrorID:        &r.ID,
		CreatedByUserID: r.CreatedByUserID,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
	r.MirrorID = &mirror.ID
	return mirror
}

// UpdateSummary updates the summary
func (r *EntityRelation) UpdateSummary(summary string) {
	r.Summary = summary
	r.UpdatedAt = time.Now()
}

// Helper functions
func copyAttributes(attrs map[string]interface{}) map[string]interface{} {
	if attrs == nil {
		return make(map[string]interface{})
	}
	cp := make(map[string]interface{})
	for k, v := range attrs {
		cp[k] = v
	}
	return cp
}

// AttributesJSON returns attributes as JSON bytes
func (r *EntityRelation) AttributesJSON() ([]byte, error) {
	return json.Marshal(r.Attributes)
}
