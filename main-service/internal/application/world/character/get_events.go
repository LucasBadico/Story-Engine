package character

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// EventReferenceDTO is a compatibility DTO for EventReference (deprecated)
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

// GetCharacterEventsUseCase handles getting events for a character
type GetCharacterEventsUseCase struct {
	listRelationsUseCase *relationapp.ListRelationsByTargetUseCase
	logger               logger.Logger
}

// NewGetCharacterEventsUseCase creates a new GetCharacterEventsUseCase
func NewGetCharacterEventsUseCase(
	listRelationsUseCase *relationapp.ListRelationsByTargetUseCase,
	logger logger.Logger,
) *GetCharacterEventsUseCase {
	return &GetCharacterEventsUseCase{
		listRelationsUseCase: listRelationsUseCase,
		logger:               logger,
	}
}

// GetCharacterEventsInput represents the input for getting character events
type GetCharacterEventsInput struct {
	TenantID    uuid.UUID
	CharacterID uuid.UUID
}

// GetCharacterEventsOutput represents the output of getting character events
type GetCharacterEventsOutput struct {
	EventReferences []*EventReferenceDTO
}

// Execute retrieves all event references for a character
func (uc *GetCharacterEventsUseCase) Execute(ctx context.Context, input GetCharacterEventsInput) (*GetCharacterEventsOutput, error) {
	output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsByTargetInput{
		TenantID:   input.TenantID,
		TargetType: "character",
		TargetID:   input.CharacterID,
		Options: repositories.ListOptions{
			Limit: 100,
		},
	})
	if err != nil {
		uc.logger.Error("failed to get character events", "error", err, "character_id", input.CharacterID)
		return nil, err
	}

	// Map EntityRelation to EventReferenceDTO (only for relations where source is event)
	references := make([]*EventReferenceDTO, 0)
	for _, rel := range output.Relations.Items {
		if rel.SourceType != "event" || rel.TargetType != "character" {
			continue
		}

		dto := &EventReferenceDTO{
			ID:         rel.ID,
			TenantID:   rel.TenantID,
			EventID:    rel.SourceID,
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

	return &GetCharacterEventsOutput{
		EventReferences: references,
	}, nil
}
