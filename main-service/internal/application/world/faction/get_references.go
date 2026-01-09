package faction

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// FactionReferenceDTO is a compatibility DTO for FactionReference (deprecated)
// Maps from EntityRelation to maintain handler compatibility
type FactionReferenceDTO struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	FactionID  uuid.UUID `json:"faction_id"`
	EntityType string    `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
	Role       *string   `json:"role"`
	Notes      string    `json:"notes"`
	CreatedAt  string    `json:"created_at"`
}

// GetReferencesUseCase handles retrieving references for a faction
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
	TenantID  uuid.UUID
	FactionID uuid.UUID
}

// GetReferencesOutput represents the output of getting references
type GetReferencesOutput struct {
	References []*FactionReferenceDTO
}

// Execute retrieves references for a faction
func (uc *GetReferencesUseCase) Execute(ctx context.Context, input GetReferencesInput) (*GetReferencesOutput, error) {
	output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
		TenantID:   input.TenantID,
		SourceType: "faction",
		SourceID:   input.FactionID,
		Options: repositories.ListOptions{
			Limit: 100, // Default limit for compatibility
		},
	})
	if err != nil {
		uc.logger.Error("failed to get faction references", "error", err, "faction_id", input.FactionID)
		return nil, err
	}

	// Map EntityRelation to FactionReferenceDTO
	references := make([]*FactionReferenceDTO, 0, len(output.Relations.Items))
	for _, rel := range output.Relations.Items {
		if rel.SourceType != "faction" || rel.SourceID != input.FactionID {
			continue
		}

		dto := &FactionReferenceDTO{
			ID:         rel.ID,
			TenantID:   rel.TenantID,
			FactionID:  input.FactionID,
			EntityType: rel.TargetType,
			EntityID:   rel.TargetID,
			CreatedAt:  rel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if rel.RelationType != "" {
			dto.Role = &rel.RelationType
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
