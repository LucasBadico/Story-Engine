package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
)

// AuditLogRepository defines the interface for audit log persistence
type AuditLogRepository interface {
	Create(ctx context.Context, log *audit.AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*audit.AuditLog, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*audit.AuditLog, error)
	ListByEntity(ctx context.Context, entityType audit.EntityType, entityID uuid.UUID, limit, offset int) ([]*audit.AuditLog, error)
	ListByActor(ctx context.Context, actorUserID uuid.UUID, limit, offset int) ([]*audit.AuditLog, error)
}

