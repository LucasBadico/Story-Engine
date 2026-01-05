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
func (r *EventRepository) Create(ctx context.Context, e *world.Event) error {
	query := `
		INSERT INTO events (id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.Exec(ctx, query,
		e.ID, e.TenantID, e.WorldID, e.Name, e.Type, e.Description, e.Timeline, e.Importance,
		e.ParentID, e.HierarchyLevel, e.TimelinePosition, e.IsEpoch, e.CreatedAt, e.UpdatedAt)
	return err
}

// GetByID retrieves an event by ID
func (r *EventRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Event, error) {
	query := `
		SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
		FROM events
		WHERE tenant_id = $1 AND id = $2
	`
	var e world.Event
	var parentID *uuid.UUID

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&e.ID, &e.TenantID, &e.WorldID, &e.Name, &e.Type, &e.Description, &e.Timeline, &e.Importance,
		&parentID, &e.HierarchyLevel, &e.TimelinePosition, &e.IsEpoch, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	e.ParentID = parentID
	return &e, nil
}

// ListByWorld lists events for a world
func (r *EventRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID) ([]*world.Event, error) {
	query := `
		SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
		FROM events
		WHERE tenant_id = $1 AND world_id = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, worldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// Update updates an event
func (r *EventRepository) Update(ctx context.Context, e *world.Event) error {
	query := `
		UPDATE events
		SET name = $2, type = $3, description = $4, timeline = $5, importance = $6, parent_id = $7, hierarchy_level = $8, timeline_position = $9, is_epoch = $10, updated_at = $11
		WHERE tenant_id = $12 AND id = $1
	`
	_, err := r.db.Exec(ctx, query, e.ID, e.Name, e.Type, e.Description, e.Timeline, e.Importance,
		e.ParentID, e.HierarchyLevel, e.TimelinePosition, e.IsEpoch, e.UpdatedAt, e.TenantID)
	return err
}

// Delete deletes an event
func (r *EventRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM events WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *EventRepository) scanEvents(rows pgx.Rows) ([]*world.Event, error) {
	events := make([]*world.Event, 0)
	for rows.Next() {
		var e world.Event
		var parentID *uuid.UUID
		err := rows.Scan(
			&e.ID, &e.TenantID, &e.WorldID, &e.Name, &e.Type, &e.Description, &e.Timeline, &e.Importance,
			&parentID, &e.HierarchyLevel, &e.TimelinePosition, &e.IsEpoch, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		e.ParentID = parentID
		events = append(events, &e)
	}
	return events, rows.Err()
}

// GetChildren returns direct children of an event (causality)
func (r *EventRepository) GetChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]*world.Event, error) {
	query := `
		SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
		FROM events
		WHERE tenant_id = $1 AND parent_id = $2
		ORDER BY timeline_position ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetAncestors returns the chain of ancestors up to the root
func (r *EventRepository) GetAncestors(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.Event, error) {
	ancestors := make([]*world.Event, 0)
	currentID := eventID

	for currentID != uuid.Nil {
		event, err := r.GetByID(ctx, tenantID, currentID)
		if err != nil {
			return nil, err
		}
		if event.ParentID == nil {
			break
		}
		parent, err := r.GetByID(ctx, tenantID, *event.ParentID)
		if err != nil {
			return nil, err
		}
		ancestors = append([]*world.Event{parent}, ancestors...)
		currentID = *event.ParentID
	}

	return ancestors, nil
}

// GetDescendants returns all descendants (complete subtree)
func (r *EventRepository) GetDescendants(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.Event, error) {
	descendants := make([]*world.Event, 0)
	queue := []uuid.UUID{eventID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		children, err := r.GetChildren(ctx, tenantID, currentID)
		if err != nil {
			return nil, err
		}

		for _, child := range children {
			descendants = append(descendants, child)
			queue = append(queue, child.ID)
		}
	}

	return descendants, nil
}

// GetEpoch returns the epoch event (time zero) for a world
func (r *EventRepository) GetEpoch(ctx context.Context, tenantID, worldID uuid.UUID) (*world.Event, error) {
	query := `
		SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
		FROM events
		WHERE tenant_id = $1 AND world_id = $2 AND is_epoch = TRUE
		LIMIT 1
	`
	var e world.Event
	var parentID *uuid.UUID

	err := r.db.QueryRow(ctx, query, tenantID, worldID).Scan(
		&e.ID, &e.TenantID, &e.WorldID, &e.Name, &e.Type, &e.Description, &e.Timeline, &e.Importance,
		&parentID, &e.HierarchyLevel, &e.TimelinePosition, &e.IsEpoch, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "epoch_event",
				ID:       worldID.String(),
			}
		}
		return nil, err
	}

	e.ParentID = parentID
	return &e, nil
}

// ListByTimeline lists events ordered by timeline position
func (r *EventRepository) ListByTimeline(ctx context.Context, tenantID, worldID uuid.UUID, fromPos, toPos *float64) ([]*world.Event, error) {
	var query string
	var args []interface{}

	if fromPos != nil && toPos != nil {
		query = `
			SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
			FROM events
			WHERE tenant_id = $1 AND world_id = $2 AND timeline_position >= $3 AND timeline_position <= $4
			ORDER BY timeline_position ASC
		`
		args = []interface{}{tenantID, worldID, *fromPos, *toPos}
	} else if fromPos != nil {
		query = `
			SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
			FROM events
			WHERE tenant_id = $1 AND world_id = $2 AND timeline_position >= $3
			ORDER BY timeline_position ASC
		`
		args = []interface{}{tenantID, worldID, *fromPos}
	} else if toPos != nil {
		query = `
			SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
			FROM events
			WHERE tenant_id = $1 AND world_id = $2 AND timeline_position <= $3
			ORDER BY timeline_position ASC
		`
		args = []interface{}{tenantID, worldID, *toPos}
	} else {
		query = `
			SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
			FROM events
			WHERE tenant_id = $1 AND world_id = $2
			ORDER BY timeline_position ASC
		`
		args = []interface{}{tenantID, worldID}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEvents(rows)
}


