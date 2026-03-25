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

var _ repositories.PlotDevelopmentRepository = (*PlotDevelopmentRepository)(nil)

// PlotDevelopmentRepository implements the plot development repository interface
type PlotDevelopmentRepository struct {
	db *DB
}

// NewPlotDevelopmentRepository creates a new plot development repository
func NewPlotDevelopmentRepository(db *DB) *PlotDevelopmentRepository {
	return &PlotDevelopmentRepository{db: db}
}

// Create creates a new plot development
func (r *PlotDevelopmentRepository) Create(ctx context.Context, dev *story.PlotDevelopment) error {
	query := `
		INSERT INTO plot_developments (
			id, tenant_id, plot_id, chapter_number, action, description, notes,
			timeline_day, timeline_confidence, timeline_event_id, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		dev.ID, dev.TenantID, dev.PlotID, dev.ChapterNumber, dev.Action, dev.Description, dev.Notes,
		dev.TimelineDay, dev.TimelineConfidence, dev.TimelineEventID, dev.CreatedAt)
	return err
}

// GetByID retrieves a plot development by ID
func (r *PlotDevelopmentRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*story.PlotDevelopment, error) {
	query := `
		SELECT id, tenant_id, plot_id, chapter_number, action, description, notes,
			timeline_day, timeline_confidence, timeline_event_id, created_at
		FROM plot_developments
		WHERE tenant_id = $1 AND id = $2
	`
	dev, err := r.scanPlotDevelopment(r.db.QueryRow(ctx, query, tenantID, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "plot_development",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return dev, nil
}

// ListByPlot lists all developments for a plot
func (r *PlotDevelopmentRepository) ListByPlot(ctx context.Context, tenantID, plotID uuid.UUID) ([]*story.PlotDevelopment, error) {
	query := `
		SELECT id, tenant_id, plot_id, chapter_number, action, description, notes,
			timeline_day, timeline_confidence, timeline_event_id, created_at
		FROM plot_developments
		WHERE tenant_id = $1 AND plot_id = $2
		ORDER BY chapter_number, created_at
	`
	rows, err := r.db.Query(ctx, query, tenantID, plotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPlotDevelopments(rows)
}

// ListByPlotUpToChapter lists developments for a plot up to a given chapter
func (r *PlotDevelopmentRepository) ListByPlotUpToChapter(ctx context.Context, tenantID, plotID uuid.UUID, asOfChapter int) ([]*story.PlotDevelopment, error) {
	query := `
		SELECT id, tenant_id, plot_id, chapter_number, action, description, notes,
			timeline_day, timeline_confidence, timeline_event_id, created_at
		FROM plot_developments
		WHERE tenant_id = $1 AND plot_id = $2 AND chapter_number <= $3
		ORDER BY chapter_number, created_at
	`
	rows, err := r.db.Query(ctx, query, tenantID, plotID, asOfChapter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPlotDevelopments(rows)
}

// Delete deletes a plot development
func (r *PlotDevelopmentRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM plot_developments WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *PlotDevelopmentRepository) scanPlotDevelopment(row pgx.Row) (*story.PlotDevelopment, error) {
	var dev story.PlotDevelopment
	var description, notes *string

	err := row.Scan(
		&dev.ID, &dev.TenantID, &dev.PlotID, &dev.ChapterNumber, &dev.Action, &description, &notes,
		&dev.TimelineDay, &dev.TimelineConfidence, &dev.TimelineEventID, &dev.CreatedAt)
	if err != nil {
		return nil, err
	}

	if description != nil {
		dev.Description = *description
	}
	if notes != nil {
		dev.Notes = *notes
	}

	return &dev, nil
}

func (r *PlotDevelopmentRepository) scanPlotDevelopments(rows pgx.Rows) ([]*story.PlotDevelopment, error) {
	devs := make([]*story.PlotDevelopment, 0)
	for rows.Next() {
		var dev story.PlotDevelopment
		var description, notes *string

		err := rows.Scan(
			&dev.ID, &dev.TenantID, &dev.PlotID, &dev.ChapterNumber, &dev.Action, &description, &notes,
			&dev.TimelineDay, &dev.TimelineConfidence, &dev.TimelineEventID, &dev.CreatedAt)
		if err != nil {
			return nil, err
		}

		if description != nil {
			dev.Description = *description
		}
		if notes != nil {
			dev.Notes = *notes
		}
		devs = append(devs, &dev)
	}

	return devs, rows.Err()
}
