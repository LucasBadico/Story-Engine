package importer

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/core/world"
)

// MainServiceAPI abstracts all main service operations the importer needs.
// Backed by an HTTP client that calls the main service REST API.
type MainServiceAPI interface {
	// Listing (for dedup)
	ListCharacters(ctx context.Context, worldID uuid.UUID) ([]*world.Character, error)
	ListLocations(ctx context.Context, worldID uuid.UUID) ([]*world.Location, error)
	ListArtifacts(ctx context.Context, worldID uuid.UUID) ([]*world.Artifact, error)
	ListChapters(ctx context.Context, storyID uuid.UUID) ([]*story.Chapter, error)

	// World entity creation
	CreateCharacter(ctx context.Context, input APICreateCharacterInput) (*world.Character, error)
	CreateLocation(ctx context.Context, input APICreateLocationInput) (*world.Location, error)
	CreateArtifact(ctx context.Context, input APICreateArtifactInput) (*world.Artifact, error)
	CreateCharacterState(ctx context.Context, input APICreateCharacterStateInput) (*world.CharacterState, error)
	CreateRelationshipState(ctx context.Context, input APICreateRelationshipStateInput) (*world.RelationshipState, error)

	// Story entity creation
	CreateChapter(ctx context.Context, input APICreateChapterInput) (*story.Chapter, error)
	CreateScene(ctx context.Context, input APICreateSceneInput) (*story.Scene, error)
	CreateBeat(ctx context.Context, input APICreateBeatInput) (*story.Beat, error)
	CreateContentBlock(ctx context.Context, input APICreateContentBlockInput) (*story.ContentBlock, error)
	CreateContentAnchor(ctx context.Context, contentBlockID uuid.UUID, entityType string, entityID uuid.UUID) (*story.ContentAnchor, error)
	CreateTimelineEvent(ctx context.Context, input APICreateTimelineEventInput) (*story.TimelineEvent, error)
	CreatePlot(ctx context.Context, input APICreatePlotInput) (*story.Plot, error)
	CreatePlotDevelopment(ctx context.Context, input APICreatePlotDevelopmentInput) (*story.PlotDevelopment, error)
}

// Input types — mirror the HTTP request bodies

type APICreateCharacterInput struct {
	WorldID     uuid.UUID
	Name        string
	Description string
}

type APICreateLocationInput struct {
	WorldID     uuid.UUID
	Name        string
	Description string
}

type APICreateArtifactInput struct {
	WorldID     uuid.UUID
	Name        string
	Description string
}

type APICreateCharacterStateInput struct {
	CharacterID        uuid.UUID
	ChapterID          *uuid.UUID
	ChapterNumber      int
	Emotion            string
	Goal               string
	Secrets            []string
	Clothes            string
	PhysicalState      string
	TimelineDay        *int
	TimelineConfidence *string
}

type APICreateRelationshipStateInput struct {
	ChapterID          uuid.UUID
	ChapterNumber      int
	TensionDescription string
	TimelineDay        *int
	TimelineConfidence *string
	Entities           []APIEntityLink
}

type APIEntityLink struct {
	Role       string `json:"role"`
	EntityType string `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
}

type APICreateChapterInput struct {
	StoryID uuid.UUID
	Number  int
	Title   string
	Status  string
}

type APICreateSceneInput struct {
	StoryID   uuid.UUID
	ChapterID *uuid.UUID
	OrderNum  int
	Goal      string
}

type APICreateBeatInput struct {
	SceneID  uuid.UUID
	OrderNum int
	Type     story.BeatType
	Intent   string
}

type APICreateContentBlockInput struct {
	ChapterID *uuid.UUID
	OrderNum  *int
	Type      story.ContentType
	Kind      story.ContentKind
	Content   string
}

type APICreateTimelineEventInput struct {
	StoryID            uuid.UUID
	ChapterID          uuid.UUID
	ChapterNumber      int
	TimelineDay        *int
	TimelineConfidence *string
	Description        string
}

type APICreatePlotInput struct {
	StoryID     uuid.UUID
	Title       string
	Type        string
	Description string
}

type APICreatePlotDevelopmentInput struct {
	PlotID        uuid.UUID
	ChapterNumber int
	Action        string
	Description   string
}
