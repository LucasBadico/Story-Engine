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

func TestFactionReferenceRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	factionRefRepo := NewFactionReferenceRepository(db)

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

	// Create faction
	testFaction, err := world.NewFaction(testTenant.ID, testWorld.ID, "Test Faction", nil)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}
	err = factionRepo.Create(ctx, testFaction)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
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
		fr := world.NewFactionReference(testFaction.ID, "character", testCharacter.ID, nil)

		err = factionRefRepo.Create(ctx, fr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify faction-reference can be retrieved
		retrieved, err := factionRefRepo.GetByID(ctx, testTenant.ID, fr.ID)
		if err != nil {
			t.Fatalf("failed to retrieve faction-reference: %v", err)
		}

		if retrieved.FactionID != testFaction.ID {
			t.Errorf("expected faction_id to be %s, got %s", testFaction.ID, retrieved.FactionID)
		}

		if retrieved.EntityType != "character" {
			t.Errorf("expected entity_type to be 'character', got '%s'", retrieved.EntityType)
		}

		if retrieved.EntityID != testCharacter.ID {
			t.Errorf("expected entity_id to be %s, got %s", testCharacter.ID, retrieved.EntityID)
		}
	})

	t.Run("successful creation with location", func(t *testing.T) {
		fr := world.NewFactionReference(testFaction.ID, "location", testLocation.ID, nil)

		err = factionRefRepo.Create(ctx, fr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := factionRefRepo.GetByID(ctx, testTenant.ID, fr.ID)
		if err != nil {
			t.Fatalf("failed to retrieve faction-reference: %v", err)
		}

		if retrieved.EntityType != "location" {
			t.Errorf("expected entity_type to be 'location', got '%s'", retrieved.EntityType)
		}

		if retrieved.EntityID != testLocation.ID {
			t.Errorf("expected entity_id to be %s, got %s", testLocation.ID, retrieved.EntityID)
		}
	})

	t.Run("successful creation with role", func(t *testing.T) {
		// Create a new character to avoid UNIQUE constraint violation
		character2, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character With Role")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		err = characterRepo.Create(ctx, character2)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		role := "Leader"
		fr := world.NewFactionReference(testFaction.ID, "character", character2.ID, &role)

		err = factionRefRepo.Create(ctx, fr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := factionRefRepo.GetByID(ctx, testTenant.ID, fr.ID)
		if err != nil {
			t.Fatalf("failed to retrieve faction-reference: %v", err)
		}

		if retrieved.Role == nil {
			t.Fatal("expected role to be set, got nil")
		}

		if *retrieved.Role != role {
			t.Errorf("expected role to be '%s', got '%s'", role, *retrieved.Role)
		}
	})

	t.Run("successful creation with notes", func(t *testing.T) {
		// Create a new character to avoid UNIQUE constraint violation
		character3, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character With Notes")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		err = characterRepo.Create(ctx, character3)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		fr := world.NewFactionReference(testFaction.ID, "character", character3.ID, nil)
		fr.UpdateNotes("Some notes")

		err = factionRefRepo.Create(ctx, fr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := factionRefRepo.GetByID(ctx, testTenant.ID, fr.ID)
		if err != nil {
			t.Fatalf("failed to retrieve faction-reference: %v", err)
		}

		if retrieved.Notes != "Some notes" {
			t.Errorf("expected notes to be 'Some notes', got '%s'", retrieved.Notes)
		}
	})
}

func TestFactionReferenceRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)
	characterRepo := NewCharacterRepository(db)
	factionRefRepo := NewFactionReferenceRepository(db)

	// Create tenant, world, faction, and character
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

	testFaction, err := world.NewFaction(testTenant.ID, testWorld.ID, "Test Faction", nil)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}
	err = factionRepo.Create(ctx, testFaction)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("existing faction-reference", func(t *testing.T) {
		fr := world.NewFactionReference(testFaction.ID, "character", testCharacter.ID, nil)
		err = factionRefRepo.Create(ctx, fr)
		if err != nil {
			t.Fatalf("failed to create faction-reference: %v", err)
		}

		retrieved, err := factionRefRepo.GetByID(ctx, testTenant.ID, fr.ID)
		if err != nil {
			t.Fatalf("failed to get faction-reference: %v", err)
		}

		if retrieved.ID != fr.ID {
			t.Errorf("expected ID to be %s, got %s", fr.ID, retrieved.ID)
		}

		if retrieved.FactionID != testFaction.ID {
			t.Errorf("expected faction_id to be %s, got %s", testFaction.ID, retrieved.FactionID)
		}
	})

	t.Run("non-existent faction-reference", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := factionRefRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent faction-reference")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "faction_reference" {
			t.Errorf("expected resource to be 'faction_reference', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestFactionReferenceRepository_ListByFaction(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	factionRefRepo := NewFactionReferenceRepository(db)

	// Create tenant, world, and faction
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

	testFaction, err := world.NewFaction(testTenant.ID, testWorld.ID, "Test Faction", nil)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}
	err = factionRepo.Create(ctx, testFaction)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}

	t.Run("empty list", func(t *testing.T) {
		refs, err := factionRefRepo.ListByFaction(ctx, testTenant.ID, testFaction.ID)
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
		fr1 := world.NewFactionReference(testFaction.ID, "character", char1.ID, nil)
		fr2 := world.NewFactionReference(testFaction.ID, "character", char2.ID, nil)
		fr3 := world.NewFactionReference(testFaction.ID, "location", loc1.ID, nil)
		factionRefRepo.Create(ctx, fr1)
		time.Sleep(10 * time.Millisecond)
		factionRefRepo.Create(ctx, fr2)
		time.Sleep(10 * time.Millisecond)
		factionRefRepo.Create(ctx, fr3)

		refs, err := factionRefRepo.ListByFaction(ctx, testTenant.ID, testFaction.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 3 {
			t.Errorf("expected 3 references, got %d", len(refs))
		}

		// Verify all belong to the faction
		for _, ref := range refs {
			if ref.FactionID != testFaction.ID {
				t.Errorf("expected faction_id to be %s, got %s", testFaction.ID, ref.FactionID)
			}
		}
	})
}

