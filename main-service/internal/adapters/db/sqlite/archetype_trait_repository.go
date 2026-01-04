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

var _ repositories.ArchetypeTraitRepository = (*ArchetypeTraitRepository)(nil)

// ArchetypeTraitRepository implements the archetype-trait repository interface for SQLite
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
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM archetypes WHERE id = ?", at.ArchetypeID.String()).Scan(&tenantIDStr); err != nil {
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO archetype_traits (id, tenant_id, archetype_id, trait_id, default_value, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.Exec(ctx, query,
		at.ID.String(),
		tenantID.String(),
		at.ArchetypeID.String(),
		at.TraitID.String(),
		at.DefaultValue,
		at.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByArchetype retrieves all traits for an archetype
func (r *ArchetypeTraitRepository) GetByArchetype(ctx context.Context, tenantID, archetypeID uuid.UUID) ([]*world.ArchetypeTrait, error) {
	query := `
		SELECT id, archetype_id, trait_id, default_value, created_at
		FROM archetype_traits
		WHERE tenant_id = ? AND archetype_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), archetypeID.String())
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
		WHERE tenant_id = ? AND trait_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), traitID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArchetypeTraits(rows)
}

// Delete deletes an archetype-trait relationship
func (r *ArchetypeTraitRepository) Delete(ctx context.Context, tenantID, archetypeID, traitID uuid.UUID) error {
	query := `DELETE FROM archetype_traits WHERE tenant_id = ? AND archetype_id = ? AND trait_id = ?`
	result, err := r.db.Exec(ctx, query, tenantID.String(), archetypeID.String(), traitID.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return &platformerrors.NotFoundError{
			Resource: "archetype_trait",
			ID:       archetypeID.String() + "/" + traitID.String(),
		}
	}
	return nil
}

// DeleteByArchetype deletes all traits for an archetype
func (r *ArchetypeTraitRepository) DeleteByArchetype(ctx context.Context, tenantID, archetypeID uuid.UUID) error {
	query := `DELETE FROM archetype_traits WHERE tenant_id = ? AND archetype_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), archetypeID.String())
	return err
}

func (r *ArchetypeTraitRepository) scanArchetypeTraits(rows *sql.Rows) ([]*world.ArchetypeTrait, error) {
	archetypeTraits := make([]*world.ArchetypeTrait, 0)
	for rows.Next() {
		var at world.ArchetypeTrait
		var idStr, archetypeIDStr, traitIDStr, createdAtStr string

		err := rows.Scan(
			&idStr, &archetypeIDStr, &traitIDStr, &at.DefaultValue, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		at.ID = parsedID

		parsedArchetypeID, err := uuid.Parse(archetypeIDStr)
		if err != nil {
			return nil, err
		}
		at.ArchetypeID = parsedArchetypeID

		parsedTraitID, err := uuid.Parse(traitIDStr)
		if err != nil {
			return nil, err
		}
		at.TraitID = parsedTraitID

		// Parse timestamp
		at.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		archetypeTraits = append(archetypeTraits, &at)
	}

	return archetypeTraits, rows.Err()
}

