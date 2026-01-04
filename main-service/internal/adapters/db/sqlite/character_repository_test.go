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

func TestCharacterRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)

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

	t.Run("successful creation", func(t *testing.T) {
		char, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		err = characterRepo.Create(ctx, char)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify character can be retrieved
		retrieved, err := characterRepo.GetByID(ctx, testTenant.ID, char.ID)
		if err != nil {
			t.Fatalf("failed to retrieve character: %v", err)
		}

		if retrieved.Name != "Test Character" {
			t.Errorf("expected name to be 'Test Character', got '%s'", retrieved.Name)
		}

		if retrieved.ClassLevel != 1 {
			t.Errorf("expected class_level to be 1, got %d", retrieved.ClassLevel)
		}

		if retrieved.TenantID != testTenant.ID {
			t.Errorf("expected tenant_id to be %s, got %s", testTenant.ID, retrieved.TenantID)
		}
	})

	t.Run("successful creation with archetype", func(t *testing.T) {
		archetypeID := uuid.New()
		char, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character With Archetype")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		char.SetArchetype(&archetypeID)

		err = characterRepo.Create(ctx, char)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := characterRepo.GetByID(ctx, testTenant.ID, char.ID)
		if err != nil {
			t.Fatalf("failed to retrieve character: %v", err)
		}

		if retrieved.ArchetypeID == nil {
			t.Fatal("expected archetype_id to be set, got nil")
		}

		if *retrieved.ArchetypeID != archetypeID {
			t.Errorf("expected archetype_id to be %s, got %s", archetypeID, *retrieved.ArchetypeID)
		}
	})

	t.Run("successful creation with class", func(t *testing.T) {
		classID := uuid.New()
		char, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character With Class")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		char.SetClass(&classID)
		char.SetClassLevel(5)

		err = characterRepo.Create(ctx, char)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := characterRepo.GetByID(ctx, testTenant.ID, char.ID)
		if err != nil {
			t.Fatalf("failed to retrieve character: %v", err)
		}

		if retrieved.CurrentClassID == nil {
			t.Fatal("expected current_class_id to be set, got nil")
		}

		if *retrieved.CurrentClassID != classID {
			t.Errorf("expected current_class_id to be %s, got %s", classID, *retrieved.CurrentClassID)
		}

		if retrieved.ClassLevel != 5 {
			t.Errorf("expected class_level to be 5, got %d", retrieved.ClassLevel)
		}
	})

	t.Run("successful creation with description", func(t *testing.T) {
		char, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character With Description")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		char.UpdateDescription("A test character description")

		err = characterRepo.Create(ctx, char)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := characterRepo.GetByID(ctx, testTenant.ID, char.ID)
		if err != nil {
			t.Fatalf("failed to retrieve character: %v", err)
		}

		if retrieved.Description != "A test character description" {
			t.Errorf("expected description to be 'A test character description', got '%s'", retrieved.Description)
		}
	})
}

func TestCharacterRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)

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

	t.Run("existing character", func(t *testing.T) {
		char, err := world.NewCharacter(testTenant.ID, testWorld.ID, "GetByID Character")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		err = characterRepo.Create(ctx, char)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		retrieved, err := characterRepo.GetByID(ctx, testTenant.ID, char.ID)
		if err != nil {
			t.Fatalf("failed to get character: %v", err)
		}

		if retrieved.ID != char.ID {
			t.Errorf("expected ID to be %s, got %s", char.ID, retrieved.ID)
		}

		if retrieved.Name != "GetByID Character" {
			t.Errorf("expected name to be 'GetByID Character', got '%s'", retrieved.Name)
		}
	})

	t.Run("non-existent character", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := characterRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent character")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "character" {
			t.Errorf("expected resource to be 'character', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestCharacterRepository_ListByWorld(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)

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

	t.Run("empty list", func(t *testing.T) {
		characters, err := characterRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(characters) != 0 {
			t.Errorf("expected empty list, got %d characters", len(characters))
		}
	})

	t.Run("list with characters", func(t *testing.T) {
		// Create multiple characters
		characterNames := []string{"Character A", "Character B", "Character C"}
		createdCharacters := make([]*world.Character, 0, len(characterNames))

		for _, name := range characterNames {
			char, err := world.NewCharacter(testTenant.ID, testWorld.ID, name)
			if err != nil {
				t.Fatalf("failed to create character: %v", err)
			}
			err = characterRepo.Create(ctx, char)
			if err != nil {
				t.Fatalf("failed to create character: %v", err)
			}
			createdCharacters = append(createdCharacters, char)
			time.Sleep(10 * time.Millisecond)
		}

		characters, err := characterRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(characters) != len(characterNames) {
			t.Errorf("expected %d characters, got %d", len(characterNames), len(characters))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		// Create 5 characters
		for i := 1; i <= 5; i++ {
			char, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Pagination Character")
			if err != nil {
				t.Fatalf("failed to create character: %v", err)
			}
			err = characterRepo.Create(ctx, char)
			if err != nil {
				t.Fatalf("failed to create character: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}

		// Get first page
		characters, err := characterRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(characters) != 2 {
			t.Errorf("expected 2 characters, got %d", len(characters))
		}

		// Get second page
		characters, err = characterRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID, 2, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(characters) != 2 {
			t.Errorf("expected 2 characters, got %d", len(characters))
		}
	})
}

func TestCharacterRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)

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

	t.Run("successful update", func(t *testing.T) {
		char, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Update Character")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		err = characterRepo.Create(ctx, char)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		// Update name
		err = char.UpdateName("Updated Name")
		if err != nil {
			t.Fatalf("failed to update name: %v", err)
		}

		// Update description
		char.UpdateDescription("Updated Description")

		// Update class level
		char.SetClassLevel(10)

		err = characterRepo.Update(ctx, char)
		if err != nil {
			t.Fatalf("failed to update character: %v", err)
		}

		// Verify update
		retrieved, err := characterRepo.GetByID(ctx, testTenant.ID, char.ID)
		if err != nil {
			t.Fatalf("failed to get character: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("expected name to be 'Updated Name', got '%s'", retrieved.Name)
		}

		if retrieved.Description != "Updated Description" {
			t.Errorf("expected description to be 'Updated Description', got '%s'", retrieved.Description)
		}

		if retrieved.ClassLevel != 10 {
			t.Errorf("expected class_level to be 10, got %d", retrieved.ClassLevel)
		}
	})

	t.Run("update archetype", func(t *testing.T) {
		char, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Update Archetype Character")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		err = characterRepo.Create(ctx, char)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		// Set archetype
		archetypeID := uuid.New()
		char.SetArchetype(&archetypeID)
		err = characterRepo.Update(ctx, char)
		if err != nil {
			t.Fatalf("failed to update character: %v", err)
		}

		retrieved, err := characterRepo.GetByID(ctx, testTenant.ID, char.ID)
		if err != nil {
			t.Fatalf("failed to get character: %v", err)
		}

		if retrieved.ArchetypeID == nil {
			t.Fatal("expected archetype_id to be set, got nil")
		}

		if *retrieved.ArchetypeID != archetypeID {
			t.Errorf("expected archetype_id to be %s, got %s", archetypeID, *retrieved.ArchetypeID)
		}

		// Clear archetype
		char.SetArchetype(nil)
		err = characterRepo.Update(ctx, char)
		if err != nil {
			t.Fatalf("failed to update character: %v", err)
		}

		retrieved, err = characterRepo.GetByID(ctx, testTenant.ID, char.ID)
		if err != nil {
			t.Fatalf("failed to get character: %v", err)
		}

		if retrieved.ArchetypeID != nil {
			t.Error("expected archetype_id to be nil, got non-nil")
		}
	})

	t.Run("update class", func(t *testing.T) {
		char, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Update Class Character")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		err = characterRepo.Create(ctx, char)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		// Set class
		classID := uuid.New()
		char.SetClass(&classID)
		char.SetClassLevel(15)
		err = characterRepo.Update(ctx, char)
		if err != nil {
			t.Fatalf("failed to update character: %v", err)
		}

		retrieved, err := characterRepo.GetByID(ctx, testTenant.ID, char.ID)
		if err != nil {
			t.Fatalf("failed to get character: %v", err)
		}

		if retrieved.CurrentClassID == nil {
			t.Fatal("expected current_class_id to be set, got nil")
		}

		if *retrieved.CurrentClassID != classID {
			t.Errorf("expected current_class_id to be %s, got %s", classID, *retrieved.CurrentClassID)
		}

		if retrieved.ClassLevel != 15 {
			t.Errorf("expected class_level to be 15, got %d", retrieved.ClassLevel)
		}

		// Clear class
		char.SetClass(nil)
		char.SetClassLevel(1)
		err = characterRepo.Update(ctx, char)
		if err != nil {
			t.Fatalf("failed to update character: %v", err)
		}

		retrieved, err = characterRepo.GetByID(ctx, testTenant.ID, char.ID)
		if err != nil {
			t.Fatalf("failed to get character: %v", err)
		}

		if retrieved.CurrentClassID != nil {
			t.Error("expected current_class_id to be nil, got non-nil")
		}
	})
}

func TestCharacterRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)

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

	t.Run("successful delete", func(t *testing.T) {
		char, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Delete Character")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		err = characterRepo.Create(ctx, char)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		err = characterRepo.Delete(ctx, testTenant.ID, char.ID)
		if err != nil {
			t.Fatalf("failed to delete character: %v", err)
		}

		// Verify character is deleted
		_, err = characterRepo.GetByID(ctx, testTenant.ID, char.ID)
		if err == nil {
			t.Fatal("expected error for deleted character")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "character" {
			t.Errorf("expected resource to be 'character', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestCharacterRepository_CountByWorld(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)

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

	t.Run("empty count", func(t *testing.T) {
		count, err := characterRepo.CountByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 0 {
			t.Errorf("expected count to be 0, got %d", count)
		}
	})

	t.Run("count with characters", func(t *testing.T) {
		// Create 3 characters
		for i := 1; i <= 3; i++ {
			char, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Count Character")
			if err != nil {
				t.Fatalf("failed to create character: %v", err)
			}
			err = characterRepo.Create(ctx, char)
			if err != nil {
				t.Fatalf("failed to create character: %v", err)
			}
		}

		count, err := characterRepo.CountByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 3 {
			t.Errorf("expected count to be 3, got %d", count)
		}
	})
}

