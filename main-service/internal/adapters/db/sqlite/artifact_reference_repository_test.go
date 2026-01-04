//go:build integration

package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/tenant"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
)

func TestArtifactReferenceRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	artifactRepo := NewArtifactRepository(db)
	artifactRefRepo := NewArtifactReferenceRepository(db)

	// Create tenant and world first
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

	// Create artifact
	testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	err = artifactRepo.Create(ctx, testArtifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	// Create character
	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	// Create location
	testLocation, err := world.NewLocation(testTenant.ID, testWorld.ID, "Test Location", nil)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}
	err = locationRepo.Create(ctx, testLocation)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}

	t.Run("successful creation with character", func(t *testing.T) {
		ref, err := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeCharacter, testCharacter.ID)
		if err != nil {
			t.Fatalf("failed to create artifact reference: %v", err)
		}

		err = artifactRefRepo.Create(ctx, ref)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify artifact reference can be retrieved
		retrieved, err := artifactRefRepo.GetByID(ctx, testTenant.ID, ref.ID)
		if err != nil {
			t.Fatalf("failed to retrieve artifact reference: %v", err)
		}

		if retrieved.ArtifactID != testArtifact.ID {
			t.Errorf("expected artifact_id to be %s, got %s", testArtifact.ID, retrieved.ArtifactID)
		}

		if retrieved.EntityType != world.ArtifactReferenceEntityTypeCharacter {
			t.Errorf("expected entity_type to be 'character', got '%s'", retrieved.EntityType)
		}

		if retrieved.EntityID != testCharacter.ID {
			t.Errorf("expected entity_id to be %s, got %s", testCharacter.ID, retrieved.EntityID)
		}
	})

	t.Run("successful creation with location", func(t *testing.T) {
		ref, err := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeLocation, testLocation.ID)
		if err != nil {
			t.Fatalf("failed to create artifact reference: %v", err)
		}

		err = artifactRefRepo.Create(ctx, ref)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := artifactRefRepo.GetByID(ctx, testTenant.ID, ref.ID)
		if err != nil {
			t.Fatalf("failed to retrieve artifact reference: %v", err)
		}

		if retrieved.EntityType != world.ArtifactReferenceEntityTypeLocation {
			t.Errorf("expected entity_type to be 'location', got '%s'", retrieved.EntityType)
		}

		if retrieved.EntityID != testLocation.ID {
			t.Errorf("expected entity_id to be %s, got %s", testLocation.ID, retrieved.EntityID)
		}
	})
}

func TestArtifactReferenceRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	artifactRepo := NewArtifactRepository(db)
	artifactRefRepo := NewArtifactReferenceRepository(db)

	// Create tenant, world, artifact, and character
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

	testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	err = artifactRepo.Create(ctx, testArtifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("existing artifact reference", func(t *testing.T) {
		ref, err := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeCharacter, testCharacter.ID)
		if err != nil {
			t.Fatalf("failed to create artifact reference: %v", err)
		}
		err = artifactRefRepo.Create(ctx, ref)
		if err != nil {
			t.Fatalf("failed to create artifact reference: %v", err)
		}

		retrieved, err := artifactRefRepo.GetByID(ctx, testTenant.ID, ref.ID)
		if err != nil {
			t.Fatalf("failed to get artifact reference: %v", err)
		}

		if retrieved.ID != ref.ID {
			t.Errorf("expected ID to be %s, got %s", ref.ID, retrieved.ID)
		}

		if retrieved.ArtifactID != testArtifact.ID {
			t.Errorf("expected artifact_id to be %s, got %s", testArtifact.ID, retrieved.ArtifactID)
		}
	})

	t.Run("non-existent artifact reference", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := artifactRefRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent artifact reference")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "artifact_reference" {
			t.Errorf("expected resource to be 'artifact_reference', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestArtifactReferenceRepository_ListByArtifact(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	artifactRepo := NewArtifactRepository(db)
	artifactRefRepo := NewArtifactReferenceRepository(db)

	// Create tenant, world, and artifact
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

	testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	err = artifactRepo.Create(ctx, testArtifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	t.Run("empty list", func(t *testing.T) {
		refs, err := artifactRefRepo.ListByArtifact(ctx, testTenant.ID, testArtifact.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 0 {
			t.Errorf("expected empty list, got %d references", len(refs))
		}
	})

	t.Run("list with references", func(t *testing.T) {
		// Create characters and locations
		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		loc1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 1", nil)
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)
		locationRepo.Create(ctx, loc1)

		// Create references
		ref1, _ := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeCharacter, char1.ID)
		ref2, _ := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeCharacter, char2.ID)
		ref3, _ := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeLocation, loc1.ID)
		artifactRefRepo.Create(ctx, ref1)
		time.Sleep(10 * time.Millisecond)
		artifactRefRepo.Create(ctx, ref2)
		time.Sleep(10 * time.Millisecond)
		artifactRefRepo.Create(ctx, ref3)

		refs, err := artifactRefRepo.ListByArtifact(ctx, testTenant.ID, testArtifact.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 3 {
			t.Errorf("expected 3 references, got %d", len(refs))
		}

		// Verify all belong to the artifact
		for _, ref := range refs {
			if ref.ArtifactID != testArtifact.ID {
				t.Errorf("expected artifact_id to be %s, got %s", testArtifact.ID, ref.ArtifactID)
			}
		}
	})
}

