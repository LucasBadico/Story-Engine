//go:build integration

package sqlite

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/core/tenant"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
)

func TestSceneRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	sceneRepo := NewSceneRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		testScene, err := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		err = sceneRepo.Create(ctx, testScene)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify scene can be retrieved
		retrieved, err := sceneRepo.GetByID(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("failed to retrieve scene: %v", err)
		}

		if retrieved.OrderNum != 1 {
			t.Errorf("expected order_num to be 1, got %d", retrieved.OrderNum)
		}

		if retrieved.StoryID != testStory.ID {
			t.Errorf("expected story_id to be %s, got %s", testStory.ID, retrieved.StoryID)
		}
	})

	t.Run("creation with chapter", func(t *testing.T) {
		testChapter, err := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}
		err = chapterRepo.Create(ctx, testChapter)
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		testScene, err := story.NewScene(testTenant.ID, testStory.ID, &testChapter.ID, 2)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}
		testScene.UpdateGoal("Test goal")
		testScene.TimeRef = "Afternoon"

		err = sceneRepo.Create(ctx, testScene)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := sceneRepo.GetByID(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("failed to retrieve scene: %v", err)
		}

		if retrieved.ChapterID == nil || *retrieved.ChapterID != testChapter.ID {
			t.Errorf("expected chapter_id to be %s, got %v", testChapter.ID, retrieved.ChapterID)
		}

		if retrieved.Goal != "Test goal" {
			t.Errorf("expected goal to be 'Test goal', got '%s'", retrieved.Goal)
		}

		if retrieved.TimeRef != "Afternoon" {
			t.Errorf("expected time_ref to be 'Afternoon', got '%s'", retrieved.TimeRef)
		}
	})
}

func TestSceneRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	sceneRepo := NewSceneRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("existing scene", func(t *testing.T) {
		testScene, err := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		err = sceneRepo.Create(ctx, testScene)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		retrieved, err := sceneRepo.GetByID(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("failed to get scene: %v", err)
		}

		if retrieved.ID != testScene.ID {
			t.Errorf("expected ID to be %s, got %s", testScene.ID, retrieved.ID)
		}
	})

	t.Run("non-existent scene", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := sceneRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent scene")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "scene" {
			t.Errorf("expected resource to be 'scene', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestSceneRepository_ListByChapter(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	sceneRepo := NewSceneRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("list with scenes", func(t *testing.T) {
		testChapter, err := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}
		err = chapterRepo.Create(ctx, testChapter)
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		// Create multiple scenes
		scene1, _ := story.NewScene(testTenant.ID, testStory.ID, &testChapter.ID, 1)
		scene2, _ := story.NewScene(testTenant.ID, testStory.ID, &testChapter.ID, 2)
		sceneRepo.Create(ctx, scene1)
		sceneRepo.Create(ctx, scene2)

		scenes, err := sceneRepo.ListByChapter(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(scenes) != 2 {
			t.Errorf("expected 2 scenes, got %d", len(scenes))
		}

		// Verify ordering (by order_num ASC)
		for i := 0; i < len(scenes)-1; i++ {
			if scenes[i].OrderNum > scenes[i+1].OrderNum {
				t.Error("expected scenes to be ordered by order_num ascending")
			}
		}
	})

	t.Run("empty list", func(t *testing.T) {
		testChapter, err := story.NewChapter(testTenant.ID, testStory.ID, 2, "Chapter 2")
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}
		err = chapterRepo.Create(ctx, testChapter)
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		scenes, err := sceneRepo.ListByChapter(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(scenes) != 0 {
			t.Errorf("expected empty list, got %d scenes", len(scenes))
		}
	})
}

