package faction

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddReferenceUseCase handles adding a reference to a faction
type AddReferenceUseCase struct {
	factionRepo          repositories.FactionRepository
	factionReferenceRepo repositories.FactionReferenceRepository
	characterRepo        repositories.CharacterRepository
	locationRepo         repositories.LocationRepository
	artifactRepo         repositories.ArtifactRepository
	eventRepo            repositories.EventRepository
	loreRepo             repositories.LoreRepository
	loreReferenceRepo    repositories.LoreReferenceRepository
	logger               logger.Logger
}

// NewAddReferenceUseCase creates a new AddReferenceUseCase
func NewAddReferenceUseCase(
	factionRepo repositories.FactionRepository,
	factionReferenceRepo repositories.FactionReferenceRepository,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	artifactRepo repositories.ArtifactRepository,
	eventRepo repositories.EventRepository,
	loreRepo repositories.LoreRepository,
	loreReferenceRepo repositories.LoreReferenceRepository,
	logger logger.Logger,
) *AddReferenceUseCase {
	return &AddReferenceUseCase{
		factionRepo:          factionRepo,
		factionReferenceRepo: factionReferenceRepo,
		characterRepo:        characterRepo,
		locationRepo:         locationRepo,
		artifactRepo:         artifactRepo,
		eventRepo:            eventRepo,
		loreRepo:             loreRepo,
		loreReferenceRepo:    loreReferenceRepo,
		logger:               logger,
	}
}

// AddReferenceInput represents the input for adding a reference
type AddReferenceInput struct {
	TenantID   uuid.UUID
	FactionID  uuid.UUID
	EntityType string
	EntityID   uuid.UUID
	Role       *string
	Notes      string
}

// Execute adds a reference to a faction
func (uc *AddReferenceUseCase) Execute(ctx context.Context, input AddReferenceInput) error {
	// Validate faction exists
	faction, err := uc.factionRepo.GetByID(ctx, input.TenantID, input.FactionID)
	if err != nil {
		return err
	}

	// Validate entity exists (basic validation - full validation depends on entity_type)
	// For cross-references (lore_reference), we validate the reference exists
	// For regular entities, we validate the entity exists
	switch input.EntityType {
	case "character":
		c, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if c.WorldID != faction.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "character must belong to the same world as faction",
			}
		}
	case "location":
		l, err := uc.locationRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if l.WorldID != faction.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "location must belong to the same world as faction",
			}
		}
	case "artifact":
		a, err := uc.artifactRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if a.WorldID != faction.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "artifact must belong to the same world as faction",
			}
		}
	case "event":
		e, err := uc.eventRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if e.WorldID != faction.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "event must belong to the same world as faction",
			}
		}
	case "lore":
		l, err := uc.loreRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if l.WorldID != faction.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "lore must belong to the same world as faction",
			}
		}
	case "faction":
		f, err := uc.factionRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if f.WorldID != faction.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "faction must belong to the same world",
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
	fr := world.NewFactionReference(input.FactionID, input.EntityType, input.EntityID, input.Role)
	if input.Notes != "" {
		fr.UpdateNotes(input.Notes)
	}

	if err := uc.factionReferenceRepo.Create(ctx, fr); err != nil {
		uc.logger.Error("failed to add faction reference", "error", err, "faction_id", input.FactionID)
		return err
	}

	uc.logger.Info("faction reference added", "faction_id", input.FactionID, "entity_type", input.EntityType, "entity_id", input.EntityID)
	return nil
}

