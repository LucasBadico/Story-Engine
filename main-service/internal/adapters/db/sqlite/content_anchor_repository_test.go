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

func TestContentAnchorRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	artifactRepo := NewArtifactRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)
	contentAnchorRepo := NewContentAnchorRepository(db)

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

	testChapter, err := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
	if err != nil {
		t.Fatalf("failed to create chapter: %v", err)
	}
	err = chapterRepo.Create(ctx, testChapter)
	if err != nil {
		t.Fatalf("failed to create chapter: %v", err)
	}

	t.Run("create with character", func(t *testing.T) {
		// Create content block and character
		orderNum := 1
		testContentBlock, err := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "Test Content")
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}
		err = contentBlockRepo.Create(ctx, testContentBlock)
		if err != nil {
			t.Fatalf("failed to create content block: %v", err)
		}

		testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		err = characterRepo.Create(ctx, testCharacter)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		// Create content anchor
		contentBlockRef, err := story.NewContentAnchor(testContentBlock.ID, story.EntityTypeCharacter, testCharacter.ID)
		if err != nil {
			t.Fatalf("failed to create content anchor: %v", err)
		}

		err = contentAnchorRepo.Create(ctx, contentBlockRef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify content anchor can be retrieved
		retrieved, err := contentAnchorRepo.GetByID(ctx, testTenant.ID, contentBlockRef.ID)
		if err != nil {
			t.Fatalf("failed to retrieve content anchor: %v", err)
		}

		if retrieved.ContentBlockID != testContentBlock.ID {
			t.Errorf("expected content_block_id to be %s, got %s", testContentBlock.ID, retrieved.ContentBlockID)
		}

		if retrieved.EntityType != story.EntityTypeCharacter {
			t.Errorf("expected entity_type to be 'character', got '%s'", retrieved.EntityType)
		}

		if retrieved.EntityID != testCharacter.ID {
			t.Errorf("expected entity_id to be %s, got %s", testCharacter.ID, retrieved.EntityID)
		}
	})

	t.Run("create with location", func(t *testing.T) {
		orderNum := 2
		testContentBlock, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "Test Content 2")
		contentBlockRepo.Create(ctx, testContentBlock)

		testLocation, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Test Location", nil)
		locationRepo.Create(ctx, testLocation)

		contentBlockRef, _ := story.NewContentAnchor(testContentBlock.ID, story.EntityTypeLocation, testLocation.ID)
		err = contentAnchorRepo.Create(ctx, contentBlockRef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := contentAnchorRepo.GetByID(ctx, testTenant.ID, contentBlockRef.ID)
		if err != nil {
			t.Fatalf("failed to retrieve content anchor: %v", err)
		}

		if retrieved.EntityType != story.EntityTypeLocation {
			t.Errorf("expected entity_type to be 'location', got '%s'", retrieved.EntityType)
		}
	})

	t.Run("create with artifact", func(t *testing.T) {
		orderNum := 3
		testContentBlock, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "Test Content 3")
		contentBlockRepo.Create(ctx, testContentBlock)

		testArtifact, _ := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
		artifactRepo.Create(ctx, testArtifact)

		contentBlockRef, _ := story.NewContentAnchor(testContentBlock.ID, story.EntityTypeArtifact, testArtifact.ID)
		err = contentAnchorRepo.Create(ctx, contentBlockRef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := contentAnchorRepo.GetByID(ctx, testTenant.ID, contentBlockRef.ID)
		if err != nil {
			t.Fatalf("failed to retrieve content anchor: %v", err)
		}

		if retrieved.EntityType != story.EntityTypeArtifact {
			t.Errorf("expected entity_type to be 'artifact', got '%s'", retrieved.EntityType)
		}
	})
}

func TestContentAnchorRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)
	contentAnchorRepo := NewContentAnchorRepository(db)

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

	testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
	chapterRepo.Create(ctx, testChapter)

	t.Run("existing content anchor", func(t *testing.T) {
		orderNum := 1
		testContentBlock, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "Test Content")
		contentBlockRepo.Create(ctx, testContentBlock)

		testCharacter, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		characterRepo.Create(ctx, testCharacter)

		contentBlockRef, _ := story.NewContentAnchor(testContentBlock.ID, story.EntityTypeCharacter, testCharacter.ID)
		contentAnchorRepo.Create(ctx, contentBlockRef)

		retrieved, err := contentAnchorRepo.GetByID(ctx, testTenant.ID, contentBlockRef.ID)
		if err != nil {
			t.Fatalf("failed to get content anchor: %v", err)
		}

		if retrieved.ID != contentBlockRef.ID {
			t.Errorf("expected ID to be %s, got %s", contentBlockRef.ID, retrieved.ID)
		}
	})

	t.Run("non-existent content anchor", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := contentAnchorRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent content anchor")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "content_anchor" {
			t.Errorf("expected resource to be 'content_anchor', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestContentAnchorRepository_ListByContentBlock(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)
	contentAnchorRepo := NewContentAnchorRepository(db)

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

	testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
	chapterRepo.Create(ctx, testChapter)

	t.Run("list by content block", func(t *testing.T) {
		orderNum := 1
		testContentBlock, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "Test Content")
		contentBlockRepo.Create(ctx, testContentBlock)

		// Create entities
		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		location1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 1", nil)
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)
		locationRepo.Create(ctx, location1)

		// Create content anchors
		ref1, _ := story.NewContentAnchor(testContentBlock.ID, story.EntityTypeCharacter, char1.ID)
		ref2, _ := story.NewContentAnchor(testContentBlock.ID, story.EntityTypeCharacter, char2.ID)
		ref3, _ := story.NewContentAnchor(testContentBlock.ID, story.EntityTypeLocation, location1.ID)
		contentAnchorRepo.Create(ctx, ref1)
		contentAnchorRepo.Create(ctx, ref2)
		contentAnchorRepo.Create(ctx, ref3)

		refs, err := contentAnchorRepo.ListByContentBlock(ctx, testTenant.ID, testContentBlock.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 3 {
			t.Errorf("expected 3 references, got %d", len(refs))
		}

		for _, ref := range refs {
			if ref.ContentBlockID != testContentBlock.ID {
				t.Errorf("expected content_block_id to be %s, got %s", testContentBlock.ID, ref.ContentBlockID)
			}
		}
	})
}

