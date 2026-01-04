package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/tenant"
	"github.com/story-engine/main-service/internal/ports/repositories"
	"github.com/story-engine/main-service/internal/platform/logger"
)

const (
	// DefaultTenantID is the fixed UUID for the default tenant in offline mode
	DefaultTenantID = "00000000-0000-0000-0000-000000000001"
	// DefaultTenantName is the name for the default tenant
	DefaultTenantName = "default"
)

// SetupDefaultTenant creates the default tenant if it doesn't exist
func SetupDefaultTenant(ctx context.Context, tenantRepo repositories.TenantRepository, log logger.Logger) (uuid.UUID, error) {
	defaultTenantUUID, err := uuid.Parse(DefaultTenantID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid default tenant ID: %w", err)
	}

	// Try to get existing tenant by ID
	existingTenant, err := tenantRepo.GetByID(ctx, defaultTenantUUID)
	if err == nil && existingTenant != nil {
		log.Info("default tenant already exists", "tenant_id", defaultTenantUUID)
		return defaultTenantUUID, nil
	}

	// Try to get by name as fallback
	existingTenant, err = tenantRepo.GetByName(ctx, DefaultTenantName)
	if err == nil && existingTenant != nil {
		log.Info("default tenant found by name", "tenant_id", existingTenant.ID, "name", DefaultTenantName)
		return existingTenant.ID, nil
	}

	// Create default tenant
	newTenant, err := tenant.NewTenant(DefaultTenantName, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create tenant entity: %w", err)
	}

	// Use the fixed UUID for consistency
	newTenant.ID = defaultTenantUUID

	if err := newTenant.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("invalid tenant: %w", err)
	}

	if err := tenantRepo.Create(ctx, newTenant); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create default tenant: %w", err)
	}

	log.Info("default tenant created", "tenant_id", defaultTenantUUID, "name", DefaultTenantName)
	return defaultTenantUUID, nil
}

