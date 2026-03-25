package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.TimelineEventRepository = (*TimelineEventRepository)(nil)

// TimelineEventRepository implements the timeline event repository interface
type TimelineEventRepository struct {
	db *DB
}

// NewTimelineEventRepository creates a new timeline event repository
func NewTimelineEventRepository(db *DB) *TimelineEventRepository {
	return &TimelineEventRepository{db: db}
}

// Create creates a new timeline event
func (r *TimelineEventRepository) Create(ctx context.Context, te *story.TimelineEvent) error {
	query := `
		INSERT INTO timeline_events (
			id, tenant_id, story_id, chapter_id, chapter_number,
			timeline_day, timeline_confidence, description, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		te.ID, te.TenantID, te.StoryID, te.ChapterID, te.ChapterNumber,
		te.TimelineDay, te.TimelineConfidence, te.Description, te.CreatedAt)
	return err
}

// GetByID retrieves a timeline event by ID
func (r *TimelineEventRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.TimelineEvent, error) {
	query := `
		SELECT id, tenant_id, story_id, chapter_id, chapter_number,
			timeline_day, timeline_confidence, description, created_at
		FROM timeline_events
		WHERE tenant_id = $1 AND id = $2
	`
	te, err := r.scanTimelineEvent(r.db.QueryRow(ctx, query, tenantID, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "timeline_event",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	// Load entities
	entities, err := r.ListEntities(ctx, te.ID)
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		te.Entities = append(te.Entities, *e)
	}

	return te, nil
}

// ListByStory lists timeline events for a story up to a given chapter
func (r *TimelineEventRepository) ListByStory(ctx context.Context, tenantID, storyID uuid.UUID, asOfChapter int, limit int) ([]*story.TimelineEvent, error) {
	query := `
		SELECT id, tenant_id, story_id, chapter_id, chapter_number,
			timeline_day, timeline_confidence, description, created_at
		FROM timeline_events
		WHERE tenant_id = $1 AND story_id = $2 AND chapter_number <= $3
		ORDER BY chapter_number, created_at
		LIMIT $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, storyID, asOfChapter, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTimelineEvents(rows)
}

// ListByChapter lists timeline events for a chapter
func (r *TimelineEventRepository) ListByChapter(ctx context.Context, tenantID, chapterID uuid.UUID) ([]*story.TimelineEvent, error) {
	query := `
		SELECT id, tenant_id, story_id, chapter_id, chapter_number,
			timeline_day, timeline_confidence, description, created_at
		FROM timeline_events
		WHERE tenant_id = $1 AND chapter_id = $2
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, tenantID, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTimelineEvents(rows)
}

// Delete deletes a timeline event
func (r *TimelineEventRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM timeline_events WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// AddEntity adds an entity link to a timeline event
func (r *TimelineEventRepository) AddEntity(ctx context.Context, entity *story.TimelineEventEntity) error {
	query := `
		INSERT INTO timeline_event_entities (id, timeline_event_id, role, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		entity.ID, entity.TimelineEventID, entity.Role, entity.EntityType, entity.EntityID, entity.CreatedAt)
	return err
}

// ListEntities lists entity links for a timeline event
func (r *TimelineEventRepository) ListEntities(ctx context.Context, timelineEventID uuid.UUID) ([]*story.TimelineEventEntity, error) {
	query := `
		SELECT id, timeline_event_id, role, entity_type, entity_id, created_at
		FROM timeline_event_entities
		WHERE timeline_event_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, timelineEventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entities := make([]*story.TimelineEventEntity, 0)
	for rows.Next() {
		var e story.TimelineEventEntity
		if err := rows.Scan(&e.ID, &e.TimelineEventID, &e.Role, &e.EntityType, &e.EntityID, &e.CreatedAt); err != nil {
			return nil, err
		}
		entities = append(entities, &e)
	}
	return entities, rows.Err()
}

func (r *TimelineEventRepository) scanTimelineEvent(row pgx.Row) (*story.TimelineEvent, error) {
	var te story.TimelineEvent

	err := row.Scan(
		&te.ID, &te.TenantID, &te.StoryID, &te.ChapterID, &te.ChapterNumber,
		&te.TimelineDay, &te.TimelineConfidence, &te.Description, &te.CreatedAt)
	if err != nil {
		return nil, err
	}

	te.Entities = []story.TimelineEventEntity{}

	return &te, nil
}

func (r *TimelineEventRepository) scanTimelineEvents(rows pgx.Rows) ([]*story.TimelineEvent, error) {
	events := make([]*story.TimelineEvent, 0)
	for rows.Next() {
		var te story.TimelineEvent

		err := rows.Scan(
			&te.ID, &te.TenantID, &te.StoryID, &te.ChapterID, &te.ChapterNumber,
			&te.TimelineDay, &te.TimelineConfidence, &te.Description, &te.CreatedAt)
		if err != nil {
			return nil, err
		}

		te.Entities = []story.TimelineEventEntity{}
		events = append(events, &te)
	}

	return events, rows.Err()
}
