package scene

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// SceneReferenceDTO is a compatibility DTO for SceneReference (deprecated)
// Maps from EntityRelation to maintain handler compatibility
type SceneReferenceDTO struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	SceneID    uuid.UUID `json:"scene_id"`
	EntityType string    `json:"entity_type"` // "character", "location", or "artifact"
	EntityID   uuid.UUID `json:"entity_id"`
	CreatedAt  string    `json:"created_at"`
}

// GetSceneReferencesUseCase handles getting references for a scene
type GetSceneReferencesUseCase struct {
	listRelationsUseCase *relationapp.ListRelationsBySourceUseCase
	logger               logger.Logger
}

// NewGetSceneReferencesUseCase creates a new GetSceneReferencesUseCase
func NewGetSceneReferencesUseCase(
	listRelationsUseCase *relationapp.ListRelationsBySourceUseCase,
	logger logger.Logger,
) *GetSceneReferencesUseCase {
	return &GetSceneReferencesUseCase{
		listRelationsUseCase: listRelationsUseCase,
		logger:               logger,
	}
}

// GetSceneReferencesInput represents the input for getting references
type GetSceneReferencesInput struct {
	TenantID uuid.UUID
	SceneID  uuid.UUID
}

// GetSceneReferencesOutput represents the output of getting references
type GetSceneReferencesOutput struct {
	References []*SceneReferenceDTO
}

// Execute retrieves all references for a scene
func (uc *GetSceneReferencesUseCase) Execute(ctx context.Context, input GetSceneReferencesInput) (*GetSceneReferencesOutput, error) {
	output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
		TenantID:   input.TenantID,
		SourceType: "scene",
		SourceID:   input.SceneID,
		Options: repositories.ListOptions{
			Limit: 100, // Default limit for compatibility
		},
	})
	if err != nil {
		uc.logger.Error("failed to get scene references", "error", err, "scene_id", input.SceneID)
		return nil, err
	}

	// Map EntityRelation to SceneReferenceDTO
	references := make([]*SceneReferenceDTO, 0, len(output.Relations.Items))
	for _, rel := range output.Relations.Items {
		if rel.SourceType != "scene" || rel.SourceID != input.SceneID {
			continue
		}

		dto := &SceneReferenceDTO{
			ID:         rel.ID,
			TenantID:   rel.TenantID,
			SceneID:    input.SceneID,
			EntityType: rel.TargetType,
			EntityID:   rel.TargetID,
			CreatedAt:  rel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		references = append(references, dto)
	}

	return &GetSceneReferencesOutput{
		References: references,
	}, nil
}
