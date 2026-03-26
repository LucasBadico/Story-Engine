package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

var _ repositories.StateSnapshotRepository = (*StateSnapshotRepository)(nil)

// StateSnapshotRepository implements the state snapshot repository interface
type StateSnapshotRepository struct {
	db *DB
}

// NewStateSnapshotRepository creates a new state snapshot repository
func NewStateSnapshotRepository(db *DB) *StateSnapshotRepository {
	return &StateSnapshotRepository{db: db}
}

// Create creates a new state snapshot
func (r *StateSnapshotRepository) Create(ctx context.Context, snapshot *memory.StateSnapshot) error {
	stateDataJSON, err := json.Marshal(snapshot.StateData)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO state_snapshots (id, tenant_id, story_id, entity_id, entity_type, chapter_number, timeline_day, timeline_confidence, state_data, scene_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = r.db.Exec(ctx, query,
		snapshot.ID, snapshot.TenantID, snapshot.StoryID, snapshot.EntityID, snapshot.EntityType, snapshot.ChapterNumber, snapshot.TimelineDay, snapshot.TimelineConfidence, stateDataJSON, snapshot.SceneID, snapshot.CreatedAt)
	return err
}

// GetLatestByEntity retrieves the latest snapshot for an entity up to a given chapter
func (r *StateSnapshotRepository) GetLatestByEntity(ctx context.Context, storyID, entityID uuid.UUID, asOfChapter int) (*memory.StateSnapshot, error) {
	query := `
		SELECT id, tenant_id, story_id, entity_id, entity_type, chapter_number, timeline_day, timeline_confidence, state_data, scene_id, created_at
		FROM state_snapshots
		WHERE story_id = $1 AND entity_id = $2 AND chapter_number <= $3
		ORDER BY chapter_number DESC
		LIMIT 1
	`
	return r.scanOne(ctx, query, storyID, entityID, asOfChapter)
}

// GetLatestByEntityAtDay retrieves the latest snapshot for an entity up to a given timeline day
func (r *StateSnapshotRepository) GetLatestByEntityAtDay(ctx context.Context, storyID, entityID uuid.UUID, asOfDay int) (*memory.StateSnapshot, error) {
	query := `
		SELECT id, tenant_id, story_id, entity_id, entity_type, chapter_number, timeline_day, timeline_confidence, state_data, scene_id, created_at
		FROM state_snapshots
		WHERE story_id = $1 AND entity_id = $2 AND timeline_day <= $3
		ORDER BY timeline_day DESC
		LIMIT 1
	`
	return r.scanOne(ctx, query, storyID, entityID, asOfDay)
}

// ListByStory lists the latest snapshot per entity for a story up to a given chapter
func (r *StateSnapshotRepository) ListByStory(ctx context.Context, storyID uuid.UUID, asOfChapter int) ([]*memory.StateSnapshot, error) {
	query := `
		SELECT DISTINCT ON (entity_id) id, tenant_id, story_id, entity_id, entity_type, chapter_number, timeline_day, timeline_confidence, state_data, scene_id, created_at
		FROM state_snapshots
		WHERE story_id = $1 AND chapter_number <= $2
		ORDER BY entity_id, chapter_number DESC
	`
	rows, err := r.db.Query(ctx, query, storyID, asOfChapter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// ListByStoryAtDay lists the latest snapshot per entity for a story up to a given timeline day
func (r *StateSnapshotRepository) ListByStoryAtDay(ctx context.Context, storyID uuid.UUID, asOfDay int) ([]*memory.StateSnapshot, error) {
	query := `
		SELECT DISTINCT ON (entity_id) id, tenant_id, story_id, entity_id, entity_type, chapter_number, timeline_day, timeline_confidence, state_data, scene_id, created_at
		FROM state_snapshots
		WHERE story_id = $1 AND timeline_day <= $2
		ORDER BY entity_id, timeline_day DESC
	`
	rows, err := r.db.Query(ctx, query, storyID, asOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// Delete deletes a state snapshot by ID
func (r *StateSnapshotRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM state_snapshots WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// scanOne scans a single state snapshot row
func (r *StateSnapshotRepository) scanOne(ctx context.Context, query string, args ...interface{}) (*memory.StateSnapshot, error) {
	var snapshot memory.StateSnapshot
	var stateDataJSON []byte

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&snapshot.ID, &snapshot.TenantID, &snapshot.StoryID, &snapshot.EntityID, &snapshot.EntityType, &snapshot.ChapterNumber, &snapshot.TimelineDay, &snapshot.TimelineConfidence, &stateDataJSON, &snapshot.SceneID, &snapshot.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	snapshot.StateData = map[string]interface{}{}
	if len(stateDataJSON) > 0 {
		if err := json.Unmarshal(stateDataJSON, &snapshot.StateData); err != nil {
			return nil, err
		}
	}
	return &snapshot, nil
}

// scanRows scans multiple state snapshot rows
func (r *StateSnapshotRepository) scanRows(rows pgx.Rows) ([]*memory.StateSnapshot, error) {
	var snapshots []*memory.StateSnapshot
	for rows.Next() {
		var snapshot memory.StateSnapshot
		var stateDataJSON []byte
		if err := rows.Scan(
			&snapshot.ID, &snapshot.TenantID, &snapshot.StoryID, &snapshot.EntityID, &snapshot.EntityType, &snapshot.ChapterNumber, &snapshot.TimelineDay, &snapshot.TimelineConfidence, &stateDataJSON, &snapshot.SceneID, &snapshot.CreatedAt); err != nil {
			return nil, err
		}
		snapshot.StateData = map[string]interface{}{}
		if len(stateDataJSON) > 0 {
			if err := json.Unmarshal(stateDataJSON, &snapshot.StateData); err != nil {
				return nil, err
			}
		}
		snapshots = append(snapshots, &snapshot)
	}
	return snapshots, rows.Err()
}
