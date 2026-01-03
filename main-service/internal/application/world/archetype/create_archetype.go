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

// CreateArchetypeUseCase handles archetype creation
type CreateArchetypeUseCase struct {
	archetypeRepo repositories.ArchetypeRepository
	tenantRepo    repositories.TenantRepository
	auditLogRepo  repositories.AuditLogRepository
	logger        logger.Logger
}

// NewCreateArchetypeUseCase creates a new CreateArchetypeUseCase
func NewCreateArchetypeUseCase(
	archetypeRepo repositories.ArchetypeRepository,
	tenantRepo repositories.TenantRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateArchetypeUseCase {
	return &CreateArchetypeUseCase{
		archetypeRepo: archetypeRepo,
		tenantRepo:    tenantRepo,
		auditLogRepo:  auditLogRepo,
		logger:        logger,
	}
}

// CreateArchetypeInput represents the input for creating an archetype
type CreateArchetypeInput struct {
	TenantID    uuid.UUID
	Name        string
	Description string
}

// CreateArchetypeOutput represents the output of creating an archetype
type CreateArchetypeOutput struct {
	Archetype *world.Archetype
}

// Execute creates a new archetype
func (uc *CreateArchetypeUseCase) Execute(ctx context.Context, input CreateArchetypeInput) (*CreateArchetypeOutput, error) {
	_, err := uc.tenantRepo.GetByID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	newArchetype, err := world.NewArchetype(input.TenantID, input.Name)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "name",
			Message: err.Error(),
		}
	}

	if input.Description != "" {
		newArchetype.UpdateDescription(input.Description)
	}

	if err := newArchetype.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "archetype",
			Message: err.Error(),
		}
	}

	if err := uc.archetypeRepo.Create(ctx, newArchetype); err != nil {
		uc.logger.Error("failed to create archetype", "error", err, "name", input.Name)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeArchetype,
		newArchetype.ID,
		map[string]interface{}{
			"name": newArchetype.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("archetype created", "archetype_id", newArchetype.ID, "name", newArchetype.Name)

	return &CreateArchetypeOutput{
		Archetype: newArchetype,
	}, nil
}


