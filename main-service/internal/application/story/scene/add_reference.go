package scene

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddSceneReferenceUseCase handles adding a reference to a scene
type AddSceneReferenceUseCase struct {
	sceneRepo         repositories.SceneRepository
	sceneReferenceRepo repositories.SceneReferenceRepository
	characterRepo     repositories.CharacterRepository
	locationRepo      repositories.LocationRepository
	artifactRepo      repositories.ArtifactRepository
	logger            logger.Logger
}

// NewAddSceneReferenceUseCase creates a new AddSceneReferenceUseCase
func NewAddSceneReferenceUseCase(
	sceneRepo repositories.SceneRepository,
	sceneReferenceRepo repositories.SceneReferenceRepository,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	artifactRepo repositories.ArtifactRepository,
	logger logger.Logger,
) *AddSceneReferenceUseCase {
	return &AddSceneReferenceUseCase{
		sceneRepo:         sceneRepo,
		sceneReferenceRepo: sceneReferenceRepo,
		characterRepo:     characterRepo,
		locationRepo:      locationRepo,
		artifactRepo:      artifactRepo,
		logger:            logger,
	}
}

// AddSceneReferenceInput represents the input for adding a reference
type AddSceneReferenceInput struct {
	SceneID   uuid.UUID
	EntityType story.SceneReferenceEntityType
	EntityID   uuid.UUID
}

// Execute adds a reference to a scene
func (uc *AddSceneReferenceUseCase) Execute(ctx context.Context, input AddSceneReferenceInput) error {
	// Validate scene exists
	s, err := uc.sceneRepo.GetByID(ctx, input.SceneID)
	if err != nil {
		return err
	}

	// Validate entity exists and belongs to same world (via story -> world)
	// We need to get the story to get the world_id
	// For now, we'll just validate the entity exists
	switch input.EntityType {
	case story.SceneReferenceEntityTypeCharacter:
		_, err := uc.characterRepo.GetByID(ctx, input.EntityID)
		if err != nil {
			return err
		}
		// TODO: Validate character belongs to same world as scene's story
	case story.SceneReferenceEntityTypeLocation:
		_, err := uc.locationRepo.GetByID(ctx, input.EntityID)
		if err != nil {
			return err
		}
		// TODO: Validate location belongs to same world as scene's story
	case story.SceneReferenceEntityTypeArtifact:
		_, err := uc.artifactRepo.GetByID(ctx, input.EntityID)
		if err != nil {
			return err
		}
		// TODO: Validate artifact belongs to same world as scene's story
	default:
		return &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "invalid entity type",
		}
	}

	// Prevent duplicate references
	existingRefs, err := uc.sceneReferenceRepo.ListByScene(ctx, input.SceneID)
	if err == nil {
		for _, ref := range existingRefs {
			if ref.EntityType == input.EntityType && ref.EntityID == input.EntityID {
				return &platformerrors.ValidationError{
					Field:   "entity_id",
					Message: "reference already exists",
				}
			}
		}
	}

	ref, err := story.NewSceneReference(input.SceneID, input.EntityType, input.EntityID)
	if err != nil {
		return err
	}

	if err := uc.sceneReferenceRepo.Create(ctx, ref); err != nil {
		uc.logger.Error("failed to add scene reference", "error", err, "scene_id", input.SceneID)
		return err
	}

	uc.logger.Info("scene reference added", "scene_id", input.SceneID, "entity_type", input.EntityType, "entity_id", input.EntityID, "story_id", s.StoryID)

	return nil
}

