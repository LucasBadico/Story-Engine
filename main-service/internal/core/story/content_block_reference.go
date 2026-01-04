package story

import (
	"time"

	"github.com/google/uuid"
)

// EntityType represents the type of entity that references a content block
type EntityType string

const (
	EntityTypeScene           EntityType = "scene"
	EntityTypeBeat            EntityType = "beat"
	EntityTypeChapter         EntityType = "chapter"
	EntityTypeCharacter       EntityType = "character"
	EntityTypeLocation        EntityType = "location"
	EntityTypeArtifact        EntityType = "artifact"
	EntityTypeEvent           EntityType = "event"
	EntityTypeWorld           EntityType = "world"
	EntityTypeRPGSystem       EntityType = "rpg_system"
	EntityTypeRPGSkill        EntityType = "rpg_skill"
	EntityTypeRPGClass        EntityType = "rpg_class"
	EntityTypeInventoryItem   EntityType = "inventory_item"
	EntityTypeFaction         EntityType = "faction"
	EntityTypeLore            EntityType = "lore"
	EntityTypeFactionReference EntityType = "faction_reference"
	EntityTypeLoreReference   EntityType = "lore_reference"
)

// ContentBlockReference represents a reference from an entity to a content block
type ContentBlockReference struct {
	ID            uuid.UUID  `json:"id"`
	ContentBlockID uuid.UUID `json:"content_block_id"`
	EntityType    EntityType `json:"entity_type"`
	EntityID      uuid.UUID  `json:"entity_id"`
	CreatedAt     time.Time  `json:"created_at"`
}

// NewContentBlockReference creates a new content block reference
func NewContentBlockReference(contentBlockID uuid.UUID, entityType EntityType, entityID uuid.UUID) (*ContentBlockReference, error) {
	if !isValidEntityType(entityType) {
		return nil, ErrInvalidEntityType
	}

	return &ContentBlockReference{
		ID:            uuid.New(),
		ContentBlockID: contentBlockID,
		EntityType:    entityType,
		EntityID:      entityID,
		CreatedAt:     time.Now(),
	}, nil
}

// Validate validates the content block reference entity
func (r *ContentBlockReference) Validate() error {
	if !isValidEntityType(r.EntityType) {
		return ErrInvalidEntityType
	}
	return nil
}

func isValidEntityType(entityType EntityType) bool {
	return entityType == EntityTypeScene ||
		entityType == EntityTypeBeat ||
		entityType == EntityTypeChapter ||
		entityType == EntityTypeCharacter ||
		entityType == EntityTypeLocation ||
		entityType == EntityTypeArtifact ||
		entityType == EntityTypeEvent ||
		entityType == EntityTypeWorld ||
		entityType == EntityTypeRPGSystem ||
		entityType == EntityTypeRPGSkill ||
		entityType == EntityTypeRPGClass ||
		entityType == EntityTypeInventoryItem ||
		entityType == EntityTypeFaction ||
		entityType == EntityTypeLore ||
		entityType == EntityTypeFactionReference ||
		entityType == EntityTypeLoreReference
}

