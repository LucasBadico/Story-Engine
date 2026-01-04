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

func TestSceneReferenceRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	artifactRepo := NewArtifactRepository(db)
	sceneRepo := NewSceneRepository(db)
	sceneRefRepo := NewSceneReferenceRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testWorld, err := world.NewWorld(testTenant.ID, "Test World", false)
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}
	err = worldRepo.Create(ctx, testWorld)
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("create with character", func(t *testing.T) {
		// Create scene and character
		testScene, err := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}
		err = sceneRepo.Create(ctx, testScene)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		err = characterRepo.Create(ctx, testCharacter)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		// Create scene reference
		sceneRef, err := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeCharacter, testCharacter.ID)
		if err != nil {
			t.Fatalf("failed to create scene reference: %v", err)
		}

		err = sceneRefRepo.Create(ctx, sceneRef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify scene reference can be retrieved
		retrieved, err := sceneRefRepo.GetByID(ctx, testTenant.ID, sceneRef.ID)
		if err != nil {
			t.Fatalf("failed to retrieve scene reference: %v", err)
		}

		if retrieved.SceneID != testScene.ID {
			t.Errorf("expected scene_id to be %s, got %s", testScene.ID, retrieved.SceneID)
		}

		if retrieved.EntityType != story.SceneReferenceEntityTypeCharacter {
			t.Errorf("expected entity_type to be 'character', got '%s'", retrieved.EntityType)
		}

		if retrieved.EntityID != testCharacter.ID {
			t.Errorf("expected entity_id to be %s, got %s", testCharacter.ID, retrieved.EntityID)
		}
	})

	t.Run("create with location", func(t *testing.T) {
		testScene, err := story.NewScene(testTenant.ID, testStory.ID, nil, 2)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}
		err = sceneRepo.Create(ctx, testScene)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		testLocation, err := world.NewLocation(testTenant.ID, testWorld.ID, "Test Location", nil)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}
		err = locationRepo.Create(ctx, testLocation)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		sceneRef, err := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeLocation, testLocation.ID)
		if err != nil {
			t.Fatalf("failed to create scene reference: %v", err)
		}

		err = sceneRefRepo.Create(ctx, sceneRef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := sceneRefRepo.GetByID(ctx, testTenant.ID, sceneRef.ID)
		if err != nil {
			t.Fatalf("failed to retrieve scene reference: %v", err)
		}

		if retrieved.EntityType != story.SceneReferenceEntityTypeLocation {
			t.Errorf("expected entity_type to be 'location', got '%s'", retrieved.EntityType)
		}
	})

	t.Run("create with artifact", func(t *testing.T) {
		testScene, err := story.NewScene(testTenant.ID, testStory.ID, nil, 3)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}
		err = sceneRepo.Create(ctx, testScene)
		if err != nil {
			t.Fatalf("failed to create scene: %v", err)
		}

		testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}
		err = artifactRepo.Create(ctx, testArtifact)
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		sceneRef, err := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeArtifact, testArtifact.ID)
		if err != nil {
			t.Fatalf("failed to create scene reference: %v", err)
		}

		err = sceneRefRepo.Create(ctx, sceneRef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := sceneRefRepo.GetByID(ctx, testTenant.ID, sceneRef.ID)
		if err != nil {
			t.Fatalf("failed to retrieve scene reference: %v", err)
		}

		if retrieved.EntityType != story.SceneReferenceEntityTypeArtifact {
			t.Errorf("expected entity_type to be 'artifact', got '%s'", retrieved.EntityType)
		}
	})
}

func TestSceneReferenceRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	sceneRepo := NewSceneRepository(db)
	sceneRefRepo := NewSceneReferenceRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testWorld, err := world.NewWorld(testTenant.ID, "Test World", false)
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}
	err = worldRepo.Create(ctx, testWorld)
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("existing scene reference", func(t *testing.T) {
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		sceneRepo.Create(ctx, testScene)

		testCharacter, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		characterRepo.Create(ctx, testCharacter)

		sceneRef, _ := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeCharacter, testCharacter.ID)
		sceneRefRepo.Create(ctx, sceneRef)

		retrieved, err := sceneRefRepo.GetByID(ctx, testTenant.ID, sceneRef.ID)
		if err != nil {
			t.Fatalf("failed to get scene reference: %v", err)
		}

		if retrieved.ID != sceneRef.ID {
			t.Errorf("expected ID to be %s, got %s", sceneRef.ID, retrieved.ID)
		}
	})

	t.Run("non-existent scene reference", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := sceneRefRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent scene reference")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "scene_reference" {
			t.Errorf("expected resource to be 'scene_reference', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestSceneReferenceRepository_ListByScene(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	sceneRepo := NewSceneRepository(db)
	sceneRefRepo := NewSceneReferenceRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testWorld, err := world.NewWorld(testTenant.ID, "Test World", false)
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}
	err = worldRepo.Create(ctx, testWorld)
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("list by scene", func(t *testing.T) {
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		sceneRepo.Create(ctx, testScene)

		// Create entities
		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		location1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 1", nil)
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)
		locationRepo.Create(ctx, location1)

		// Create scene references
		ref1, _ := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeCharacter, char1.ID)
		ref2, _ := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeCharacter, char2.ID)
		ref3, _ := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeLocation, location1.ID)
		sceneRefRepo.Create(ctx, ref1)
		sceneRefRepo.Create(ctx, ref2)
		sceneRefRepo.Create(ctx, ref3)

		refs, err := sceneRefRepo.ListByScene(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 3 {
			t.Errorf("expected 3 references, got %d", len(refs))
		}

		for _, ref := range refs {
			if ref.SceneID != testScene.ID {
				t.Errorf("expected scene_id to be %s, got %s", testScene.ID, ref.SceneID)
			}
		}
	})
}

