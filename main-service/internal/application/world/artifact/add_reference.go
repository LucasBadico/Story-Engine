package artifact

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddArtifactReferenceUseCase handles adding a reference to an artifact
type AddArtifactReferenceUseCase struct {
	artifactRepo          repositories.ArtifactRepository
	relationRepo          repositories.EntityRelationRepository
	createRelationUseCase *relationapp.CreateRelationUseCase
	characterRepo         repositories.CharacterRepository
	locationRepo          repositories.LocationRepository
	logger                logger.Logger
}

// NewAddArtifactReferenceUseCase creates a new AddArtifactReferenceUseCase
func NewAddArtifactReferenceUseCase(
	artifactRepo repositories.ArtifactRepository,
	relationRepo repositories.EntityRelationRepository,
	createRelationUseCase *relationapp.CreateRelationUseCase,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	logger logger.Logger,
) *AddArtifactReferenceUseCase {
	return &AddArtifactReferenceUseCase{
		artifactRepo:          artifactRepo,
		relationRepo:          relationRepo,
		createRelationUseCase: createRelationUseCase,
		characterRepo:         characterRepo,
		locationRepo:          locationRepo,
		logger:                logger,
	}
}

// AddArtifactReferenceInput represents the input for adding a reference
type AddArtifactReferenceInput struct {
	TenantID   uuid.UUID
	ArtifactID uuid.UUID
	EntityType string // "character" or "location"
	EntityID   uuid.UUID
}

// Execute adds a reference to an artifact
func (uc *AddArtifactReferenceUseCase) Execute(ctx context.Context, input AddArtifactReferenceInput) error {
	// Validate artifact exists
	a, err := uc.artifactRepo.GetByID(ctx, input.TenantID, input.ArtifactID)
	if err != nil {
		return err
	}

	// Validate entity exists and belongs to same world
	switch input.EntityType {
	case "character":
		c, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if c.WorldID != a.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "character must belong to the same world as artifact",
			}
		}
	case "location":
		l, err := uc.locationRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		if l.WorldID != a.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "location must belong to the same world as artifact",
			}
		}
	default:
		return &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "invalid entity type (must be 'character' or 'location')",
		}
	}

	// Create relation using entity_relations
	_, err = uc.createRelationUseCase.Execute(ctx, relationapp.CreateRelationInput{
		TenantID:     input.TenantID,
		WorldID:      a.WorldID,
		SourceType:   "artifact",
		SourceID:     input.ArtifactID,
		TargetType:   input.EntityType,
		TargetID:     input.EntityID,
		RelationType: "mentions",
		Attributes:   make(map[string]interface{}),
		CreateMirror: false, // Artifact references are one-way
	})
	if err != nil {
		uc.logger.Error("failed to add artifact reference", "error", err, "artifact_id", input.ArtifactID)
		return err
	}

	uc.logger.Info("artifact reference added", "artifact_id", input.ArtifactID, "entity_type", input.EntityType, "entity_id", input.EntityID)

	return nil
}
