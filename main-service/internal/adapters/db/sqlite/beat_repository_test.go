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

func TestBeatRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	sceneRepo := NewSceneRepository(db)
	beatRepo := NewBeatRepository(db)

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
			t.Fatalf("failed to create scene: %v", err)
		}

		testBeat, err := story.NewBeat(testTenant.ID, testScene.ID, 1, story.BeatTypeSetup)
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}

		err = beatRepo.Create(ctx, testBeat)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify beat can be retrieved
		retrieved, err := beatRepo.GetByID(ctx, testTenant.ID, testBeat.ID)
		if err != nil {
			t.Fatalf("failed to retrieve beat: %v", err)
		}

		if retrieved.OrderNum != 1 {
			t.Errorf("expected order_num to be 1, got %d", retrieved.OrderNum)
		}

		if retrieved.Type != story.BeatTypeSetup {
			t.Errorf("expected type to be 'setup', got '%s'", retrieved.Type)
		}

		if retrieved.SceneID != testScene.ID {
			t.Errorf("expected scene_id to be %s, got %s", testScene.ID, retrieved.SceneID)
		}
	})

	t.Run("creation with intent and outcome", func(t *testing.T) {
		testScene, err := story.NewScene(testTenant.ID, testStory.ID, nil, 2)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}
		err = sceneRepo.Create(ctx, testScene)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		testBeat, err := story.NewBeat(testTenant.ID, testScene.ID, 1, story.BeatTypeConflict)
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}
		testBeat.UpdateIntent("Test intent")
		testBeat.UpdateOutcome("Test outcome")

		err = beatRepo.Create(ctx, testBeat)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := beatRepo.GetByID(ctx, testTenant.ID, testBeat.ID)
		if err != nil {
			t.Fatalf("failed to retrieve beat: %v", err)
		}

		if retrieved.Intent != "Test intent" {
			t.Errorf("expected intent to be 'Test intent', got '%s'", retrieved.Intent)
		}

		if retrieved.Outcome != "Test outcome" {
			t.Errorf("expected outcome to be 'Test outcome', got '%s'", retrieved.Outcome)
		}
	})
}

func TestBeatRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	sceneRepo := NewSceneRepository(db)
	beatRepo := NewBeatRepository(db)

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

	t.Run("existing beat", func(t *testing.T) {
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		sceneRepo.Create(ctx, testScene)

		testBeat, err := story.NewBeat(testTenant.ID, testScene.ID, 1, story.BeatTypeSetup)
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}

		err = beatRepo.Create(ctx, testBeat)
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}

		retrieved, err := beatRepo.GetByID(ctx, testTenant.ID, testBeat.ID)
		if err != nil {
			t.Fatalf("failed to get beat: %v", err)
		}

		if retrieved.ID != testBeat.ID {
			t.Errorf("expected ID to be %s, got %s", testBeat.ID, retrieved.ID)
		}
	})

	t.Run("non-existent beat", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := beatRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent beat")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "beat" {
			t.Errorf("expected resource to be 'beat', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestBeatRepository_ListByScene(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	sceneRepo := NewSceneRepository(db)
	beatRepo := NewBeatRepository(db)

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

	t.Run("list with beats", func(t *testing.T) {
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		sceneRepo.Create(ctx, testScene)

		// Create multiple beats
		beat1, _ := story.NewBeat(testTenant.ID, testScene.ID, 1, story.BeatTypeSetup)
		beat2, _ := story.NewBeat(testTenant.ID, testScene.ID, 2, story.BeatTypeConflict)
		beatRepo.Create(ctx, beat1)
		beatRepo.Create(ctx, beat2)

		beats, err := beatRepo.ListByScene(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(beats) != 2 {
			t.Errorf("expected 2 beats, got %d", len(beats))
		}

		// Verify ordering (by order_num ASC)
		for i := 0; i < len(beats)-1; i++ {
			if beats[i].OrderNum > beats[i+1].OrderNum {
				t.Error("expected beats to be ordered by order_num ascending")
			}
		}
	})

	t.Run("empty list", func(t *testing.T) {
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 2)
		sceneRepo.Create(ctx, testScene)

		beats, err := beatRepo.ListByScene(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(beats) != 0 {
			t.Errorf("expected empty list, got %d beats", len(beats))
		}
	})
}

func TestBeatRepository_ListBySceneOrdered(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	sceneRepo := NewSceneRepository(db)
	beatRepo := NewBeatRepository(db)

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
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		sceneRepo.Create(ctx, testScene)

		// Create beats in non-sequential order
		beat3, _ := story.NewBeat(testTenant.ID, testScene.ID, 3, story.BeatTypeResolution)
		beat1, _ := story.NewBeat(testTenant.ID, testScene.ID, 1, story.BeatTypeSetup)
		beat2, _ := story.NewBeat(testTenant.ID, testScene.ID, 2, story.BeatTypeConflict)
		beatRepo.Create(ctx, beat3)
		beatRepo.Create(ctx, beat1)
		beatRepo.Create(ctx, beat2)

		beats, err := beatRepo.ListBySceneOrdered(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(beats) != 3 {
			t.Errorf("expected 3 beats, got %d", len(beats))
		}

		// Verify ordering (by order_num ASC)
		for i := 0; i < len(beats)-1; i++ {
			if beats[i].OrderNum > beats[i+1].OrderNum {
				t.Error("expected beats to be ordered by order_num ascending")
			}
		}

		// Verify correct order
		if beats[0].OrderNum != 1 {
			t.Errorf("expected first beat to be order_num 1, got %d", beats[0].OrderNum)
		}
		if beats[1].OrderNum != 2 {
			t.Errorf("expected second beat to be order_num 2, got %d", beats[1].OrderNum)
		}
		if beats[2].OrderNum != 3 {
			t.Errorf("expected third beat to be order_num 3, got %d", beats[2].OrderNum)
		}
	})
}

func TestBeatRepository_ListByStory(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	sceneRepo := NewSceneRepository(db)
	beatRepo := NewBeatRepository(db)

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
		// Create multiple scenes
		scene1, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		scene2, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 2)
		sceneRepo.Create(ctx, scene1)
		sceneRepo.Create(ctx, scene2)

		// Create beats in different scenes
		beat1, _ := story.NewBeat(testTenant.ID, scene1.ID, 1, story.BeatTypeSetup)
		beat2, _ := story.NewBeat(testTenant.ID, scene1.ID, 2, story.BeatTypeConflict)
		beat3, _ := story.NewBeat(testTenant.ID, scene2.ID, 1, story.BeatTypeReveal)
		beatRepo.Create(ctx, beat1)
		beatRepo.Create(ctx, beat2)
		beatRepo.Create(ctx, beat3)

		beats, err := beatRepo.ListByStory(ctx, testTenant.ID, testStory.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(beats) != 3 {
			t.Errorf("expected 3 beats, got %d", len(beats))
		}
	})
}

func TestBeatRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	sceneRepo := NewSceneRepository(db)
	beatRepo := NewBeatRepository(db)

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
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		sceneRepo.Create(ctx, testScene)

		testBeat, err := story.NewBeat(testTenant.ID, testScene.ID, 1, story.BeatTypeSetup)
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}

		err = beatRepo.Create(ctx, testBeat)
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}

		// Update
		testBeat.UpdateIntent("Updated intent")
		testBeat.UpdateOutcome("Updated outcome")
		testBeat.Type = story.BeatTypeClimax
		testBeat.OrderNum = 2

		err = beatRepo.Update(ctx, testBeat)
		if err != nil {
			t.Fatalf("failed to update beat: %v", err)
		}

		// Verify update
		retrieved, err := beatRepo.GetByID(ctx, testTenant.ID, testBeat.ID)
		if err != nil {
			t.Fatalf("failed to get beat: %v", err)
		}

		if retrieved.Intent != "Updated intent" {
			t.Errorf("expected intent to be 'Updated intent', got '%s'", retrieved.Intent)
		}

		if retrieved.Outcome != "Updated outcome" {
			t.Errorf("expected outcome to be 'Updated outcome', got '%s'", retrieved.Outcome)
		}

		if retrieved.Type != story.BeatTypeClimax {
			t.Errorf("expected type to be 'climax', got '%s'", retrieved.Type)
		}

		if retrieved.OrderNum != 2 {
			t.Errorf("expected order_num to be 2, got %d", retrieved.OrderNum)
		}
	})
}

func TestBeatRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	sceneRepo := NewSceneRepository(db)
	beatRepo := NewBeatRepository(db)

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
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		sceneRepo.Create(ctx, testScene)

		testBeat, err := story.NewBeat(testTenant.ID, testScene.ID, 1, story.BeatTypeSetup)
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}

		err = beatRepo.Create(ctx, testBeat)
		if err != nil {
			t.Fatalf("failed to create beat: %v", err)
		}

		err = beatRepo.Delete(ctx, testTenant.ID, testBeat.ID)
		if err != nil {
			t.Fatalf("failed to delete beat: %v", err)
		}

		// Verify beat is deleted
		_, err = beatRepo.GetByID(ctx, testTenant.ID, testBeat.ID)
		if err == nil {
			t.Fatal("expected error for deleted beat")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "beat" {
			t.Errorf("expected resource to be 'beat', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestBeatRepository_DeleteByScene(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	sceneRepo := NewSceneRepository(db)
	beatRepo := NewBeatRepository(db)

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

	t.Run("delete all beats for scene", func(t *testing.T) {
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		sceneRepo.Create(ctx, testScene)

		// Create multiple beats
		beat1, _ := story.NewBeat(testTenant.ID, testScene.ID, 1, story.BeatTypeSetup)
		beat2, _ := story.NewBeat(testTenant.ID, testScene.ID, 2, story.BeatTypeConflict)
		beatRepo.Create(ctx, beat1)
		beatRepo.Create(ctx, beat2)

		// Verify they exist
		beats, err := beatRepo.ListByScene(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(beats) != 2 {
			t.Fatalf("expected 2 beats, got %d", len(beats))
		}

		// Delete all
		err = beatRepo.DeleteByScene(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("failed to delete by scene: %v", err)
		}

		// Verify all deleted
		beats, err = beatRepo.ListByScene(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(beats) != 0 {
			t.Errorf("expected no beats, got %d", len(beats))
		}
	})
}

