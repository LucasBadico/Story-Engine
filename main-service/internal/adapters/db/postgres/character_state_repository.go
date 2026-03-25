package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.CharacterStateRepository = (*CharacterStateRepository)(nil)

// CharacterStateRepository implements the character state repository interface
type CharacterStateRepository struct {
	db *DB
}

// NewCharacterStateRepository creates a new character state repository
func NewCharacterStateRepository(db *DB) *CharacterStateRepository {
	return &CharacterStateRepository{db: db}
}

// Create creates a new character state
func (r *CharacterStateRepository) Create(ctx context.Context, cs *world.CharacterState) error {
	metadataJSON, err := json.Marshal(cs.Metadata)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO character_states (
			id, tenant_id, character_id, chapter_id, chapter_number,
			emotion, goal, secrets, clothes, physical_state,
			timeline_day, timeline_confidence, metadata, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err = r.db.Exec(ctx, query,
		cs.ID, cs.TenantID, cs.CharacterID, cs.ChapterID, cs.ChapterNumber,
		cs.Emotion, cs.Goal, cs.Secrets, cs.Clothes, cs.PhysicalState,
		cs.TimelineDay, cs.TimelineConfidence, metadataJSON, cs.CreatedAt, cs.UpdatedAt)
	return err
}

// GetByID retrieves a character state by ID
func (r *CharacterStateRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.CharacterState, error) {
	query := `
		SELECT id, tenant_id, character_id, chapter_id, chapter_number,
			emotion, goal, secrets, clothes, physical_state,
			timeline_day, timeline_confidence, metadata, created_at, updated_at
		FROM character_states
		WHERE tenant_id = $1 AND id = $2
	`
	cs, err := r.scanCharacterState(r.db.QueryRow(ctx, query, tenantID, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_state",
				ID:       id.String(),
			}
		}
		return nil, err
	}
	return cs, nil
}

// GetLatestByCharacter retrieves the latest character state for a character as of a given chapter
func (r *CharacterStateRepository) GetLatestByCharacter(ctx context.Context, tenantID, characterID uuid.UUID, asOfChapter int) (*world.CharacterState, error) {
	query := `
		SELECT id, tenant_id, character_id, chapter_id, chapter_number,
			emotion, goal, secrets, clothes, physical_state,
			timeline_day, timeline_confidence, metadata, created_at, updated_at
		FROM character_states
		WHERE tenant_id = $1 AND character_id = $2 AND chapter_number <= $3
		ORDER BY chapter_number DESC LIMIT 1
	`
	cs, err := r.scanCharacterState(r.db.QueryRow(ctx, query, tenantID, characterID, asOfChapter))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "character_state",
				ID:       characterID.String(),
			}
		}
		return nil, err
	}
	return cs, nil
}

// ListByChapter lists character states for a chapter
func (r *CharacterStateRepository) ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*world.CharacterState, error) {
	query := `
		SELECT id, tenant_id, character_id, chapter_id, chapter_number,
			emotion, goal, secrets, clothes, physical_state,
			timeline_day, timeline_confidence, metadata, created_at, updated_at
		FROM character_states
		WHERE tenant_id = $1 AND chapter_id = $2
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, tenantID, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacterStates(rows)
}

// ListByCharacter lists all character states for a character
func (r *CharacterStateRepository) ListByCharacter(ctx context.Context, tenantID, characterID uuid.UUID) ([]*world.CharacterState, error) {
	query := `
		SELECT id, tenant_id, character_id, chapter_id, chapter_number,
			emotion, goal, secrets, clothes, physical_state,
			timeline_day, timeline_confidence, metadata, created_at, updated_at
		FROM character_states
		WHERE tenant_id = $1 AND character_id = $2
		ORDER BY chapter_number, created_at
	`
	rows, err := r.db.Query(ctx, query, tenantID, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanCharacterStates(rows)
}

// Delete deletes a character state
func (r *CharacterStateRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM character_states WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *CharacterStateRepository) scanCharacterState(row pgx.Row) (*world.CharacterState, error) {
	var cs world.CharacterState
	var chapterID sql.NullString
	var metadataJSON []byte

	err := row.Scan(
		&cs.ID, &cs.TenantID, &cs.CharacterID, &chapterID, &cs.ChapterNumber,
		&cs.Emotion, &cs.Goal, &cs.Secrets, &cs.Clothes, &cs.PhysicalState,
		&cs.TimelineDay, &cs.TimelineConfidence, &metadataJSON, &cs.CreatedAt, &cs.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if chapterID.Valid {
		parsedID, err := uuid.Parse(chapterID.String)
		if err == nil {
			cs.ChapterID = &parsedID
		}
	}

	if cs.Secrets == nil {
		cs.Secrets = []string{}
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &cs.Metadata); err != nil {
			return nil, err
		}
	} else {
		cs.Metadata = map[string]interface{}{}
	}

	return &cs, nil
}

func (r *CharacterStateRepository) scanCharacterStates(rows pgx.Rows) ([]*world.CharacterState, error) {
	states := make([]*world.CharacterState, 0)
	for rows.Next() {
		var cs world.CharacterState
		var chapterID sql.NullString
		var metadataJSON []byte

		err := rows.Scan(
			&cs.ID, &cs.TenantID, &cs.CharacterID, &chapterID, &cs.ChapterNumber,
			&cs.Emotion, &cs.Goal, &cs.Secrets, &cs.Clothes, &cs.PhysicalState,
			&cs.TimelineDay, &cs.TimelineConfidence, &metadataJSON, &cs.CreatedAt, &cs.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if chapterID.Valid {
			parsedID, err := uuid.Parse(chapterID.String)
			if err == nil {
				cs.ChapterID = &parsedID
			}
		}

		if cs.Secrets == nil {
			cs.Secrets = []string{}
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &cs.Metadata); err != nil {
				return nil, err
			}
		} else {
			cs.Metadata = map[string]interface{}{}
		}

		states = append(states, &cs)
	}

	return states, rows.Err()
}
