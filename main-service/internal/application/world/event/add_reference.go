package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddReferenceUseCase handles adding a reference to an event
type AddReferenceUseCase struct {
	eventRepo            repositories.EventRepository
	eventReferenceRepo   repositories.EventReferenceRepository
	characterRepo        repositories.CharacterRepository
	locationRepo         repositories.LocationRepository
	artifactRepo         repositories.ArtifactRepository
	factionRepo          repositories.FactionRepository
	loreRepo             repositories.LoreRepository
	factionReferenceRepo repositories.FactionReferenceRepository
	loreReferenceRepo    repositories.LoreReferenceRepository
	logger               logger.Logger
}

// NewAddReferenceUseCase creates a new AddReferenceUseCase
func NewAddReferenceUseCase(
	eventRepo repositories.EventRepository,
	eventReferenceRepo repositories.EventReferenceRepository,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	artifactRepo repositories.ArtifactRepository,
	factionRepo repositories.FactionRepository,
	loreRepo repositories.LoreRepository,
	factionReferenceRepo repositories.FactionReferenceRepository,
	loreReferenceRepo repositories.LoreReferenceRepository,
	logger logger.Logger,
) *AddReferenceUseCase {
	return &AddReferenceUseCase{
		eventRepo:            eventRepo,
		eventReferenceRepo:   eventReferenceRepo,
		characterRepo:        characterRepo,
		locationRepo:         locationRepo,
		artifactRepo:         artifactRepo,
		factionRepo:          factionRepo,
		loreRepo:             loreRepo,
		factionReferenceRepo: factionReferenceRepo,
		loreReferenceRepo:    loreReferenceRepo,
		logger:               logger,
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
	case "faction_reference":
		// Cross-reference: validate the faction_reference exists
		_, err := uc.factionReferenceRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "faction_reference not found",
			}
		}
	case "lore_reference":
		// Cross-reference: validate the lore_reference exists
		_, err := uc.loreReferenceRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "lore_reference not found",
			}
		}
	default:
		return &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "invalid entity type",
		}
	}

	// Create relationship
	er := world.NewEventReference(input.EventID, input.EntityType, input.EntityID, input.RelationshipType)
	if input.Notes != "" {
		er.UpdateNotes(input.Notes)
	}

	if err := uc.eventReferenceRepo.Create(ctx, er); err != nil {
		uc.logger.Error("failed to add event reference", "error", err, "event_id", input.EventID)
		return err
	}

	uc.logger.Info("event reference added", "event_id", input.EventID, "entity_type", input.EntityType, "entity_id", input.EntityID)
	return nil
}

