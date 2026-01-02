package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.ArchetypeTraitRepository = (*ArchetypeTraitRepository)(nil)

// ArchetypeTraitRepository implements the archetype-trait repository interface
type ArchetypeTraitRepository struct {
	db *DB
}

// NewArchetypeTraitRepository creates a new archetype-trait repository
func NewArchetypeTraitRepository(db *DB) *ArchetypeTraitRepository {
	return &ArchetypeTraitRepository{db: db}
}

// Create creates a new archetype-trait relationship
func (r *ArchetypeTraitRepository) Create(ctx context.Context, at *world.ArchetypeTrait) error {
	// Get tenant_id from archetype
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM archetypes WHERE id = $1", at.ArchetypeID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		INSERT INTO archetype_traits (id, tenant_id, archetype_id, trait_id, default_value, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		at.ID, tenantID, at.ArchetypeID, at.TraitID, at.DefaultValue, at.CreatedAt)
	return err
}

// GetByArchetype retrieves all traits for an archetype
func (r *ArchetypeTraitRepository) GetByArchetype(ctx context.Context, tenantID, archetypeID uuid.UUID) ([]*world.ArchetypeTrait, error) {
	query := `
		SELECT id, archetype_id, trait_id, default_value, created_at
		FROM archetype_traits
		WHERE tenant_id = $1 AND archetype_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, archetypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArchetypeTraits(rows)
}

// GetByTrait retrieves all archetypes that use a trait
func (r *ArchetypeTraitRepository) GetByTrait(ctx context.Context, tenantID, traitID uuid.UUID) ([]*world.ArchetypeTrait, error) {
	query := `
		SELECT id, archetype_id, trait_id, default_value, created_at
		FROM archetype_traits
		WHERE tenant_id = $1 AND trait_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, traitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArchetypeTraits(rows)
}

// Delete deletes an archetype-trait relationship
func (r *ArchetypeTraitRepository) Delete(ctx context.Context, tenantID, archetypeID, traitID uuid.UUID) error {
	query := `DELETE FROM archetype_traits WHERE tenant_id = $1 AND archetype_id = $2 AND trait_id = $3`
	result, err := r.db.Exec(ctx, query, tenantID, archetypeID, traitID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &platformerrors.NotFoundError{
			Resource: "archetype_trait",
			ID:       archetypeID.String() + "/" + traitID.String(),
		}
	}
	return nil
}

// DeleteByArchetype deletes all traits for an archetype
func (r *ArchetypeTraitRepository) DeleteByArchetype(ctx context.Context, tenantID, archetypeID uuid.UUID) error {
	query := `DELETE FROM archetype_traits WHERE tenant_id = $1 AND archetype_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, archetypeID)
	return err
}

func (r *ArchetypeTraitRepository) scanArchetypeTraits(rows pgx.Rows) ([]*world.ArchetypeTrait, error) {
	archetypeTraits := make([]*world.ArchetypeTrait, 0)
	for rows.Next() {
		var at world.ArchetypeTrait

		err := rows.Scan(
			&at.ID, &at.ArchetypeID, &at.TraitID, &at.DefaultValue, &at.CreatedAt)
		if err != nil {
			return nil, err
		}

		archetypeTraits = append(archetypeTraits, &at)
	}

	return archetypeTraits, rows.Err()
}

