package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

var _ repositories.EventRepository = (*EventRepository)(nil)

// EventRepository implements the event repository interface
type EventRepository struct {
	db *DB
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *DB) *EventRepository {
	return &EventRepository{db: db}
}

// Create creates a new event
func (r *EventRepository) Create(ctx context.Context, event *memory.Event) error {
	entitiesJSON, err := json.Marshal(event.EntitiesInvolved)
	if err != nil {
		return fmt.Errorf("failed to marshal entities_involved: %w", err)
	}
	if len(event.EntitiesInvolved) == 0 {
		entitiesJSON = []byte("[]")
	}
	query := `
		INSERT INTO events (id, tenant_id, story_id, chapter_number, timeline_day, timeline_confidence, scene_id, description, entities_involved, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = r.db.Exec(ctx, query,
		event.ID, event.TenantID, event.StoryID, event.ChapterNumber, event.TimelineDay, event.TimelineConfidence, event.SceneID, event.Description, entitiesJSON, event.CreatedAt)
	return err
}

// ListByStory lists events for a story up to a given chapter
func (r *EventRepository) ListByStory(ctx context.Context, storyID uuid.UUID, asOfChapter int, limit int) ([]*memory.Event, error) {
	query := `
		SELECT id, tenant_id, story_id, chapter_number, timeline_day, timeline_confidence, scene_id, description, entities_involved, created_at
		FROM events
		WHERE story_id = $1 AND chapter_number <= $2
		ORDER BY chapter_number, created_at
		LIMIT $3
	`
	rows, err := r.db.Query(ctx, query, storyID, asOfChapter, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// ListByStoryAtDay lists events for a story up to a given timeline day
func (r *EventRepository) ListByStoryAtDay(ctx context.Context, storyID uuid.UUID, asOfDay int, limit int) ([]*memory.Event, error) {
	query := `
		SELECT id, tenant_id, story_id, chapter_number, timeline_day, timeline_confidence, scene_id, description, entities_involved, created_at
		FROM events
		WHERE story_id = $1 AND timeline_day <= $2
		ORDER BY timeline_day, created_at
		LIMIT $3
	`
	rows, err := r.db.Query(ctx, query, storyID, asOfDay, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// Delete deletes an event by ID
func (r *EventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// scanRows scans multiple event rows
func (r *EventRepository) scanRows(rows pgx.Rows) ([]*memory.Event, error) {
	var events []*memory.Event
	for rows.Next() {
		var event memory.Event
		var entitiesJSON []byte
		if err := rows.Scan(
			&event.ID, &event.TenantID, &event.StoryID, &event.ChapterNumber, &event.TimelineDay, &event.TimelineConfidence, &event.SceneID, &event.Description, &entitiesJSON, &event.CreatedAt); err != nil {
			return nil, err
		}
		event.EntitiesInvolved = []string{}
		if len(entitiesJSON) > 0 {
			if err := json.Unmarshal(entitiesJSON, &event.EntitiesInvolved); err != nil {
				return nil, fmt.Errorf("failed to unmarshal entities_involved: %w", err)
			}
		}
		events = append(events, &event)
	}
	return events, rows.Err()
}
