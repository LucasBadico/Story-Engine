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

var _ repositories.PlotRepository = (*PlotRepository)(nil)

// PlotRepository implements the plot repository interface
type PlotRepository struct {
	db *DB
}

// NewPlotRepository creates a new plot repository
func NewPlotRepository(db *DB) *PlotRepository {
	return &PlotRepository{db: db}
}

// Create creates a new plot
func (r *PlotRepository) Create(ctx context.Context, p *story.Plot) error {
	query := `
		INSERT INTO plots (id, tenant_id, story_id, title, type, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.TenantID, p.StoryID, p.Title, p.Type, p.Description, p.Status, p.CreatedAt, p.UpdatedAt)
	return err
}

// GetByID retrieves a plot by ID
func (r *PlotRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.Plot, error) {
	query := `
		SELECT id, tenant_id, story_id, title, type, description, status, created_at, updated_at
		FROM plots
		WHERE tenant_id = $1 AND id = $2
	`
	p, err := r.scanPlot(r.db.QueryRow(ctx, query, tenantID, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "plot",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	// Load entities
	entities, err := r.ListEntities(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		p.Entities = append(p.Entities, *e)
	}

	return p, nil
}

// ListByStory lists all plots for a story
func (r *PlotRepository) ListByStory(ctx context.Context, tenantID, storyID uuid.UUID) ([]*story.Plot, error) {
	query := `
		SELECT id, tenant_id, story_id, title, type, description, status, created_at, updated_at
		FROM plots
		WHERE tenant_id = $1 AND story_id = $2
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, tenantID, storyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPlots(rows)
}

// ListByStoryAndStatus lists plots for a story filtered by status
func (r *PlotRepository) ListByStoryAndStatus(ctx context.Context, tenantID, storyID uuid.UUID, status string) ([]*story.Plot, error) {
	query := `
		SELECT id, tenant_id, story_id, title, type, description, status, created_at, updated_at
		FROM plots
		WHERE tenant_id = $1 AND story_id = $2 AND status = $3
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, tenantID, storyID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPlots(rows)
}

// Update updates a plot
func (r *PlotRepository) Update(ctx context.Context, p *story.Plot) error {
	query := `
		UPDATE plots
		SET title = $3, type = $4, description = $5, status = $6, updated_at = $7
		WHERE tenant_id = $1 AND id = $2
	`
	_, err := r.db.Exec(ctx, query,
		p.TenantID, p.ID, p.Title, p.Type, p.Description, p.Status, p.UpdatedAt)
	return err
}

// Delete deletes a plot
func (r *PlotRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM plots WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

// AddEntity adds an entity link to a plot
func (r *PlotRepository) AddEntity(ctx context.Context, entity *story.PlotEntity) error {
	query := `
		INSERT INTO plot_entities (id, plot_id, role, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		entity.ID, entity.PlotID, entity.Role, entity.EntityType, entity.EntityID, entity.CreatedAt)
	return err
}

// ListEntities lists entity links for a plot
func (r *PlotRepository) ListEntities(ctx context.Context, plotID uuid.UUID) ([]*story.PlotEntity, error) {
	query := `
		SELECT id, plot_id, role, entity_type, entity_id, created_at
		FROM plot_entities
		WHERE plot_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, plotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entities := make([]*story.PlotEntity, 0)
	for rows.Next() {
		var e story.PlotEntity
		if err := rows.Scan(&e.ID, &e.PlotID, &e.Role, &e.EntityType, &e.EntityID, &e.CreatedAt); err != nil {
			return nil, err
		}
		entities = append(entities, &e)
	}
	return entities, rows.Err()
}

// DeleteEntities deletes all entity links for a plot
func (r *PlotRepository) DeleteEntities(ctx context.Context, plotID uuid.UUID) error {
	query := `DELETE FROM plot_entities WHERE plot_id = $1`
	_, err := r.db.Exec(ctx, query, plotID)
	return err
}

func (r *PlotRepository) scanPlot(row pgx.Row) (*story.Plot, error) {
	var p story.Plot
	var description *string

	err := row.Scan(
		&p.ID, &p.TenantID, &p.StoryID, &p.Title, &p.Type, &description, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if description != nil {
		p.Description = *description
	}
	p.Entities = []story.PlotEntity{}

	return &p, nil
}

func (r *PlotRepository) scanPlots(rows pgx.Rows) ([]*story.Plot, error) {
	plots := make([]*story.Plot, 0)
	for rows.Next() {
		var p story.Plot
		var description *string

		err := rows.Scan(
			&p.ID, &p.TenantID, &p.StoryID, &p.Title, &p.Type, &description, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if description != nil {
			p.Description = *description
		}
		p.Entities = []story.PlotEntity{}
		plots = append(plots, &p)
	}

	return plots, rows.Err()
}
