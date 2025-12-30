package versioning

import (
	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
)

// CloneResult holds the result of a story clone operation
type CloneResult struct {
	NewStoryID      uuid.UUID
	NewChapterIDs   map[uuid.UUID]uuid.UUID // old chapter ID -> new chapter ID
	NewSceneIDs     map[uuid.UUID]uuid.UUID // old scene ID -> new scene ID
	NewBeatIDs      map[uuid.UUID]uuid.UUID // old beat ID -> new beat ID
	NewProseBlockIDs map[uuid.UUID]uuid.UUID // old prose block ID -> new prose block ID
}

// CloneStory creates a deep copy of a story for versioning
func CloneStory(sourceStory *story.Story, versionNumber int) (*story.Story, error) {
	if sourceStory == nil {
		return nil, ErrSourceStoryRequired
	}
	if versionNumber < 1 {
		return nil, ErrInvalidVersionNumber
	}

	newStory := &story.Story{
		ID:              uuid.New(),
		TenantID:        sourceStory.TenantID,
		Title:           sourceStory.Title,
		Status:          sourceStory.Status,
		VersionNumber:   versionNumber,
		RootStoryID:     sourceStory.RootStoryID, // Keep same root
		PreviousStoryID: &sourceStory.ID,          // Point to source
		CreatedByUserID: sourceStory.CreatedByUserID,
		CreatedAt:       sourceStory.CreatedAt,
		UpdatedAt:       sourceStory.UpdatedAt,
	}

	return newStory, nil
}

// CloneChapter creates a deep copy of a chapter
func CloneChapter(sourceChapter *story.Chapter, newStoryID uuid.UUID) *story.Chapter {
	return &story.Chapter{
		ID:        uuid.New(),
		StoryID:   newStoryID,
		Number:    sourceChapter.Number,
		Title:     sourceChapter.Title,
		Status:    sourceChapter.Status,
		CreatedAt: sourceChapter.CreatedAt,
		UpdatedAt: sourceChapter.UpdatedAt,
	}
}

// CloneScene creates a deep copy of a scene
func CloneScene(sourceScene *story.Scene, newStoryID, newChapterID uuid.UUID) *story.Scene {
	return &story.Scene{
		ID:             uuid.New(),
		StoryID:        newStoryID,
		ChapterID:      newChapterID,
		OrderNum:       sourceScene.OrderNum,
		POVCharacterID: sourceScene.POVCharacterID,
		LocationID:     sourceScene.LocationID,
		TimeRef:        sourceScene.TimeRef,
		Goal:           sourceScene.Goal,
		CreatedAt:      sourceScene.CreatedAt,
		UpdatedAt:      sourceScene.UpdatedAt,
	}
}

// CloneBeat creates a deep copy of a beat
func CloneBeat(sourceBeat *story.Beat, newSceneID uuid.UUID) *story.Beat {
	return &story.Beat{
		ID:        uuid.New(),
		SceneID:   newSceneID,
		OrderNum:  sourceBeat.OrderNum,
		Type:      sourceBeat.Type,
		Intent:    sourceBeat.Intent,
		Outcome:   sourceBeat.Outcome,
		CreatedAt: sourceBeat.CreatedAt,
		UpdatedAt: sourceBeat.UpdatedAt,
	}
}

// CloneProseBlock creates a deep copy of a prose block
func CloneProseBlock(sourceProse *story.ProseBlock, newSceneID uuid.UUID) *story.ProseBlock {
	return &story.ProseBlock{
		ID:        uuid.New(),
		SceneID:   newSceneID,
		Kind:      sourceProse.Kind,
		Content:   sourceProse.Content,
		WordCount: sourceProse.WordCount,
		CreatedAt: sourceProse.CreatedAt,
		UpdatedAt: sourceProse.UpdatedAt,
	}
}