func TestSceneRepository_ListByChapterOrdered(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	sceneRepo := NewSceneRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("list ordered by order_num", func(t *testing.T) {
		testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapterRepo.Create(ctx, testChapter)

		// Create scenes in non-sequential order
		scene3, _ := story.NewScene(testTenant.ID, testStory.ID, &testChapter.ID, 3)
		scene1, _ := story.NewScene(testTenant.ID, testStory.ID, &testChapter.ID, 1)
		scene2, _ := story.NewScene(testTenant.ID, testStory.ID, &testChapter.ID, 2)
		sceneRepo.Create(ctx, scene3)
		sceneRepo.Create(ctx, scene1)
		sceneRepo.Create(ctx, scene2)

		scenes, err := sceneRepo.ListByChapterOrdered(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(scenes) != 3 {
			t.Errorf("expected 3 scenes, got %d", len(scenes))
		}

		// Verify ordering (by order_num ASC)
		for i := 0; i < len(scenes)-1; i++ {
			if scenes[i].OrderNum > scenes[i+1].OrderNum {
				t.Error("expected scenes to be ordered by order_num ascending")
			}
		}

		// Verify correct order
		if scenes[0].OrderNum != 1 {
			t.Errorf("expected first scene to be order_num 1, got %d", scenes[0].OrderNum)
		}
		if scenes[1].OrderNum != 2 {
			t.Errorf("expected second scene to be order_num 2, got %d", scenes[1].OrderNum)
		}
		if scenes[2].OrderNum != 3 {
			t.Errorf("expected third scene to be order_num 3, got %d", scenes[2].OrderNum)
		}
	})
}

func TestSceneRepository_ListByStory(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	sceneRepo := NewSceneRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("list by story", func(t *testing.T) {
		// Create chapters
		chapter1, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapter2, _ := story.NewChapter(testTenant.ID, testStory.ID, 2, "Chapter 2")
		chapterRepo.Create(ctx, chapter1)
		chapterRepo.Create(ctx, chapter2)

		// Create scenes in different chapters
		scene1, _ := story.NewScene(testTenant.ID, testStory.ID, &chapter1.ID, 1)
		scene2, _ := story.NewScene(testTenant.ID, testStory.ID, &chapter1.ID, 2)
		scene3, _ := story.NewScene(testTenant.ID, testStory.ID, &chapter2.ID, 1)
		sceneRepo.Create(ctx, scene1)
		sceneRepo.Create(ctx, scene2)
		sceneRepo.Create(ctx, scene3)

		scenes, err := sceneRepo.ListByStory(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(scenes) != 3 {
			t.Errorf("expected 3 scenes, got %d", len(scenes))
		}

		// Verify ordering (by chapter_id, then order_num)
		for i := 0; i < len(scenes)-1; i++ {
			if scenes[i].ChapterID != nil && scenes[i+1].ChapterID != nil {
				chapter1ID := scenes[i].ChapterID.String()
				chapter2ID := scenes[i+1].ChapterID.String()
				if chapter1ID > chapter2ID {
					t.Error("expected scenes to be ordered by chapter_id")
				}
				if chapter1ID == chapter2ID && scenes[i].OrderNum > scenes[i+1].OrderNum {
					t.Error("expected scenes within same chapter to be ordered by order_num")
				}
			}
		}
	})
}

func TestSceneRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	sceneRepo := NewSceneRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		testScene, err := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		err = sceneRepo.Create(ctx, testScene)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		// Update
		testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapterRepo.Create(ctx, testChapter)

		testScene.UpdateChapter(&testChapter.ID)
		testScene.UpdateGoal("Updated goal")
		testScene.TimeRef = "Updated time"
		testScene.OrderNum = 2

		err = sceneRepo.Update(ctx, testScene)
		if err != nil {
			t.Fatalf("failed to update scene: %v", err)
		}

		// Verify update
		retrieved, err := sceneRepo.GetByID(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("failed to get scene: %v", err)
		}

		if retrieved.ChapterID == nil || *retrieved.ChapterID != testChapter.ID {
			t.Errorf("expected chapter_id to be %s, got %v", testChapter.ID, retrieved.ChapterID)
		}

		if retrieved.Goal != "Updated goal" {
			t.Errorf("expected goal to be 'Updated goal', got '%s'", retrieved.Goal)
		}

		if retrieved.TimeRef != "Updated time" {
			t.Errorf("expected time_ref to be 'Updated time', got '%s'", retrieved.TimeRef)
		}

		if retrieved.OrderNum != 2 {
			t.Errorf("expected order_num to be 2, got %d", retrieved.OrderNum)
		}
	})
}

func TestSceneRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	sceneRepo := NewSceneRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		testScene, err := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		err = sceneRepo.Create(ctx, testScene)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		err = sceneRepo.Delete(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("failed to delete scene: %v", err)
		}

		// Verify scene is deleted
		_, err = sceneRepo.GetByID(ctx, testTenant.ID, testScene.ID)
		if err == nil {
			t.Fatal("expected error for deleted scene")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "scene" {
			t.Errorf("expected resource to be 'scene', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestSceneRepository_DeleteByChapter(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	sceneRepo := NewSceneRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("delete all scenes for chapter", func(t *testing.T) {
		testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapterRepo.Create(ctx, testChapter)

		// Create multiple scenes
		scene1, _ := story.NewScene(testTenant.ID, testStory.ID, &testChapter.ID, 1)
		scene2, _ := story.NewScene(testTenant.ID, testStory.ID, &testChapter.ID, 2)
		sceneRepo.Create(ctx, scene1)
		sceneRepo.Create(ctx, scene2)

		// Verify they exist
		scenes, err := sceneRepo.ListByChapter(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(scenes) != 2 {
			t.Fatalf("expected 2 scenes, got %d", len(scenes))
		}

		// Delete all
		err = sceneRepo.DeleteByChapter(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("failed to delete by chapter: %v", err)
		}

		// Verify all deleted
		scenes, err = sceneRepo.ListByChapter(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(scenes) != 0 {
			t.Errorf("expected no scenes, got %d", len(scenes))
		}
	})
}

func TestSceneRepository_DeleteByStory(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	sceneRepo := NewSceneRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("delete all scenes for story", func(t *testing.T) {
		testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapterRepo.Create(ctx, testChapter)

		// Create multiple scenes
		scene1, _ := story.NewScene(testTenant.ID, testStory.ID, &testChapter.ID, 1)
		scene2, _ := story.NewScene(testTenant.ID, testStory.ID, &testChapter.ID, 2)
		sceneRepo.Create(ctx, scene1)
		sceneRepo.Create(ctx, scene2)

		// Verify they exist
		scenes, err := sceneRepo.ListByStory(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(scenes) != 2 {
			t.Fatalf("expected 2 scenes, got %d", len(scenes))
		}

		// Delete all
		err = sceneRepo.DeleteByStory(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("failed to delete by story: %v", err)
		}

		// Verify all deleted
		scenes, err = sceneRepo.ListByStory(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(scenes) != 0 {
			t.Errorf("expected no scenes, got %d", len(scenes))
		}
	})
}

