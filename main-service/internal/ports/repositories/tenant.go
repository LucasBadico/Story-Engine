package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/tenant"
)

// TenantRepository defines the interface for tenant persistence
type TenantRepository interface {
	Create(ctx context.Context, t *tenant.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error)
	GetByName(ctx context.Context, name string) (*tenant.Tenant, error)
	Update(ctx context.Context, t *tenant.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*tenant.Tenant, error)
}

