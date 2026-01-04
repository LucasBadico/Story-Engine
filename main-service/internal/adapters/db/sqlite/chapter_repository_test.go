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

func TestChapterRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)

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
		testChapter, err := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		err = chapterRepo.Create(ctx, testChapter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify chapter can be retrieved
		retrieved, err := chapterRepo.GetByID(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("failed to retrieve chapter: %v", err)
		}

		if retrieved.Title != "Chapter 1" {
			t.Errorf("expected title to be 'Chapter 1', got '%s'", retrieved.Title)
		}

		if retrieved.Number != 1 {
			t.Errorf("expected number to be 1, got %d", retrieved.Number)
		}

		if retrieved.Status != story.ChapterStatusDraft {
			t.Errorf("expected status to be 'draft', got '%s'", retrieved.Status)
		}

		if retrieved.StoryID != testStory.ID {
			t.Errorf("expected story_id to be %s, got %s", testStory.ID, retrieved.StoryID)
		}
	})

	t.Run("creation without title", func(t *testing.T) {
		testChapter, err := story.NewChapter(testTenant.ID, testStory.ID, 2, "")
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		err = chapterRepo.Create(ctx, testChapter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := chapterRepo.GetByID(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("failed to retrieve chapter: %v", err)
		}

		if retrieved.Title != "" {
			t.Errorf("expected title to be empty, got '%s'", retrieved.Title)
		}
	})
}

func TestChapterRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)

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

	t.Run("existing chapter", func(t *testing.T) {
		testChapter, err := story.NewChapter(testTenant.ID, testStory.ID, 1, "GetByID Chapter")
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		err = chapterRepo.Create(ctx, testChapter)
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		retrieved, err := chapterRepo.GetByID(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("failed to get chapter: %v", err)
		}

		if retrieved.ID != testChapter.ID {
			t.Errorf("expected ID to be %s, got %s", testChapter.ID, retrieved.ID)
		}

		if retrieved.Title != "GetByID Chapter" {
			t.Errorf("expected title to be 'GetByID Chapter', got '%s'", retrieved.Title)
		}
	})

	t.Run("non-existent chapter", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := chapterRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent chapter")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "chapter" {
			t.Errorf("expected resource to be 'chapter', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestChapterRepository_ListByStory(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)

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

	t.Run("empty list", func(t *testing.T) {
		chapters, err := chapterRepo.ListByStory(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(chapters) != 0 {
			t.Errorf("expected empty list, got %d chapters", len(chapters))
		}
	})

	t.Run("list with chapters", func(t *testing.T) {
		// Create multiple chapters
		chapterNumbers := []int{1, 2, 3}
		createdChapters := make([]*story.Chapter, 0, len(chapterNumbers))

		for _, num := range chapterNumbers {
			testChapter, err := story.NewChapter(testTenant.ID, testStory.ID, num, "Chapter")
			if err != nil {
				t.Fatalf("failed to create chapter: %v", err)
			}
			err = chapterRepo.Create(ctx, testChapter)
			if err != nil {
				t.Fatalf("failed to create chapter: %v", err)
			}
			createdChapters = append(createdChapters, testChapter)
		}

		chapters, err := chapterRepo.ListByStory(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(chapters) != len(chapterNumbers) {
			t.Errorf("expected %d chapters, got %d", len(chapterNumbers), len(chapters))
		}

		// Verify ordering (by number ASC)
		for i := 0; i < len(chapters)-1; i++ {
			if chapters[i].Number > chapters[i+1].Number {
				t.Error("expected chapters to be ordered by number ascending")
			}
		}
	})
}

func TestChapterRepository_ListByStoryOrdered(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)

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

	t.Run("list ordered by number", func(t *testing.T) {
		// Create chapters in non-sequential order
		chapter3, _ := story.NewChapter(testTenant.ID, testStory.ID, 3, "Chapter 3")
		chapter1, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapter2, _ := story.NewChapter(testTenant.ID, testStory.ID, 2, "Chapter 2")
		chapterRepo.Create(ctx, chapter3)
		chapterRepo.Create(ctx, chapter1)
		chapterRepo.Create(ctx, chapter2)

		chapters, err := chapterRepo.ListByStoryOrdered(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(chapters) != 3 {
			t.Errorf("expected 3 chapters, got %d", len(chapters))
		}

		// Verify ordering (by number ASC)
		for i := 0; i < len(chapters)-1; i++ {
			if chapters[i].Number > chapters[i+1].Number {
				t.Error("expected chapters to be ordered by number ascending")
			}
		}

		// Verify correct order
		if chapters[0].Number != 1 {
			t.Errorf("expected first chapter to be number 1, got %d", chapters[0].Number)
		}
		if chapters[1].Number != 2 {
			t.Errorf("expected second chapter to be number 2, got %d", chapters[1].Number)
		}
		if chapters[2].Number != 3 {
			t.Errorf("expected third chapter to be number 3, got %d", chapters[2].Number)
		}
	})
}

func TestChapterRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)

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
		testChapter, err := story.NewChapter(testTenant.ID, testStory.ID, 1, "Update Chapter")
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		err = chapterRepo.Create(ctx, testChapter)
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		// Update title
		testChapter.UpdateTitle("Updated Title")

		// Update status
		err = testChapter.UpdateStatus(story.ChapterStatusPublished)
		if err != nil {
			t.Fatalf("failed to update status: %v", err)
		}

		// Update number
		testChapter.Number = 2

		err = chapterRepo.Update(ctx, testChapter)
		if err != nil {
			t.Fatalf("failed to update chapter: %v", err)
		}

		// Verify update
		retrieved, err := chapterRepo.GetByID(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("failed to get chapter: %v", err)
		}

		if retrieved.Title != "Updated Title" {
			t.Errorf("expected title to be 'Updated Title', got '%s'", retrieved.Title)
		}

		if retrieved.Status != story.ChapterStatusPublished {
			t.Errorf("expected status to be 'published', got '%s'", retrieved.Status)
		}

		if retrieved.Number != 2 {
			t.Errorf("expected number to be 2, got %d", retrieved.Number)
		}
	})
}

func TestChapterRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)

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
		testChapter, err := story.NewChapter(testTenant.ID, testStory.ID, 1, "Delete Chapter")
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		err = chapterRepo.Create(ctx, testChapter)
		if err != nil {
			t.Fatalf("failed to create chapter: %v", err)
		}

		err = chapterRepo.Delete(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("failed to delete chapter: %v", err)
		}

		// Verify chapter is deleted
		_, err = chapterRepo.GetByID(ctx, testTenant.ID, testChapter.ID)
		if err == nil {
			t.Fatal("expected error for deleted chapter")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "chapter" {
			t.Errorf("expected resource to be 'chapter', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestChapterRepository_DeleteByStory(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)

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

	t.Run("delete all chapters for story", func(t *testing.T) {
		// Create multiple chapters
		chapter1, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapter2, _ := story.NewChapter(testTenant.ID, testStory.ID, 2, "Chapter 2")
		chapterRepo.Create(ctx, chapter1)
		chapterRepo.Create(ctx, chapter2)

		// Verify they exist
		chapters, err := chapterRepo.ListByStory(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(chapters) != 2 {
			t.Fatalf("expected 2 chapters, got %d", len(chapters))
		}

		// Delete all
		err = chapterRepo.DeleteByStory(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("failed to delete by story: %v", err)
		}

		// Verify all deleted
		chapters, err = chapterRepo.ListByStory(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(chapters) != 0 {
			t.Errorf("expected no chapters, got %d", len(chapters))
		}
	})
}