func TestArtifactReferenceRepository_ListByEntity(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	artifactRepo := NewArtifactRepository(db)
	artifactRefRepo := NewArtifactReferenceRepository(db)

	// Create tenant, world, character, and location
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

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	testLocation, err := world.NewLocation(testTenant.ID, testWorld.ID, "Test Location", nil)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}
	err = locationRepo.Create(ctx, testLocation)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}

	t.Run("list by character", func(t *testing.T) {
		// Create artifacts
		artifact1, _ := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact 1")
		artifact2, _ := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact 2")
		artifactRepo.Create(ctx, artifact1)
		artifactRepo.Create(ctx, artifact2)

		// Create references
		ref1, _ := world.NewArtifactReference(artifact1.ID, world.ArtifactReferenceEntityTypeCharacter, testCharacter.ID)
		ref2, _ := world.NewArtifactReference(artifact2.ID, world.ArtifactReferenceEntityTypeCharacter, testCharacter.ID)
		artifactRefRepo.Create(ctx, ref1)
		artifactRefRepo.Create(ctx, ref2)

		refs, err := artifactRefRepo.ListByEntity(ctx, testTenant.ID, world.ArtifactReferenceEntityTypeCharacter, testCharacter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 2 {
			t.Errorf("expected 2 references, got %d", len(refs))
		}

		for _, ref := range refs {
			if ref.EntityType != world.ArtifactReferenceEntityTypeCharacter {
				t.Errorf("expected entity_type to be 'character', got '%s'", ref.EntityType)
			}
			if ref.EntityID != testCharacter.ID {
				t.Errorf("expected entity_id to be %s, got %s", testCharacter.ID, ref.EntityID)
			}
		}
	})

	t.Run("list by location", func(t *testing.T) {
		// Create artifact
		artifact3, _ := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact 3")
		artifactRepo.Create(ctx, artifact3)

		// Create reference
		ref3, _ := world.NewArtifactReference(artifact3.ID, world.ArtifactReferenceEntityTypeLocation, testLocation.ID)
		artifactRefRepo.Create(ctx, ref3)

		refs, err := artifactRefRepo.ListByEntity(ctx, testTenant.ID, world.ArtifactReferenceEntityTypeLocation, testLocation.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 1 {
			t.Errorf("expected 1 reference, got %d", len(refs))
		}

		for _, ref := range refs {
			if ref.EntityType != world.ArtifactReferenceEntityTypeLocation {
				t.Errorf("expected entity_type to be 'location', got '%s'", ref.EntityType)
			}
			if ref.EntityID != testLocation.ID {
				t.Errorf("expected entity_id to be %s, got %s", testLocation.ID, ref.EntityID)
			}
		}
	})

	t.Run("empty list", func(t *testing.T) {
		nonExistentID := uuid.New()
		refs, err := artifactRefRepo.ListByEntity(ctx, testTenant.ID, world.ArtifactReferenceEntityTypeCharacter, nonExistentID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 0 {
			t.Errorf("expected empty list, got %d references", len(refs))
		}
	})
}

func TestArtifactReferenceRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	artifactRepo := NewArtifactRepository(db)
	artifactRefRepo := NewArtifactReferenceRepository(db)

	// Create tenant, world, artifact, and character
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

	testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	err = artifactRepo.Create(ctx, testArtifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		ref, err := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeCharacter, testCharacter.ID)
		if err != nil {
			t.Fatalf("failed to create artifact reference: %v", err)
		}
		err = artifactRefRepo.Create(ctx, ref)
		if err != nil {
			t.Fatalf("failed to create artifact reference: %v", err)
		}

		err = artifactRefRepo.Delete(ctx, testTenant.ID, ref.ID)
		if err != nil {
			t.Fatalf("failed to delete artifact reference: %v", err)
		}

		// Verify artifact reference is deleted
		_, err = artifactRefRepo.GetByID(ctx, testTenant.ID, ref.ID)
		if err == nil {
			t.Fatal("expected error for deleted artifact reference")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "artifact_reference" {
			t.Errorf("expected resource to be 'artifact_reference', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestArtifactReferenceRepository_DeleteByArtifact(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	artifactRepo := NewArtifactRepository(db)
	artifactRefRepo := NewArtifactReferenceRepository(db)

	// Create tenant, world, and artifact
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

	testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	err = artifactRepo.Create(ctx, testArtifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	t.Run("delete all references for artifact", func(t *testing.T) {
		// Create characters and locations
		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		loc1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 1", nil)
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)
		locationRepo.Create(ctx, loc1)

		// Create references
		ref1, _ := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeCharacter, char1.ID)
		ref2, _ := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeCharacter, char2.ID)
		ref3, _ := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeLocation, loc1.ID)
		artifactRefRepo.Create(ctx, ref1)
		artifactRefRepo.Create(ctx, ref2)
		artifactRefRepo.Create(ctx, ref3)

		// Verify they exist
		refs, err := artifactRefRepo.ListByArtifact(ctx, testTenant.ID, testArtifact.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(refs) != 3 {
			t.Errorf("expected 3 references, got %d", len(refs))
		}

		// Delete all by artifact
		err = artifactRefRepo.DeleteByArtifact(ctx, testTenant.ID, testArtifact.ID)
		if err != nil {
			t.Fatalf("failed to delete artifact references: %v", err)
		}

		// Verify all are deleted
		refs, err = artifactRefRepo.ListByArtifact(ctx, testTenant.ID, testArtifact.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 0 {
			t.Errorf("expected no references, got %d", len(refs))
		}
	})
}

func TestArtifactReferenceRepository_DeleteByArtifactAndEntity(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	artifactRepo := NewArtifactRepository(db)
	artifactRefRepo := NewArtifactReferenceRepository(db)

	// Create tenant, world, artifact, and character
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

	testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	err = artifactRepo.Create(ctx, testArtifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("delete specific artifact-entity reference", func(t *testing.T) {
		// Create multiple references
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		characterRepo.Create(ctx, char2)

		ref1, _ := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeCharacter, testCharacter.ID)
		ref2, _ := world.NewArtifactReference(testArtifact.ID, world.ArtifactReferenceEntityTypeCharacter, char2.ID)
		artifactRefRepo.Create(ctx, ref1)
		artifactRefRepo.Create(ctx, ref2)

		// Verify they exist
		refs, err := artifactRefRepo.ListByArtifact(ctx, testTenant.ID, testArtifact.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(refs) != 2 {
			t.Errorf("expected 2 references, got %d", len(refs))
		}

		// Delete specific reference
		err = artifactRefRepo.DeleteByArtifactAndEntity(ctx, testTenant.ID, testArtifact.ID, world.ArtifactReferenceEntityTypeCharacter, testCharacter.ID)
		if err != nil {
			t.Fatalf("failed to delete artifact reference: %v", err)
		}

		// Verify only one is deleted
		refs, err = artifactRefRepo.ListByArtifact(ctx, testTenant.ID, testArtifact.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 1 {
			t.Errorf("expected 1 reference, got %d", len(refs))
		}

		if refs[0].EntityID != char2.ID {
			t.Errorf("expected remaining reference to have entity_id %s, got %s", char2.ID, refs[0].EntityID)
		}
	})
}

