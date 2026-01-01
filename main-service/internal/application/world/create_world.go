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

// CreateWorldUseCase handles world creation
type CreateWorldUseCase struct {
	worldRepo    repositories.WorldRepository
	tenantRepo   repositories.TenantRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewCreateWorldUseCase creates a new CreateWorldUseCase
func NewCreateWorldUseCase(
	worldRepo repositories.WorldRepository,
	tenantRepo repositories.TenantRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateWorldUseCase {
	return &CreateWorldUseCase{
		worldRepo:    worldRepo,
		tenantRepo:   tenantRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// CreateWorldInput represents the input for creating a world
type CreateWorldInput struct {
	TenantID    uuid.UUID
	Name        string
	Description string
	Genre       string
	IsImplicit  bool
}

// CreateWorldOutput represents the output of creating a world
type CreateWorldOutput struct {
	World *world.World
}

// Execute creates a new world
func (uc *CreateWorldUseCase) Execute(ctx context.Context, input CreateWorldInput) (*CreateWorldOutput, error) {
	// Validate tenant exists
	_, err := uc.tenantRepo.GetByID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}

	// Create world
	newWorld, err := world.NewWorld(input.TenantID, input.Name, input.IsImplicit)
	if err != nil {
		return nil, err
	}

	if input.Description != "" {
		newWorld.UpdateDescription(input.Description)
	}
	if input.Genre != "" {
		newWorld.UpdateGenre(input.Genre)
	}

	if err := newWorld.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "world",
			Message: err.Error(),
		}
	}

	if err := uc.worldRepo.Create(ctx, newWorld); err != nil {
		uc.logger.Error("failed to create world", "error", err, "name", input.Name)
		return nil, err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		input.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeWorld,
		newWorld.ID,
		map[string]interface{}{
			"name":        newWorld.Name,
			"is_implicit": newWorld.IsImplicit,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("world created", "world_id", newWorld.ID, "name", newWorld.Name, "tenant_id", input.TenantID)

	return &CreateWorldOutput{
		World: newWorld,
	}, nil
}

