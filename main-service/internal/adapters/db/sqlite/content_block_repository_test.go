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

func TestContentBlockRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)

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
			t.Fatalf("failed to create chapter: %v", err)
		}

		orderNum := 1
		testContentBlock, err := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "Test content")
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		err = contentBlockRepo.Create(ctx, testContentBlock)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify content block can be retrieved
		retrieved, err := contentBlockRepo.GetByID(ctx, testTenant.ID, testContentBlock.ID)
		if err != nil {
			t.Fatalf("failed to retrieve content block: %v", err)
		}

		if retrieved.Content != "Test content" {
			t.Errorf("expected content to be 'Test content', got '%s'", retrieved.Content)
		}

		if retrieved.Type != story.ContentTypeText {
			t.Errorf("expected type to be 'text', got '%s'", retrieved.Type)
		}

		if retrieved.Kind != story.ContentKindFinal {
			t.Errorf("expected kind to be 'final', got '%s'", retrieved.Kind)
		}

		if retrieved.OrderNum == nil || *retrieved.OrderNum != 1 {
			t.Errorf("expected order_num to be 1, got %v", retrieved.OrderNum)
		}
	})

	t.Run("creation without chapter", func(t *testing.T) {
		testContentBlock, err := story.NewContentBlock(testTenant.ID, nil, nil, story.ContentTypeText, story.ContentKindDraft, "Standalone content")
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		err = contentBlockRepo.Create(ctx, testContentBlock)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := contentBlockRepo.GetByID(ctx, testTenant.ID, testContentBlock.ID)
		if err != nil {
			t.Fatalf("failed to retrieve content block: %v", err)
		}

		if retrieved.ChapterID != nil {
			t.Error("expected chapter_id to be nil, got non-nil")
		}
	})
}

func TestContentBlockRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)

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

	t.Run("existing content block", func(t *testing.T) {
		testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapterRepo.Create(ctx, testChapter)

		orderNum := 1
		testContentBlock, err := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "GetByID Content")
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		err = contentBlockRepo.Create(ctx, testContentBlock)
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		retrieved, err := contentBlockRepo.GetByID(ctx, testTenant.ID, testContentBlock.ID)
		if err != nil {
			t.Fatalf("failed to get content block: %v", err)
		}

		if retrieved.ID != testContentBlock.ID {
			t.Errorf("expected ID to be %s, got %s", testContentBlock.ID, retrieved.ID)
		}

		if retrieved.Content != "GetByID Content" {
			t.Errorf("expected content to be 'GetByID Content', got '%s'", retrieved.Content)
		}
	})

	t.Run("non-existent content block", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := contentBlockRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent content block")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "content_block" {
			t.Errorf("expected resource to be 'content_block', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestContentBlockRepository_ListByChapter(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)

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

	t.Run("list with content blocks", func(t *testing.T) {
		testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapterRepo.Create(ctx, testChapter)

		// Create multiple content blocks
		orderNum1 := 1
		orderNum2 := 2
		block1, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum1, story.ContentTypeText, story.ContentKindFinal, "Content 1")
		block2, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum2, story.ContentTypeText, story.ContentKindFinal, "Content 2")
		contentBlockRepo.Create(ctx, block1)
		contentBlockRepo.Create(ctx, block2)

		blocks, err := contentBlockRepo.ListByChapter(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(blocks) != 2 {
			t.Errorf("expected 2 content blocks, got %d", len(blocks))
		}
	})

	t.Run("empty list", func(t *testing.T) {
		testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 2, "Chapter 2")
		chapterRepo.Create(ctx, testChapter)

		blocks, err := contentBlockRepo.ListByChapter(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(blocks) != 0 {
			t.Errorf("expected empty list, got %d content blocks", len(blocks))
		}
	})
}

func TestContentBlockRepository_GetByChapterAndKind(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)

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

	t.Run("get by chapter and kind", func(t *testing.T) {
		testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapterRepo.Create(ctx, testChapter)

		orderNum := 1
		testContentBlock, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "Final Content")
		contentBlockRepo.Create(ctx, testContentBlock)

		retrieved, err := contentBlockRepo.GetByChapterAndKind(ctx, testTenant.ID, testChapter.ID, story.ContentKindFinal)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if retrieved.Kind != story.ContentKindFinal {
			t.Errorf("expected kind to be 'final', got '%s'", retrieved.Kind)
		}

		if retrieved.Content != "Final Content" {
			t.Errorf("expected content to be 'Final Content', got '%s'", retrieved.Content)
		}
	})
}

func TestContentBlockRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)

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
		testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapterRepo.Create(ctx, testChapter)

		orderNum := 1
		testContentBlock, err := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "Update Content")
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		err = contentBlockRepo.Create(ctx, testContentBlock)
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		// Update
		err = testContentBlock.UpdateContent("Updated Content")
		if err != nil {
			t.Fatalf("failed to update content: %v", err)
		}

		orderNum2 := 2
		testContentBlock.OrderNum = &orderNum2

		err = contentBlockRepo.Update(ctx, testContentBlock)
		if err != nil {
			t.Fatalf("failed to update content block: %v", err)
		}

		// Verify update
		retrieved, err := contentBlockRepo.GetByID(ctx, testTenant.ID, testContentBlock.ID)
		if err != nil {
			t.Fatalf("failed to get content block: %v", err)
		}

		if retrieved.Content != "Updated Content" {
			t.Errorf("expected content to be 'Updated Content', got '%s'", retrieved.Content)
		}

		if retrieved.OrderNum == nil || *retrieved.OrderNum != 2 {
			t.Errorf("expected order_num to be 2, got %v", retrieved.OrderNum)
		}
	})
}

func TestContentBlockRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)

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
		testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapterRepo.Create(ctx, testChapter)

		orderNum := 1
		testContentBlock, err := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "Delete Content")
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		err = contentBlockRepo.Create(ctx, testContentBlock)
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		err = contentBlockRepo.Delete(ctx, testTenant.ID, testContentBlock.ID)
		if err != nil {
			t.Fatalf("failed to delete content block: %v", err)
		}

		// Verify content block is deleted
		_, err = contentBlockRepo.GetByID(ctx, testTenant.ID, testContentBlock.ID)
		if err == nil {
			t.Fatal("expected error for deleted content block")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "content_block" {
			t.Errorf("expected resource to be 'content_block', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestContentBlockRepository_DeleteByChapter(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)

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

	t.Run("delete all content blocks for chapter", func(t *testing.T) {
		testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
		chapterRepo.Create(ctx, testChapter)

		// Create multiple content blocks
		orderNum1 := 1
		orderNum2 := 2
		block1, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum1, story.ContentTypeText, story.ContentKindFinal, "Content 1")
		block2, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum2, story.ContentTypeText, story.ContentKindFinal, "Content 2")
		contentBlockRepo.Create(ctx, block1)
		contentBlockRepo.Create(ctx, block2)

		// Verify they exist
		blocks, err := contentBlockRepo.ListByChapter(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(blocks) != 2 {
			t.Fatalf("expected 2 content blocks, got %d", len(blocks))
		}

		// Delete all
		err = contentBlockRepo.DeleteByChapter(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("failed to delete by chapter: %v", err)
		}

		// Verify all deleted
		blocks, err = contentBlockRepo.ListByChapter(ctx, testTenant.ID, testChapter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(blocks) != 0 {
			t.Errorf("expected no content blocks, got %d", len(blocks))
		}
	})
}

