package artifact

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// ArtifactReferenceDTO is a compatibility DTO for ArtifactReference (deprecated)
// Maps from EntityRelation to maintain handler compatibility
type ArtifactReferenceDTO struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	ArtifactID uuid.UUID `json:"artifact_id"`
	EntityType string    `json:"entity_type"` // "character" or "location"
	EntityID   uuid.UUID `json:"entity_id"`
	CreatedAt  string    `json:"created_at"`
}

// GetArtifactReferencesUseCase handles getting references for an artifact
type GetArtifactReferencesUseCase struct {
	listRelationsUseCase *relationapp.ListRelationsBySourceUseCase
	logger               logger.Logger
}

// NewGetArtifactReferencesUseCase creates a new GetArtifactReferencesUseCase
func NewGetArtifactReferencesUseCase(
	listRelationsUseCase *relationapp.ListRelationsBySourceUseCase,
	logger logger.Logger,
) *GetArtifactReferencesUseCase {
	return &GetArtifactReferencesUseCase{
		listRelationsUseCase: listRelationsUseCase,
		logger:               logger,
	}
}

// GetArtifactReferencesInput represents the input for getting references
type GetArtifactReferencesInput struct {
	TenantID   uuid.UUID
	ArtifactID uuid.UUID
}

// GetArtifactReferencesOutput represents the output of getting references
type GetArtifactReferencesOutput struct {
	References []*ArtifactReferenceDTO
}

// Execute retrieves all references for an artifact
func (uc *GetArtifactReferencesUseCase) Execute(ctx context.Context, input GetArtifactReferencesInput) (*GetArtifactReferencesOutput, error) {
	output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
		TenantID:   input.TenantID,
		SourceType: "artifact",
		SourceID:   input.ArtifactID,
		Options: repositories.ListOptions{
			Limit: 100, // Default limit for compatibility
		},
	})
	if err != nil {
		uc.logger.Error("failed to get artifact references", "error", err, "artifact_id", input.ArtifactID)
		return nil, err
	}

	// Map EntityRelation to ArtifactReferenceDTO
	references := make([]*ArtifactReferenceDTO, 0, len(output.Relations.Items))
	for _, rel := range output.Relations.Items {
		if rel.SourceType != "artifact" || rel.SourceID != input.ArtifactID {
			continue
		}

		dto := &ArtifactReferenceDTO{
			ID:         rel.ID,
			TenantID:   rel.TenantID,
			ArtifactID: input.ArtifactID,
			EntityType: rel.TargetType,
			EntityID:   rel.TargetID,
			CreatedAt:  rel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		references = append(references, dto)
	}

	return &GetArtifactReferencesOutput{
		References: references,
	}, nil
}
