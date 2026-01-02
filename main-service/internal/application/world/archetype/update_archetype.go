package archetype

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateArchetypeUseCase handles archetype updates
type UpdateArchetypeUseCase struct {
	archetypeRepo repositories.ArchetypeRepository
	auditLogRepo  repositories.AuditLogRepository
	logger        logger.Logger
}

// NewUpdateArchetypeUseCase creates a new UpdateArchetypeUseCase
func NewUpdateArchetypeUseCase(
	archetypeRepo repositories.ArchetypeRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdateArchetypeUseCase {
	return &UpdateArchetypeUseCase{
		archetypeRepo: archetypeRepo,
		auditLogRepo:  auditLogRepo,
		logger:        logger,
	}
}

// UpdateArchetypeInput represents the input for updating an archetype
type UpdateArchetypeInput struct {
	ID          uuid.UUID
	Name        *string
	Description *string
}

// UpdateArchetypeOutput represents the output of updating an archetype
type UpdateArchetypeOutput struct {
	Archetype *world.Archetype
}

// Execute updates an archetype
func (uc *UpdateArchetypeUseCase) Execute(ctx context.Context, input UpdateArchetypeInput) (*UpdateArchetypeOutput, error) {
	a, err := uc.archetypeRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		if err := a.UpdateName(*input.Name); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "name",
				Message: err.Error(),
			}
		}
	}
	if input.Description != nil {
		a.UpdateDescription(*input.Description)
	}

	if err := a.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "archetype",
			Message: err.Error(),
		}
	}

	if err := uc.archetypeRepo.Update(ctx, a); err != nil {
		uc.logger.Error("failed to update archetype", "error", err, "archetype_id", input.ID)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		a.TenantID,
		nil,
		audit.ActionUpdate,
		audit.EntityTypeArchetype,
		a.ID,
		map[string]interface{}{
			"name": a.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("archetype updated", "archetype_id", a.ID, "name", a.Name)

	return &UpdateArchetypeOutput{
		Archetype: a,
	}, nil
}


