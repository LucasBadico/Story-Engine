package scene

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// AddSceneReferenceUseCase handles adding a reference to a scene
type AddSceneReferenceUseCase struct {
	sceneRepo             repositories.SceneRepository
	storyRepo             repositories.StoryRepository
	createRelationUseCase *relationapp.CreateRelationUseCase
	listRelationsUseCase  *relationapp.ListRelationsBySourceUseCase
	characterRepo         repositories.CharacterRepository
	locationRepo          repositories.LocationRepository
	artifactRepo          repositories.ArtifactRepository
	logger                logger.Logger
}

// NewAddSceneReferenceUseCase creates a new AddSceneReferenceUseCase
func NewAddSceneReferenceUseCase(
	sceneRepo repositories.SceneRepository,
	storyRepo repositories.StoryRepository,
	createRelationUseCase *relationapp.CreateRelationUseCase,
	listRelationsUseCase *relationapp.ListRelationsBySourceUseCase,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	artifactRepo repositories.ArtifactRepository,
	logger logger.Logger,
) *AddSceneReferenceUseCase {
	return &AddSceneReferenceUseCase{
		sceneRepo:             sceneRepo,
		storyRepo:             storyRepo,
		createRelationUseCase: createRelationUseCase,
		listRelationsUseCase:  listRelationsUseCase,
		characterRepo:         characterRepo,
		locationRepo:          locationRepo,
		artifactRepo:          artifactRepo,
		logger:                logger,
	}
}

// AddSceneReferenceInput represents the input for adding a reference
type AddSceneReferenceInput struct {
	TenantID   uuid.UUID
	SceneID    uuid.UUID
	EntityType string // "character", "location", or "artifact"
	EntityID   uuid.UUID
}

// Execute adds a reference to a scene
func (uc *AddSceneReferenceUseCase) Execute(ctx context.Context, input AddSceneReferenceInput) error {
	// Validate scene exists
	s, err := uc.sceneRepo.GetByID(ctx, input.TenantID, input.SceneID)
	if err != nil {
		return err
	}

	// Validate entity exists and belongs to same world (via story -> world)
	// We need to get the story to get the world_id
	// For now, we'll just validate the entity exists
	switch input.EntityType {
	case "character":
		_, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		// TODO: Validate character belongs to same world as scene's story
	case "location":
		_, err := uc.locationRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		// TODO: Validate location belongs to same world as scene's story
	case "artifact":
		_, err := uc.artifactRepo.GetByID(ctx, input.TenantID, input.EntityID)
		if err != nil {
			return err
		}
		// TODO: Validate artifact belongs to same world as scene's story
	default:
		return &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "invalid entity type (must be 'character', 'location', or 'artifact')",
		}
	}

	// Get story to get world_id
	st, err := uc.storyRepo.GetByID(ctx, input.TenantID, s.StoryID)
	if err != nil {
		return err
	}

	if st.WorldID == nil {
		return &platformerrors.ValidationError{
			Field:   "story",
			Message: "story must have a world_id to create scene references",
		}
	}

	// Prevent duplicate references
	output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
		TenantID:   input.TenantID,
		SourceType: "scene",
		SourceID:   input.SceneID,
		Options: repositories.ListOptions{
			Limit: 100,
		},
	})
	if err == nil {
		for _, rel := range output.Relations.Items {
			if rel.TargetType == input.EntityType && rel.TargetID == input.EntityID {
				return &platformerrors.ValidationError{
					Field:   "entity_id",
					Message: "reference already exists",
				}
			}
		}
	}

	// Create relation using entity_relations
	_, err = uc.createRelationUseCase.Execute(ctx, relationapp.CreateRelationInput{
		TenantID:     input.TenantID,
		WorldID:      *st.WorldID,
		SourceType:   "scene",
		SourceID:     input.SceneID,
		TargetType:   input.EntityType,
		TargetID:     input.EntityID,
		RelationType: "mentions",
		Attributes:   make(map[string]interface{}),
		CreateMirror: false, // Scene references are one-way
	})
	if err != nil {
		uc.logger.Error("failed to add scene reference", "error", err, "scene_id", input.SceneID)
		return err
	}

	uc.logger.Info("scene reference added", "scene_id", input.SceneID, "entity_type", input.EntityType, "entity_id", input.EntityID, "story_id", s.StoryID)

	return nil
}
