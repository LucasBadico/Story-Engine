package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.ArchetypeRepository = (*ArchetypeRepository)(nil)

// ArchetypeRepository implements the archetype repository interface
type ArchetypeRepository struct {
	db *DB
}

// NewArchetypeRepository creates a new archetype repository
func NewArchetypeRepository(db *DB) *ArchetypeRepository {
	return &ArchetypeRepository{db: db}
}

// Create creates a new archetype
func (r *ArchetypeRepository) Create(ctx context.Context, a *world.Archetype) error {
	query := `
		INSERT INTO archetypes (id, tenant_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		a.ID, a.TenantID, a.Name, a.Description, a.CreatedAt, a.UpdatedAt)
	return err
}

// GetByID retrieves an archetype by ID
func (r *ArchetypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*world.Archetype, error) {
	query := `
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM archetypes
		WHERE id = $1
	`
	var a world.Archetype

	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.TenantID, &a.Name, &a.Description, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "archetype",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &a, nil
}

// ListByTenant lists archetypes for a tenant
func (r *ArchetypeRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*world.Archetype, error) {
	query := `
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM archetypes
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArchetypes(rows)
}

// Update updates an archetype
func (r *ArchetypeRepository) Update(ctx context.Context, a *world.Archetype) error {
	query := `
		UPDATE archetypes
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, a.ID, a.Name, a.Description, a.UpdatedAt)
	return err
}

// Delete deletes an archetype
func (r *ArchetypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM archetypes WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CountByTenant counts archetypes for a tenant
func (r *ArchetypeRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM archetypes WHERE tenant_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID).Scan(&count)
	return count, err
}

func (r *ArchetypeRepository) scanArchetypes(rows pgx.Rows) ([]*world.Archetype, error) {
	archetypes := make([]*world.Archetype, 0)
	for rows.Next() {
		var a world.Archetype

		err := rows.Scan(
			&a.ID, &a.TenantID, &a.Name, &a.Description, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}

		archetypes = append(archetypes, &a)
	}

	return archetypes, rows.Err()
}

