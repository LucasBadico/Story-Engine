package world

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEventNameRequired = errors.New("event name is required")
	ErrInvalidImportance = errors.New("importance must be between 1 and 10")
)

// Event represents an event in a world
type Event struct {
	ID               uuid.UUID  `json:"id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	WorldID          uuid.UUID  `json:"world_id"`
	Name             string     `json:"name"`
	Type             *string    `json:"type,omitempty"`
	Description      *string    `json:"description,omitempty"`
	Timeline         *string    `json:"timeline,omitempty"`
	Importance       int        `json:"importance"`
	ParentID         *uuid.UUID `json:"parent_id,omitempty"`         // evento causador
	HierarchyLevel   int        `json:"hierarchy_level"`              // nível na árvore de causalidade
	TimelinePosition float64    `json:"timeline_position"`             // posição absoluta na timeline (anos desde epoch)
	IsEpoch          bool       `json:"is_epoch"`                      // marca evento como "tempo zero"
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// NewEvent creates a new event
func NewEvent(tenantID, worldID uuid.UUID, name string) (*Event, error) {
	if name == "" {
		return nil, ErrEventNameRequired
	}

	return &Event{
		ID:               uuid.New(),
		TenantID:         tenantID,
		WorldID:          worldID,
		Name:             name,
		Importance:       5,
		HierarchyLevel:   0,
		TimelinePosition: 0.0,
		IsEpoch:          false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}, nil
}

// Validate validates the event entity
func (e *Event) Validate() error {
	if e.Name == "" {
		return ErrEventNameRequired
	}
	if e.Importance < 1 || e.Importance > 10 {
		return ErrInvalidImportance
	}
	return nil
}

// UpdateName updates the event name
func (e *Event) UpdateName(name string) error {
	if name == "" {
		return ErrEventNameRequired
	}
	e.Name = name
	e.UpdatedAt = time.Now()
	return nil
}

// UpdateType updates the event type
func (e *Event) UpdateType(eventType *string) {
	e.Type = eventType
	e.UpdatedAt = time.Now()
}

// UpdateDescription updates the event description
func (e *Event) UpdateDescription(description *string) {
	e.Description = description
	e.UpdatedAt = time.Now()
}

// UpdateTimeline updates the event timeline
func (e *Event) UpdateTimeline(timeline *string) {
	e.Timeline = timeline
	e.UpdatedAt = time.Now()
}

// UpdateImportance updates the event importance
func (e *Event) UpdateImportance(importance int) error {
	if importance < 1 || importance > 10 {
		return ErrInvalidImportance
	}
	e.Importance = importance
	e.UpdatedAt = time.Now()
	return nil
}

// SetParent sets the parent event (causer) and updates hierarchy level
func (e *Event) SetParent(parentID *uuid.UUID, parentLevel int) {
	e.ParentID = parentID
	if parentID != nil {
		e.HierarchyLevel = parentLevel + 1
	} else {
		e.HierarchyLevel = 0
	}
	e.UpdatedAt = time.Now()
}

// SetHierarchyLevel sets the hierarchy level directly
func (e *Event) SetHierarchyLevel(level int) {
	e.HierarchyLevel = level
	e.UpdatedAt = time.Now()
}

// SetTimelinePosition sets the absolute timeline position
func (e *Event) SetTimelinePosition(position float64) {
	e.TimelinePosition = position
	e.UpdatedAt = time.Now()
}

// SetAsEpoch marks this event as the epoch (time zero) of the world
func (e *Event) SetAsEpoch(isEpoch bool) {
	e.IsEpoch = isEpoch
	if isEpoch {
		e.TimelinePosition = 0.0
	}
	e.UpdatedAt = time.Now()
}

// WorldDate represents a date in the world's calendar
type WorldDate struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
}

// GetWorldDate converts timeline position to a readable date using TimeConfig
func (e *Event) GetWorldDate(timeConfig *TimeConfig) WorldDate {
	if timeConfig == nil {
		timeConfig = DefaultTimeConfig()
	}

	// timeline_position is in years (base unit)
	years := int(e.TimelinePosition)
	fractionalYear := e.TimelinePosition - float64(years)

	// Calculate day of year
	dayOfYear := int(fractionalYear * float64(timeConfig.DaysPerYear))
	if dayOfYear < 0 {
		dayOfYear = 0
	}
	if dayOfYear >= timeConfig.DaysPerYear {
		dayOfYear = timeConfig.DaysPerYear - 1
	}

	// Find month and day
	month := 1
	day := 1
	daysPassed := 0
	for m := 1; m <= timeConfig.MonthsPerYear; m++ {
		daysInMonth := timeConfig.DaysInMonth(m)
		if daysPassed+daysInMonth > dayOfYear {
			month = m
			day = dayOfYear - daysPassed + 1
			break
		}
		daysPassed += daysInMonth
	}

	// Calculate hour and minute (simplified)
	hoursInYear := float64(timeConfig.DaysPerYear) * timeConfig.HoursPerDay
	hoursInFractionalYear := fractionalYear * hoursInYear
	hour := int(hoursInFractionalYear) % int(timeConfig.HoursPerDay)
	minute := int((hoursInFractionalYear - float64(hour)) * 60)

	return WorldDate{
		Year:   years,
		Month:  month,
		Day:    day,
		Hour:   hour,
		Minute: minute,
	}
}


