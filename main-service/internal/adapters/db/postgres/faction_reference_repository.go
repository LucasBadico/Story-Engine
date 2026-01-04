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

var _ repositories.FactionReferenceRepository = (*FactionReferenceRepository)(nil)

// FactionReferenceRepository implements the faction reference repository interface
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
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM factions WHERE id = $1", fr.FactionID).Scan(&tenantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &platformerrors.NotFoundError{
				Resource: "faction",
				ID:       fr.FactionID.String(),
			}
		}
		return err
	}

	query := `
		INSERT INTO faction_references (id, tenant_id, faction_id, entity_type, entity_id, role, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (faction_id, entity_type, entity_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query,
		fr.ID, tenantID, fr.FactionID, fr.EntityType, fr.EntityID, fr.Role, fr.Notes, fr.CreatedAt)
	return err
}

// GetByID retrieves a faction reference by ID
func (r *FactionReferenceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.FactionReference, error) {
	query := `
		SELECT id, faction_id, entity_type, entity_id, role, notes, created_at
		FROM faction_references
		WHERE tenant_id = $1 AND id = $2
	`
	var fr world.FactionReference

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&fr.ID, &fr.FactionID, &fr.EntityType, &fr.EntityID, &fr.Role, &fr.Notes, &fr.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "faction_reference",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &fr, nil
}

// ListByFaction lists faction references for a faction
func (r *FactionReferenceRepository) ListByFaction(ctx context.Context, tenantID, factionID uuid.UUID) ([]*world.FactionReference, error) {
	query := `
		SELECT id, faction_id, entity_type, entity_id, role, notes, created_at
		FROM faction_references
		WHERE tenant_id = $1 AND faction_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, factionID)
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
		WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanFactionReferences(rows)
}

// Update updates a faction reference
func (r *FactionReferenceRepository) Update(ctx context.Context, fr *world.FactionReference) error {
	// Get tenant_id from faction
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM factions WHERE id = $1", fr.FactionID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		UPDATE faction_references
		SET role = $2, notes = $3
		WHERE tenant_id = $4 AND id = $1
	`
	result, err := r.db.Exec(ctx, query, fr.ID, fr.Role, fr.Notes, tenantID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &platformerrors.NotFoundError{
			Resource: "faction_reference",
			ID:       fr.ID.String(),
		}
	}
	return nil
}

// Delete deletes a faction reference
func (r *FactionReferenceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM faction_references WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByFactionAndEntity deletes a specific faction reference
func (r *FactionReferenceRepository) DeleteByFactionAndEntity(ctx context.Context, tenantID, factionID uuid.UUID, entityType string, entityID uuid.UUID) error {
	query := `DELETE FROM faction_references WHERE tenant_id = $1 AND faction_id = $2 AND entity_type = $3 AND entity_id = $4`
	result, err := r.db.Exec(ctx, query, tenantID, factionID, entityType, entityID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &platformerrors.NotFoundError{
			Resource: "faction_reference",
			ID:       factionID.String() + "/" + entityType + "/" + entityID.String(),
		}
	}
	return nil
}

// DeleteByFaction deletes all faction references for a faction
func (r *FactionReferenceRepository) DeleteByFaction(ctx context.Context, tenantID, factionID uuid.UUID) error {
	query := `DELETE FROM faction_references WHERE tenant_id = $1 AND faction_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, factionID)
	return err
}

func (r *FactionReferenceRepository) scanFactionReferences(rows pgx.Rows) ([]*world.FactionReference, error) {
	references := make([]*world.FactionReference, 0)
	for rows.Next() {
		var fr world.FactionReference
		err := rows.Scan(
			&fr.ID, &fr.FactionID, &fr.EntityType, &fr.EntityID, &fr.Role, &fr.Notes, &fr.CreatedAt)
		if err != nil {
			return nil, err
		}
		references = append(references, &fr)
	}
	return references, rows.Err()
}

