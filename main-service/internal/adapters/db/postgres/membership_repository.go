package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/auth"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.MembershipRepository = (*MembershipRepository)(nil)

// MembershipRepository implements the membership repository interface
type MembershipRepository struct {
	db *DB
}

// NewMembershipRepository creates a new membership repository
func NewMembershipRepository(db *DB) *MembershipRepository {
	return &MembershipRepository{db: db}
}

// Create creates a new membership
func (r *MembershipRepository) Create(ctx context.Context, m *auth.Membership) error {
	query := `
		INSERT INTO memberships (id, tenant_id, user_id, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		m.ID, m.TenantID, m.UserID, string(m.Role), string(m.Status), m.CreatedAt, m.UpdatedAt)
	return err
}

// GetByID retrieves a membership by ID
func (r *MembershipRepository) GetByID(ctx context.Context, id uuid.UUID) (*auth.Membership, error) {
	query := `
		SELECT id, tenant_id, user_id, role, status, created_at, updated_at
		FROM memberships
		WHERE id = $1
	`
	var m auth.Membership

	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.TenantID, &m.UserID, &m.Role, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "membership",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &m, nil
}

// GetByTenantAndUser retrieves a membership by tenant ID and user ID
func (r *MembershipRepository) GetByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (*auth.Membership, error) {
	query := `
		SELECT id, tenant_id, user_id, role, status, created_at, updated_at
		FROM memberships
		WHERE tenant_id = $1 AND user_id = $2
	`
	var m auth.Membership

	err := r.db.QueryRow(ctx, query, tenantID, userID).Scan(
		&m.ID, &m.TenantID, &m.UserID, &m.Role, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "membership",
				ID:       tenantID.String() + "/" + userID.String(),
			}
		}
		return nil, err
	}

	return &m, nil
}

// ListByTenant lists memberships for a tenant
func (r *MembershipRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*auth.Membership, error) {
	query := `
		SELECT id, tenant_id, user_id, role, status, created_at, updated_at
		FROM memberships
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMemberships(rows)
}

// ListByUser lists memberships for a user
func (r *MembershipRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*auth.Membership, error) {
	query := `
		SELECT id, tenant_id, user_id, role, status, created_at, updated_at
		FROM memberships
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMemberships(rows)
}

// Update updates a membership
func (r *MembershipRepository) Update(ctx context.Context, m *auth.Membership) error {
	query := `
		UPDATE memberships
		SET role = $2, status = $3, updated_at = $4
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, m.ID, string(m.Role), string(m.Status), m.UpdatedAt)
	return err
}

// Delete deletes a membership
func (r *MembershipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM memberships WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CountOwnersByTenant counts the number of owner memberships for a tenant
func (r *MembershipRepository) CountOwnersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM memberships
		WHERE tenant_id = $1 AND role = 'owner' AND status = 'active'
	`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID).Scan(&count)
	return count, err
}

// scanMemberships scans rows into a slice of memberships
func (r *MembershipRepository) scanMemberships(rows pgx.Rows) ([]*auth.Membership, error) {
	var memberships []*auth.Membership
	for rows.Next() {
		var m auth.Membership

		err := rows.Scan(&m.ID, &m.TenantID, &m.UserID, &m.Role, &m.Status, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}

		memberships = append(memberships, &m)
	}

	return memberships, rows.Err()
}

