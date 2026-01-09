package lore

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddReferenceUseCase handles adding a reference to a lore
type AddReferenceUseCase struct {
	loreRepo              repositories.LoreRepository
	relationRepo          repositories.EntityRelationRepository
	createRelationUseCase *relationapp.CreateRelationUseCase
	characterRepo         repositories.CharacterRepository
	locationRepo          repositories.LocationRepository
	artifactRepo          repositories.ArtifactRepository
	eventRepo             repositories.EventRepository
	factionRepo           repositories.FactionRepository
	logger                logger.Logger
}

// NewAddReferenceUseCase creates a new AddReferenceUseCase
func NewAddReferenceUseCase(
	loreRepo repositories.LoreRepository,
	relationRepo repositories.EntityRelationRepository,
	createRelationUseCase *relationapp.CreateRelationUseCase,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	artifactRepo repositories.ArtifactRepository,
	eventRepo repositories.EventRepository,
	factionRepo repositories.FactionRepository,
	logger logger.Logger,
) *AddReferenceUseCase {
	return &AddReferenceUseCase{
		loreRepo:              loreRepo,
		relationRepo:          relationRepo,
		createRelationUseCase: createRelationUseCase,
		characterRepo:         characterRepo,
		locationRepo:          locationRepo,
		artifactRepo:          artifactRepo,
		eventRepo:             eventRepo,
		factionRepo:           factionRepo,
		logger:                logger,
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
		WorldID:      lore.WorldID,
		SourceType:   "lore",
		SourceID:     input.LoreID,
		TargetType:   input.EntityType,
		TargetID:     input.EntityID,
		RelationType: relationType,
		Attributes:   attributes,
		CreateMirror: false, // Lore references are one-way
	})
	if err != nil {
		uc.logger.Error("failed to add lore reference", "error", err, "lore_id", input.LoreID)
		return err
	}

	uc.logger.Info("lore reference added", "lore_id", input.LoreID, "entity_type", input.EntityType, "entity_id", input.EntityID)
	return nil
}
