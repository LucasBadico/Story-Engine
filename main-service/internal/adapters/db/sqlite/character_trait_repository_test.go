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

func TestCharacterTraitRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	traitRepo := NewTraitRepository(db)
	characterTraitRepo := NewCharacterTraitRepository(db)

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

	// Create character
	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	// Create trait
	testTrait, err := world.NewTrait(testTenant.ID, "Test Trait")
	if err != nil {
		t.Fatalf("failed to create trait: %v", err)
	}
	err = traitRepo.Create(ctx, testTrait)
	if err != nil {
		t.Fatalf("failed to create trait: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		ct := world.NewCharacterTrait(testCharacter.ID, testTrait.ID, testTrait, "Default Value")

		err = characterTraitRepo.Create(ctx, ct)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify character-trait can be retrieved
		retrieved, err := characterTraitRepo.GetByID(ctx, testTenant.ID, ct.ID)
		if err != nil {
			t.Fatalf("failed to retrieve character-trait: %v", err)
		}

		if retrieved.CharacterID != testCharacter.ID {
			t.Errorf("expected character_id to be %s, got %s", testCharacter.ID, retrieved.CharacterID)
		}

		if retrieved.TraitID != testTrait.ID {
			t.Errorf("expected trait_id to be %s, got %s", testTrait.ID, retrieved.TraitID)
		}

		if retrieved.TraitName != testTrait.Name {
			t.Errorf("expected trait_name to be '%s', got '%s'", testTrait.Name, retrieved.TraitName)
		}

		if retrieved.Value != "Default Value" {
			t.Errorf("expected value to be 'Default Value', got '%s'", retrieved.Value)
		}
	})

	t.Run("successful creation with notes", func(t *testing.T) {
		trait2, err := world.NewTrait(testTenant.ID, "Trait With Notes")
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}
		err = traitRepo.Create(ctx, trait2)
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		ct := world.NewCharacterTrait(testCharacter.ID, trait2.ID, trait2, "Value")
		ct.UpdateNotes("Some notes")

		err = characterTraitRepo.Create(ctx, ct)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := characterTraitRepo.GetByID(ctx, testTenant.ID, ct.ID)
		if err != nil {
			t.Fatalf("failed to retrieve character-trait: %v", err)
		}

		if retrieved.Notes != "Some notes" {
			t.Errorf("expected notes to be 'Some notes', got '%s'", retrieved.Notes)
		}
	})

	t.Run("trait snapshot copied", func(t *testing.T) {
		trait3, err := world.NewTrait(testTenant.ID, "Snapshot Trait")
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}
		trait3.UpdateCategory("Category")
		trait3.UpdateDescription("Description")
		err = traitRepo.Create(ctx, trait3)
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		ct := world.NewCharacterTrait(testCharacter.ID, trait3.ID, trait3, "Value")

		err = characterTraitRepo.Create(ctx, ct)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := characterTraitRepo.GetByID(ctx, testTenant.ID, ct.ID)
		if err != nil {
			t.Fatalf("failed to retrieve character-trait: %v", err)
		}

		if retrieved.TraitCategory != "Category" {
			t.Errorf("expected trait_category to be 'Category', got '%s'", retrieved.TraitCategory)
		}

		if retrieved.TraitDescription != "Description" {
			t.Errorf("expected trait_description to be 'Description', got '%s'", retrieved.TraitDescription)
		}
	})
}

func TestCharacterTraitRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	traitRepo := NewTraitRepository(db)
	characterTraitRepo := NewCharacterTraitRepository(db)

	// Create tenant, world, character, and trait
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

	testTrait, err := world.NewTrait(testTenant.ID, "Test Trait")
	if err != nil {
		t.Fatalf("failed to create trait: %v", err)
	}
	err = traitRepo.Create(ctx, testTrait)
	if err != nil {
		t.Fatalf("failed to create trait: %v", err)
	}

	t.Run("existing character-trait", func(t *testing.T) {
		ct := world.NewCharacterTrait(testCharacter.ID, testTrait.ID, testTrait, "Value")
		err = characterTraitRepo.Create(ctx, ct)
		if err != nil {
			t.Fatalf("failed to create character-trait: %v", err)
		}

		retrieved, err := characterTraitRepo.GetByID(ctx, testTenant.ID, ct.ID)
		if err != nil {
			t.Fatalf("failed to get character-trait: %v", err)
		}

		if retrieved.ID != ct.ID {
			t.Errorf("expected ID to be %s, got %s", ct.ID, retrieved.ID)
		}

		if retrieved.CharacterID != testCharacter.ID {
			t.Errorf("expected character_id to be %s, got %s", testCharacter.ID, retrieved.CharacterID)
		}
	})

	t.Run("non-existent character-trait", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := characterTraitRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent character-trait")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "character_trait" {
			t.Errorf("expected resource to be 'character_trait', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestCharacterTraitRepository_GetByCharacter(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	traitRepo := NewTraitRepository(db)
	characterTraitRepo := NewCharacterTraitRepository(db)

	// Create tenant, world, and character
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

	t.Run("empty list", func(t *testing.T) {
		characterTraits, err := characterTraitRepo.GetByCharacter(ctx, testTenant.ID, testCharacter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(characterTraits) != 0 {
			t.Errorf("expected empty list, got %d character-traits", len(characterTraits))
		}
	})

	t.Run("list with character-traits", func(t *testing.T) {
		// Create multiple traits
		trait1, _ := world.NewTrait(testTenant.ID, "Trait 1")
		trait2, _ := world.NewTrait(testTenant.ID, "Trait 2")
		trait3, _ := world.NewTrait(testTenant.ID, "Trait 3")
		traitRepo.Create(ctx, trait1)
		traitRepo.Create(ctx, trait2)
		traitRepo.Create(ctx, trait3)

		// Create character-traits
		ct1 := world.NewCharacterTrait(testCharacter.ID, trait1.ID, trait1, "Value 1")
		ct2 := world.NewCharacterTrait(testCharacter.ID, trait2.ID, trait2, "Value 2")
		ct3 := world.NewCharacterTrait(testCharacter.ID, trait3.ID, trait3, "Value 3")
		characterTraitRepo.Create(ctx, ct1)
		time.Sleep(10 * time.Millisecond)
		characterTraitRepo.Create(ctx, ct2)
		time.Sleep(10 * time.Millisecond)
		characterTraitRepo.Create(ctx, ct3)

		characterTraits, err := characterTraitRepo.GetByCharacter(ctx, testTenant.ID, testCharacter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(characterTraits) != 3 {
			t.Errorf("expected 3 character-traits, got %d", len(characterTraits))
		}

		// Verify all belong to the character
		for _, ct := range characterTraits {
			if ct.CharacterID != testCharacter.ID {
				t.Errorf("expected character_id to be %s, got %s", testCharacter.ID, ct.CharacterID)
			}
		}
	})
}

func TestCharacterTraitRepository_GetByCharacterAndTrait(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	traitRepo := NewTraitRepository(db)
	characterTraitRepo := NewCharacterTraitRepository(db)

	// Create tenant, world, character, and trait
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

	testTrait, err := world.NewTrait(testTenant.ID, "Test Trait")
	if err != nil {
		t.Fatalf("failed to create trait: %v", err)
	}
	err = traitRepo.Create(ctx, testTrait)
	if err != nil {
		t.Fatalf("failed to create trait: %v", err)
	}

	t.Run("existing character-trait", func(t *testing.T) {
		ct := world.NewCharacterTrait(testCharacter.ID, testTrait.ID, testTrait, "Value")
		err = characterTraitRepo.Create(ctx, ct)
		if err != nil {
			t.Fatalf("failed to create character-trait: %v", err)
		}

		retrieved, err := characterTraitRepo.GetByCharacterAndTrait(ctx, testTenant.ID, testCharacter.ID, testTrait.ID)
		if err != nil {
			t.Fatalf("failed to get character-trait: %v", err)
		}

		if retrieved.ID != ct.ID {
			t.Errorf("expected ID to be %s, got %s", ct.ID, retrieved.ID)
		}

		if retrieved.CharacterID != testCharacter.ID {
			t.Errorf("expected character_id to be %s, got %s", testCharacter.ID, retrieved.CharacterID)
		}

		if retrieved.TraitID != testTrait.ID {
			t.Errorf("expected trait_id to be %s, got %s", testTrait.ID, retrieved.TraitID)
		}
	})

	t.Run("non-existent character-trait", func(t *testing.T) {
		nonExistentTraitID := uuid.New()

		_, err := characterTraitRepo.GetByCharacterAndTrait(ctx, testTenant.ID, testCharacter.ID, nonExistentTraitID)
		if err == nil {
			t.Fatal("expected error for non-existent character-trait")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "character_trait" {
			t.Errorf("expected resource to be 'character_trait', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestCharacterTraitRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	traitRepo := NewTraitRepository(db)
	characterTraitRepo := NewCharacterTraitRepository(db)

	// Create tenant, world, character, and trait
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

	testTrait, err := world.NewTrait(testTenant.ID, "Test Trait")
	if err != nil {
		t.Fatalf("failed to create trait: %v", err)
	}
	err = traitRepo.Create(ctx, testTrait)
	if err != nil {
		t.Fatalf("failed to create trait: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		ct := world.NewCharacterTrait(testCharacter.ID, testTrait.ID, testTrait, "Initial Value")
		err = characterTraitRepo.Create(ctx, ct)
		if err != nil {
			t.Fatalf("failed to create character-trait: %v", err)
		}

		// Update value
		ct.UpdateValue("Updated Value")

		// Update notes
		ct.UpdateNotes("Updated Notes")

		err = characterTraitRepo.Update(ctx, ct)
		if err != nil {
			t.Fatalf("failed to update character-trait: %v", err)
		}

		// Verify update
		retrieved, err := characterTraitRepo.GetByID(ctx, testTenant.ID, ct.ID)
		if err != nil {
			t.Fatalf("failed to get character-trait: %v", err)
		}

		if retrieved.Value != "Updated Value" {
			t.Errorf("expected value to be 'Updated Value', got '%s'", retrieved.Value)
		}

		if retrieved.Notes != "Updated Notes" {
			t.Errorf("expected notes to be 'Updated Notes', got '%s'", retrieved.Notes)
		}
	})

	t.Run("update trait snapshot", func(t *testing.T) {
		trait2, err := world.NewTrait(testTenant.ID, "Trait Snapshot")
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}
		err = traitRepo.Create(ctx, trait2)
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		ct := world.NewCharacterTrait(testCharacter.ID, trait2.ID, trait2, "Value")
		err = characterTraitRepo.Create(ctx, ct)
		if err != nil {
			t.Fatalf("failed to create character-trait: %v", err)
		}

		// Update trait
		trait2.UpdateName("Updated Trait Name")
		trait2.UpdateCategory("Updated Category")
		trait2.UpdateDescription("Updated Description")
		err = traitRepo.Update(ctx, trait2)
		if err != nil {
			t.Fatalf("failed to update trait: %v", err)
		}

		// Update snapshot in character-trait
		ct.UpdateTraitSnapshot(trait2)
		err = characterTraitRepo.Update(ctx, ct)
		if err != nil {
			t.Fatalf("failed to update character-trait: %v", err)
		}

		// Verify update
		retrieved, err := characterTraitRepo.GetByID(ctx, testTenant.ID, ct.ID)
		if err != nil {
			t.Fatalf("failed to get character-trait: %v", err)
		}

		if retrieved.TraitName != "Updated Trait Name" {
			t.Errorf("expected trait_name to be 'Updated Trait Name', got '%s'", retrieved.TraitName)
		}

		if retrieved.TraitCategory != "Updated Category" {
			t.Errorf("expected trait_category to be 'Updated Category', got '%s'", retrieved.TraitCategory)
		}

		if retrieved.TraitDescription != "Updated Description" {
			t.Errorf("expected trait_description to be 'Updated Description', got '%s'", retrieved.TraitDescription)
		}
	})
}

func TestCharacterTraitRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	traitRepo := NewTraitRepository(db)
	characterTraitRepo := NewCharacterTraitRepository(db)

	// Create tenant, world, character, and trait
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

	testTrait, err := world.NewTrait(testTenant.ID, "Test Trait")
	if err != nil {
		t.Fatalf("failed to create trait: %v", err)
	}
	err = traitRepo.Create(ctx, testTrait)
	if err != nil {
		t.Fatalf("failed to create trait: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		ct := world.NewCharacterTrait(testCharacter.ID, testTrait.ID, testTrait, "Value")
		err = characterTraitRepo.Create(ctx, ct)
		if err != nil {
			t.Fatalf("failed to create character-trait: %v", err)
		}

		err = characterTraitRepo.Delete(ctx, testTenant.ID, ct.ID)
		if err != nil {
			t.Fatalf("failed to delete character-trait: %v", err)
		}

		// Verify character-trait is deleted
		_, err = characterTraitRepo.GetByID(ctx, testTenant.ID, ct.ID)
		if err == nil {
			t.Fatal("expected error for deleted character-trait")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "character_trait" {
			t.Errorf("expected resource to be 'character_trait', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestCharacterTraitRepository_DeleteByCharacter(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	traitRepo := NewTraitRepository(db)
	characterTraitRepo := NewCharacterTraitRepository(db)

	// Create tenant, world, and character
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

	t.Run("delete all character-traits for character", func(t *testing.T) {
		// Create multiple traits
		trait1, _ := world.NewTrait(testTenant.ID, "Trait 1")
		trait2, _ := world.NewTrait(testTenant.ID, "Trait 2")
		trait3, _ := world.NewTrait(testTenant.ID, "Trait 3")
		traitRepo.Create(ctx, trait1)
		traitRepo.Create(ctx, trait2)
		traitRepo.Create(ctx, trait3)

		// Create character-traits
		ct1 := world.NewCharacterTrait(testCharacter.ID, trait1.ID, trait1, "Value 1")
		ct2 := world.NewCharacterTrait(testCharacter.ID, trait2.ID, trait2, "Value 2")
		ct3 := world.NewCharacterTrait(testCharacter.ID, trait3.ID, trait3, "Value 3")
		characterTraitRepo.Create(ctx, ct1)
		characterTraitRepo.Create(ctx, ct2)
		characterTraitRepo.Create(ctx, ct3)

		// Verify they exist
		characterTraits, err := characterTraitRepo.GetByCharacter(ctx, testTenant.ID, testCharacter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(characterTraits) != 3 {
			t.Errorf("expected 3 character-traits, got %d", len(characterTraits))
		}

		// Delete all by character
		err = characterTraitRepo.DeleteByCharacter(ctx, testTenant.ID, testCharacter.ID)
		if err != nil {
			t.Fatalf("failed to delete character-traits: %v", err)
		}

		// Verify all are deleted
		characterTraits, err = characterTraitRepo.GetByCharacter(ctx, testTenant.ID, testCharacter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(characterTraits) != 0 {
			t.Errorf("expected no character-traits, got %d", len(characterTraits))
		}
	})

	t.Run("delete for character with no traits", func(t *testing.T) {
		char2, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character No Traits")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		err = characterRepo.Create(ctx, char2)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		err = characterTraitRepo.DeleteByCharacter(ctx, testTenant.ID, char2.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

