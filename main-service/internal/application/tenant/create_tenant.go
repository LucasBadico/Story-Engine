package tenant

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/tenant"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateTenantUseCase handles tenant creation
type CreateTenantUseCase struct {
	tenantRepo    repositories.TenantRepository
	auditLogRepo  repositories.AuditLogRepository
	logger        logger.Logger
}

// NewCreateTenantUseCase creates a new CreateTenantUseCase
func NewCreateTenantUseCase(
	tenantRepo repositories.TenantRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateTenantUseCase {
	return &CreateTenantUseCase{
		tenantRepo:   tenantRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// CreateTenantInput represents the input for creating a tenant
type CreateTenantInput struct {
	Name      string
	CreatedBy *uuid.UUID
}

// CreateTenantOutput represents the output of creating a tenant
type CreateTenantOutput struct {
	Tenant *tenant.Tenant
}

// Execute creates a new tenant
func (uc *CreateTenantUseCase) Execute(ctx context.Context, input CreateTenantInput) (*CreateTenantOutput, error) {
	// Validate tenant name
	if input.Name == "" {
		return nil, &platformerrors.ValidationError{
			Field:   "name",
			Message: "tenant name is required",
		}
	}

	// Check if tenant with same name already exists
	existing, err := uc.tenantRepo.GetByName(ctx, input.Name)
	if err == nil && existing != nil {
		return nil, &platformerrors.AlreadyExistsError{
			Resource: "tenant",
			Field:    "name",
			Value:    input.Name,
		}
	}

	// Create tenant
	newTenant, err := tenant.NewTenant(input.Name, input.CreatedBy)
	if err != nil {
		return nil, err
	}

	if err := newTenant.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "tenant",
			Message: err.Error(),
		}
	}

	if err := uc.tenantRepo.Create(ctx, newTenant); err != nil {
		uc.logger.Error("failed to create tenant", "error", err, "name", input.Name)
		return nil, err
	}

	// Log audit event
	auditLog := audit.NewAuditLog(
		newTenant.ID,
		input.CreatedBy,
		audit.ActionCreate,
		audit.EntityTypeTenant,
		newTenant.ID,
		map[string]interface{}{
			"name": newTenant.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
		// Don't fail the operation if audit logging fails
	}

	uc.logger.Info("tenant created", "tenant_id", newTenant.ID, "name", newTenant.Name)

	return &CreateTenantOutput{
		Tenant: newTenant,
	}, nil
}

