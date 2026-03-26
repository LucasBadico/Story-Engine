package story

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PlotType constants
const (
	PlotTypeMain      = "main"
	PlotTypeSecondary = "secondary"
	PlotTypeMinor     = "minor"
)

// PlotStatus constants
const (
	PlotStatusOpen   = "open"
	PlotStatusClosed = "closed"
)

// PlotAction constants
const (
	PlotActionOpen     = "open"
	PlotActionDevelop  = "develop"
	PlotActionEscalate = "escalate"
	PlotActionClose    = "close"
)

var (
	ErrPlotTitleRequired             = errors.New("title is required for plot")
	ErrPlotTypeRequired              = errors.New("type is required for plot")
	ErrPlotDevelopmentActionRequired = errors.New("action is required for plot development")
	ErrPlotDevelopmentPlotIDRequired = errors.New("plot_id is required for plot development")
)

// PlotEntity represents an entity linked to a plot
type PlotEntity struct {
	ID         uuid.UUID
	PlotID     uuid.UUID
	Role       string
	EntityType string
	EntityID   uuid.UUID
	CreatedAt  time.Time
}

// Plot represents a narrative plot thread in a story
type Plot struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	StoryID     uuid.UUID
	Title       string
	Type        string // main, secondary, minor
	Description string
	Status      string // open, closed
	Entities    []PlotEntity
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PlotDevelopment represents a development (event) in a plot at a specific chapter
type PlotDevelopment struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	PlotID             uuid.UUID
	ChapterNumber      int
	Action             string // open, develop, escalate, close
	Description        string
	Notes              string
	TimelineDay        *int
	TimelineConfidence *string
	TimelineEventID    *uuid.UUID
	CreatedAt          time.Time
}

// NewPlot creates a new plot
func NewPlot(tenantID, storyID uuid.UUID, title, plotType string) (*Plot, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant_id is required")
	}
	if storyID == uuid.Nil {
		return nil, errors.New("story_id is required")
	}
	if title == "" {
		return nil, ErrPlotTitleRequired
	}
	if plotType == "" {
		plotType = PlotTypeSecondary
	}

	now := time.Now()
	return &Plot{
		ID:        uuid.New(),
		TenantID:  tenantID,
		StoryID:   storyID,
		Title:     title,
		Type:      plotType,
		Status:    PlotStatusOpen,
		Entities:  []PlotEntity{},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// NewPlotDevelopment creates a new plot development
func NewPlotDevelopment(tenantID, plotID uuid.UUID, chapterNumber int, action, description string) (*PlotDevelopment, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant_id is required")
	}
	if plotID == uuid.Nil {
		return nil, ErrPlotDevelopmentPlotIDRequired
	}
	if action == "" {
		return nil, ErrPlotDevelopmentActionRequired
	}

	return &PlotDevelopment{
		ID:            uuid.New(),
		TenantID:      tenantID,
		PlotID:        plotID,
		ChapterNumber: chapterNumber,
		Action:        action,
		Description:   description,
		CreatedAt:     time.Now(),
	}, nil
}

// AddEntity adds an entity link to this plot
func (p *Plot) AddEntity(role, entityType string, entityID uuid.UUID) {
	p.Entities = append(p.Entities, PlotEntity{
		ID:         uuid.New(),
		PlotID:     p.ID,
		Role:       role,
		EntityType: entityType,
		EntityID:   entityID,
		CreatedAt:  time.Now(),
	})
}

// SetTimelineEvent links this development to a timeline event
func (d *PlotDevelopment) SetTimelineEvent(eventID uuid.UUID) {
	d.TimelineEventID = &eventID
}

// Close sets the plot status to closed
func (p *Plot) Close() {
	p.Status = PlotStatusClosed
	p.UpdatedAt = time.Now()
}

// SetTimeline sets the timeline day and confidence on a development
func (d *PlotDevelopment) SetTimeline(day int, confidence string) {
	d.TimelineDay = &day
	d.TimelineConfidence = &confidence
}

// Validate validates the plot entity
func (p *Plot) Validate() error {
	if p.Title == "" {
		return ErrPlotTitleRequired
	}
	return nil
}
