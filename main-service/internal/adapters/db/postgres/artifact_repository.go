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

var _ repositories.ArtifactRepository = (*ArtifactRepository)(nil)

// ArtifactRepository implements the artifact repository interface
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
		INSERT INTO artifacts (id, world_id, character_id, location_id, name, description, rarity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		a.ID, a.WorldID, a.CharacterID, a.LocationID, a.Name, a.Description, a.Rarity, a.CreatedAt, a.UpdatedAt)
	return err
}

// GetByID retrieves an artifact by ID
func (r *ArtifactRepository) GetByID(ctx context.Context, id uuid.UUID) (*world.Artifact, error) {
	query := `
		SELECT id, world_id, character_id, location_id, name, description, rarity, created_at, updated_at
		FROM artifacts
		WHERE id = $1
	`
	var a world.Artifact
	var characterID *uuid.UUID
	var locationID *uuid.UUID

	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.WorldID, &characterID, &locationID, &a.Name, &a.Description, &a.Rarity, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "artifact",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	a.CharacterID = characterID
	a.LocationID = locationID
	return &a, nil
}

// ListByWorld lists artifacts for a world
func (r *ArtifactRepository) ListByWorld(ctx context.Context, worldID uuid.UUID, limit, offset int) ([]*world.Artifact, error) {
	query := `
		SELECT id, world_id, character_id, location_id, name, description, rarity, created_at, updated_at
		FROM artifacts
		WHERE world_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, worldID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArtifacts(rows)
}

// ListByCharacter lists artifacts for a character
func (r *ArtifactRepository) ListByCharacter(ctx context.Context, characterID uuid.UUID) ([]*world.Artifact, error) {
	query := `
		SELECT id, world_id, character_id, location_id, name, description, rarity, created_at, updated_at
		FROM artifacts
		WHERE character_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArtifacts(rows)
}

// ListByLocation lists artifacts for a location
func (r *ArtifactRepository) ListByLocation(ctx context.Context, locationID uuid.UUID) ([]*world.Artifact, error) {
	query := `
		SELECT id, world_id, character_id, location_id, name, description, rarity, created_at, updated_at
		FROM artifacts
		WHERE location_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, locationID)
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
		SET name = $2, description = $3, rarity = $4, character_id = $5, location_id = $6, updated_at = $7
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, a.ID, a.Name, a.Description, a.Rarity, a.CharacterID, a.LocationID, a.UpdatedAt)
	return err
}

// Delete deletes an artifact
func (r *ArtifactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM artifacts WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CountByWorld counts artifacts for a world
func (r *ArtifactRepository) CountByWorld(ctx context.Context, worldID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM artifacts WHERE world_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, worldID).Scan(&count)
	return count, err
}

func (r *ArtifactRepository) scanArtifacts(rows pgx.Rows) ([]*world.Artifact, error) {
	artifacts := make([]*world.Artifact, 0)
	for rows.Next() {
		var a world.Artifact
		var characterID *uuid.UUID
		var locationID *uuid.UUID

		err := rows.Scan(
			&a.ID, &a.WorldID, &characterID, &locationID, &a.Name, &a.Description, &a.Rarity, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}

		a.CharacterID = characterID
		a.LocationID = locationID
		artifacts = append(artifacts, &a)
	}

	return artifacts, rows.Err()
}

