//go:build integration

package sqlite

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/core/tenant"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
)

func TestStoryRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("successful creation root story", func(t *testing.T) {
		testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		err = storyRepo.Create(ctx, testStory)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify story can be retrieved
		retrieved, err := storyRepo.GetByID(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("failed to retrieve story: %v", err)
		}

		if retrieved.Title != "Test Story" {
			t.Errorf("expected title to be 'Test Story', got '%s'", retrieved.Title)
		}

		if retrieved.Status != story.StoryStatusDraft {
			t.Errorf("expected status to be 'draft', got '%s'", retrieved.Status)
		}

		if retrieved.VersionNumber != 1 {
			t.Errorf("expected version_number to be 1, got %d", retrieved.VersionNumber)
		}

		if retrieved.RootStoryID != testStory.ID {
			t.Errorf("expected root_story_id to be %s, got %s", testStory.ID, retrieved.RootStoryID)
		}

		if retrieved.PreviousStoryID != nil {
			t.Error("expected previous_story_id to be nil, got non-nil")
		}

		if !retrieved.IsRoot() {
			t.Error("expected story to be root")
		}

		if !retrieved.IsFirstVersion() {
			t.Error("expected story to be first version")
		}
	})

	t.Run("creation with world and user", func(t *testing.T) {
		worldRepo := NewWorldRepository(db)
		testWorld, err := world.NewWorld(testTenant.ID, "Test World", false)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}
		err = worldRepo.Create(ctx, testWorld)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		userID := uuid.New()
		testStory, err := story.NewStory(testTenant.ID, "Story With World", &userID)
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}
		testStory.WorldID = &testWorld.ID

		err = storyRepo.Create(ctx, testStory)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := storyRepo.GetByID(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("failed to retrieve story: %v", err)
		}

		if retrieved.WorldID == nil || *retrieved.WorldID != testWorld.ID {
			t.Errorf("expected world_id to be %s, got %v", testWorld.ID, retrieved.WorldID)
		}

		if retrieved.CreatedByUserID == nil || *retrieved.CreatedByUserID != userID {
			t.Errorf("expected created_by_user_id to be %s, got %v", userID, retrieved.CreatedByUserID)
		}
	})
}

func TestStoryRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("existing story", func(t *testing.T) {
		testStory, err := story.NewStory(testTenant.ID, "GetByID Story", nil)
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		err = storyRepo.Create(ctx, testStory)
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		retrieved, err := storyRepo.GetByID(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("failed to get story: %v", err)
		}

		if retrieved.ID != testStory.ID {
			t.Errorf("expected ID to be %s, got %s", testStory.ID, retrieved.ID)
		}

		if retrieved.Title != "GetByID Story" {
			t.Errorf("expected title to be 'GetByID Story', got '%s'", retrieved.Title)
		}
	})

	t.Run("non-existent story", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := storyRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent story")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "story" {
			t.Errorf("expected resource to be 'story', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestStoryRepository_ListByTenant(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("empty list", func(t *testing.T) {
		stories, err := storyRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(stories) != 0 {
			t.Errorf("expected empty list, got %d stories", len(stories))
		}
	})

	t.Run("list with stories", func(t *testing.T) {
		// Create multiple stories
		storyTitles := []string{"Story A", "Story B", "Story C"}
		createdStories := make([]*story.Story, 0, len(storyTitles))

		for _, title := range storyTitles {
			testStory, err := story.NewStory(testTenant.ID, title, nil)
			if err != nil {
				t.Fatalf("failed to create story: %v", err)
			}
			err = storyRepo.Create(ctx, testStory)
			if err != nil {
				t.Fatalf("failed to create story: %v", err)
			}
			createdStories = append(createdStories, testStory)
		}

		stories, err := storyRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(stories) != len(storyTitles) {
			t.Errorf("expected %d stories, got %d", len(storyTitles), len(stories))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		// Create more stories
		storyTitles := []string{"Story D", "Story E"}
		for _, title := range storyTitles {
			testStory, _ := story.NewStory(testTenant.ID, title, nil)
			storyRepo.Create(ctx, testStory)
		}

		// Test limit
		stories, err := storyRepo.ListByTenant(ctx, testTenant.ID, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(stories) > 2 {
			t.Errorf("expected at most 2 stories, got %d", len(stories))
		}

		// Test offset
		allStories, err := storyRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(allStories) < 3 {
			t.Skip("not enough stories to test pagination")
		}

		offsetStories, err := storyRepo.ListByTenant(ctx, testTenant.ID, 10, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(offsetStories) >= len(allStories) {
			t.Error("expected offset to reduce number of stories")
		}
	})
}

func TestStoryRepository_ListVersionsByRoot(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("list versions", func(t *testing.T) {
		// Create root story (version 1)
		rootStory, err := story.NewStory(testTenant.ID, "Root Story", nil)
		if err != nil {
			t.Fatalf("failed to create root story: %v", err)
		}
		err = storyRepo.Create(ctx, rootStory)
		if err != nil {
			t.Fatalf("failed to create root story: %v", err)
		}

		// Create version 2
		version2 := &story.Story{
			ID:              uuid.New(),
			TenantID:        testTenant.ID,
			Title:           "Root Story v2",
			Status:          story.StoryStatusDraft,
			VersionNumber:   2,
			RootStoryID:     rootStory.RootStoryID,
			PreviousStoryID: &rootStory.ID,
			CreatedAt:       rootStory.CreatedAt,
			UpdatedAt:       rootStory.UpdatedAt,
		}
		err = storyRepo.Create(ctx, version2)
		if err != nil {
			t.Fatalf("failed to create version 2: %v", err)
		}

		// Create version 3
		version3 := &story.Story{
			ID:              uuid.New(),
			TenantID:        testTenant.ID,
			Title:           "Root Story v3",
			Status:          story.StoryStatusPublished,
			VersionNumber:   3,
			RootStoryID:     rootStory.RootStoryID,
			PreviousStoryID: &version2.ID,
			CreatedAt:       rootStory.CreatedAt,
			UpdatedAt:       rootStory.UpdatedAt,
		}
		err = storyRepo.Create(ctx, version3)
		if err != nil {
			t.Fatalf("failed to create version 3: %v", err)
		}

		// List versions
		versions, err := storyRepo.ListVersionsByRoot(ctx, testTenant.ID, rootStory.RootStoryID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(versions) != 3 {
			t.Errorf("expected 3 versions, got %d", len(versions))
		}

		// Verify ordering (by version_number ASC)
		for i := 0; i < len(versions)-1; i++ {
			if versions[i].VersionNumber > versions[i+1].VersionNumber {
				t.Error("expected versions to be ordered by version_number ascending")
			}
		}

		// Verify all versions share same root
		for _, v := range versions {
			if v.RootStoryID != rootStory.RootStoryID {
				t.Errorf("expected root_story_id to be %s, got %s", rootStory.RootStoryID, v.RootStoryID)
			}
		}
	})

	t.Run("single version", func(t *testing.T) {
		singleStory, err := story.NewStory(testTenant.ID, "Single Story", nil)
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}
		err = storyRepo.Create(ctx, singleStory)
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		versions, err := storyRepo.ListVersionsByRoot(ctx, testTenant.ID, singleStory.RootStoryID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(versions) != 1 {
			t.Errorf("expected 1 version, got %d", len(versions))
		}
	})
}

func TestStoryRepository_GetVersionGraph(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("get version graph", func(t *testing.T) {
		rootStory, err := story.NewStory(testTenant.ID, "Graph Story", nil)
		if err != nil {
			t.Fatalf("failed to create root story: %v", err)
		}
		storyRepo.Create(ctx, rootStory)

		version2 := &story.Story{
			ID:              uuid.New(),
			TenantID:        testTenant.ID,
			Title:           "Graph Story v2",
			Status:          story.StoryStatusDraft,
			VersionNumber:   2,
			RootStoryID:     rootStory.RootStoryID,
			PreviousStoryID: &rootStory.ID,
			CreatedAt:       rootStory.CreatedAt,
			UpdatedAt:       rootStory.UpdatedAt,
		}
		storyRepo.Create(ctx, version2)

		graph, err := storyRepo.GetVersionGraph(ctx, testTenant.ID, rootStory.RootStoryID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(graph) != 2 {
			t.Errorf("expected 2 stories in graph, got %d", len(graph))
		}
	})
}

func TestStoryRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		testStory, err := story.NewStory(testTenant.ID, "Update Story", nil)
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		err = storyRepo.Create(ctx, testStory)
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		// Update title
		err = testStory.UpdateTitle("Updated Title")
		if err != nil {
			t.Fatalf("failed to update title: %v", err)
		}

		// Update status
		err = testStory.UpdateStatus(story.StoryStatusPublished)
		if err != nil {
			t.Fatalf("failed to update status: %v", err)
		}

		err = storyRepo.Update(ctx, testStory)
		if err != nil {
			t.Fatalf("failed to update story: %v", err)
		}

		// Verify update
		retrieved, err := storyRepo.GetByID(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("failed to get story: %v", err)
		}

		if retrieved.Title != "Updated Title" {
			t.Errorf("expected title to be 'Updated Title', got '%s'", retrieved.Title)
		}

		if retrieved.Status != story.StoryStatusPublished {
			t.Errorf("expected status to be 'published', got '%s'", retrieved.Status)
		}
	})
}

func TestStoryRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		testStory, err := story.NewStory(testTenant.ID, "Delete Story", nil)
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		err = storyRepo.Create(ctx, testStory)
		if err != nil {
			t.Fatalf("failed to create story: %v", err)
		}

		err = storyRepo.Delete(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("failed to delete story: %v", err)
		}

		// Verify story is deleted
		_, err = storyRepo.GetByID(ctx, testTenant.ID, testStory.ID)
		if err == nil {
			t.Fatal("expected error for deleted story")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "story" {
			t.Errorf("expected resource to be 'story', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestStoryRepository_CountByTenant(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("empty count", func(t *testing.T) {
		count, err := storyRepo.CountByTenant(ctx, testTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 0 {
			t.Errorf("expected count to be 0, got %d", count)
		}
	})

	t.Run("count with stories", func(t *testing.T) {
		// Create multiple stories
		storyTitles := []string{"Story 1", "Story 2", "Story 3"}
		for _, title := range storyTitles {
			testStory, err := story.NewStory(testTenant.ID, title, nil)
			if err != nil {
				t.Fatalf("failed to create story: %v", err)
			}
			err = storyRepo.Create(ctx, testStory)
			if err != nil {
				t.Fatalf("failed to create story: %v", err)
			}
		}

		count, err := storyRepo.CountByTenant(ctx, testTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count < len(storyTitles) {
			t.Errorf("expected count to be at least %d, got %d", len(storyTitles), count)
		}
	})
}