func TestContentAnchorRepository_ListByEntity(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)
	contentAnchorRepo := NewContentAnchorRepository(db)

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

	testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
	chapterRepo.Create(ctx, testChapter)

	t.Run("list by entity", func(t *testing.T) {
		testCharacter, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		characterRepo.Create(ctx, testCharacter)

		// Create multiple content blocks
		orderNum1 := 1
		orderNum2 := 2
		block1, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum1, story.ContentTypeText, story.ContentKindFinal, "Content 1")
		block2, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum2, story.ContentTypeText, story.ContentKindFinal, "Content 2")
		contentBlockRepo.Create(ctx, block1)
		contentBlockRepo.Create(ctx, block2)

		// Create references from different content blocks to same character
		ref1, _ := story.NewContentAnchor(block1.ID, story.EntityTypeCharacter, testCharacter.ID)
		ref2, _ := story.NewContentAnchor(block2.ID, story.EntityTypeCharacter, testCharacter.ID)
		contentAnchorRepo.Create(ctx, ref1)
		contentAnchorRepo.Create(ctx, ref2)

		refs, err := contentAnchorRepo.ListByEntity(ctx, testTenant.ID, story.EntityTypeCharacter, testCharacter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 2 {
			t.Errorf("expected 2 references, got %d", len(refs))
		}

		for _, ref := range refs {
			if ref.EntityType != story.EntityTypeCharacter {
				t.Errorf("expected entity_type to be 'character', got '%s'", ref.EntityType)
			}

			if ref.EntityID != testCharacter.ID {
				t.Errorf("expected entity_id to be %s, got %s", testCharacter.ID, ref.EntityID)
			}
		}
	})
}

func TestContentAnchorRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	storyRepo := NewStoryRepository(db)
	chapterRepo := NewChapterRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	contentBlockRepo := NewContentBlockRepository(db)
	contentAnchorRepo := NewContentAnchorRepository(db)

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

	testChapter, _ := story.NewChapter(testTenant.ID, testStory.ID, 1, "Chapter 1")
	chapterRepo.Create(ctx, testChapter)

	t.Run("delete by id", func(t *testing.T) {
		orderNum := 1
		testContentBlock, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "Test Content")
		contentBlockRepo.Create(ctx, testContentBlock)

		testCharacter, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		characterRepo.Create(ctx, testCharacter)

		contentBlockRef, _ := story.NewContentAnchor(testContentBlock.ID, story.EntityTypeCharacter, testCharacter.ID)
		contentAnchorRepo.Create(ctx, contentBlockRef)

		err = contentAnchorRepo.Delete(ctx, testTenant.ID, contentBlockRef.ID)
		if err != nil {
			t.Fatalf("failed to delete content anchor: %v", err)
		}

		// Verify deletion
		_, err = contentAnchorRepo.GetByID(ctx, testTenant.ID, contentBlockRef.ID)
		if err == nil {
			t.Fatal("expected error for deleted content anchor")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "content_anchor" {
			t.Errorf("expected resource to be 'content_anchor', got '%s'", notFoundErr.Resource)
		}
	})

	t.Run("delete by content block", func(t *testing.T) {
		orderNum := 2
		testContentBlock, _ := story.NewContentBlock(testTenant.ID, &testChapter.ID, &orderNum, story.ContentTypeText, story.ContentKindFinal, "Test Content 2")
		contentBlockRepo.Create(ctx, testContentBlock)

		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)

		ref1, _ := story.NewContentAnchor(testContentBlock.ID, story.EntityTypeCharacter, char1.ID)
		ref2, _ := story.NewContentAnchor(testContentBlock.ID, story.EntityTypeCharacter, char2.ID)
		contentAnchorRepo.Create(ctx, ref1)
		contentAnchorRepo.Create(ctx, ref2)

		err = contentAnchorRepo.DeleteByContentBlock(ctx, testTenant.ID, testContentBlock.ID)
		if err != nil {
			t.Fatalf("failed to delete by content block: %v", err)
		}

		// Verify all references deleted
		refs, err := contentAnchorRepo.ListByContentBlock(ctx, testTenant.ID, testContentBlock.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 0 {
			t.Errorf("expected no references, got %d", len(refs))
		}
	})
}

