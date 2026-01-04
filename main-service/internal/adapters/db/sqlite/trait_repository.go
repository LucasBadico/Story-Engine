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

var _ repositories.TraitRepository = (*TraitRepository)(nil)

// TraitRepository implements the trait repository interface for SQLite
type TraitRepository struct {
	db *DB
}

// NewTraitRepository creates a new trait repository
func NewTraitRepository(db *DB) *TraitRepository {
	return &TraitRepository{db: db}
}

// Create creates a new trait
func (r *TraitRepository) Create(ctx context.Context, t *world.Trait) error {
	query := `
		INSERT INTO traits (id, tenant_id, name, category, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(ctx, query,
		t.ID.String(),
		t.TenantID.String(),
		t.Name,
		t.Category,
		t.Description,
		t.CreatedAt.Format(time.RFC3339),
		t.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a trait by ID
func (r *TraitRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Trait, error) {
	query := `
		SELECT id, tenant_id, name, category, description, created_at, updated_at
		FROM traits
		WHERE tenant_id = ? AND id = ?
	`
	var t world.Trait
	var idStr, tenantIDStr, createdAtStr, updatedAtStr string

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &t.Name, &t.Category, &t.Description, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "trait",
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
	t.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	t.TenantID = parsedTenantID

	// Parse timestamps
	t.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	t.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// ListByTenant lists traits for a tenant
func (r *TraitRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*world.Trait, error) {
	query := `
		SELECT id, tenant_id, name, category, description, created_at, updated_at
		FROM traits
		WHERE tenant_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTraits(rows)
}

// Update updates a trait
func (r *TraitRepository) Update(ctx context.Context, t *world.Trait) error {
	query := `
		UPDATE traits
		SET name = ?, category = ?, description = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`
	_, err := r.db.Exec(ctx, query,
		t.Name,
		t.Category,
		t.Description,
		t.UpdatedAt.Format(time.RFC3339),
		t.TenantID.String(),
		t.ID.String(),
	)
	return err
}

// Delete deletes a trait
func (r *TraitRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM traits WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// CountByTenant counts traits for a tenant
func (r *TraitRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM traits WHERE tenant_id = ?`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID.String()).Scan(&count)
	return count, err
}

func (r *TraitRepository) scanTraits(rows *sql.Rows) ([]*world.Trait, error) {
	traits := make([]*world.Trait, 0)
	for rows.Next() {
		var t world.Trait
		var idStr, tenantIDStr, createdAtStr, updatedAtStr string

		err := rows.Scan(
			&idStr, &tenantIDStr, &t.Name, &t.Category, &t.Description, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		t.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		t.TenantID = parsedTenantID

		// Parse timestamps
		t.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		t.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		traits = append(traits, &t)
	}

	return traits, rows.Err()
}

