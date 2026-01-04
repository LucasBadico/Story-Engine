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

var _ repositories.EventArtifactRepository = (*EventArtifactRepository)(nil)

// EventArtifactRepository implements the event-artifact repository interface for SQLite
type EventArtifactRepository struct {
	db *DB
}

// NewEventArtifactRepository creates a new event-artifact repository
func NewEventArtifactRepository(db *DB) *EventArtifactRepository {
	return &EventArtifactRepository{db: db}
}

// Create creates a new event-artifact relationship
func (r *EventArtifactRepository) Create(ctx context.Context, ea *world.EventArtifact) error {
	// Get tenant_id from event
	var tenantIDStr string
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM events WHERE id = ?", ea.EventID.String()).Scan(&tenantIDStr); err != nil {
		return err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO event_artifacts (id, tenant_id, event_id, artifact_id, role, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var role sql.NullString
	if ea.Role != nil {
		role = sql.NullString{String: *ea.Role, Valid: true}
	}

	_, err = r.db.Exec(ctx, query,
		ea.ID.String(),
		tenantID.String(),
		ea.EventID.String(),
		ea.ArtifactID.String(),
		role,
		ea.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetByID retrieves an event-artifact relationship by ID
func (r *EventArtifactRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.EventArtifact, error) {
	query := `
		SELECT id, event_id, artifact_id, role, created_at
		FROM event_artifacts
		WHERE tenant_id = ? AND id = ?
	`
	var ea world.EventArtifact
	var idStr, eventIDStr, artifactIDStr, createdAtStr string
	var role sql.NullString

	err := r.db.QueryRow(ctx, query, tenantID.String(), id.String()).Scan(
		&idStr, &eventIDStr, &artifactIDStr, &role, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event_artifact",
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
	ea.ID = parsedID

	parsedEventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return nil, err
	}
	ea.EventID = parsedEventID

	parsedArtifactID, err := uuid.Parse(artifactIDStr)
	if err != nil {
		return nil, err
	}
	ea.ArtifactID = parsedArtifactID

	// Parse nullable string
	if role.Valid {
		ea.Role = &role.String
	}

	// Parse timestamp
	ea.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}

	return &ea, nil
}

// ListByEvent lists event-artifact relationships for an event
func (r *EventArtifactRepository) ListByEvent(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.EventArtifact, error) {
	query := `
		SELECT id, event_id, artifact_id, role, created_at
		FROM event_artifacts
		WHERE tenant_id = ? AND event_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), eventID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventArtifacts(rows)
}

// ListByArtifact lists event-artifact relationships for an artifact
func (r *EventArtifactRepository) ListByArtifact(ctx context.Context, tenantID, artifactID uuid.UUID) ([]*world.EventArtifact, error) {
	query := `
		SELECT id, event_id, artifact_id, role, created_at
		FROM event_artifacts
		WHERE tenant_id = ? AND artifact_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID.String(), artifactID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventArtifacts(rows)
}

// Delete deletes an event-artifact relationship
func (r *EventArtifactRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM event_artifacts WHERE tenant_id = ? AND id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), id.String())
	return err
}

// DeleteByEventAndArtifact deletes an event-artifact relationship
func (r *EventArtifactRepository) DeleteByEventAndArtifact(ctx context.Context, tenantID, eventID, artifactID uuid.UUID) error {
	query := `DELETE FROM event_artifacts WHERE tenant_id = ? AND event_id = ? AND artifact_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), eventID.String(), artifactID.String())
	return err
}

// DeleteByEvent deletes all event-artifact relationships for an event
func (r *EventArtifactRepository) DeleteByEvent(ctx context.Context, tenantID, eventID uuid.UUID) error {
	query := `DELETE FROM event_artifacts WHERE tenant_id = ? AND event_id = ?`
	_, err := r.db.Exec(ctx, query, tenantID.String(), eventID.String())
	return err
}

func (r *EventArtifactRepository) scanEventArtifacts(rows *sql.Rows) ([]*world.EventArtifact, error) {
	eventArtifacts := make([]*world.EventArtifact, 0)
	for rows.Next() {
		var ea world.EventArtifact
		var idStr, eventIDStr, artifactIDStr, createdAtStr string
		var role sql.NullString

		err := rows.Scan(
			&idStr, &eventIDStr, &artifactIDStr, &role, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		ea.ID = parsedID

		parsedEventID, err := uuid.Parse(eventIDStr)
		if err != nil {
			return nil, err
		}
		ea.EventID = parsedEventID

		parsedArtifactID, err := uuid.Parse(artifactIDStr)
		if err != nil {
			return nil, err
		}
		ea.ArtifactID = parsedArtifactID

		// Parse nullable string
		if role.Valid {
			ea.Role = &role.String
		}

		// Parse timestamp
		ea.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		eventArtifacts = append(eventArtifacts, &ea)
	}
	return eventArtifacts, rows.Err()
}

