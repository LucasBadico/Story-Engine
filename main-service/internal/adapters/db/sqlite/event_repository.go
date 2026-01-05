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

var _ repositories.EventRepository = (*EventRepository)(nil)

// EventRepository implements the event repository interface for SQLite
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var eventType sql.NullString
	if e.Type != nil {
		eventType = sql.NullString{String: *e.Type, Valid: true}
	}

	var description sql.NullString
	if e.Description != nil {
		description = sql.NullString{String: *e.Description, Valid: true}
	}

	var timeline sql.NullString
	if e.Timeline != nil {
		timeline = sql.NullString{String: *e.Timeline, Valid: true}
	}

	var parentID sql.NullString
	if e.ParentID != nil {
		parentID = sql.NullString{String: e.ParentID.String(), Valid: true}
	}

	var isEpoch int
	if e.IsEpoch {
		isEpoch = 1
	}

	_, err := r.db.Exec(ctx, query,
		e.ID.String(),
		e.TenantID.String(),
		e.WorldID.String(),
		e.Name,
		eventType,
		description,
		timeline,
		e.Importance,
		parentID,
		e.HierarchyLevel,
		e.TimelinePosition,
		isEpoch,
		e.CreatedAt.Format(time.RFC3339),
		e.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves an event by ID
func (r *EventRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.Event, error) {
	query := `
		SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
		FROM events
		WHERE tenant_id = ? AND id = ?
	`
	var e world.Event
	var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
	var eventType, description, timeline, parentIDStr sql.NullString
	var isEpoch int

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &tenantIDStr, &worldIDStr, &e.Name, &eventType, &description, &timeline, &e.Importance,
		&parentIDStr, &e.HierarchyLevel, &e.TimelinePosition, &isEpoch, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event",
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
	e.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	e.TenantID = parsedTenantID

	parsedWorldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		return nil, err
	}
	e.WorldID = parsedWorldID

	// Parse parent ID
	if parentIDStr.Valid {
		parsedParentID, err := uuid.Parse(parentIDStr.String)
		if err != nil {
			return nil, err
		}
		e.ParentID = &parsedParentID
	}

	// Parse is_epoch
	e.IsEpoch = isEpoch == 1

	// Parse timestamps
	e.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	e.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable strings
	if eventType.Valid {
		e.Type = &eventType.String
	}
	if description.Valid {
		e.Description = &description.String
	}
	if timeline.Valid {
		e.Timeline = &timeline.String
	}

	return &e, nil
}

// ListByWorld lists events for a world
func (r *EventRepository) ListByWorld(ctx context.Context, tenantID, worldID uuid.UUID) ([]*world.Event, error) {
	query := `
		SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
		FROM events
		WHERE tenant_id = ? AND world_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), worldID.String())
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
		SET name = ?, type = ?, description = ?, timeline = ?, importance = ?, parent_id = ?, hierarchy_level = ?, timeline_position = ?, is_epoch = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?
	`

	var eventType sql.NullString
	if e.Type != nil {
		eventType = sql.NullString{String: *e.Type, Valid: true}
	}

	var description sql.NullString
	if e.Description != nil {
		description = sql.NullString{String: *e.Description, Valid: true}
	}

	var timeline sql.NullString
	if e.Timeline != nil {
		timeline = sql.NullString{String: *e.Timeline, Valid: true}
	}

	var parentID sql.NullString
	if e.ParentID != nil {
		parentID = sql.NullString{String: e.ParentID.String(), Valid: true}
	}

	var isEpoch int
	if e.IsEpoch {
		isEpoch = 1
	}

	_, err := r.db.Exec(ctx, query,
		e.Name,
		eventType,
		description,
		timeline,
		e.Importance,
		parentID,
		e.HierarchyLevel,
		e.TimelinePosition,
		isEpoch,
		e.UpdatedAt.Format(time.RFC3339),
		e.TenantID.String(),
		e.ID.String(),
	)
	return err
}

