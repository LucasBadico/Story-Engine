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

func TestCharacterRelationshipRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	characterRelationshipRepo := NewCharacterRelationshipRepository(db)

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

	// Create two characters
	char1, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
	if err != nil {
		t.Fatalf("failed to create character 1: %v", err)
	}
	err = characterRepo.Create(ctx, char1)
	if err != nil {
		t.Fatalf("failed to create character 1: %v", err)
	}

	char2, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	if err != nil {
		t.Fatalf("failed to create character 2: %v", err)
	}
	err = characterRepo.Create(ctx, char2)
	if err != nil {
		t.Fatalf("failed to create character 2: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		cr, err := world.NewCharacterRelationship(testTenant.ID, char1.ID, char2.ID, "ally")
		if err != nil {
			t.Fatalf("failed to create character relationship: %v", err)
		}
		cr.Description = "Best friends"
		cr.Bidirectional = true

		err = characterRelationshipRepo.Create(ctx, cr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify relationship can be retrieved
		retrieved, err := characterRelationshipRepo.GetByID(ctx, testTenant.ID, cr.ID)
		if err != nil {
			t.Fatalf("failed to retrieve relationship: %v", err)
		}

		if retrieved.Character1ID != char1.ID {
			t.Errorf("expected character1_id to be %s, got %s", char1.ID, retrieved.Character1ID)
		}

		if retrieved.Character2ID != char2.ID {
			t.Errorf("expected character2_id to be %s, got %s", char2.ID, retrieved.Character2ID)
		}

		if retrieved.RelationshipType != "ally" {
			t.Errorf("expected relationship_type to be 'ally', got '%s'", retrieved.RelationshipType)
		}

		if retrieved.Description != "Best friends" {
			t.Errorf("expected description to be 'Best friends', got '%s'", retrieved.Description)
		}

		if !retrieved.Bidirectional {
			t.Error("expected bidirectional to be true")
		}
	})

	t.Run("successful creation without description", func(t *testing.T) {
		char3, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 3")
		if err != nil {
			t.Fatalf("failed to create character 3: %v", err)
		}
		err = characterRepo.Create(ctx, char3)
		if err != nil {
			t.Fatalf("failed to create character 3: %v", err)
		}

		cr, err := world.NewCharacterRelationship(testTenant.ID, char1.ID, char3.ID, "enemy")
		if err != nil {
			t.Fatalf("failed to create character relationship: %v", err)
		}

		err = characterRelationshipRepo.Create(ctx, cr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := characterRelationshipRepo.GetByID(ctx, testTenant.ID, cr.ID)
		if err != nil {
			t.Fatalf("failed to retrieve relationship: %v", err)
		}

		if retrieved.Description != "" {
			t.Errorf("expected empty description, got '%s'", retrieved.Description)
		}

		if retrieved.Bidirectional {
			t.Error("expected bidirectional to be false by default")
		}
	})
}

func TestCharacterRelationshipRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	characterRelationshipRepo := NewCharacterRelationshipRepository(db)

	// Create tenant, world, and characters
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

	char1, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
	if err != nil {
		t.Fatalf("failed to create character 1: %v", err)
	}
	err = characterRepo.Create(ctx, char1)
	if err != nil {
		t.Fatalf("failed to create character 1: %v", err)
	}

	char2, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	if err != nil {
		t.Fatalf("failed to create character 2: %v", err)
	}
	err = characterRepo.Create(ctx, char2)
	if err != nil {
		t.Fatalf("failed to create character 2: %v", err)
	}

	t.Run("existing relationship", func(t *testing.T) {
		cr, err := world.NewCharacterRelationship(testTenant.ID, char1.ID, char2.ID, "mentor")
		if err != nil {
			t.Fatalf("failed to create character relationship: %v", err)
		}
		err = characterRelationshipRepo.Create(ctx, cr)
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}

		retrieved, err := characterRelationshipRepo.GetByID(ctx, testTenant.ID, cr.ID)
		if err != nil {
			t.Fatalf("failed to get relationship: %v", err)
		}

		if retrieved.ID != cr.ID {
			t.Errorf("expected ID to be %s, got %s", cr.ID, retrieved.ID)
		}

		if retrieved.TenantID != testTenant.ID {
			t.Errorf("expected tenant_id to be %s, got %s", testTenant.ID, retrieved.TenantID)
		}

		if retrieved.RelationshipType != "mentor" {
			t.Errorf("expected relationship_type to be 'mentor', got '%s'", retrieved.RelationshipType)
		}
	})

	t.Run("non-existent relationship", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := characterRelationshipRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent relationship")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "character_relationship" {
			t.Errorf("expected resource to be 'character_relationship', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestCharacterRelationshipRepository_ListByCharacter(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	characterRelationshipRepo := NewCharacterRelationshipRepository(db)

	// Create tenant and world
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

	// Create multiple characters
	mainChar, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Main Character")
	if err != nil {
		t.Fatalf("failed to create main character: %v", err)
	}
	err = characterRepo.Create(ctx, mainChar)
	if err != nil {
		t.Fatalf("failed to create main character: %v", err)
	}

	char2, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	if err != nil {
		t.Fatalf("failed to create character 2: %v", err)
	}
	err = characterRepo.Create(ctx, char2)
	if err != nil {
		t.Fatalf("failed to create character 2: %v", err)
	}

	char3, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 3")
	if err != nil {
		t.Fatalf("failed to create character 3: %v", err)
	}
	err = characterRepo.Create(ctx, char3)
	if err != nil {
		t.Fatalf("failed to create character 3: %v", err)
	}

	t.Run("empty list", func(t *testing.T) {
		otherChar, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Lonely Character")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		err = characterRepo.Create(ctx, otherChar)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		relationships, err := characterRelationshipRepo.ListByCharacter(ctx, testTenant.ID, otherChar.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(relationships) != 0 {
			t.Errorf("expected empty list, got %d relationships", len(relationships))
		}
	})

	t.Run("list with relationships as character1", func(t *testing.T) {
		// Create relationships where mainChar is character1
		cr1, err := world.NewCharacterRelationship(testTenant.ID, mainChar.ID, char2.ID, "friend")
		if err != nil {
			t.Fatalf("failed to create relationship 1: %v", err)
		}
		err = characterRelationshipRepo.Create(ctx, cr1)
		if err != nil {
			t.Fatalf("failed to create relationship 1: %v", err)
		}

		cr2, err := world.NewCharacterRelationship(testTenant.ID, mainChar.ID, char3.ID, "rival")
		if err != nil {
			t.Fatalf("failed to create relationship 2: %v", err)
		}
		err = characterRelationshipRepo.Create(ctx, cr2)
		if err != nil {
			t.Fatalf("failed to create relationship 2: %v", err)
		}

		relationships, err := characterRelationshipRepo.ListByCharacter(ctx, testTenant.ID, mainChar.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(relationships) < 2 {
			t.Errorf("expected at least 2 relationships, got %d", len(relationships))
		}
	})

	t.Run("list with relationships as character2", func(t *testing.T) {
		// Create a relationship where mainChar is character2
		cr, err := world.NewCharacterRelationship(testTenant.ID, char2.ID, mainChar.ID, "student")
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}
		err = characterRelationshipRepo.Create(ctx, cr)
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}

		// ListByCharacter should find relationships where character is either character1 or character2
		relationships, err := characterRelationshipRepo.ListByCharacter(ctx, testTenant.ID, mainChar.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should include all relationships involving mainChar
		if len(relationships) < 3 {
			t.Errorf("expected at least 3 relationships (both as char1 and char2), got %d", len(relationships))
		}

		// Verify the new relationship is in the list
		found := false
		for _, r := range relationships {
			if r.ID == cr.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("relationship where mainChar is character2 not found in list")
		}
	})
}

func TestCharacterRelationshipRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	characterRelationshipRepo := NewCharacterRelationshipRepository(db)

	// Create tenant, world, and characters
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

	char1, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
	if err != nil {
		t.Fatalf("failed to create character 1: %v", err)
	}
	err = characterRepo.Create(ctx, char1)
	if err != nil {
		t.Fatalf("failed to create character 1: %v", err)
	}

	char2, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	if err != nil {
		t.Fatalf("failed to create character 2: %v", err)
	}
	err = characterRepo.Create(ctx, char2)
	if err != nil {
		t.Fatalf("failed to create character 2: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		cr, err := world.NewCharacterRelationship(testTenant.ID, char1.ID, char2.ID, "acquaintance")
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}
		err = characterRelationshipRepo.Create(ctx, cr)
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}

		// Update relationship
		err = cr.UpdateRelationshipType("friend")
		if err != nil {
			t.Fatalf("failed to update relationship type: %v", err)
		}
		cr.UpdateDescription("They became good friends")
		cr.UpdateBidirectional(true)

		err = characterRelationshipRepo.Update(ctx, cr)
		if err != nil {
			t.Fatalf("failed to update relationship: %v", err)
		}

		// Verify update
		retrieved, err := characterRelationshipRepo.GetByID(ctx, testTenant.ID, cr.ID)
		if err != nil {
			t.Fatalf("failed to get relationship: %v", err)
		}

		if retrieved.RelationshipType != "friend" {
			t.Errorf("expected relationship_type to be 'friend', got '%s'", retrieved.RelationshipType)
		}

		if retrieved.Description != "They became good friends" {
			t.Errorf("expected description to be 'They became good friends', got '%s'", retrieved.Description)
		}

		if !retrieved.Bidirectional {
			t.Error("expected bidirectional to be true")
		}
	})

	t.Run("update description only", func(t *testing.T) {
		// Create new characters for this test case to avoid UNIQUE constraint violation
		char3, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 3")
		if err != nil {
			t.Fatalf("failed to create character 3: %v", err)
		}
		err = characterRepo.Create(ctx, char3)
		if err != nil {
			t.Fatalf("failed to create character 3: %v", err)
		}

		char4, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 4")
		if err != nil {
			t.Fatalf("failed to create character 4: %v", err)
		}
		err = characterRepo.Create(ctx, char4)
		if err != nil {
			t.Fatalf("failed to create character 4: %v", err)
		}

		cr, err := world.NewCharacterRelationship(testTenant.ID, char3.ID, char4.ID, "enemy")
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}
		cr.Bidirectional = true
		err = characterRelationshipRepo.Create(ctx, cr)
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}

		// Update only description
		cr.UpdateDescription("Mortal enemies")

		err = characterRelationshipRepo.Update(ctx, cr)
		if err != nil {
			t.Fatalf("failed to update relationship: %v", err)
		}

		retrieved, err := characterRelationshipRepo.GetByID(ctx, testTenant.ID, cr.ID)
		if err != nil {
			t.Fatalf("failed to get relationship: %v", err)
		}

		// RelationshipType and Bidirectional should remain unchanged
		if retrieved.RelationshipType != "enemy" {
			t.Errorf("expected relationship_type to remain 'enemy', got '%s'", retrieved.RelationshipType)
		}

		if !retrieved.Bidirectional {
			t.Error("expected bidirectional to remain true")
		}

		if retrieved.Description != "Mortal enemies" {
			t.Errorf("expected description to be 'Mortal enemies', got '%s'", retrieved.Description)
		}
	})
}

func TestCharacterRelationshipRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	characterRelationshipRepo := NewCharacterRelationshipRepository(db)

	// Create tenant, world, and characters
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

	char1, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
	if err != nil {
		t.Fatalf("failed to create character 1: %v", err)
	}
	err = characterRepo.Create(ctx, char1)
	if err != nil {
		t.Fatalf("failed to create character 1: %v", err)
	}

	char2, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	if err != nil {
		t.Fatalf("failed to create character 2: %v", err)
	}
	err = characterRepo.Create(ctx, char2)
	if err != nil {
		t.Fatalf("failed to create character 2: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		cr, err := world.NewCharacterRelationship(testTenant.ID, char1.ID, char2.ID, "temporary")
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}
		err = characterRelationshipRepo.Create(ctx, cr)
		if err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}

		err = characterRelationshipRepo.Delete(ctx, testTenant.ID, cr.ID)
		if err != nil {
			t.Fatalf("failed to delete relationship: %v", err)
		}

		// Verify relationship is deleted
		_, err = characterRelationshipRepo.GetByID(ctx, testTenant.ID, cr.ID)
		if err == nil {
			t.Fatal("expected error for deleted relationship")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "character_relationship" {
			t.Errorf("expected resource to be 'character_relationship', got '%s'", notFoundErr.Resource)
		}
	})

	t.Run("delete non-existent relationship", func(t *testing.T) {
		nonExistentID := uuid.New()

		// Delete should not error even for non-existent ID (idempotent)
		err := characterRelationshipRepo.Delete(ctx, testTenant.ID, nonExistentID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("deleting one relationship doesn't affect others", func(t *testing.T) {
		char3, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 3")
		if err != nil {
			t.Fatalf("failed to create character 3: %v", err)
		}
		err = characterRepo.Create(ctx, char3)
		if err != nil {
			t.Fatalf("failed to create character 3: %v", err)
		}

		// Create two relationships
		cr1, err := world.NewCharacterRelationship(testTenant.ID, char1.ID, char2.ID, "ally")
		if err != nil {
			t.Fatalf("failed to create relationship 1: %v", err)
		}
		err = characterRelationshipRepo.Create(ctx, cr1)
		if err != nil {
			t.Fatalf("failed to create relationship 1: %v", err)
		}

		cr2, err := world.NewCharacterRelationship(testTenant.ID, char1.ID, char3.ID, "enemy")
		if err != nil {
			t.Fatalf("failed to create relationship 2: %v", err)
		}
		err = characterRelationshipRepo.Create(ctx, cr2)
		if err != nil {
			t.Fatalf("failed to create relationship 2: %v", err)
		}

		// Delete first relationship
		err = characterRelationshipRepo.Delete(ctx, testTenant.ID, cr1.ID)
		if err != nil {
			t.Fatalf("failed to delete relationship 1: %v", err)
		}

		// Second relationship should still exist
		retrieved, err := characterRelationshipRepo.GetByID(ctx, testTenant.ID, cr2.ID)
		if err != nil {
			t.Fatalf("relationship 2 should still exist: %v", err)
		}

		if retrieved.ID != cr2.ID {
			t.Errorf("expected ID to be %s, got %s", cr2.ID, retrieved.ID)
		}
	})
}

