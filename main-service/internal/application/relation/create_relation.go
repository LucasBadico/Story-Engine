package relation

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/relation"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateRelationUseCase handles creating entity relations
type CreateRelationUseCase struct {
	relationRepo     repositories.EntityRelationRepository
	summaryGenerator *SummaryGenerator
	ingestionQueue   queue.IngestionQueue
	logger           logger.Logger
}

// NewCreateRelationUseCase creates a new CreateRelationUseCase
func NewCreateRelationUseCase(
	relationRepo repositories.EntityRelationRepository,
	summaryGenerator *SummaryGenerator,
	ingestionQueue queue.IngestionQueue,
	logger logger.Logger,
) *CreateRelationUseCase {
	return &CreateRelationUseCase{
		relationRepo:     relationRepo,
		summaryGenerator: summaryGenerator,
		ingestionQueue:   ingestionQueue,
		logger:           logger,
	}
}

// CreateRelationInput represents the input for creating a relation
// Both SourceID and TargetID are required - entities must exist
type CreateRelationInput struct {
	TenantID        uuid.UUID
	WorldID         uuid.UUID
	SourceType      string
	SourceID        uuid.UUID // Required - entity must exist
	TargetType      string
	TargetID        uuid.UUID // Required - entity must exist
	RelationType    string
	ContextType     *string
	ContextID       *uuid.UUID
	Attributes      map[string]interface{}
	Summary         string
	CreatedByUserID *uuid.UUID
	CreateMirror    bool // If true, creates the inverse relation automatically
}

// CreateRelationOutput represents the output of creating a relation
type CreateRelationOutput struct {
	Relation *relation.EntityRelation
	Mirror   *relation.EntityRelation // Only set if CreateMirror is true
}

// Execute creates a new entity relation
func (uc *CreateRelationUseCase) Execute(ctx context.Context, input CreateRelationInput) (*CreateRelationOutput, error) {
	// Create the relation
	rel, err := relation.NewEntityRelation(
		input.TenantID,
		input.WorldID,
		input.SourceType,
		input.SourceID,
		input.TargetType,
		input.TargetID,
		input.RelationType,
	)
	if err != nil {
		return nil, err
	}

	// Set optional fields
	rel.ContextType = input.ContextType
	rel.ContextID = input.ContextID
	if input.Attributes != nil {
		rel.Attributes = input.Attributes
	}
	rel.CreatedByUserID = input.CreatedByUserID

	// Generate summary if not provided
	if input.Summary == "" {
		rel.Summary = uc.summaryGenerator.GenerateSummary(rel)
	} else {
		rel.Summary = input.Summary
	}

	// Validate
	if err := rel.Validate(); err != nil {
		return nil, err
	}

	// Check for duplicate relation (uniqueness check at application level)
	existingRelations, err := uc.relationRepo.ListBySource(ctx, input.TenantID, input.SourceType, input.SourceID, repositories.ListOptions{
		Limit:        100,
		RelationType: &input.RelationType,
	})
	if err == nil {
		for _, existing := range existingRelations.Items {
			// Check if it's a duplicate based on: source_id, target_id, relation_type, context_type, context_id
			if existing.SourceID == input.SourceID &&
				existing.TargetID == input.TargetID &&
				existing.RelationType == input.RelationType {
				// Compare context (both nil or both equal)
				contextMatch := false
				if existing.ContextType == nil && input.ContextType == nil {
					if existing.ContextID == nil && input.ContextID == nil {
						contextMatch = true
					} else if existing.ContextID != nil && input.ContextID != nil && *existing.ContextID == *input.ContextID {
						contextMatch = true
					}
				} else if existing.ContextType != nil && input.ContextType != nil &&
					*existing.ContextType == *input.ContextType {
					if existing.ContextID == nil && input.ContextID == nil {
						contextMatch = true
					} else if existing.ContextID != nil && input.ContextID != nil && *existing.ContextID == *input.ContextID {
						contextMatch = true
					}
				}
				if contextMatch {
					return nil, &platformerrors.ValidationError{
						Field:   "relation",
						Message: "a relation with the same source, target, type, and context already exists",
					}
				}
			}
		}
	}

	output := &CreateRelationOutput{
		Relation: rel,
	}

	// Create with or without mirror
	if input.CreateMirror {
		mirror, err := uc.relationRepo.CreateWithMirror(ctx, rel)
		if err != nil {
			uc.logger.Error("failed to create relation with mirror", "error", err)
			return nil, err
		}
		output.Mirror = mirror
	} else {
		if err := uc.relationRepo.Create(ctx, rel); err != nil {
			uc.logger.Error("failed to create relation", "error", err)
			return nil, err
		}
	}

	uc.logger.Info("relation created", "relation_id", rel.ID, "source_id", rel.SourceID, "target_id", rel.TargetID)

	// Enqueue embedding ingestion (all relations created via API are confirmed)
	uc.enqueueIngestion(ctx, rel.TenantID, rel.ID)

	return output, nil
}

// enqueueIngestion enqueues a relation for embedding ingestion
func (uc *CreateRelationUseCase) enqueueIngestion(ctx context.Context, tenantID, relationID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "relation", relationID); err != nil {
		uc.logger.Error("failed to enqueue relation ingestion", "error", err, "relation_id", relationID, "tenant_id", tenantID)
	}
}
