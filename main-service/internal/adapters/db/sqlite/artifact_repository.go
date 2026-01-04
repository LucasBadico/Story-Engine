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

var _ repositories.ArtifactRepository = (*ArtifactRepository)(nil)

// ArtifactRepository implements the artifact repository interface for SQLite
type ArtifactRepository struct {
	db *DB
}

// NewArtifactRepository creates a new artifact repository
func NewArtifactRepository(db *DB) *ArtifactRepository {
	return &ArtifactRepository{db: db}
}

// Create creates a new artifact
func (r *ArtifactRepository) Create(ctx context.Context, a *world.Artifact) error {
	query := `
		INSERT INTO artifacts (id, tenant_id, world_id, name, description, rarity, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(ctx, query,
		a.ID.String(),
		a.TenantID.String(),
		a.WorldID.String(),
		a.Name,
		a.Description,
		a.Rarity,
		a.CreatedAt.Format(time.RFC3339),
		a.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves an artifact by ID
func (r *ArtifactRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Artifact, error) {
	query := `
		SELECT id, tenant_id, world_id, name, description, rarity, created_at, updated_at
		FROM artifacts
		WHERE tenant_id = ? AND id = ?
	`
	var a world.Artifact
	var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &worldIDStr, &a.Name, &a.Description, &a.Rarity, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "artifact",
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
	a.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	a.TenantID = parsedTenantID

	parsedWorldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		return nil, err
	}
	a.WorldID = parsedWorldID

	// Parse timestamps
	a.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	a.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// ListByWorld lists artifacts for a world
func (r *ArtifactRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID, limit, offset int) ([]*world.Artifact, error) {
	query := `
		SELECT id, tenant_id, world_id, name, description, rarity, created_at, updated_at
		FROM artifacts
		WHERE tenant_id = ? AND world_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), worldID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArtifacts(rows)
}

// Update updates an artifact
func (r *ArtifactRepository) Update(ctx context.Context, a *world.Artifact) error {
	query := `
		UPDATE artifacts
		SET name = ?, description = ?, rarity = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`
	_, err := r.db.Exec(ctx, query,
		a.Name,
		a.Description,
		a.Rarity,
		a.UpdatedAt.Format(time.RFC3339),
		a.TenantID.String(),
		a.ID.String(),
	)
	return err
}

// Delete deletes an artifact
func (r *ArtifactRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM artifacts WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// CountByWorld counts artifacts for a world
func (r *ArtifactRepository) CountByWorld(ctx context.Context, tenantID, worldID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM artifacts WHERE tenant_id = ? AND world_id = ?`
	var count int
	err := r.db.QueryRow(ctx, query, tenantID.String(), worldID.String()).Scan(&count)
	return count, err
}

func (r *ArtifactRepository) scanArtifacts(rows *sql.Rows) ([]*world.Artifact, error) {
	artifacts := make([]*world.Artifact, 0)
	for rows.Next() {
		var a world.Artifact
		var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string

		err := rows.Scan(
			&idStr, &tenantIDStr, &worldIDStr, &a.Name, &a.Description, &a.Rarity, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		a.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		a.TenantID = parsedTenantID

		parsedWorldID, err := uuid.Parse(worldIDStr)
		if err != nil {
			return nil, err
		}
		a.WorldID = parsedWorldID

		// Parse timestamps
		a.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		a.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		artifacts = append(artifacts, &a)
	}

	return artifacts, rows.Err()
}

