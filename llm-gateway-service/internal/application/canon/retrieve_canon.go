package canon

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

// RetrievalMode controls what data is returned
type RetrievalMode string

const (
	ModeRevealedNarrative  RetrievalMode = "revealed_narrative"
	ModeTimelineConsistent RetrievalMode = "timeline_consistent"
	ModeOmniscient         RetrievalMode = "omniscient"
)

type RetrieveCanonInput struct {
	StoryID     uuid.UUID
	AsOfChapter int
	AsOfDay     *int
	Mode        RetrievalMode
	EventLimit  int
}

type CanonLayer struct {
	Type    string      `json:"type"`
	Entries []CanonEntry `json:"entries"`
}

type CanonEntry struct {
	ID         uuid.UUID              `json:"id"`
	SourceType string                 `json:"source_type"`
	SourceID   uuid.UUID              `json:"source_id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Chapter    int                    `json:"chapter,omitempty"`
	Metadata   map[string]string      `json:"metadata,omitempty"`
}

type StateEntry struct {
	ID                 uuid.UUID              `json:"id"`
	EntityID           uuid.UUID              `json:"entity_id"`
	EntityType         string                 `json:"entity_type"`
	ChapterNumber      int                    `json:"chapter_number"`
	TimelineDay        *int                   `json:"timeline_day,omitempty"`
	TimelineConfidence *string                `json:"timeline_confidence,omitempty"`
	StateData          map[string]interface{} `json:"state_data"`
}

type EventEntry struct {
	ID                 uuid.UUID `json:"id"`
	ChapterNumber      int       `json:"chapter_number"`
	TimelineDay        *int      `json:"timeline_day,omitempty"`
	TimelineConfidence *string   `json:"timeline_confidence,omitempty"`
	Description        string    `json:"description"`
	EntitiesInvolved   []string  `json:"entities_involved,omitempty"`
}

type RetrieveCanonOutput struct {
	StoryID     uuid.UUID     `json:"story_id"`
	AsOfChapter int           `json:"as_of_chapter"`
	AsOfDay     *int          `json:"as_of_day,omitempty"`
	Mode        RetrievalMode `json:"mode"`
	Canon       []CanonEntry  `json:"canon"`
	States      []StateEntry  `json:"states"`
	Events      []EventEntry  `json:"events"`
}

type RetrieveCanonUseCase struct {
	documentRepo      repositories.DocumentRepository
	stateSnapshotRepo repositories.StateSnapshotRepository
	eventRepo         repositories.EventRepository
	logger            *logger.Logger
}

func NewRetrieveCanonUseCase(
	documentRepo repositories.DocumentRepository,
	stateSnapshotRepo repositories.StateSnapshotRepository,
	eventRepo repositories.EventRepository,
	logger *logger.Logger,
) *RetrieveCanonUseCase {
	return &RetrieveCanonUseCase{
		documentRepo:      documentRepo,
		stateSnapshotRepo: stateSnapshotRepo,
		eventRepo:         eventRepo,
		logger:            logger,
	}
}

func (uc *RetrieveCanonUseCase) Execute(ctx context.Context, input RetrieveCanonInput) (*RetrieveCanonOutput, error) {
	if input.Mode == "" {
		input.Mode = ModeRevealedNarrative
	}
	if input.EventLimit <= 0 {
		input.EventLimit = 100
	}

	output := &RetrieveCanonOutput{
		StoryID:     input.StoryID,
		AsOfChapter: input.AsOfChapter,
		AsOfDay:     input.AsOfDay,
		Mode:        input.Mode,
		Canon:       []CanonEntry{},
		States:      []StateEntry{},
		Events:      []EventEntry{},
	}

	// 1. Cold + Warm: canonical documents
	if err := uc.loadCanon(ctx, input, output); err != nil {
		return nil, fmt.Errorf("failed to load canon: %w", err)
	}

	// 2. Hot: state snapshots
	if err := uc.loadStates(ctx, input, output); err != nil {
		return nil, fmt.Errorf("failed to load states: %w", err)
	}

	// 3. Timeline events
	if err := uc.loadEvents(ctx, input, output); err != nil {
		return nil, fmt.Errorf("failed to load events: %w", err)
	}

	uc.logger.Info("Canon retrieved",
		"story_id", input.StoryID,
		"mode", input.Mode,
		"as_of_chapter", input.AsOfChapter,
		"canon", len(output.Canon),
		"states", len(output.States),
		"events", len(output.Events),
	)

	return output, nil
}

func (uc *RetrieveCanonUseCase) loadCanon(ctx context.Context, input RetrieveCanonInput, output *RetrieveCanonOutput) error {
	var asOfChapter int
	switch input.Mode {
	case ModeOmniscient:
		asOfChapter = 999999 // effectively all
	case ModeTimelineConsistent:
		asOfChapter = 999999 // canon is always fully visible in timeline mode
	default: // revealed_narrative
		asOfChapter = input.AsOfChapter
	}

	docs, err := uc.documentRepo.ListCanonical(ctx, input.StoryID, asOfChapter)
	if err != nil {
		return err
	}

	for _, doc := range docs {
		entry := CanonEntry{
			ID:         doc.ID,
			SourceType: string(doc.SourceType),
			SourceID:   doc.SourceID,
			Title:      doc.Title,
			Content:    doc.Content,
			Chapter:    doc.ValidFromChapter,
			Metadata:   doc.Metadata,
		}
		output.Canon = append(output.Canon, entry)
	}
	return nil
}

func (uc *RetrieveCanonUseCase) loadStates(ctx context.Context, input RetrieveCanonInput, output *RetrieveCanonOutput) error {
	var snapshots []*memory.StateSnapshot
	var err error

	switch input.Mode {
	case ModeOmniscient:
		snapshots, err = uc.stateSnapshotRepo.ListByStory(ctx, input.StoryID, 999999)
	case ModeTimelineConsistent:
		if input.AsOfDay != nil {
			snapshots, err = uc.stateSnapshotRepo.ListByStoryAtDay(ctx, input.StoryID, *input.AsOfDay)
		} else {
			// Fallback to chapter filtering when no day specified
			snapshots, err = uc.stateSnapshotRepo.ListByStory(ctx, input.StoryID, input.AsOfChapter)
		}
	default: // revealed_narrative
		snapshots, err = uc.stateSnapshotRepo.ListByStory(ctx, input.StoryID, input.AsOfChapter)
	}

	if err != nil {
		return err
	}

	for _, s := range snapshots {
		output.States = append(output.States, StateEntry{
			ID:                 s.ID,
			EntityID:           s.EntityID,
			EntityType:         s.EntityType,
			ChapterNumber:      s.ChapterNumber,
			TimelineDay:        s.TimelineDay,
			TimelineConfidence: s.TimelineConfidence,
			StateData:          s.StateData,
		})
	}
	return nil
}

func (uc *RetrieveCanonUseCase) loadEvents(ctx context.Context, input RetrieveCanonInput, output *RetrieveCanonOutput) error {
	var events []*memory.Event
	var err error

	switch input.Mode {
	case ModeOmniscient:
		events, err = uc.eventRepo.ListByStory(ctx, input.StoryID, 999999, input.EventLimit)
	case ModeTimelineConsistent:
		if input.AsOfDay != nil {
			events, err = uc.eventRepo.ListByStoryAtDay(ctx, input.StoryID, *input.AsOfDay, input.EventLimit)
		} else {
			events, err = uc.eventRepo.ListByStory(ctx, input.StoryID, input.AsOfChapter, input.EventLimit)
		}
	default: // revealed_narrative
		events, err = uc.eventRepo.ListByStory(ctx, input.StoryID, input.AsOfChapter, input.EventLimit)
	}

	if err != nil {
		return err
	}

	for _, e := range events {
		output.Events = append(output.Events, EventEntry{
			ID:                 e.ID,
			ChapterNumber:      e.ChapterNumber,
			TimelineDay:        e.TimelineDay,
			TimelineConfidence: e.TimelineConfidence,
			Description:        e.Description,
			EntitiesInvolved:   e.EntitiesInvolved,
		})
	}
	return nil
}

