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

var _ repositories.TraitRepository = (*TraitRepository)(nil)

// TraitRepository implements the trait repository interface
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
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		t.ID, t.TenantID, t.Name, t.Category, t.Description, t.CreatedAt, t.UpdatedAt)
	return err
}

// GetByID retrieves a trait by ID
func (r *TraitRepository) GetByID(ctx context.Context, id uuid.UUID) (*world.Trait, error) {
	query := `
		SELECT id, tenant_id, name, category, description, created_at, updated_at
		FROM traits
		WHERE id = $1
	`
	var t world.Trait

	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.TenantID, &t.Name, &t.Category, &t.Description, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "trait",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &t, nil
}

// ListByTenant lists traits for a tenant
func (r *TraitRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*world.Trait, error) {
	query := `
		SELECT id, tenant_id, name, category, description, created_at, updated_at
		FROM traits
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
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
		SET name = $2, category = $3, description = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, t.ID, t.Name, t.Category, t.Description, t.UpdatedAt)
	return err
}

// Delete deletes a trait
func (r *TraitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM traits WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CountByTenant counts traits for a tenant
func (r *TraitRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM traits WHERE tenant_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID).Scan(&count)
	return count, err
}

func (r *TraitRepository) scanTraits(rows pgx.Rows) ([]*world.Trait, error) {
	traits := make([]*world.Trait, 0)
	for rows.Next() {
		var t world.Trait

		err := rows.Scan(
			&t.ID, &t.TenantID, &t.Name, &t.Category, &t.Description, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}

		traits = append(traits, &t)
	}

	return traits, rows.Err()
}

