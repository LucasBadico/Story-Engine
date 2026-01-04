//go:build integration

package sqlite

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/tenant"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
)

func TestLoreReferenceRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	loreRefRepo := NewLoreReferenceRepository(db)

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

	t.Run("create with character", func(t *testing.T) {
		// Create lore and character
		testLore, err := world.NewLore(testTenant.ID, testWorld.ID, "Test Lore", nil)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}
		err = loreRepo.Create(ctx, testLore)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}

		testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		err = characterRepo.Create(ctx, testCharacter)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		// Create lore reference
		loreRef := world.NewLoreReference(testLore.ID, "character", testCharacter.ID, nil)
		loreRef.UpdateNotes("Test notes")

		err = loreRefRepo.Create(ctx, loreRef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify lore reference can be retrieved
		retrieved, err := loreRefRepo.GetByID(ctx, testTenant.ID, loreRef.ID)
		if err != nil {
			t.Fatalf("failed to retrieve lore reference: %v", err)
		}

		if retrieved.LoreID != testLore.ID {
			t.Errorf("expected lore_id to be %s, got %s", testLore.ID, retrieved.LoreID)
		}

		if retrieved.EntityType != "character" {
			t.Errorf("expected entity_type to be 'character', got '%s'", retrieved.EntityType)
		}

		if retrieved.EntityID != testCharacter.ID {
			t.Errorf("expected entity_id to be %s, got %s", testCharacter.ID, retrieved.EntityID)
		}

		if retrieved.Notes != "Test notes" {
			t.Errorf("expected notes to be 'Test notes', got '%s'", retrieved.Notes)
		}
	})

	t.Run("create with location", func(t *testing.T) {
		testLore, err := world.NewLore(testTenant.ID, testWorld.ID, "Test Lore 2", nil)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}
		err = loreRepo.Create(ctx, testLore)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}

		testLocation, err := world.NewLocation(testTenant.ID, testWorld.ID, "Test Location", nil)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}
		err = locationRepo.Create(ctx, testLocation)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		relationshipType := "location_of"
		loreRef := world.NewLoreReference(testLore.ID, "location", testLocation.ID, &relationshipType)

		err = loreRefRepo.Create(ctx, loreRef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := loreRefRepo.GetByID(ctx, testTenant.ID, loreRef.ID)
		if err != nil {
			t.Fatalf("failed to retrieve lore reference: %v", err)
		}

		if retrieved.RelationshipType == nil || *retrieved.RelationshipType != relationshipType {
			t.Errorf("expected relationship_type to be '%s', got %v", relationshipType, retrieved.RelationshipType)
		}
	})
}

func TestLoreReferenceRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)
	characterRepo := NewCharacterRepository(db)
	loreRefRepo := NewLoreReferenceRepository(db)

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

	t.Run("existing lore reference", func(t *testing.T) {
		testLore, _ := world.NewLore(testTenant.ID, testWorld.ID, "Test Lore", nil)
		loreRepo.Create(ctx, testLore)

		testCharacter, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		characterRepo.Create(ctx, testCharacter)

		loreRef := world.NewLoreReference(testLore.ID, "character", testCharacter.ID, nil)
		loreRefRepo.Create(ctx, loreRef)

		retrieved, err := loreRefRepo.GetByID(ctx, testTenant.ID, loreRef.ID)
		if err != nil {
			t.Fatalf("failed to get lore reference: %v", err)
		}

		if retrieved.ID != loreRef.ID {
			t.Errorf("expected ID to be %s, got %s", loreRef.ID, retrieved.ID)
		}
	})

	t.Run("non-existent lore reference", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := loreRefRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent lore reference")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "lore_reference" {
			t.Errorf("expected resource to be 'lore_reference', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestLoreReferenceRepository_ListByLore(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)
	characterRepo := NewCharacterRepository(db)
	loreRefRepo := NewLoreReferenceRepository(db)

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

	t.Run("list by lore", func(t *testing.T) {
		testLore, _ := world.NewLore(testTenant.ID, testWorld.ID, "Test Lore", nil)
		loreRepo.Create(ctx, testLore)

		// Create characters
		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)

		// Create lore references
		ref1 := world.NewLoreReference(testLore.ID, "character", char1.ID, nil)
		ref2 := world.NewLoreReference(testLore.ID, "character", char2.ID, nil)
		loreRefRepo.Create(ctx, ref1)
		loreRefRepo.Create(ctx, ref2)

		refs, err := loreRefRepo.ListByLore(ctx, testTenant.ID, testLore.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 2 {
			t.Errorf("expected 2 references, got %d", len(refs))
		}

		for _, ref := range refs {
			if ref.LoreID != testLore.ID {
				t.Errorf("expected lore_id to be %s, got %s", testLore.ID, ref.LoreID)
			}
		}
	})
}

func TestLoreReferenceRepository_ListByEntity(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)
	characterRepo := NewCharacterRepository(db)
	loreRefRepo := NewLoreReferenceRepository(db)

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

	t.Run("list by entity", func(t *testing.T) {
		testCharacter, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		characterRepo.Create(ctx, testCharacter)

		// Create multiple lores
		lore1, _ := world.NewLore(testTenant.ID, testWorld.ID, "Lore 1", nil)
		lore2, _ := world.NewLore(testTenant.ID, testWorld.ID, "Lore 2", nil)
		loreRepo.Create(ctx, lore1)
		loreRepo.Create(ctx, lore2)

		// Create references from different lores to same character
		ref1 := world.NewLoreReference(lore1.ID, "character", testCharacter.ID, nil)
		ref2 := world.NewLoreReference(lore2.ID, "character", testCharacter.ID, nil)
		loreRefRepo.Create(ctx, ref1)
		loreRefRepo.Create(ctx, ref2)

		refs, err := loreRefRepo.ListByEntity(ctx, testTenant.ID, "character", testCharacter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 2 {
			t.Errorf("expected 2 references, got %d", len(refs))
		}

		for _, ref := range refs {
			if ref.EntityType != "character" {
				t.Errorf("expected entity_type to be 'character', got '%s'", ref.EntityType)
			}

			if ref.EntityID != testCharacter.ID {
				t.Errorf("expected entity_id to be %s, got %s", testCharacter.ID, ref.EntityID)
			}
		}
	})
}

func TestLoreReferenceRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)
	characterRepo := NewCharacterRepository(db)
	loreRefRepo := NewLoreReferenceRepository(db)

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

	t.Run("update notes and relationship type", func(t *testing.T) {
		testLore, _ := world.NewLore(testTenant.ID, testWorld.ID, "Test Lore", nil)
		loreRepo.Create(ctx, testLore)

		testCharacter, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		characterRepo.Create(ctx, testCharacter)

		loreRef := world.NewLoreReference(testLore.ID, "character", testCharacter.ID, nil)
		loreRefRepo.Create(ctx, loreRef)

		// Update
		newRelationshipType := "related_to"
		loreRef.UpdateRelationshipType(&newRelationshipType)
		loreRef.UpdateNotes("Updated notes")

		err = loreRefRepo.Update(ctx, loreRef)
		if err != nil {
			t.Fatalf("failed to update lore reference: %v", err)
		}

		// Verify update
		retrieved, err := loreRefRepo.GetByID(ctx, testTenant.ID, loreRef.ID)
		if err != nil {
			t.Fatalf("failed to get lore reference: %v", err)
		}

		if retrieved.RelationshipType == nil || *retrieved.RelationshipType != newRelationshipType {
			t.Errorf("expected relationship_type to be '%s', got %v", newRelationshipType, retrieved.RelationshipType)
		}

		if retrieved.Notes != "Updated notes" {
			t.Errorf("expected notes to be 'Updated notes', got '%s'", retrieved.Notes)
		}
	})
}

func TestLoreReferenceRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)
	characterRepo := NewCharacterRepository(db)
	loreRefRepo := NewLoreReferenceRepository(db)

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

	t.Run("delete by id", func(t *testing.T) {
		testLore, _ := world.NewLore(testTenant.ID, testWorld.ID, "Test Lore", nil)
		loreRepo.Create(ctx, testLore)

		testCharacter, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		characterRepo.Create(ctx, testCharacter)

		loreRef := world.NewLoreReference(testLore.ID, "character", testCharacter.ID, nil)
		loreRefRepo.Create(ctx, loreRef)

		err = loreRefRepo.Delete(ctx, testTenant.ID, loreRef.ID)
		if err != nil {
			t.Fatalf("failed to delete lore reference: %v", err)
		}

		// Verify deletion
		_, err = loreRefRepo.GetByID(ctx, testTenant.ID, loreRef.ID)
		if err == nil {
			t.Fatal("expected error for deleted lore reference")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "lore_reference" {
			t.Errorf("expected resource to be 'lore_reference', got '%s'", notFoundErr.Resource)
		}
	})

	t.Run("delete by lore", func(t *testing.T) {
		testLore, _ := world.NewLore(testTenant.ID, testWorld.ID, "Test Lore 2", nil)
		loreRepo.Create(ctx, testLore)

		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)

		ref1 := world.NewLoreReference(testLore.ID, "character", char1.ID, nil)
		ref2 := world.NewLoreReference(testLore.ID, "character", char2.ID, nil)
		loreRefRepo.Create(ctx, ref1)
		loreRefRepo.Create(ctx, ref2)

		err = loreRefRepo.DeleteByLore(ctx, testTenant.ID, testLore.ID)
		if err != nil {
			t.Fatalf("failed to delete by lore: %v", err)
		}

		// Verify all references deleted
		refs, err := loreRefRepo.ListByLore(ctx, testTenant.ID, testLore.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 0 {
			t.Errorf("expected no references, got %d", len(refs))
		}
	})

	t.Run("delete by lore and entity", func(t *testing.T) {
		testLore, _ := world.NewLore(testTenant.ID, testWorld.ID, "Test Lore 3", nil)
		loreRepo.Create(ctx, testLore)

		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)

		ref1 := world.NewLoreReference(testLore.ID, "character", char1.ID, nil)
		ref2 := world.NewLoreReference(testLore.ID, "character", char2.ID, nil)
		loreRefRepo.Create(ctx, ref1)
		loreRefRepo.Create(ctx, ref2)

		err = loreRefRepo.DeleteByLoreAndEntity(ctx, testTenant.ID, testLore.ID, "character", char1.ID)
		if err != nil {
			t.Fatalf("failed to delete by lore and entity: %v", err)
		}

		// Verify only ref1 was deleted
		refs, err := loreRefRepo.ListByLore(ctx, testTenant.ID, testLore.ID)
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

