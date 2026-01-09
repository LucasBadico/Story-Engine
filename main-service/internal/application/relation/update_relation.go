package relation

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/relation"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateRelationUseCase handles updating entity relations
type UpdateRelationUseCase struct {
	relationRepo     repositories.EntityRelationRepository
	summaryGenerator *SummaryGenerator
	ingestionQueue   queue.IngestionQueue
	logger           logger.Logger
}

// NewUpdateRelationUseCase creates a new UpdateRelationUseCase
func NewUpdateRelationUseCase(
	relationRepo repositories.EntityRelationRepository,
	summaryGenerator *SummaryGenerator,
	ingestionQueue queue.IngestionQueue,
	logger logger.Logger,
) *UpdateRelationUseCase {
	return &UpdateRelationUseCase{
		relationRepo:     relationRepo,
		summaryGenerator: summaryGenerator,
		ingestionQueue:   ingestionQueue,
		logger:           logger,
	}
}

// UpdateRelationInput represents the input for updating a relation
type UpdateRelationInput struct {
	TenantID     uuid.UUID
	ID           uuid.UUID
	RelationType *string
	Attributes   *map[string]interface{}
	Summary      *string
}

// UpdateRelationOutput represents the output of updating a relation
type UpdateRelationOutput struct {
	Relation *relation.EntityRelation
}

// Execute updates an entity relation
func (uc *UpdateRelationUseCase) Execute(ctx context.Context, input UpdateRelationInput) (*UpdateRelationOutput, error) {
	// Get existing relation
	rel, err := uc.relationRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		uc.logger.Error("failed to get relation for update", "error", err, "relation_id", input.ID)
		return nil, err
	}

	// Update fields
	if input.RelationType != nil {
		rel.RelationType = *input.RelationType
	}
	if input.Attributes != nil {
		rel.Attributes = *input.Attributes
	}
	if input.Summary != nil {
		rel.UpdateSummary(*input.Summary)
	} else if input.RelationType != nil {
		// If relation type changed, regenerate summary
		rel.Summary = uc.summaryGenerator.GenerateSummary(rel)
	}

	// Validate
	if err := rel.Validate(); err != nil {
		return nil, err
	}

	// Update in repository
	if err := uc.relationRepo.Update(ctx, rel); err != nil {
		uc.logger.Error("failed to update relation", "error", err, "relation_id", input.ID)
		return nil, err
	}

	uc.logger.Info("relation updated", "relation_id", rel.ID)

	// Enqueue embedding ingestion
	uc.enqueueIngestion(ctx, rel.TenantID, rel.ID)

	return &UpdateRelationOutput{
		Relation: rel,
	}, nil
}

// enqueueIngestion enqueues a relation for embedding ingestion
func (uc *UpdateRelationUseCase) enqueueIngestion(ctx context.Context, tenantID, relationID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "relation", relationID); err != nil {
		uc.logger.Error("failed to enqueue relation ingestion", "error", err, "relation_id", relationID, "tenant_id", tenantID)
	}
}
