package audit

import (
	"time"

	"github.com/google/uuid"
)

// Action represents the type of action performed
type Action string

const (
	ActionCreate   Action = "create"
	ActionUpdate   Action = "update"
	ActionDelete   Action = "delete"
	ActionClone    Action = "clone"
	ActionPromote  Action = "promote"
	ActionActivate Action = "activate"
	ActionSuspend  Action = "suspend"
)

// EntityType represents the type of entity
type EntityType string

const (
	EntityTypeTenant     EntityType = "tenant"
	EntityTypeUser       EntityType = "user"
	EntityTypeMembership EntityType = "membership"
	EntityTypeStory      EntityType = "story"
	EntityTypeChapter    EntityType = "chapter"
	EntityTypeScene      EntityType = "scene"
	EntityTypeBeat       EntityType = "beat"
	EntityTypeProseBlock EntityType = "prose_block"
	EntityTypeWorld      EntityType = "world"
	EntityTypeTrait      EntityType = "trait"
	EntityTypeArchetype  EntityType = "archetype"
	EntityTypeLocation   EntityType = "location"
	EntityTypeCharacter  EntityType = "character"
	EntityTypeArtifact   EntityType = "artifact"
	EntityTypeEvent      EntityType = "event"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	ActorUserID *uuid.UUID // nullable (can be API key action)
	Action      Action
	EntityType  EntityType
	EntityID    uuid.UUID
	Metadata    map[string]interface{} // JSONB in DB
	CreatedAt   time.Time
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(tenantID uuid.UUID, actorUserID *uuid.UUID, action Action, entityType EntityType, entityID uuid.UUID, metadata map[string]interface{}) *AuditLog {
	return &AuditLog{
		ID:          uuid.New(),
		TenantID:    tenantID,
		ActorUserID: actorUserID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
	}
}

// Validate validates the audit log entity
func (a *AuditLog) Validate() error {
	if a.Action == "" {
		return ErrActionRequired
	}
	if a.EntityType == "" {
		return ErrEntityTypeRequired
	}
	return nil
}