func TestSceneReferenceRepository_ListByEntity(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	sceneRepo := NewSceneRepository(db)
	sceneRefRepo := NewSceneReferenceRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testWorld, err := world.NewWorld(testTenant.ID, "Test World", false)
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}
	err = worldRepo.Create(ctx, testWorld)
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("list by entity", func(t *testing.T) {
		testCharacter, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		characterRepo.Create(ctx, testCharacter)

		// Create multiple scenes
		scene1, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		scene2, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 2)
		sceneRepo.Create(ctx, scene1)
		sceneRepo.Create(ctx, scene2)

		// Create references from different scenes to same character
		ref1, _ := story.NewSceneReference(scene1.ID, story.SceneReferenceEntityTypeCharacter, testCharacter.ID)
		ref2, _ := story.NewSceneReference(scene2.ID, story.SceneReferenceEntityTypeCharacter, testCharacter.ID)
		sceneRefRepo.Create(ctx, ref1)
		sceneRefRepo.Create(ctx, ref2)

		refs, err := sceneRefRepo.ListByEntity(ctx, testTenant.ID, story.SceneReferenceEntityTypeCharacter, testCharacter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 2 {
			t.Errorf("expected 2 references, got %d", len(refs))
		}

		for _, ref := range refs {
			if ref.EntityType != story.SceneReferenceEntityTypeCharacter {
				t.Errorf("expected entity_type to be 'character', got '%s'", ref.EntityType)
			}

			if ref.EntityID != testCharacter.ID {
				t.Errorf("expected entity_id to be %s, got %s", testCharacter.ID, ref.EntityID)
			}
		}
	})
}

func TestSceneReferenceRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	sceneRepo := NewSceneRepository(db)
	sceneRefRepo := NewSceneReferenceRepository(db)

	// Create tenant and story first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testWorld, err := world.NewWorld(testTenant.ID, "Test World", false)
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}
	err = worldRepo.Create(ctx, testWorld)
	if err != nil {
		t.Fatalf("failed to create world: %v", err)
	}

	testStory, err := story.NewStory(testTenant.ID, "Test Story", nil)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}
	err = storyRepo.Create(ctx, testStory)
	if err != nil {
		t.Fatalf("failed to create story: %v", err)
	}

	t.Run("delete by id", func(t *testing.T) {
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 1)
		sceneRepo.Create(ctx, testScene)

		testCharacter, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		characterRepo.Create(ctx, testCharacter)

		sceneRef, _ := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeCharacter, testCharacter.ID)
		sceneRefRepo.Create(ctx, sceneRef)

		err = sceneRefRepo.Delete(ctx, testTenant.ID, sceneRef.ID)
		if err != nil {
			t.Fatalf("failed to delete scene reference: %v", err)
		}

		// Verify deletion
		_, err = sceneRefRepo.GetByID(ctx, testTenant.ID, sceneRef.ID)
		if err == nil {
			t.Fatal("expected error for deleted scene reference")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "scene_reference" {
			t.Errorf("expected resource to be 'scene_reference', got '%s'", notFoundErr.Resource)
		}
	})

	t.Run("delete by scene", func(t *testing.T) {
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 2)
		sceneRepo.Create(ctx, testScene)

		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)

		ref1, _ := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeCharacter, char1.ID)
		ref2, _ := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeCharacter, char2.ID)
		sceneRefRepo.Create(ctx, ref1)
		sceneRefRepo.Create(ctx, ref2)

		err = sceneRefRepo.DeleteByScene(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("failed to delete by scene: %v", err)
		}

		// Verify all references deleted
		refs, err := sceneRefRepo.ListByScene(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 0 {
			t.Errorf("expected no references, got %d", len(refs))
		}
	})

	t.Run("delete by scene and entity", func(t *testing.T) {
		testScene, _ := story.NewScene(testTenant.ID, testStory.ID, nil, 3)
		sceneRepo.Create(ctx, testScene)

		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)

		ref1, _ := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeCharacter, char1.ID)
		ref2, _ := story.NewSceneReference(testScene.ID, story.SceneReferenceEntityTypeCharacter, char2.ID)
		sceneRefRepo.Create(ctx, ref1)
		sceneRefRepo.Create(ctx, ref2)

		err = sceneRefRepo.DeleteBySceneAndEntity(ctx, testTenant.ID, testScene.ID, story.SceneReferenceEntityTypeCharacter, char1.ID)
		if err != nil {
			t.Fatalf("failed to delete by scene and entity: %v", err)
		}

		// Verify only ref1 was deleted
		refs, err := sceneRefRepo.ListByScene(ctx, testTenant.ID, testScene.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 1 {
			t.Errorf("expected 1 reference, got %d", len(refs))
		}

		if refs[0].EntityID != char2.ID {
			t.Errorf("expected remaining entity_id to be %s, got %s", char2.ID, refs[0].EntityID)
		}
	})
}

