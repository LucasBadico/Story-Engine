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
	query := `
		INSERT INTO event_artifacts (id, event_id, artifact_id, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (event_id, artifact_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, ea.ID, ea.EventID, ea.ArtifactID, ea.Role, ea.CreatedAt)
	return err
}

// GetByID retrieves an event-artifact relationship by ID
func (r *EventArtifactRepository) GetByID(ctx context.Context, id uuid.UUID) (*world.EventArtifact, error) {
	query := `
		SELECT id, event_id, artifact_id, role, created_at
		FROM event_artifacts
		WHERE id = $1
	`
	var ea world.EventArtifact

	err := r.db.QueryRow(ctx, query, id).Scan(
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
func (r *EventArtifactRepository) ListByEvent(ctx context.Context, eventID uuid.UUID) ([]*world.EventArtifact, error) {
	query := `
		SELECT id, event_id, artifact_id, role, created_at
		FROM event_artifacts
		WHERE event_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventArtifacts(rows)
}

// ListByArtifact lists event-artifact relationships for an artifact
func (r *EventArtifactRepository) ListByArtifact(ctx context.Context, artifactID uuid.UUID) ([]*world.EventArtifact, error) {
	query := `
		SELECT id, event_id, artifact_id, role, created_at
		FROM event_artifacts
		WHERE artifact_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, artifactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEventArtifacts(rows)
}

// Delete deletes an event-artifact relationship
func (r *EventArtifactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM event_artifacts WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByEventAndArtifact deletes an event-artifact relationship
func (r *EventArtifactRepository) DeleteByEventAndArtifact(ctx context.Context, eventID, artifactID uuid.UUID) error {
	query := `DELETE FROM event_artifacts WHERE event_id = $1 AND artifact_id = $2`
	_, err := r.db.Exec(ctx, query, eventID, artifactID)
	return err
}

// DeleteByEvent deletes all event-artifact relationships for an event
func (r *EventArtifactRepository) DeleteByEvent(ctx context.Context, eventID uuid.UUID) error {
	query := `DELETE FROM event_artifacts WHERE event_id = $1`
	_, err := r.db.Exec(ctx, query, eventID)
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


