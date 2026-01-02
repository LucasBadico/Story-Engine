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

var _ repositories.EventArtifactRepository = (*EventArtifactRepository)(nil)

// EventArtifactRepository implements the event-artifact repository interface
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
	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, "SELECT tenant_id FROM events WHERE id = $1", ea.EventID).Scan(&tenantID); err != nil {
		return err
	}

	query := `
		INSERT INTO event_artifacts (id, tenant_id, event_id, artifact_id, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (event_id, artifact_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, ea.ID, tenantID, ea.EventID, ea.ArtifactID, ea.Role, ea.CreatedAt)
	return err
}

// GetByID retrieves an event-artifact relationship by ID
func (r *EventArtifactRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*world.EventArtifact, error) {
	query := `
		SELECT id, event_id, artifact_id, role, created_at
		FROM event_artifacts
		WHERE tenant_id = $1 AND id = $2
	`
	var ea world.EventArtifact

	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(
		&ea.ID, &ea.EventID, &ea.ArtifactID, &ea.Role, &ea.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "event_artifact",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &ea, nil
}

// ListByEvent lists event-artifact relationships for an event
func (r *EventArtifactRepository) ListByEvent(ctx context.Context, tenantID, eventID uuid.UUID) ([]*world.EventArtifact, error) {
	query := `
		SELECT id, event_id, artifact_id, role, created_at
		FROM event_artifacts
		WHERE tenant_id = $1 AND event_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, eventID)
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
		WHERE tenant_id = $1 AND artifact_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, artifactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventArtifacts(rows)
}

// Delete deletes an event-artifact relationship
func (r *EventArtifactRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM event_artifacts WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// DeleteByEventAndArtifact deletes an event-artifact relationship
func (r *EventArtifactRepository) DeleteByEventAndArtifact(ctx context.Context, tenantID, eventID, artifactID uuid.UUID) error {
	query := `DELETE FROM event_artifacts WHERE tenant_id = $1 AND event_id = $2 AND artifact_id = $3`
	_, err := r.db.Exec(ctx, query, tenantID, eventID, artifactID)
	return err
}

// DeleteByEvent deletes all event-artifact relationships for an event
func (r *EventArtifactRepository) DeleteByEvent(ctx context.Context, tenantID, eventID uuid.UUID) error {
	query := `DELETE FROM event_artifacts WHERE tenant_id = $1 AND event_id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, eventID)
	return err
}

func (r *EventArtifactRepository) scanEventArtifacts(rows pgx.Rows) ([]*world.EventArtifact, error) {
	eventArtifacts := make([]*world.EventArtifact, 0)
	for rows.Next() {
		var ea world.EventArtifact
		err := rows.Scan(
			&ea.ID, &ea.EventID, &ea.ArtifactID, &ea.Role, &ea.CreatedAt)
		if err != nil {
			return nil, err
		}
		eventArtifacts = append(eventArtifacts, &ea)
	}
	return eventArtifacts, rows.Err()
}


