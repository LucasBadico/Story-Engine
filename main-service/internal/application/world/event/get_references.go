package event

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// EventReferenceDTO is a compatibility DTO for EventReference (deprecated)
// Maps from EntityRelation to maintain handler compatibility
type EventReferenceDTO struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenant_id"`
	EventID          uuid.UUID `json:"event_id"`
	EntityType       string    `json:"entity_type"`
	EntityID         uuid.UUID `json:"entity_id"`
	RelationshipType *string   `json:"relationship_type"`
	Notes            string    `json:"notes"`
	CreatedAt        string    `json:"created_at"`
}

// GetReferencesUseCase handles retrieving references for an event
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
	EventID  uuid.UUID
}

// GetReferencesOutput represents the output of getting references
type GetReferencesOutput struct {
	References []*EventReferenceDTO
}

// Execute retrieves references for an event
func (uc *GetReferencesUseCase) Execute(ctx context.Context, input GetReferencesInput) (*GetReferencesOutput, error) {
	output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
		TenantID:   input.TenantID,
		SourceType: "event",
		SourceID:   input.EventID,
		Options: repositories.ListOptions{
			Limit: 100, // Default limit for compatibility
		},
	})
	if err != nil {
		uc.logger.Error("failed to get event references", "error", err, "event_id", input.EventID)
		return nil, err
	}

	// Map EntityRelation to EventReferenceDTO
	references := make([]*EventReferenceDTO, 0, len(output.Relations.Items))
	for _, rel := range output.Relations.Items {
		if rel.SourceType != "event" || rel.SourceID != input.EventID {
			continue
		}

		dto := &EventReferenceDTO{
			ID:         rel.ID,
			TenantID:   rel.TenantID,
			EventID:    input.EventID,
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
