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

// CreateTraitUseCase handles trait creation
type CreateTraitUseCase struct {
	traitRepo    repositories.TraitRepository
	tenantRepo   repositories.TenantRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewCreateTraitUseCase creates a new CreateTraitUseCase
func NewCreateTraitUseCase(
	traitRepo repositories.TraitRepository,
	tenantRepo repositories.TenantRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateTraitUseCase {
	return &CreateTraitUseCase{
		traitRepo:    traitRepo,
		tenantRepo:   tenantRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// CreateTraitInput represents the input for creating a trait
type CreateTraitInput struct {
	TenantID    uuid.UUID
	Name        string
	Category    string
	Description string
}

// CreateTraitOutput represents the output of creating a trait
type CreateTraitOutput struct {
	Trait *world.Trait
}

// Execute creates a new trait
func (uc *CreateTraitUseCase) Execute(ctx context.Context, input CreateTraitInput) (*CreateTraitOutput, error) {
	_, err := uc.tenantRepo.GetByID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	newTrait, err := world.NewTrait(input.TenantID, input.Name)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "name",
			Message: err.Error(),
		}
	}

	if input.Category != "" {
		newTrait.UpdateCategory(input.Category)
	}
	if input.Description != "" {
		newTrait.UpdateDescription(input.Description)
	}

	if err := newTrait.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "trait",
			Message: err.Error(),
		}
	}

	if err := uc.traitRepo.Create(ctx, newTrait); err != nil {
		uc.logger.Error("failed to create trait", "error", err, "name", input.Name)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeTrait,
		newTrait.ID,
		map[string]interface{}{
			"name":     newTrait.Name,
			"category": newTrait.Category,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("trait created", "trait_id", newTrait.ID, "name", newTrait.Name)

	return &CreateTraitOutput{
		Trait: newTrait,
	}, nil
}


