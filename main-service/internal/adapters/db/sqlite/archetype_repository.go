package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.ArchetypeRepository = (*ArchetypeRepository)(nil)

// ArchetypeRepository implements the archetype repository interface for SQLite
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
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(ctx, query,
		a.ID.String(),
		a.TenantID.String(),
		a.Name,
		a.Description,
		a.CreatedAt.Format(time.RFC3339),
		a.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves an archetype by ID
func (r *ArchetypeRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Archetype, error) {
	query := `
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM archetypes
		WHERE tenant_id = ? AND id = ?
	`
	var a world.Archetype
	var idStr, tenantIDStr, createdAtStr, updatedAtStr string

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &a.Name, &a.Description, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "archetype",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	// Parse UUIDs
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	a.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	a.TenantID = parsedTenantID

	// Parse timestamps
	a.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	a.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// ListByTenant lists archetypes for a tenant
func (r *ArchetypeRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*world.Archetype, error) {
	query := `
		SELECT id, tenant_id, name, description, created_at, updated_at
		FROM archetypes
		WHERE tenant_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), limit, offset)
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
		SET name = ?, description = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`
	_, err := r.db.Exec(ctx, query,
		a.Name,
		a.Description,
		a.UpdatedAt.Format(time.RFC3339),
		a.TenantID.String(),
		a.ID.String(),
	)
	return err
}

// Delete deletes an archetype
func (r *ArchetypeRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM archetypes WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// CountByTenant counts archetypes for a tenant
func (r *ArchetypeRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM archetypes WHERE tenant_id = ?`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID.String()).Scan(&count)
	return count, err
}

func (r *ArchetypeRepository) scanArchetypes(rows *sql.Rows) ([]*world.Archetype, error) {
	archetypes := make([]*world.Archetype, 0)
	for rows.Next() {
		var a world.Archetype
		var idStr, tenantIDStr, createdAtStr, updatedAtStr string

		err := rows.Scan(
			&idStr, &tenantIDStr, &a.Name, &a.Description, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		a.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		a.TenantID = parsedTenantID

		// Parse timestamps
		a.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		a.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		archetypes = append(archetypes, &a)
	}

	return archetypes, rows.Err()
}

