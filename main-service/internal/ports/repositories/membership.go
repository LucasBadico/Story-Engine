package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/auth"
)

// MembershipRepository defines the interface for membership persistence
type MembershipRepository interface {
	Create(ctx context.Context, m *auth.Membership) error
	GetByID(ctx context.Context, id uuid.UUID) (*auth.Membership, error)
	GetByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (*auth.Membership, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*auth.Membership, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*auth.Membership, error)
	Update(ctx context.Context, m *auth.Membership) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountOwnersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
}

