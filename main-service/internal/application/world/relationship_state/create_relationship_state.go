package relationship_state

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// EntityLinkInput represents an entity link to be created
type EntityLinkInput struct {
	Role       string
	EntityType string
	EntityID   uuid.UUID
}

// CreateRelationshipStateUseCase handles relationship state creation
type CreateRelationshipStateUseCase struct {
	relationshipStateRepo repositories.RelationshipStateRepository
	auditLogRepo          repositories.AuditLogRepository
	ingestionQueue        queue.IngestionQueue
	logger                logger.Logger
}

// NewCreateRelationshipStateUseCase creates a new CreateRelationshipStateUseCase
func NewCreateRelationshipStateUseCase(
	relationshipStateRepo repositories.RelationshipStateRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateRelationshipStateUseCase {
	return &CreateRelationshipStateUseCase{
		relationshipStateRepo: relationshipStateRepo,
		auditLogRepo:          auditLogRepo,
		ingestionQueue:        nil,
		logger:                logger,
	}
}

func (uc *CreateRelationshipStateUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
}

// CreateRelationshipStateInput represents the input for creating a relationship state
type CreateRelationshipStateInput struct {
	TenantID           uuid.UUID
	ChapterID          uuid.UUID
	ChapterNumber      int
	TensionDescription string
	TimelineDay        *int
	TimelineConfidence *string
	Entities           []EntityLinkInput
}

// CreateRelationshipStateOutput represents the output of creating a relationship state
type CreateRelationshipStateOutput struct {
	RelationshipState *world.RelationshipState
}

// Execute creates a new relationship state
func (uc *CreateRelationshipStateUseCase) Execute(ctx context.Context, input CreateRelationshipStateInput) (*CreateRelationshipStateOutput, error) {
	newState, err := world.NewRelationshipState(input.TenantID, input.ChapterID, input.ChapterNumber, input.TensionDescription)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "relationship_state",
			Message: err.Error(),
		}
	}

	if input.TimelineDay != nil && input.TimelineConfidence != nil {
		newState.SetTimeline(*input.TimelineDay, *input.TimelineConfidence)
	}

	// Add entity links
	for _, e := range input.Entities {
		newState.AddEntity(e.Role, e.EntityType, e.EntityID)
	}

	if err := newState.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "relationship_state",
			Message: err.Error(),
		}
	}

	if err := uc.relationshipStateRepo.Create(ctx, newState); err != nil {
		uc.logger.Error("failed to create relationship state", "error", err)
		return nil, err
	}

	// Create entity links in DB
	for _, entity := range newState.Entities {
		if err := uc.relationshipStateRepo.AddEntity(ctx, &entity); err != nil {
			uc.logger.Error("failed to add entity to relationship state", "error", err, "entity_id", entity.EntityID)
		}
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeRelationshipState,
		newState.ID,
		map[string]interface{}{
			"chapter_number": input.ChapterNumber,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("relationship state created", "relationship_state_id", newState.ID)
	uc.enqueueIngestion(ctx, input.TenantID, newState.ID)

	return &CreateRelationshipStateOutput{
		RelationshipState: newState,
	}, nil
}

func (uc *CreateRelationshipStateUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, relationshipStateID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "relationship_state", relationshipStateID); err != nil {
		uc.logger.Error("failed to enqueue relationship state ingestion", "error", err, "relationship_state_id", relationshipStateID, "tenant_id", tenantID)
	}
}
