package trait

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateTraitUseCase handles trait updates
type UpdateTraitUseCase struct {
	traitRepo    repositories.TraitRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewUpdateTraitUseCase creates a new UpdateTraitUseCase
func NewUpdateTraitUseCase(
	traitRepo repositories.TraitRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdateTraitUseCase {
	return &UpdateTraitUseCase{
		traitRepo:    traitRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// UpdateTraitInput represents the input for updating a trait
type UpdateTraitInput struct {
	ID          uuid.UUID
	Name        *string
	Category    *string
	Description *string
}

// UpdateTraitOutput represents the output of updating a trait
type UpdateTraitOutput struct {
	Trait *world.Trait
}

// Execute updates a trait
func (uc *UpdateTraitUseCase) Execute(ctx context.Context, input UpdateTraitInput) (*UpdateTraitOutput, error) {
	t, err := uc.traitRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		if err := t.UpdateName(*input.Name); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "name",
				Message: err.Error(),
			}
		}
	}
	if input.Category != nil {
		t.UpdateCategory(*input.Category)
	}
	if input.Description != nil {
		t.UpdateDescription(*input.Description)
	}

	if err := t.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "trait",
			Message: err.Error(),
		}
	}

	if err := uc.traitRepo.Update(ctx, t); err != nil {
		uc.logger.Error("failed to update trait", "error", err, "trait_id", input.ID)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		t.TenantID,
		nil,
		audit.ActionUpdate,
		audit.EntityTypeTrait,
		t.ID,
		map[string]interface{}{
			"name": t.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("trait updated", "trait_id", t.ID, "name", t.Name)

	return &UpdateTraitOutput{
		Trait: t,
	}, nil
}


