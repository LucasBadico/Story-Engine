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

var _ repositories.LoreReferenceRepository = (*LoreReferenceRepository)(nil)

// LoreReferenceRepository implements the lore reference repository interface for SQLite
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
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM lores WHERE id = ?", lr.LoreID.String()).Scan(&tenantIDStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &platformerrors.NotFoundError{
				Resource: "lore",
				ID:       lr.LoreID.String(),
			}
		}
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO lore_references (id, tenant_id, lore_id, entity_type, entity_id, relationship_type, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	var relationshipType sql.NullString
	if lr.RelationshipType != nil {
		relationshipType = sql.NullString{String: *lr.RelationshipType, Valid: true}
	}

	_, err = r.db.Exec(ctx, query,
		lr.ID.String(),
		tenantID.String(),
		lr.LoreID.String(),
		lr.EntityType,
		lr.EntityID.String(),
		relationshipType,
		lr.Notes,
		lr.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a lore reference by ID
func (r *LoreReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.LoreReference, error) {
	query := `
		SELECT id, lore_id, entity_type, entity_id, relationship_type, notes, created_at
		FROM lore_references
		WHERE tenant_id = ? AND id = ?
	`
	var lr world.LoreReference
	var idStr, loreIDStr, entityIDStr, createdAtStr string
	var relationshipType sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &loreIDStr, &lr.EntityType, &entityIDStr, &relationshipType, &lr.Notes, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "lore_reference",
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
	lr.ID = parsedID

	parsedLoreID, err := uuid.Parse(loreIDStr)
	if err != nil {
		return nil, err
	}
	lr.LoreID = parsedLoreID

	parsedEntityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return nil, err
	}
	lr.EntityID = parsedEntityID

	// Parse nullable string
	if relationshipType.Valid {
		lr.RelationshipType = &relationshipType.String
	}

	// Parse timestamp
	lr.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}

	return &lr, nil
}

// ListByLore lists lore references for a lore
func (r *LoreReferenceRepository) ListByLore(ctx context.Context, tenantID, loreID uuid.UUID) ([]*world.LoreReference, error) {
	query := `
		SELECT id, lore_id, entity_type, entity_id, relationship_type, notes, created_at
		FROM lore_references
		WHERE tenant_id = ? AND lore_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), loreID.String())
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
		WHERE tenant_id = ? AND entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), entityType, entityID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLoreReferences(rows)
}

// Update updates a lore reference
func (r *LoreReferenceRepository) Update(ctx context.Context, lr *world.LoreReference) error {
	// Get tenant_id from lore
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM lores WHERE id = ?", lr.LoreID.String()).Scan(&tenantIDStr); err != nil {
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		UPDATE lore_references
		SET relationship_type = ?, notes = ?
		WHERE tenant_id = ? AND id = ?
	`

	var relationshipType sql.NullString
	if lr.RelationshipType != nil {
		relationshipType = sql.NullString{String: *lr.RelationshipType, Valid: true}
	}

	result, err := r.db.Exec(ctx, query, relationshipType, lr.Notes, tenantID.String(), lr.ID.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return &platformerrors.NotFoundError{
			Resource: "lore_reference",
			ID:       lr.ID.String(),
		}
	}
	return nil
}

// Delete deletes a lore reference
func (r *LoreReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM lore_references WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByLoreAndEntity deletes a specific lore reference
func (r *LoreReferenceRepository) DeleteByLoreAndEntity(ctx context.Context, tenantID, loreID uuid.UUID, entityType string, entityID uuid.UUID) error {
	query := `DELETE FROM lore_references WHERE tenant_id = ? AND lore_id = ? AND entity_type = ? AND entity_id = ?`
	result, err := r.db.Exec(ctx, query, tenantID.String(), loreID.String(), entityType, entityID.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return &platformerrors.NotFoundError{
			Resource: "lore_reference",
			ID:       loreID.String() + "/" + entityType + "/" + entityID.String(),
		}
	}
	return nil
}

// DeleteByLore deletes all lore references for a lore
func (r *LoreReferenceRepository) DeleteByLore(ctx context.Context, tenantID, loreID uuid.UUID) error {
	query := `DELETE FROM lore_references WHERE tenant_id = ? AND lore_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), loreID.String())
	return err
}

func (r *LoreReferenceRepository) scanLoreReferences(rows *sql.Rows) ([]*world.LoreReference, error) {
	references := make([]*world.LoreReference, 0)
	for rows.Next() {
		var lr world.LoreReference
		var idStr, loreIDStr, entityIDStr, createdAtStr string
		var relationshipType sql.NullString

		err := rows.Scan(
			&idStr, &loreIDStr, &lr.EntityType, &entityIDStr, &relationshipType, &lr.Notes, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		lr.ID = parsedID

		parsedLoreID, err := uuid.Parse(loreIDStr)
		if err != nil {
			return nil, err
		}
		lr.LoreID = parsedLoreID

		parsedEntityID, err := uuid.Parse(entityIDStr)
		if err != nil {
			return nil, err
		}
		lr.EntityID = parsedEntityID

		// Parse nullable string
		if relationshipType.Valid {
			lr.RelationshipType = &relationshipType.String
		}

		// Parse timestamp
		lr.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		references = append(references, &lr)
	}

	return references, rows.Err()
}

