package event

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddReferenceUseCase handles adding a reference to an event
type AddReferenceUseCase struct {
	eventRepo             repositories.EventRepository
	relationRepo          repositories.EntityRelationRepository
	createRelationUseCase *relationapp.CreateRelationUseCase
	characterRepo         repositories.CharacterRepository
	locationRepo          repositories.LocationRepository
	artifactRepo          repositories.ArtifactRepository
	factionRepo           repositories.FactionRepository
	loreRepo              repositories.LoreRepository
	logger                logger.Logger
}

// NewAddReferenceUseCase creates a new AddReferenceUseCase
func NewAddReferenceUseCase(
	eventRepo repositories.EventRepository,
	relationRepo repositories.EntityRelationRepository,
	createRelationUseCase *relationapp.CreateRelationUseCase,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	artifactRepo repositories.ArtifactRepository,
	factionRepo repositories.FactionRepository,
	loreRepo repositories.LoreRepository,
	logger logger.Logger,
) *AddReferenceUseCase {
	return &AddReferenceUseCase{
		eventRepo:             eventRepo,
		relationRepo:          relationRepo,
		createRelationUseCase: createRelationUseCase,
		characterRepo:         characterRepo,
		locationRepo:          locationRepo,
		artifactRepo:          artifactRepo,
		factionRepo:           factionRepo,
		loreRepo:              loreRepo,
		logger:                logger,
	}
}

// AddReferenceInput represents the input for adding a reference
type AddReferenceInput struct {
	TenantID         uuid.UUID
	EventID          uuid.UUID
	EntityType       string
	EntityID         uuid.UUID
	RelationshipType *string // "role" para character/artifact, "significance" para location
	Notes            string
}

// Execute adds a reference to an event
func (uc *AddReferenceUseCase) Execute(ctx context.Context, input AddReferenceInput) error {
	// Validate event exists
	event, err := uc.eventRepo.GetByID(ctx, input.TenantID, input.EventID)
	if err != nil {
		return err
	}

	// Validate entity exists (basic validation - full validation depends on entity_type)
	// For cross-references (faction_reference, lore_reference), we validate the reference exists
	// For regular entities, we validate the entity exists
	switch input.EntityType {
	case "character":
		c, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if c.WorldID != event.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "character must belong to the same world as event",
			}
		}
	case "location":
		l, err := uc.locationRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if l.WorldID != event.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "location must belong to the same world as event",
			}
		}
	case "artifact":
		a, err := uc.artifactRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if a.WorldID != event.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "artifact must belong to the same world as event",
			}
		}
	case "faction":
		f, err := uc.factionRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if f.WorldID != event.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "faction must belong to the same world as event",
			}
		}
	case "lore":
		l, err := uc.loreRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if l.WorldID != event.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "lore must belong to the same world as event",
			}
		}
	case "faction_reference", "lore_reference":
		// Cross-references are now handled as regular entity_relations
		// No special validation needed - they will be validated when the relation is created
	default:
		return &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "invalid entity type",
		}
	}

	// Create relation using entity_relations
	relationType := "mentions"
	if input.RelationshipType != nil && *input.RelationshipType != "" {
		relationType = *input.RelationshipType
	}

	attributes := make(map[string]interface{})
	if input.Notes != "" {
		attributes["notes"] = input.Notes
	}

	_, err = uc.createRelationUseCase.Execute(ctx, relationapp.CreateRelationInput{
		TenantID:     input.TenantID,
		WorldID:      event.WorldID,
		SourceType:   "event",
		SourceID:     input.EventID,
		TargetType:   input.EntityType,
		TargetID:     input.EntityID,
		RelationType: relationType,
		Attributes:   attributes,
		CreateMirror: false, // Event references are one-way
	})
	if err != nil {
		uc.logger.Error("failed to add event reference", "error", err, "event_id", input.EventID)
		return err
	}

	uc.logger.Info("event reference added", "event_id", input.EventID, "entity_type", input.EntityType, "entity_id", input.EntityID)
	return nil
}
