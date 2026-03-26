package character_state

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

// CreateCharacterStateUseCase handles character state creation
type CreateCharacterStateUseCase struct {
	characterStateRepo repositories.CharacterStateRepository
	characterRepo      repositories.CharacterRepository
	auditLogRepo       repositories.AuditLogRepository
	ingestionQueue     queue.IngestionQueue
	logger             logger.Logger
}

// NewCreateCharacterStateUseCase creates a new CreateCharacterStateUseCase
func NewCreateCharacterStateUseCase(
	characterStateRepo repositories.CharacterStateRepository,
	characterRepo repositories.CharacterRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateCharacterStateUseCase {
	return &CreateCharacterStateUseCase{
		characterStateRepo: characterStateRepo,
		characterRepo:      characterRepo,
		auditLogRepo:       auditLogRepo,
		ingestionQueue:     nil,
		logger:             logger,
	}
}

func (uc *CreateCharacterStateUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
}

// CreateCharacterStateInput represents the input for creating a character state
type CreateCharacterStateInput struct {
	TenantID           uuid.UUID
	CharacterID        uuid.UUID
	ChapterID          *uuid.UUID
	ChapterNumber      int
	Emotion            string
	Goal               string
	Secrets            []string
	Clothes            string
	PhysicalState      string
	TimelineDay        *int
	TimelineConfidence *string
}

// CreateCharacterStateOutput represents the output of creating a character state
type CreateCharacterStateOutput struct {
	CharacterState *world.CharacterState
}

// Execute creates a new character state
func (uc *CreateCharacterStateUseCase) Execute(ctx context.Context, input CreateCharacterStateInput) (*CreateCharacterStateOutput, error) {
	// Validate character exists
	_, err := uc.characterRepo.GetByID(ctx, input.TenantID, input.CharacterID)
	if err != nil {
		return nil, err
	}

	newState, err := world.NewCharacterState(input.TenantID, input.CharacterID, input.ChapterNumber)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "character_state",
			Message: err.Error(),
		}
	}

	if input.ChapterID != nil {
		newState.ChapterID = input.ChapterID
	}
	if input.Emotion != "" {
		newState.SetEmotion(input.Emotion)
	}
	if input.Goal != "" {
		newState.SetGoal(input.Goal)
	}
	if input.Secrets != nil {
		newState.SetSecrets(input.Secrets)
	}
	if input.Clothes != "" {
		newState.SetClothes(input.Clothes)
	}
	if input.PhysicalState != "" {
		newState.SetPhysicalState(input.PhysicalState)
	}
	if input.TimelineDay != nil && input.TimelineConfidence != nil {
		newState.SetTimeline(*input.TimelineDay, *input.TimelineConfidence)
	}

	if err := newState.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "character_state",
			Message: err.Error(),
		}
	}

	if err := uc.characterStateRepo.Create(ctx, newState); err != nil {
		uc.logger.Error("failed to create character state", "error", err, "character_id", input.CharacterID)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeCharacterState,
		newState.ID,
		map[string]interface{}{
			"character_id":   input.CharacterID.String(),
			"chapter_number": input.ChapterNumber,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("character state created", "character_state_id", newState.ID, "character_id", input.CharacterID)
	uc.enqueueIngestion(ctx, input.TenantID, newState.ID)

	return &CreateCharacterStateOutput{
		CharacterState: newState,
	}, nil
}

func (uc *CreateCharacterStateUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, characterStateID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "character_state", characterStateID); err != nil {
		uc.logger.Error("failed to enqueue character state ingestion", "error", err, "character_state_id", characterStateID, "tenant_id", tenantID)
	}
}
