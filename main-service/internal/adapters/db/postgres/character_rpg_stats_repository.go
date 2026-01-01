package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/rpg"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.CharacterRPGStatsRepository = (*CharacterRPGStatsRepository)(nil)

// CharacterRPGStatsRepository implements the character RPG stats repository interface
type CharacterRPGStatsRepository struct {
	db *DB
}

// NewCharacterRPGStatsRepository creates a new character RPG stats repository
func NewCharacterRPGStatsRepository(db *DB) *CharacterRPGStatsRepository {
	return &CharacterRPGStatsRepository{db: db}
}

// Create creates new character RPG stats
func (r *CharacterRPGStatsRepository) Create(ctx context.Context, stats *rpg.CharacterRPGStats) error {
	query := `
		INSERT INTO character_rpg_stats (id, character_id, event_id, base_stats, derived_stats, progression, is_active, version, reason, timeline, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		stats.ID, stats.CharacterID, stats.EventID,
		stats.BaseStats, stats.DerivedStats, stats.Progression,
		stats.IsActive, stats.Version, stats.Reason, stats.Timeline,
		stats.CreatedAt)
	return err
}

// GetByID retrieves character RPG stats by ID
func (r *CharacterRPGStatsRepository) GetByID(ctx context.Context, id uuid.UUID) (*rpg.CharacterRPGStats, error) {
	query := `
		SELECT id, character_id, event_id, base_stats, derived_stats, progression, is_active, version, reason, timeline, created_at
		FROM character_rpg_stats
		WHERE id = $1
	`
	var stats rpg.CharacterRPGStats
	var eventID sql.NullString
	var derivedStats sql.NullString
	var progression sql.NullString
	var reason sql.NullString
	var timeline sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&stats.ID, &stats.CharacterID, &eventID,
		&stats.BaseStats, &derivedStats, &progression,
		&stats.IsActive, &stats.Version, &reason, &timeline,
		&stats.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_rpg_stats",
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
	if derivedStats.Valid {
		derived := json.RawMessage(derivedStats.String)
		stats.DerivedStats = &derived
	}
	if progression.Valid {
		prog := json.RawMessage(progression.String)
		stats.Progression = &prog
	}
	if reason.Valid {
		stats.Reason = &reason.String
	}
	if timeline.Valid {
		stats.Timeline = &timeline.String
	}

	return &stats, nil
}

// GetActiveByCharacter retrieves the active stats for a character
func (r *CharacterRPGStatsRepository) GetActiveByCharacter(ctx context.Context, characterID uuid.UUID) (*rpg.CharacterRPGStats, error) {
	query := `
		SELECT id, character_id, event_id, base_stats, derived_stats, progression, is_active, version, reason, timeline, created_at
		FROM character_rpg_stats
		WHERE character_id = $1 AND is_active = TRUE
		ORDER BY version DESC
		LIMIT 1
	`
	var stats rpg.CharacterRPGStats
	var eventID sql.NullString
	var derivedStats sql.NullString
	var progression sql.NullString
	var reason sql.NullString
	var timeline sql.NullString

	err := r.db.QueryRow(ctx, query, characterID).Scan(
		&stats.ID, &stats.CharacterID, &eventID,
		&stats.BaseStats, &derivedStats, &progression,
		&stats.IsActive, &stats.Version, &reason, &timeline,
		&stats.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_rpg_stats",
				ID:       characterID.String(),
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
	if derivedStats.Valid {
		derived := json.RawMessage(derivedStats.String)
		stats.DerivedStats = &derived
	}
	if progression.Valid {
		prog := json.RawMessage(progression.String)
		stats.Progression = &prog
	}
	if reason.Valid {
		stats.Reason = &reason.String
	}
	if timeline.Valid {
		stats.Timeline = &timeline.String
	}

	return &stats, nil
}

// ListByCharacter lists all stats versions for a character
func (r *CharacterRPGStatsRepository) ListByCharacter(ctx context.Context, characterID uuid.UUID) ([]*rpg.CharacterRPGStats, error) {
	query := `
		SELECT id, character_id, event_id, base_stats, derived_stats, progression, is_active, version, reason, timeline, created_at
		FROM character_rpg_stats
		WHERE character_id = $1
		ORDER BY version ASC
	`
	rows, err := r.db.Query(ctx, query, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacterRPGStats(rows)
}

// ListByEvent lists all stats versions caused by an event
func (r *CharacterRPGStatsRepository) ListByEvent(ctx context.Context, eventID uuid.UUID) ([]*rpg.CharacterRPGStats, error) {
	query := `
		SELECT id, character_id, event_id, base_stats, derived_stats, progression, is_active, version, reason, timeline, created_at
		FROM character_rpg_stats
		WHERE event_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacterRPGStats(rows)
}

// Update updates character RPG stats
func (r *CharacterRPGStatsRepository) Update(ctx context.Context, stats *rpg.CharacterRPGStats) error {
	query := `
		UPDATE character_rpg_stats
		SET event_id = $2, base_stats = $3, derived_stats = $4, progression = $5, is_active = $6, version = $7, reason = $8, timeline = $9
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		stats.ID, stats.EventID, stats.BaseStats, stats.DerivedStats, stats.Progression,
		stats.IsActive, stats.Version, stats.Reason, stats.Timeline)
	return err
}

// Delete deletes character RPG stats
func (r *CharacterRPGStatsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM character_rpg_stats WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByCharacter deletes all stats for a character
func (r *CharacterRPGStatsRepository) DeleteByCharacter(ctx context.Context, characterID uuid.UUID) error {
	query := `DELETE FROM character_rpg_stats WHERE character_id = $1`
	_, err := r.db.Exec(ctx, query, characterID)
	return err
}

// DeactivateAllByCharacter deactivates all stats for a character
func (r *CharacterRPGStatsRepository) DeactivateAllByCharacter(ctx context.Context, characterID uuid.UUID) error {
	query := `UPDATE character_rpg_stats SET is_active = FALSE WHERE character_id = $1`
	_, err := r.db.Exec(ctx, query, characterID)
	return err
}

// GetNextVersion gets the next version number for a character
func (r *CharacterRPGStatsRepository) GetNextVersion(ctx context.Context, characterID uuid.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(version), 0) + 1 FROM character_rpg_stats WHERE character_id = $1`
	var version int
	err := r.db.QueryRow(ctx, query, characterID).Scan(&version)
	return version, err
}

func (r *CharacterRPGStatsRepository) scanCharacterRPGStats(rows pgx.Rows) ([]*rpg.CharacterRPGStats, error) {
	statsList := make([]*rpg.CharacterRPGStats, 0)
	for rows.Next() {
		var stats rpg.CharacterRPGStats
		var eventID sql.NullString
		var derivedStats sql.NullString
		var progression sql.NullString
		var reason sql.NullString
		var timeline sql.NullString

		err := rows.Scan(
			&stats.ID, &stats.CharacterID, &eventID,
			&stats.BaseStats, &derivedStats, &progression,
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
		if derivedStats.Valid {
			derived := json.RawMessage(derivedStats.String)
			stats.DerivedStats = &derived
		}
		if progression.Valid {
			prog := json.RawMessage(progression.String)
			stats.Progression = &prog
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

