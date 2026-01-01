package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/rpg"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.ArtifactRPGStatsRepository = (*ArtifactRPGStatsRepository)(nil)

// ArtifactRPGStatsRepository implements the artifact RPG stats repository interface
type ArtifactRPGStatsRepository struct {
	db *DB
}

// NewArtifactRPGStatsRepository creates a new artifact RPG stats repository
func NewArtifactRPGStatsRepository(db *DB) *ArtifactRPGStatsRepository {
	return &ArtifactRPGStatsRepository{db: db}
}

// Create creates new artifact RPG stats
func (r *ArtifactRPGStatsRepository) Create(ctx context.Context, stats *rpg.ArtifactRPGStats) error {
	query := `
		INSERT INTO artifact_rpg_stats (id, artifact_id, event_id, stats, is_active, version, reason, timeline, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		stats.ID, stats.ArtifactID, stats.EventID, stats.Stats,
		stats.IsActive, stats.Version, stats.Reason, stats.Timeline,
		stats.CreatedAt)
	return err
}

// GetByID retrieves artifact RPG stats by ID
func (r *ArtifactRPGStatsRepository) GetByID(ctx context.Context, id uuid.UUID) (*rpg.ArtifactRPGStats, error) {
	query := `
		SELECT id, artifact_id, event_id, stats, is_active, version, reason, timeline, created_at
		FROM artifact_rpg_stats
		WHERE id = $1
	`
	var stats rpg.ArtifactRPGStats
	var eventID sql.NullString
	var reason sql.NullString
	var timeline sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&stats.ID, &stats.ArtifactID, &eventID, &stats.Stats,
		&stats.IsActive, &stats.Version, &reason, &timeline,
		&stats.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "artifact_rpg_stats",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	if eventID.Valid {
		parsedID, err := uuid.Parse(eventID.String)
		if err == nil {
			stats.EventID = &parsedID
		}
	}
	if reason.Valid {
		stats.Reason = &reason.String
	}
	if timeline.Valid {
		stats.Timeline = &timeline.String
	}

	return &stats, nil
}

// GetActiveByArtifact retrieves the active stats for an artifact
func (r *ArtifactRPGStatsRepository) GetActiveByArtifact(ctx context.Context, artifactID uuid.UUID) (*rpg.ArtifactRPGStats, error) {
	query := `
		SELECT id, artifact_id, event_id, stats, is_active, version, reason, timeline, created_at
		FROM artifact_rpg_stats
		WHERE artifact_id = $1 AND is_active = TRUE
		ORDER BY version DESC
		LIMIT 1
	`
	var stats rpg.ArtifactRPGStats
	var eventID sql.NullString
	var reason sql.NullString
	var timeline sql.NullString

	err := r.db.QueryRow(ctx, query, artifactID).Scan(
		&stats.ID, &stats.ArtifactID, &eventID, &stats.Stats,
		&stats.IsActive, &stats.Version, &reason, &timeline,
		&stats.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "artifact_rpg_stats",
				ID:       artifactID.String(),
			}
		}
		return nil, err
	}

	if eventID.Valid {
		parsedID, err := uuid.Parse(eventID.String)
		if err == nil {
			stats.EventID = &parsedID
		}
	}
	if reason.Valid {
		stats.Reason = &reason.String
	}
	if timeline.Valid {
		stats.Timeline = &timeline.String
	}

	return &stats, nil
}

// ListByArtifact lists all stats versions for an artifact
func (r *ArtifactRPGStatsRepository) ListByArtifact(ctx context.Context, artifactID uuid.UUID) ([]*rpg.ArtifactRPGStats, error) {
	query := `
		SELECT id, artifact_id, event_id, stats, is_active, version, reason, timeline, created_at
		FROM artifact_rpg_stats
		WHERE artifact_id = $1
		ORDER BY version ASC
	`
	rows, err := r.db.Query(ctx, query, artifactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArtifactRPGStats(rows)
}

// ListByEvent lists all stats versions caused by an event
func (r *ArtifactRPGStatsRepository) ListByEvent(ctx context.Context, eventID uuid.UUID) ([]*rpg.ArtifactRPGStats, error) {
	query := `
		SELECT id, artifact_id, event_id, stats, is_active, version, reason, timeline, created_at
		FROM artifact_rpg_stats
		WHERE event_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArtifactRPGStats(rows)
}

// Update updates artifact RPG stats
func (r *ArtifactRPGStatsRepository) Update(ctx context.Context, stats *rpg.ArtifactRPGStats) error {
	query := `
		UPDATE artifact_rpg_stats
		SET event_id = $2, stats = $3, is_active = $4, version = $5, reason = $6, timeline = $7
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		stats.ID, stats.EventID, stats.Stats,
		stats.IsActive, stats.Version, stats.Reason, stats.Timeline)
	return err
}

// Delete deletes artifact RPG stats
func (r *ArtifactRPGStatsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM artifact_rpg_stats WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByArtifact deletes all stats for an artifact
func (r *ArtifactRPGStatsRepository) DeleteByArtifact(ctx context.Context, artifactID uuid.UUID) error {
	query := `DELETE FROM artifact_rpg_stats WHERE artifact_id = $1`
	_, err := r.db.Exec(ctx, query, artifactID)
	return err
}

// DeactivateAllByArtifact deactivates all stats for an artifact
func (r *ArtifactRPGStatsRepository) DeactivateAllByArtifact(ctx context.Context, artifactID uuid.UUID) error {
	query := `UPDATE artifact_rpg_stats SET is_active = FALSE WHERE artifact_id = $1`
	_, err := r.db.Exec(ctx, query, artifactID)
	return err
}

// GetNextVersion gets the next version number for an artifact
func (r *ArtifactRPGStatsRepository) GetNextVersion(ctx context.Context, artifactID uuid.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(version), 0) + 1 FROM artifact_rpg_stats WHERE artifact_id = $1`
	var version int
	err := r.db.QueryRow(ctx, query, artifactID).Scan(&version)
	return version, err
}

func (r *ArtifactRPGStatsRepository) scanArtifactRPGStats(rows pgx.Rows) ([]*rpg.ArtifactRPGStats, error) {
	statsList := make([]*rpg.ArtifactRPGStats, 0)
	for rows.Next() {
		var stats rpg.ArtifactRPGStats
		var eventID sql.NullString
		var reason sql.NullString
		var timeline sql.NullString

		err := rows.Scan(
			&stats.ID, &stats.ArtifactID, &eventID, &stats.Stats,
			&stats.IsActive, &stats.Version, &reason, &timeline,
			&stats.CreatedAt)
		if err != nil {
			return nil, err
		}

		if eventID.Valid {
			parsedID, err := uuid.Parse(eventID.String)
			if err == nil {
				stats.EventID = &parsedID
			}
		}
		if reason.Valid {
			stats.Reason = &reason.String
		}
		if timeline.Valid {
			stats.Timeline = &timeline.String
		}

		statsList = append(statsList, &stats)
	}
	return statsList, rows.Err()
}

