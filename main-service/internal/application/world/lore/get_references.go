package lore

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// LoreReferenceDTO is a compatibility DTO for LoreReference (deprecated)
// Maps from EntityRelation to maintain handler compatibility
type LoreReferenceDTO struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenant_id"`
	LoreID           uuid.UUID `json:"lore_id"`
	EntityType       string    `json:"entity_type"`
	EntityID         uuid.UUID `json:"entity_id"`
	RelationshipType *string   `json:"relationship_type"`
	Notes            string    `json:"notes"`
	CreatedAt        string    `json:"created_at"`
}

// GetReferencesUseCase handles retrieving references for a lore
type GetReferencesUseCase struct {
	listRelationsUseCase *relationapp.ListRelationsBySourceUseCase
	logger               logger.Logger
}

// NewGetReferencesUseCase creates a new GetReferencesUseCase
func NewGetReferencesUseCase(
	listRelationsUseCase *relationapp.ListRelationsBySourceUseCase,
	logger logger.Logger,
) *GetReferencesUseCase {
	return &GetReferencesUseCase{
		listRelationsUseCase: listRelationsUseCase,
		logger:               logger,
	}
}

// GetReferencesInput represents the input for getting references
type GetReferencesInput struct {
	TenantID uuid.UUID
	LoreID   uuid.UUID
}

// GetReferencesOutput represents the output of getting references
type GetReferencesOutput struct {
	References []*LoreReferenceDTO
}

// Execute retrieves references for a lore
func (uc *GetReferencesUseCase) Execute(ctx context.Context, input GetReferencesInput) (*GetReferencesOutput, error) {
	output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
		TenantID:   input.TenantID,
		SourceType: "lore",
		SourceID:   input.LoreID,
		Options: repositories.ListOptions{
			Limit: 100, // Default limit for compatibility
		},
	})
	if err != nil {
		uc.logger.Error("failed to get lore references", "error", err, "lore_id", input.LoreID)
		return nil, err
	}

	// Map EntityRelation to LoreReferenceDTO
	references := make([]*LoreReferenceDTO, 0, len(output.Relations.Items))
	for _, rel := range output.Relations.Items {
		if rel.SourceType != "lore" || rel.SourceID != input.LoreID {
			continue
		}

		dto := &LoreReferenceDTO{
			ID:         rel.ID,
			TenantID:   rel.TenantID,
			LoreID:     input.LoreID,
			EntityType: rel.TargetType,
			EntityID:   rel.TargetID,
			CreatedAt:  rel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if rel.RelationType != "" {
			dto.RelationshipType = &rel.RelationType
		}

		if notes, ok := rel.Attributes["notes"].(string); ok {
			dto.Notes = notes
		}

		references = append(references, dto)
	}

	return &GetReferencesOutput{
		References: references,
	}, nil
}
