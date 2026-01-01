package artifact

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddArtifactReferenceUseCase handles adding a reference to an artifact
type AddArtifactReferenceUseCase struct {
	artifactRepo         repositories.ArtifactRepository
	artifactReferenceRepo repositories.ArtifactReferenceRepository
	characterRepo        repositories.CharacterRepository
	locationRepo         repositories.LocationRepository
	logger               logger.Logger
}

// NewAddArtifactReferenceUseCase creates a new AddArtifactReferenceUseCase
func NewAddArtifactReferenceUseCase(
	artifactRepo repositories.ArtifactRepository,
	artifactReferenceRepo repositories.ArtifactReferenceRepository,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	logger logger.Logger,
) *AddArtifactReferenceUseCase {
	return &AddArtifactReferenceUseCase{
		artifactRepo:         artifactRepo,
		artifactReferenceRepo: artifactReferenceRepo,
		characterRepo:        characterRepo,
		locationRepo:         locationRepo,
		logger:               logger,
	}
}

// AddArtifactReferenceInput represents the input for adding a reference
type AddArtifactReferenceInput struct {
	ArtifactID uuid.UUID
	EntityType world.ArtifactReferenceEntityType
	EntityID   uuid.UUID
}

// Execute adds a reference to an artifact
func (uc *AddArtifactReferenceUseCase) Execute(ctx context.Context, input AddArtifactReferenceInput) error {
	// Validate artifact exists
	a, err := uc.artifactRepo.GetByID(ctx, input.ArtifactID)
	if err != nil {
		return err
	}

	// Validate entity exists and belongs to same world
	switch input.EntityType {
	case world.ArtifactReferenceEntityTypeCharacter:
		c, err := uc.characterRepo.GetByID(ctx, input.EntityID)
		if err != nil {
			return err
		}
		if c.WorldID != a.WorldID {
			return &platformerrors.ValidationError{
				Field:   "entity_id",
				Message: "character must belong to the same world as artifact",
			}
		}
	case world.ArtifactReferenceEntityTypeLocation:
		l, err := uc.locationRepo.GetByID(ctx, input.EntityID)
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
			Message: "invalid entity type",
		}
	}

	ref, err := world.NewArtifactReference(input.ArtifactID, input.EntityType, input.EntityID)
	if err != nil {
		return err
	}

	if err := uc.artifactReferenceRepo.Create(ctx, ref); err != nil {
		uc.logger.Error("failed to add artifact reference", "error", err, "artifact_id", input.ArtifactID)
		return err
	}

	uc.logger.Info("artifact reference added", "artifact_id", input.ArtifactID, "entity_type", input.EntityType, "entity_id", input.EntityID)

	return nil
}

