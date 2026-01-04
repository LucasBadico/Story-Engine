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

var _ repositories.FactionReferenceRepository = (*FactionReferenceRepository)(nil)

// FactionReferenceRepository implements the faction reference repository interface for SQLite
type FactionReferenceRepository struct {
	db *DB
}

// NewFactionReferenceRepository creates a new faction reference repository
func NewFactionReferenceRepository(db *DB) *FactionReferenceRepository {
	return &FactionReferenceRepository{db: db}
}

// Create creates a new faction reference
func (r *FactionReferenceRepository) Create(ctx context.Context, fr *world.FactionReference) error {
	// Get tenant_id from faction
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM factions WHERE id = ?", fr.FactionID.String()).Scan(&tenantIDStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &platformerrors.NotFoundError{
				Resource: "faction",
				ID:       fr.FactionID.String(),
			}
		}
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO faction_references (id, tenant_id, faction_id, entity_type, entity_id, role, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	var role sql.NullString
	if fr.Role != nil {
		role = sql.NullString{String: *fr.Role, Valid: true}
	}

	_, err = r.db.Exec(ctx, query,
		fr.ID.String(),
		tenantID.String(),
		fr.FactionID.String(),
		fr.EntityType,
		fr.EntityID.String(),
		role,
		fr.Notes,
		fr.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a faction reference by ID
func (r *FactionReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.FactionReference, error) {
	query := `
		SELECT id, faction_id, entity_type, entity_id, role, notes, created_at
		FROM faction_references
		WHERE tenant_id = ? AND id = ?
	`
	var fr world.FactionReference
	var idStr, factionIDStr, entityIDStr, createdAtStr string
	var role sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &factionIDStr, &fr.EntityType, &entityIDStr, &role, &fr.Notes, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "faction_reference",
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
	fr.ID = parsedID

	parsedFactionID, err := uuid.Parse(factionIDStr)
	if err != nil {
		return nil, err
	}
	fr.FactionID = parsedFactionID

	parsedEntityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return nil, err
	}
	fr.EntityID = parsedEntityID

	// Parse nullable string
	if role.Valid {
		fr.Role = &role.String
	}

	// Parse timestamp
	fr.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}

	return &fr, nil
}

// ListByFaction lists faction references for a faction
func (r *FactionReferenceRepository) ListByFaction(ctx context.Context, tenantID, factionID uuid.UUID) ([]*world.FactionReference, error) {
	query := `
		SELECT id, faction_id, entity_type, entity_id, role, notes, created_at
		FROM faction_references
		WHERE tenant_id = ? AND faction_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), factionID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanFactionReferences(rows)
}

// ListByEntity lists faction references for an entity
func (r *FactionReferenceRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*world.FactionReference, error) {
	query := `
		SELECT id, faction_id, entity_type, entity_id, role, notes, created_at
		FROM faction_references
		WHERE tenant_id = ? AND entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), entityType, entityID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanFactionReferences(rows)
}

// Update updates a faction reference
func (r *FactionReferenceRepository) Update(ctx context.Context, fr *world.FactionReference) error {
	// Get tenant_id from faction
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM factions WHERE id = ?", fr.FactionID.String()).Scan(&tenantIDStr); err != nil {
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		UPDATE faction_references
		SET role = ?, notes = ?
		WHERE tenant_id = ? AND id = ?
	`

	var role sql.NullString
	if fr.Role != nil {
		role = sql.NullString{String: *fr.Role, Valid: true}
	}

	result, err := r.db.Exec(ctx, query, role, fr.Notes, tenantID.String(), fr.ID.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return &platformerrors.NotFoundError{
			Resource: "faction_reference",
			ID:       fr.ID.String(),
		}
	}
	return nil
}

// Delete deletes a faction reference
func (r *FactionReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM faction_references WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByFactionAndEntity deletes a specific faction reference
func (r *FactionReferenceRepository) DeleteByFactionAndEntity(ctx context.Context, tenantID, factionID uuid.UUID, entityType string, entityID uuid.UUID) error {
	query := `DELETE FROM faction_references WHERE tenant_id = ? AND faction_id = ? AND entity_type = ? AND entity_id = ?`
	result, err := r.db.Exec(ctx, query, tenantID.String(), factionID.String(), entityType, entityID.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return &platformerrors.NotFoundError{
			Resource: "faction_reference",
			ID:       factionID.String() + "/" + entityType + "/" + entityID.String(),
		}
	}
	return nil
}

// DeleteByFaction deletes all faction references for a faction
func (r *FactionReferenceRepository) DeleteByFaction(ctx context.Context, tenantID, factionID uuid.UUID) error {
	query := `DELETE FROM faction_references WHERE tenant_id = ? AND faction_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), factionID.String())
	return err
}

func (r *FactionReferenceRepository) scanFactionReferences(rows *sql.Rows) ([]*world.FactionReference, error) {
	references := make([]*world.FactionReference, 0)
	for rows.Next() {
		var fr world.FactionReference
		var idStr, factionIDStr, entityIDStr, createdAtStr string
		var role sql.NullString

		err := rows.Scan(
			&idStr, &factionIDStr, &fr.EntityType, &entityIDStr, &role, &fr.Notes, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		fr.ID = parsedID

		parsedFactionID, err := uuid.Parse(factionIDStr)
		if err != nil {
			return nil, err
		}
		fr.FactionID = parsedFactionID

		parsedEntityID, err := uuid.Parse(entityIDStr)
		if err != nil {
			return nil, err
		}
		fr.EntityID = parsedEntityID

		// Parse nullable string
		if role.Valid {
			fr.Role = &role.String
		}

		// Parse timestamp
		fr.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		references = append(references, &fr)
	}

	return references, rows.Err()
}