func TestFactionReferenceRepository_ListByEntity(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	factionRefRepo := NewFactionReferenceRepository(db)

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
		// Create factions
		faction1, _ := world.NewFaction(testTenant.ID, testWorld.ID, "Faction 1", nil)
		faction2, _ := world.NewFaction(testTenant.ID, testWorld.ID, "Faction 2", nil)
		factionRepo.Create(ctx, faction1)
		factionRepo.Create(ctx, faction2)

		// Create references
		fr1 := world.NewFactionReference(faction1.ID, "character", testCharacter.ID, nil)
		fr2 := world.NewFactionReference(faction2.ID, "character", testCharacter.ID, nil)
		factionRefRepo.Create(ctx, fr1)
		factionRefRepo.Create(ctx, fr2)

		refs, err := factionRefRepo.ListByEntity(ctx, testTenant.ID, "character", testCharacter.ID)
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

	t.Run("list by location", func(t *testing.T) {
		// Create faction
		faction3, _ := world.NewFaction(testTenant.ID, testWorld.ID, "Faction 3", nil)
		factionRepo.Create(ctx, faction3)

		// Create reference
		fr3 := world.NewFactionReference(faction3.ID, "location", testLocation.ID, nil)
		factionRefRepo.Create(ctx, fr3)

		refs, err := factionRefRepo.ListByEntity(ctx, testTenant.ID, "location", testLocation.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 1 {
			t.Errorf("expected 1 reference, got %d", len(refs))
		}

		for _, ref := range refs {
			if ref.EntityType != "location" {
				t.Errorf("expected entity_type to be 'location', got '%s'", ref.EntityType)
			}
			if ref.EntityID != testLocation.ID {
				t.Errorf("expected entity_id to be %s, got %s", testLocation.ID, ref.EntityID)
			}
		}
	})

	t.Run("empty list", func(t *testing.T) {
		nonExistentID := uuid.New()
		refs, err := factionRefRepo.ListByEntity(ctx, testTenant.ID, "character", nonExistentID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 0 {
			t.Errorf("expected empty list, got %d references", len(refs))
		}
	})
}

func TestFactionReferenceRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)
	characterRepo := NewCharacterRepository(db)
	factionRefRepo := NewFactionReferenceRepository(db)

	// Create tenant, world, faction, and character
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

	testFaction, err := world.NewFaction(testTenant.ID, testWorld.ID, "Test Faction", nil)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}
	err = factionRepo.Create(ctx, testFaction)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		fr := world.NewFactionReference(testFaction.ID, "character", testCharacter.ID, nil)
		err = factionRefRepo.Create(ctx, fr)
		if err != nil {
			t.Fatalf("failed to create faction-reference: %v", err)
		}

		// Update role
		role := "Updated Role"
		fr.UpdateRole(&role)

		// Update notes
		fr.UpdateNotes("Updated Notes")

		err = factionRefRepo.Update(ctx, fr)
		if err != nil {
			t.Fatalf("failed to update faction-reference: %v", err)
		}

		// Verify update
		retrieved, err := factionRefRepo.GetByID(ctx, testTenant.ID, fr.ID)
		if err != nil {
			t.Fatalf("failed to get faction-reference: %v", err)
		}

		if retrieved.Role == nil || *retrieved.Role != role {
			t.Errorf("expected role to be '%s', got %v", role, retrieved.Role)
		}

		if retrieved.Notes != "Updated Notes" {
			t.Errorf("expected notes to be 'Updated Notes', got '%s'", retrieved.Notes)
		}
	})

	t.Run("update role to nil", func(t *testing.T) {
		// Create a new character to avoid UNIQUE constraint violation
		character2, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character for nil update")
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}
		err = characterRepo.Create(ctx, character2)
		if err != nil {
			t.Fatalf("failed to create character: %v", err)
		}

		role := "Initial Role"
		fr := world.NewFactionReference(testFaction.ID, "character", character2.ID, &role)
		err = factionRefRepo.Create(ctx, fr)
		if err != nil {
			t.Fatalf("failed to create faction-reference: %v", err)
		}

		// Clear role
		fr.UpdateRole(nil)

		err = factionRefRepo.Update(ctx, fr)
		if err != nil {
			t.Fatalf("failed to update faction-reference: %v", err)
		}

		retrieved, err := factionRefRepo.GetByID(ctx, testTenant.ID, fr.ID)
		if err != nil {
			t.Fatalf("failed to get faction-reference: %v", err)
		}

		if retrieved.Role != nil {
			t.Error("expected role to be nil, got non-nil")
		}
	})
}

func TestFactionReferenceRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)
	characterRepo := NewCharacterRepository(db)
	factionRefRepo := NewFactionReferenceRepository(db)

	// Create tenant, world, faction, and character
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

	testFaction, err := world.NewFaction(testTenant.ID, testWorld.ID, "Test Faction", nil)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}
	err = factionRepo.Create(ctx, testFaction)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
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
		fr := world.NewFactionReference(testFaction.ID, "character", testCharacter.ID, nil)
		err = factionRefRepo.Create(ctx, fr)
		if err != nil {
			t.Fatalf("failed to create faction-reference: %v", err)
		}

		err = factionRefRepo.Delete(ctx, testTenant.ID, fr.ID)
		if err != nil {
			t.Fatalf("failed to delete faction-reference: %v", err)
		}

		// Verify faction-reference is deleted
		_, err = factionRefRepo.GetByID(ctx, testTenant.ID, fr.ID)
		if err == nil {
			t.Fatal("expected error for deleted faction-reference")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "faction_reference" {
			t.Errorf("expected resource to be 'faction_reference', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestFactionReferenceRepository_DeleteByFaction(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)
	characterRepo := NewCharacterRepository(db)
	locationRepo := NewLocationRepository(db)
	factionRefRepo := NewFactionReferenceRepository(db)

	// Create tenant, world, and faction
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

	testFaction, err := world.NewFaction(testTenant.ID, testWorld.ID, "Test Faction", nil)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}
	err = factionRepo.Create(ctx, testFaction)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}

	t.Run("delete all references for faction", func(t *testing.T) {
		// Create characters and locations
		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		loc1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 1", nil)
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)
		locationRepo.Create(ctx, loc1)

		// Create references
		fr1 := world.NewFactionReference(testFaction.ID, "character", char1.ID, nil)
		fr2 := world.NewFactionReference(testFaction.ID, "character", char2.ID, nil)
		fr3 := world.NewFactionReference(testFaction.ID, "location", loc1.ID, nil)
		factionRefRepo.Create(ctx, fr1)
		factionRefRepo.Create(ctx, fr2)
		factionRefRepo.Create(ctx, fr3)

		// Verify they exist
		refs, err := factionRefRepo.ListByFaction(ctx, testTenant.ID, testFaction.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(refs) != 3 {
			t.Errorf("expected 3 references, got %d", len(refs))
		}

		// Delete all by faction
		err = factionRefRepo.DeleteByFaction(ctx, testTenant.ID, testFaction.ID)
		if err != nil {
			t.Fatalf("failed to delete faction-references: %v", err)
		}

		// Verify all are deleted
		refs, err = factionRefRepo.ListByFaction(ctx, testTenant.ID, testFaction.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refs) != 0 {
			t.Errorf("expected no references, got %d", len(refs))
		}
	})
}

func TestFactionReferenceRepository_DeleteByFactionAndEntity(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)
	characterRepo := NewCharacterRepository(db)
	factionRefRepo := NewFactionReferenceRepository(db)

	// Create tenant, world, faction, and character
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

	testFaction, err := world.NewFaction(testTenant.ID, testWorld.ID, "Test Faction", nil)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}
	err = factionRepo.Create(ctx, testFaction)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("delete specific faction-entity reference", func(t *testing.T) {
		// Create multiple characters
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		characterRepo.Create(ctx, char2)

		// Create references
		fr1 := world.NewFactionReference(testFaction.ID, "character", testCharacter.ID, nil)
		fr2 := world.NewFactionReference(testFaction.ID, "character", char2.ID, nil)
		factionRefRepo.Create(ctx, fr1)
		factionRefRepo.Create(ctx, fr2)

		// Verify they exist
		refs, err := factionRefRepo.ListByFaction(ctx, testTenant.ID, testFaction.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(refs) != 2 {
			t.Errorf("expected 2 references, got %d", len(refs))
		}

		// Delete specific reference
		err = factionRefRepo.DeleteByFactionAndEntity(ctx, testTenant.ID, testFaction.ID, "character", testCharacter.ID)
		if err != nil {
			t.Fatalf("failed to delete faction-reference: %v", err)
		}

		// Verify only one is deleted
		refs, err = factionRefRepo.ListByFaction(ctx, testTenant.ID, testFaction.ID)
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

