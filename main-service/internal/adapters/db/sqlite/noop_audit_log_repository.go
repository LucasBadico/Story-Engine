package sqlite

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.AuditLogRepository = (*NoopAuditLogRepository)(nil)

// NoopAuditLogRepository is a no-op implementation of AuditLogRepository for offline mode
// It implements all methods but does nothing (no persistence)
type NoopAuditLogRepository struct{}

// NewNoopAuditLogRepository creates a new no-op audit log repository
func NewNoopAuditLogRepository() *NoopAuditLogRepository {
	return &NoopAuditLogRepository{}
}

// Create does nothing (no-op)
func (r *NoopAuditLogRepository) Create(ctx context.Context, log *audit.AuditLog) error {
	return nil
}

// GetByID returns an error (not supported in no-op mode)
func (r *NoopAuditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*audit.AuditLog, error) {
	return nil, errors.New("audit log not found")
}

// ListByTenant returns empty list
func (r *NoopAuditLogRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*audit.AuditLog, error) {
	return []*audit.AuditLog{}, nil
}

// ListByEntity returns empty list
func (r *NoopAuditLogRepository) ListByEntity(ctx context.Context, entityType audit.EntityType, entityID uuid.UUID, limit, offset int) ([]*audit.AuditLog, error) {
	return []*audit.AuditLog{}, nil
}

// ListByActor returns empty list
func (r *NoopAuditLogRepository) ListByActor(ctx context.Context, actorUserID uuid.UUID, limit, offset int) ([]*audit.AuditLog, error) {
	return []*audit.AuditLog{}, nil
}

