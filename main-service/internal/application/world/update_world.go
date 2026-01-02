package world

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateWorldUseCase handles world updates
type UpdateWorldUseCase struct {
	worldRepo    repositories.WorldRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewUpdateWorldUseCase creates a new UpdateWorldUseCase
func NewUpdateWorldUseCase(
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdateWorldUseCase {
	return &UpdateWorldUseCase{
		worldRepo:    worldRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// UpdateWorldInput represents the input for updating a world
type UpdateWorldInput struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	Genre       *string
	IsImplicit  *bool
}

// UpdateWorldOutput represents the output of updating a world
type UpdateWorldOutput struct {
	World *world.World
}

// Execute updates a world
func (uc *UpdateWorldUseCase) Execute(ctx context.Context, input UpdateWorldInput) (*UpdateWorldOutput, error) {
	w, err := uc.worldRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		if err := w.UpdateName(*input.Name); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "name",
				Message: err.Error(),
			}
		}
	}
	if input.Description != nil {
		w.UpdateDescription(*input.Description)
	}
	if input.Genre != nil {
		w.UpdateGenre(*input.Genre)
	}
	if input.IsImplicit != nil {
		w.SetImplicit(*input.IsImplicit)
	}

	if err := w.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "world",
			Message: err.Error(),
		}
	}

	if err := uc.worldRepo.Update(ctx, w); err != nil {
		uc.logger.Error("failed to update world", "error", err, "world_id", input.ID)
		return nil, err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		w.TenantID,
		nil,
		audit.ActionUpdate,
		audit.EntityTypeWorld,
		w.ID,
		map[string]interface{}{
			"name":        w.Name,
			"is_implicit": w.IsImplicit,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("world updated", "world_id", w.ID, "name", w.Name)

	return &UpdateWorldOutput{
		World: w,
	}, nil
}


