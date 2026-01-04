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

var _ repositories.FactionRepository = (*FactionRepository)(nil)

// FactionRepository implements the faction repository interface for SQLite
type FactionRepository struct {
	db *DB
}

// NewFactionRepository creates a new faction repository
func NewFactionRepository(db *DB) *FactionRepository {
	return &FactionRepository{db: db}
}

// Create creates a new faction
func (r *FactionRepository) Create(ctx context.Context, f *world.Faction) error {
	// If parent_id is provided, get parent's level to calculate hierarchy_level
	if f.ParentID != nil {
		parent, err := r.GetByID(ctx, f.TenantID, *f.ParentID)
		if err != nil {
			return err
		}
		f.SetHierarchyLevel(parent.HierarchyLevel + 1)
	}

	query := `
		INSERT INTO factions (id, tenant_id, world_id, parent_id, name, type, description, beliefs, structure, symbols, hierarchy_level, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var parentID sql.NullString
	if f.ParentID != nil {
		parentID = sql.NullString{String: f.ParentID.String(), Valid: true}
	}

	var factionType sql.NullString
	if f.Type != nil {
		factionType = sql.NullString{String: *f.Type, Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		f.ID.String(),
		f.TenantID.String(),
		f.WorldID.String(),
		parentID,
		f.Name,
		factionType,
		f.Description,
		f.Beliefs,
		f.Structure,
		f.Symbols,
		f.HierarchyLevel,
		f.CreatedAt.Format(time.RFC3339),
		f.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a faction by ID
func (r *FactionRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Faction, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, beliefs, structure, symbols, hierarchy_level, created_at, updated_at
		FROM factions
		WHERE tenant_id = ? AND id = ?
	`
	var f world.Faction
	var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
	var parentID, factionType sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &worldIDStr, &parentID, &f.Name, &factionType, &f.Description, &f.Beliefs, &f.Structure, &f.Symbols, &f.HierarchyLevel, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "faction",
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
	f.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	f.TenantID = parsedTenantID

	parsedWorldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		return nil, err
	}
	f.WorldID = parsedWorldID

	// Parse timestamps
	f.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	f.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable UUID
	if parentID.Valid {
		if parsedParentID, err := uuid.Parse(parentID.String); err == nil {
			f.ParentID = &parsedParentID
		}
	}

	// Parse nullable string
	if factionType.Valid {
		f.Type = &factionType.String
	}

	return &f, nil
}

// ListByWorld lists factions for a world
func (r *FactionRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID) ([]*world.Faction, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, beliefs, structure, symbols, hierarchy_level, created_at, updated_at
		FROM factions
		WHERE tenant_id = ? AND world_id = ?
		ORDER BY hierarchy_level ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), worldID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanFactions(rows)
}

// Update updates a faction
func (r *FactionRepository) Update(ctx context.Context, f *world.Faction) error {
	query := `
		UPDATE factions
		SET name = ?, type = ?, description = ?, beliefs = ?, structure = ?, symbols = ?, parent_id = ?, hierarchy_level = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	var parentID sql.NullString
	if f.ParentID != nil {
		parentID = sql.NullString{String: f.ParentID.String(), Valid: true}
	}

	var factionType sql.NullString
	if f.Type != nil {
		factionType = sql.NullString{String: *f.Type, Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		f.Name,
		factionType,
		f.Description,
		f.Beliefs,
		f.Structure,
		f.Symbols,
		parentID,
		f.HierarchyLevel,
		f.UpdatedAt.Format(time.RFC3339),
		f.TenantID.String(),
		f.ID.String(),
	)
	return err
}

// Delete deletes a faction
func (r *FactionRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM factions WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// GetChildren retrieves direct children of a faction
func (r *FactionRepository) GetChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]*world.Faction, error) {
	query := `
		SELECT id, tenant_id, world_id, parent_id, name, type, description, beliefs, structure, symbols, hierarchy_level, created_at, updated_at
		FROM factions
		WHERE tenant_id = ? AND parent_id = ?
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), parentID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanFactions(rows)
}

func (r *FactionRepository) scanFactions(rows *sql.Rows) ([]*world.Faction, error) {
	factions := make([]*world.Faction, 0)
	for rows.Next() {
		var f world.Faction
		var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
		var parentID, factionType sql.NullString

		err := rows.Scan(
			&idStr, &tenantIDStr, &worldIDStr, &parentID, &f.Name, &factionType, &f.Description, &f.Beliefs, &f.Structure, &f.Symbols, &f.HierarchyLevel, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		f.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		f.TenantID = parsedTenantID

		parsedWorldID, err := uuid.Parse(worldIDStr)
		if err != nil {
			return nil, err
		}
		f.WorldID = parsedWorldID

		// Parse timestamps
		f.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		f.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse nullable UUID
		if parentID.Valid {
			if parsedParentID, err := uuid.Parse(parentID.String); err == nil {
				f.ParentID = &parsedParentID
			}
		}

		// Parse nullable string
		if factionType.Valid {
			f.Type = &factionType.String
		}

		factions = append(factions, &f)
	}

	return factions, rows.Err()
}

