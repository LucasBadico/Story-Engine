package grpc

import (
	"context"

	"github.com/google/uuid"
)

// MainServiceClient defines the interface for communicating with main-service via gRPC
type MainServiceClient interface {
	// GetStory retrieves a story by ID
	GetStory(ctx context.Context, storyID uuid.UUID) (*Story, error)
	
	// GetChapter retrieves a chapter by ID
	GetChapter(ctx context.Context, chapterID uuid.UUID) (*Chapter, error)
	
	// GetScene retrieves a scene by ID
	GetScene(ctx context.Context, sceneID uuid.UUID) (*Scene, error)
	
	// GetBeat retrieves a beat by ID
	GetBeat(ctx context.Context, beatID uuid.UUID) (*Beat, error)
	
	// GetProseBlock retrieves a prose block by ID
	GetProseBlock(ctx context.Context, proseBlockID uuid.UUID) (*ProseBlock, error)
	
	// ListProseBlocksByChapter lists prose blocks for a chapter
	ListProseBlocksByChapter(ctx context.Context, chapterID uuid.UUID) ([]*ProseBlock, error)
}

// Story represents a story from main-service
type Story struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Title     string
	Status    string
	CreatedAt int64
	UpdatedAt int64
}

// Chapter represents a chapter from main-service
type Chapter struct {
	ID        uuid.UUID
	StoryID   uuid.UUID
	Number    int
	Title     string
	Status    string
	CreatedAt int64
	UpdatedAt int64
}

// Scene represents a scene from main-service
type Scene struct {
	ID             uuid.UUID
	StoryID        uuid.UUID
	ChapterID      *uuid.UUID
	OrderNum       int
	Goal           string
	TimeRef        string
	POVCharacterID *uuid.UUID
	LocationID     *uuid.UUID
	CreatedAt      int64
	UpdatedAt      int64
}

// Beat represents a beat from main-service
type Beat struct {
	ID        uuid.UUID
	SceneID   uuid.UUID
	OrderNum  int
	Type      string
	Intent    string
	Outcome   string
	CreatedAt int64
	UpdatedAt int64
}

// ProseBlock represents a prose block from main-service
type ProseBlock struct {
	ID        uuid.UUID
	ChapterID uuid.UUID
	OrderNum  int
	Kind      string
	Content   string
	WordCount int
	CreatedAt int64
	UpdatedAt int64
}

