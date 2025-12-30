//go:build integration

package story

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestCloneStoryUseCase_Execute(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := postgres.TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	// Setup repositories
	tenantRepo := postgres.NewTenantRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	beatRepo := postgres.NewBeatRepository(db)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	log := logger.New()

	// Create a tenant
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	tenantOutput, err := createTenantUseCase.Execute(ctx, tenant.CreateTenantInput{
		Name:      "Test Tenant",
		CreatedBy: nil,
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	// Create a story
	createStoryUseCase := NewCreateStoryUseCase(storyRepo, tenantRepo, auditLogRepo, log)
	storyOutput, err := createStoryUseCase.Execute(ctx, CreateStoryInput{
		TenantID:       tenantOutput.Tenant.ID,
		Title:          "Test Story",
		CreatedByUserID: nil,
	})
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	cloneUseCase := NewCloneStoryUseCase(
		storyRepo, chapterRepo, sceneRepo, beatRepo, proseBlockRepo,
		auditLogRepo, transactionRepo, log,
	)

	t.Run("clone simple story (no chapters)", func(t *testing.T) {
		input := CloneStoryInput{
			SourceStoryID:  storyOutput.Story.ID,
			CreatedByUserID: nil,
		}

		output, err := cloneUseCase.Execute(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if output.NewStoryID == storyOutput.Story.ID {
			t.Error("new story should have different ID")
		}

		// Verify new story exists
		newStory, err := storyRepo.GetByID(ctx, output.NewStoryID)
		if err != nil {
			t.Fatalf("failed to retrieve cloned story: %v", err)
		}

		if newStory.Title != storyOutput.Story.Title {
			t.Errorf("expected title to match, got '%s'", newStory.Title)
		}

		if newStory.VersionNumber != 2 {
			t.Errorf("expected version number to be 2, got %d", newStory.VersionNumber)
		}

		if newStory.RootStoryID != storyOutput.Story.RootStoryID {
			t.Error("root_story_id should match original")
		}

		if newStory.PreviousStoryID == nil || *newStory.PreviousStoryID != storyOutput.Story.ID {
			t.Error("previous_story_id should point to source story")
		}
	})

	t.Run("clone story with full hierarchy", func(t *testing.T) {
		// Create a story with chapters, scenes, beats, and prose
		sourceStory, err := story.NewStory(tenantOutput.Tenant.ID, "Full Story", nil)
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}
		if err := storyRepo.Create(ctx, sourceStory); err != nil {
			t.Fatalf("failed to save story: %v", err)
		}

		// Create chapter
		chapter, err := story.NewChapter(sourceStory.ID, 1, "Chapter 1")
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}
		if err := chapterRepo.Create(ctx, chapter); err != nil {
			t.Fatalf("failed to save chapter: %v", err)
		}

		// Create scene
		scene, err := story.NewScene(sourceStory.ID, chapter.ID, 1)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}
		if err := sceneRepo.Create(ctx, scene); err != nil {
			t.Fatalf("failed to save scene: %v", err)
		}

		// Create beat
		beat, err := story.NewBeat(scene.ID, 1, story.BeatTypeSetup)
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}
		if err := beatRepo.Create(ctx, beat); err != nil {
			t.Fatalf("failed to save beat: %v", err)
		}

		// Create prose block
		prose, err := story.NewProseBlock(scene.ID, story.ProseKindFinal, "This is the prose content.")
		if err != nil {
			t.Fatalf("failed to create prose block: %v", err)
		}
		if err := proseBlockRepo.Create(ctx, prose); err != nil {
			t.Fatalf("failed to save prose block: %v", err)
		}

		// Clone the story
		input := CloneStoryInput{
			SourceStoryID:  sourceStory.ID,
			CreatedByUserID: nil,
		}

		output, err := cloneUseCase.Execute(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify all entities were copied
		newStory, err := storyRepo.GetByID(ctx, output.NewStoryID)
		if err != nil {
			t.Fatalf("failed to retrieve cloned story: %v", err)
		}

		// Verify chapters
		newChapters, err := chapterRepo.ListByStory(ctx, newStory.ID)
		if err != nil {
			t.Fatalf("failed to list chapters: %v", err)
		}
		if len(newChapters) != 1 {
			t.Errorf("expected 1 chapter, got %d", len(newChapters))
		}
		if newChapters[0].Number != chapter.Number {
			t.Errorf("expected chapter number to match")
		}

		// Verify scenes
		newScenes, err := sceneRepo.ListByStory(ctx, newStory.ID)
		if err != nil {
			t.Fatalf("failed to list scenes: %v", err)
		}
		if len(newScenes) != 1 {
			t.Errorf("expected 1 scene, got %d", len(newScenes))
		}
		if newScenes[0].OrderNum != scene.OrderNum {
			t.Errorf("expected scene order to match")
		}

		// Verify beats
		newBeats, err := beatRepo.ListByScene(ctx, newScenes[0].ID)
		if err != nil {
			t.Fatalf("failed to list beats: %v", err)
		}
		if len(newBeats) != 1 {
			t.Errorf("expected 1 beat, got %d", len(newBeats))
		}
		if newBeats[0].Type != beat.Type {
			t.Errorf("expected beat type to match")
		}

		// Verify prose blocks
		newProseBlocks, err := proseBlockRepo.ListByScene(ctx, newScenes[0].ID)
		if err != nil {
			t.Fatalf("failed to list prose blocks: %v", err)
		}
		if len(newProseBlocks) != 1 {
			t.Errorf("expected 1 prose block, got %d", len(newProseBlocks))
		}
		if newProseBlocks[0].Content != prose.Content {
			t.Errorf("expected prose content to match")
		}

		// Verify versioning fields
		if newStory.RootStoryID != sourceStory.RootStoryID {
			t.Error("root_story_id should match original")
		}

		if newStory.PreviousStoryID == nil || *newStory.PreviousStoryID != sourceStory.ID {
			t.Error("previous_story_id should point to source story")
		}
	})
}

