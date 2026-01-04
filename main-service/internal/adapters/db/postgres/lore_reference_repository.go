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

var _ repositories.LoreReferenceRepository = (*LoreReferenceRepository)(nil)

// LoreReferenceRepository implements the lore reference repository interface
type LoreReferenceRepository struct {
	db *DB
}

// NewLoreReferenceRepository creates a new lore reference repository
func NewLoreReferenceRepository(db *DB) *LoreReferenceRepository {
	return &LoreReferenceRepository{db: db}
}

// Create creates a new lore reference
func (r *LoreReferenceRepository) Create(ctx context.Context, lr *world.LoreReference) error {
	// Get tenant_id from lore
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM lores WHERE id = $1", lr.LoreID).Scan(&tenantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &platformerrors.NotFoundError{
				Resource: "lore",
				ID:       lr.LoreID.String(),
			}
		}
		return err
	}

	query := `
		INSERT INTO lore_references (id, tenant_id, lore_id, entity_type, entity_id, relationship_type, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (lore_id, entity_type, entity_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query,
		lr.ID, tenantID, lr.LoreID, lr.EntityType, lr.EntityID, lr.RelationshipType, lr.Notes, lr.CreatedAt)
	return err
}

// GetByID retrieves a lore reference by ID
func (r *LoreReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.LoreReference, error) {
	query := `
		SELECT id, lore_id, entity_type, entity_id, relationship_type, notes, created_at
		FROM lore_references
		WHERE tenant_id = $1 AND id = $2
	`
	var lr world.LoreReference

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&lr.ID, &lr.LoreID, &lr.EntityType, &lr.EntityID, &lr.RelationshipType, &lr.Notes, &lr.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "lore_reference",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &lr, nil
}

// ListByLore lists lore references for a lore
func (r *LoreReferenceRepository) ListByLore(ctx context.Context, tenantID, loreID uuid.UUID) ([]*world.LoreReference, error) {
	query := `
		SELECT id, lore_id, entity_type, entity_id, relationship_type, notes, created_at
		FROM lore_references
		WHERE tenant_id = $1 AND lore_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, loreID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLoreReferences(rows)
}

// ListByEntity lists lore references for an entity
func (r *LoreReferenceRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*world.LoreReference, error) {
	query := `
		SELECT id, lore_id, entity_type, entity_id, relationship_type, notes, created_at
		FROM lore_references
		WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLoreReferences(rows)
}

// Update updates a lore reference
func (r *LoreReferenceRepository) Update(ctx context.Context, lr *world.LoreReference) error {
	// Get tenant_id from lore
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM lores WHERE id = $1", lr.LoreID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		UPDATE lore_references
		SET relationship_type = $2, notes = $3
		WHERE tenant_id = $4 AND id = $1
	`
	result, err := r.db.Exec(ctx, query, lr.ID, lr.RelationshipType, lr.Notes, tenantID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &platformerrors.NotFoundError{
			Resource: "lore_reference",
			ID:       lr.ID.String(),
		}
	}
	return nil
}

// Delete deletes a lore reference
func (r *LoreReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM lore_references WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByLoreAndEntity deletes a specific lore reference
func (r *LoreReferenceRepository) DeleteByLoreAndEntity(ctx context.Context, tenantID, loreID uuid.UUID, entityType string, entityID uuid.UUID) error {
	query := `DELETE FROM lore_references WHERE tenant_id = $1 AND lore_id = $2 AND entity_type = $3 AND entity_id = $4`
	result, err := r.db.Exec(ctx, query, tenantID, loreID, entityType, entityID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &platformerrors.NotFoundError{
			Resource: "lore_reference",
			ID:       loreID.String() + "/" + entityType + "/" + entityID.String(),
		}
	}
	return nil
}

// DeleteByLore deletes all lore references for a lore
func (r *LoreReferenceRepository) DeleteByLore(ctx context.Context, tenantID, loreID uuid.UUID) error {
	query := `DELETE FROM lore_references WHERE tenant_id = $1 AND lore_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, loreID)
	return err
}

func (r *LoreReferenceRepository) scanLoreReferences(rows pgx.Rows) ([]*world.LoreReference, error) {
	references := make([]*world.LoreReference, 0)
	for rows.Next() {
		var lr world.LoreReference
		err := rows.Scan(
			&lr.ID, &lr.LoreID, &lr.EntityType, &lr.EntityID, &lr.RelationshipType, &lr.Notes, &lr.CreatedAt)
		if err != nil {
			return nil, err
		}
		references = append(references, &lr)
	}
	return references, rows.Err()
}

