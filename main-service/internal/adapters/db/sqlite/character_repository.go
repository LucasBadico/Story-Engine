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

var _ repositories.CharacterRepository = (*CharacterRepository)(nil)

// CharacterRepository implements the character repository interface for SQLite
type CharacterRepository struct {
	db *DB
}

// NewCharacterRepository creates a new character repository
func NewCharacterRepository(db *DB) *CharacterRepository {
	return &CharacterRepository{db: db}
}

// Create creates a new character
func (r *CharacterRepository) Create(ctx context.Context, c *world.Character) error {
	query := `
		INSERT INTO characters (id, tenant_id, world_id, archetype_id, current_class_id, class_level, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var archetypeID sql.NullString
	if c.ArchetypeID != nil {
		archetypeID = sql.NullString{String: c.ArchetypeID.String(), Valid: true}
	}

	var currentClassID sql.NullString
	if c.CurrentClassID != nil {
		currentClassID = sql.NullString{String: c.CurrentClassID.String(), Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		c.ID.String(),
		c.TenantID.String(),
		c.WorldID.String(),
		archetypeID,
		currentClassID,
		c.ClassLevel,
		c.Name,
		c.Description,
		c.CreatedAt.Format(time.RFC3339),
		c.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves a character by ID
func (r *CharacterRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Character, error) {
	query := `
		SELECT id, tenant_id, world_id, archetype_id, current_class_id, class_level, name, description, created_at, updated_at
		FROM characters
		WHERE tenant_id = ? AND id = ?
	`
	var c world.Character
	var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
	var archetypeID, currentClassID sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &worldIDStr, &archetypeID, &currentClassID, &c.ClassLevel, &c.Name, &c.Description, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character",
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
	c.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	c.TenantID = parsedTenantID

	parsedWorldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		return nil, err
	}
	c.WorldID = parsedWorldID

	// Parse timestamps
	c.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	c.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable UUIDs
	if archetypeID.Valid {
		if parsedArchetypeID, err := uuid.Parse(archetypeID.String); err == nil {
			c.ArchetypeID = &parsedArchetypeID
		}
	}
	if currentClassID.Valid {
		if parsedCurrentClassID, err := uuid.Parse(currentClassID.String); err == nil {
			c.CurrentClassID = &parsedCurrentClassID
		}
	}

	return &c, nil
}

// ListByWorld lists characters for a world
func (r *CharacterRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID, limit, offset int) ([]*world.Character, error) {
	query := `
		SELECT id, tenant_id, world_id, archetype_id, current_class_id, class_level, name, description, created_at, updated_at
		FROM characters
		WHERE tenant_id = ? AND world_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), worldID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacters(rows)
}

// Update updates a character
func (r *CharacterRepository) Update(ctx context.Context, c *world.Character) error {
	query := `
		UPDATE characters
		SET name = ?, description = ?, archetype_id = ?, current_class_id = ?, class_level = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	var archetypeID sql.NullString
	if c.ArchetypeID != nil {
		archetypeID = sql.NullString{String: c.ArchetypeID.String(), Valid: true}
	}

	var currentClassID sql.NullString
	if c.CurrentClassID != nil {
		currentClassID = sql.NullString{String: c.CurrentClassID.String(), Valid: true}
	}

	_, err := r.db.Exec(ctx, query,
		c.Name,
		c.Description,
		archetypeID,
		currentClassID,
		c.ClassLevel,
		c.UpdatedAt.Format(time.RFC3339),
		c.TenantID.String(),
		c.ID.String(),
	)
	return err
}

// Delete deletes a character
func (r *CharacterRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM characters WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// CountByWorld counts characters for a world
func (r *CharacterRepository) CountByWorld(ctx context.Context, tenantID, worldID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM characters WHERE tenant_id = ? AND world_id = ?`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID.String(), worldID.String()).Scan(&count)
	return count, err
}

func (r *CharacterRepository) scanCharacters(rows *sql.Rows) ([]*world.Character, error) {
	characters := make([]*world.Character, 0)
	for rows.Next() {
		var c world.Character
		var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
		var archetypeID, currentClassID sql.NullString

		err := rows.Scan(
			&idStr, &tenantIDStr, &worldIDStr, &archetypeID, &currentClassID, &c.ClassLevel, &c.Name, &c.Description, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		c.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		c.TenantID = parsedTenantID

		parsedWorldID, err := uuid.Parse(worldIDStr)
		if err != nil {
			return nil, err
		}
		c.WorldID = parsedWorldID

		// Parse timestamps
		c.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		c.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse nullable UUIDs
		if archetypeID.Valid {
			if parsedArchetypeID, err := uuid.Parse(archetypeID.String); err == nil {
				c.ArchetypeID = &parsedArchetypeID
			}
		}
		if currentClassID.Valid {
			if parsedCurrentClassID, err := uuid.Parse(currentClassID.String); err == nil {
				c.CurrentClassID = &parsedCurrentClassID
			}
		}

		characters = append(characters, &c)
	}

	return characters, rows.Err()
}

