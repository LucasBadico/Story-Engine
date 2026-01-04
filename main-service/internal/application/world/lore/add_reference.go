package lore

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddReferenceUseCase handles adding a reference to a lore
type AddReferenceUseCase struct {
	loreRepo          repositories.LoreRepository
	loreReferenceRepo repositories.LoreReferenceRepository
	characterRepo     repositories.CharacterRepository
	locationRepo      repositories.LocationRepository
	artifactRepo      repositories.ArtifactRepository
	eventRepo         repositories.EventRepository
	factionRepo       repositories.FactionRepository
	factionReferenceRepo repositories.FactionReferenceRepository
	logger            logger.Logger
}

// NewAddReferenceUseCase creates a new AddReferenceUseCase
func NewAddReferenceUseCase(
	loreRepo repositories.LoreRepository,
	loreReferenceRepo repositories.LoreReferenceRepository,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	artifactRepo repositories.ArtifactRepository,
	eventRepo repositories.EventRepository,
	factionRepo repositories.FactionRepository,
	factionReferenceRepo repositories.FactionReferenceRepository,
	logger logger.Logger,
) *AddReferenceUseCase {
	return &AddReferenceUseCase{
		loreRepo:          loreRepo,
		loreReferenceRepo: loreReferenceRepo,
		characterRepo:     characterRepo,
		locationRepo:      locationRepo,
		artifactRepo:      artifactRepo,
		eventRepo:         eventRepo,
		factionRepo:       factionRepo,
		factionReferenceRepo: factionReferenceRepo,
		logger:            logger,
	}
}

// AddReferenceInput represents the input for adding a reference
type AddReferenceInput struct {
	TenantID         uuid.UUID
	LoreID           uuid.UUID
	EntityType       string
	EntityID         uuid.UUID
	RelationshipType *string
	Notes            string
}

// Execute adds a reference to a lore
func (uc *AddReferenceUseCase) Execute(ctx context.Context, input AddReferenceInput) error {
	// Validate lore exists
	lore, err := uc.loreRepo.GetByID(ctx, input.TenantID, input.LoreID)
	if err != nil {
		return err
	}

	// Validate entity exists
	switch input.EntityType {
	case "character":
		c, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if c.WorldID != lore.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "character must belong to the same world as lore",
			}
		}
	case "location":
		l, err := uc.locationRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if l.WorldID != lore.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "location must belong to the same world as lore",
			}
		}
	case "artifact":
		a, err := uc.artifactRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if a.WorldID != lore.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "artifact must belong to the same world as lore",
			}
		}
	case "event":
		e, err := uc.eventRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if e.WorldID != lore.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "event must belong to the same world as lore",
			}
		}
	case "faction":
		f, err := uc.factionRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if f.WorldID != lore.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "faction must belong to the same world as lore",
			}
		}
	case "lore":
		l, err := uc.loreRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if l.WorldID != lore.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "lore must belong to the same world",
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
	default:
		return &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "invalid entity type",
		}
	}

	// Create relationship
	lr := world.NewLoreReference(input.LoreID, input.EntityType, input.EntityID, input.RelationshipType)
	if input.Notes != "" {
		lr.UpdateNotes(input.Notes)
	}

	if err := uc.loreReferenceRepo.Create(ctx, lr); err != nil {
		uc.logger.Error("failed to add lore reference", "error", err, "lore_id", input.LoreID)
		return err
	}

	uc.logger.Info("lore reference added", "lore_id", input.LoreID, "entity_type", input.EntityType, "entity_id", input.EntityID)
	return nil
}

