package faction

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddReferenceUseCase handles adding a reference to a faction
type AddReferenceUseCase struct {
	factionRepo           repositories.FactionRepository
	relationRepo          repositories.EntityRelationRepository
	createRelationUseCase *relationapp.CreateRelationUseCase
	characterRepo         repositories.CharacterRepository
	locationRepo          repositories.LocationRepository
	artifactRepo          repositories.ArtifactRepository
	eventRepo             repositories.EventRepository
	loreRepo              repositories.LoreRepository
	logger                logger.Logger
}

// NewAddReferenceUseCase creates a new AddReferenceUseCase
func NewAddReferenceUseCase(
	factionRepo repositories.FactionRepository,
	relationRepo repositories.EntityRelationRepository,
	createRelationUseCase *relationapp.CreateRelationUseCase,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	artifactRepo repositories.ArtifactRepository,
	eventRepo repositories.EventRepository,
	loreRepo repositories.LoreRepository,
	logger logger.Logger,
) *AddReferenceUseCase {
	return &AddReferenceUseCase{
		factionRepo:           factionRepo,
		relationRepo:          relationRepo,
		createRelationUseCase: createRelationUseCase,
		characterRepo:         characterRepo,
		locationRepo:          locationRepo,
		artifactRepo:          artifactRepo,
		eventRepo:             eventRepo,
		loreRepo:              loreRepo,
		logger:                logger,
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
	if input.Role != nil && *input.Role != "" {
		relationType = *input.Role
	}

	attributes := make(map[string]interface{})
	if input.Notes != "" {
		attributes["notes"] = input.Notes
	}

	_, err = uc.createRelationUseCase.Execute(ctx, relationapp.CreateRelationInput{
		TenantID:     input.TenantID,
		WorldID:      faction.WorldID,
		SourceType:   "faction",
		SourceID:     input.FactionID,
		TargetType:   input.EntityType,
		TargetID:     input.EntityID,
		RelationType: relationType,
		Attributes:   attributes,
		CreateMirror: false, // Faction references are one-way
	})
	if err != nil {
		uc.logger.Error("failed to add faction reference", "error", err, "faction_id", input.FactionID)
		return err
	}

	uc.logger.Info("faction reference added", "faction_id", input.FactionID, "entity_type", input.EntityType, "entity_id", input.EntityID)
	return nil
}