// Delete deletes an event
func (r *EventRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM events WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

func (r *EventRepository) scanEvents(rows *sql.Rows) ([]*world.Event, error) {
	events := make([]*world.Event, 0)
	for rows.Next() {
		var e world.Event
		var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
		var eventType, description, timeline, parentIDStr sql.NullString
		var isEpoch int

		err := rows.Scan(
			&idStr, &tenantIDStr, &worldIDStr, &e.Name, &eventType, &description, &timeline, &e.Importance,
			&parentIDStr, &e.HierarchyLevel, &e.TimelinePosition, &isEpoch, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		e.ID = parsedID

		parsedTenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, err
		}
		e.TenantID = parsedTenantID

		parsedWorldID, err := uuid.Parse(worldIDStr)
		if err != nil {
			return nil, err
		}
		e.WorldID = parsedWorldID

		// Parse parent ID
		if parentIDStr.Valid {
			parsedParentID, err := uuid.Parse(parentIDStr.String)
			if err != nil {
				return nil, err
			}
			e.ParentID = &parsedParentID
		}

		// Parse is_epoch
		e.IsEpoch = isEpoch == 1

		// Parse timestamps
		e.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		e.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse nullable strings
		if eventType.Valid {
			e.Type = &eventType.String
		}
		if description.Valid {
			e.Description = &description.String
		}
		if timeline.Valid {
			e.Timeline = &timeline.String
		}

		events = append(events, &e)
	}

	return events, rows.Err()
}

// GetChildren returns direct children of an event (causality)
func (r *EventRepository) GetChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]*world.Event, error) {
	query := `
		SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
		FROM events
		WHERE tenant_id = ? AND parent_id = ?
		ORDER BY timeline_position ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), parentID.String())
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
		WHERE tenant_id = ? AND world_id = ? AND is_epoch = 1
		LIMIT 1
	`
	var e world.Event
	var idStr, tenantIDStr, worldIDStr, createdAtStr, updatedAtStr string
	var eventType, description, timeline, parentIDStr sql.NullString
	var isEpoch int

	err := r.db.QueryRow(ctx, query, tenantID.String(), worldID.String()).Scan(
		&idStr, &tenantIDStr, &worldIDStr, &e.Name, &eventType, &description, &timeline, &e.Importance,
		&parentIDStr, &e.HierarchyLevel, &e.TimelinePosition, &isEpoch, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "epoch_event",
				ID:       worldID.String(),
			}
		}
		return nil, err
	}

	// Parse UUIDs
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	e.ID = parsedID

	parsedTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	e.TenantID = parsedTenantID

	parsedWorldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		return nil, err
	}
	e.WorldID = parsedWorldID

	// Parse parent ID
	if parentIDStr.Valid {
		parsedParentID, err := uuid.Parse(parentIDStr.String)
		if err != nil {
			return nil, err
		}
		e.ParentID = &parsedParentID
	}

	// Parse is_epoch
	e.IsEpoch = isEpoch == 1

	// Parse timestamps
	e.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	e.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse nullable strings
	if eventType.Valid {
		e.Type = &eventType.String
	}
	if description.Valid {
		e.Description = &description.String
	}
	if timeline.Valid {
		e.Timeline = &timeline.String
	}

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
			WHERE tenant_id = ? AND world_id = ? AND timeline_position >= ? AND timeline_position <= ?
			ORDER BY timeline_position ASC
		`
		args = []interface{}{tenantID.String(), worldID.String(), *fromPos, *toPos}
	} else if fromPos != nil {
		query = `
			SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
			FROM events
			WHERE tenant_id = ? AND world_id = ? AND timeline_position >= ?
			ORDER BY timeline_position ASC
		`
		args = []interface{}{tenantID.String(), worldID.String(), *fromPos}
	} else if toPos != nil {
		query = `
			SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
			FROM events
			WHERE tenant_id = ? AND world_id = ? AND timeline_position <= ?
			ORDER BY timeline_position ASC
		`
		args = []interface{}{tenantID.String(), worldID.String(), *toPos}
	} else {
		query = `
			SELECT id, tenant_id, world_id, name, type, description, timeline, importance, parent_id, hierarchy_level, timeline_position, is_epoch, created_at, updated_at
			FROM events
			WHERE tenant_id = ? AND world_id = ?
			ORDER BY timeline_position ASC
		`
		args = []interface{}{tenantID.String(), worldID.String()}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

