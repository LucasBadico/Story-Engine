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

func TestFactionRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)

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

	t.Run("successful creation root faction", func(t *testing.T) {
		faction, err := world.NewFaction(testTenant.ID, testWorld.ID, "Root Faction", nil)
		if err != nil {
			t.Fatalf("failed to create faction: %v", err)
		}

		err = factionRepo.Create(ctx, faction)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify faction can be retrieved
		retrieved, err := factionRepo.GetByID(ctx, testTenant.ID, faction.ID)
		if err != nil {
			t.Fatalf("failed to retrieve faction: %v", err)
		}

		if retrieved.Name != "Root Faction" {
			t.Errorf("expected name to be 'Root Faction', got '%s'", retrieved.Name)
		}

		if retrieved.HierarchyLevel != 0 {
			t.Errorf("expected hierarchy level to be 0, got %d", retrieved.HierarchyLevel)
		}

		if retrieved.ParentID != nil {
			t.Error("expected parent_id to be nil, got non-nil")
		}
	})

	t.Run("successful creation with parent", func(t *testing.T) {
		// Create parent faction first
		parent, err := world.NewFaction(testTenant.ID, testWorld.ID, "Parent Faction", nil)
		if err != nil {
			t.Fatalf("failed to create parent faction: %v", err)
		}
		err = factionRepo.Create(ctx, parent)
		if err != nil {
			t.Fatalf("failed to create parent faction: %v", err)
		}

		// Create child faction
		child, err := world.NewFaction(testTenant.ID, testWorld.ID, "Child Faction", &parent.ID)
		if err != nil {
			t.Fatalf("failed to create child faction: %v", err)
		}

		err = factionRepo.Create(ctx, child)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify child faction
		retrieved, err := factionRepo.GetByID(ctx, testTenant.ID, child.ID)
		if err != nil {
			t.Fatalf("failed to retrieve faction: %v", err)
		}

		if retrieved.Name != "Child Faction" {
			t.Errorf("expected name to be 'Child Faction', got '%s'", retrieved.Name)
		}

		if retrieved.HierarchyLevel != 1 {
			t.Errorf("expected hierarchy level to be 1, got %d", retrieved.HierarchyLevel)
		}

		if retrieved.ParentID == nil {
			t.Fatal("expected parent_id to be set, got nil")
		}

		if *retrieved.ParentID != parent.ID {
			t.Errorf("expected parent_id to be %s, got %s", parent.ID, *retrieved.ParentID)
		}
	})

	t.Run("creation with type and fields", func(t *testing.T) {
		factionType := "Guild"
		faction, err := world.NewFaction(testTenant.ID, testWorld.ID, "Typed Faction", nil)
		if err != nil {
			t.Fatalf("failed to create faction: %v", err)
		}
		faction.UpdateType(&factionType)
		faction.UpdateDescription("A test faction description")
		faction.UpdateBeliefs("Test beliefs")
		faction.UpdateStructure("Test structure")
		faction.UpdateSymbols("Test symbols")

		err = factionRepo.Create(ctx, faction)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := factionRepo.GetByID(ctx, testTenant.ID, faction.ID)
		if err != nil {
			t.Fatalf("failed to retrieve faction: %v", err)
		}

		if retrieved.Type == nil || *retrieved.Type != factionType {
			t.Errorf("expected type to be '%s', got %v", factionType, retrieved.Type)
		}

		if retrieved.Description != "A test faction description" {
			t.Errorf("expected description to be 'A test faction description', got '%s'", retrieved.Description)
		}

		if retrieved.Beliefs != "Test beliefs" {
			t.Errorf("expected beliefs to be 'Test beliefs', got '%s'", retrieved.Beliefs)
		}

		if retrieved.Structure != "Test structure" {
			t.Errorf("expected structure to be 'Test structure', got '%s'", retrieved.Structure)
		}

		if retrieved.Symbols != "Test symbols" {
			t.Errorf("expected symbols to be 'Test symbols', got '%s'", retrieved.Symbols)
		}
	})
}

func TestFactionRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)

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

	t.Run("existing faction", func(t *testing.T) {
		faction, err := world.NewFaction(testTenant.ID, testWorld.ID, "GetByID Faction", nil)
		if err != nil {
			t.Fatalf("failed to create faction: %v", err)
		}

		err = factionRepo.Create(ctx, faction)
		if err != nil {
			t.Fatalf("failed to create faction: %v", err)
		}

		retrieved, err := factionRepo.GetByID(ctx, testTenant.ID, faction.ID)
		if err != nil {
			t.Fatalf("failed to get faction: %v", err)
		}

		if retrieved.ID != faction.ID {
			t.Errorf("expected ID to be %s, got %s", faction.ID, retrieved.ID)
		}

		if retrieved.Name != "GetByID Faction" {
			t.Errorf("expected name to be 'GetByID Faction', got '%s'", retrieved.Name)
		}
	})

	t.Run("non-existent faction", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := factionRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent faction")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "faction" {
			t.Errorf("expected resource to be 'faction', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestFactionRepository_ListByWorld(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)

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
		factions, err := factionRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(factions) != 0 {
			t.Errorf("expected empty list, got %d factions", len(factions))
		}
	})

	t.Run("list with factions", func(t *testing.T) {
		// Create multiple factions
		factionNames := []string{"Faction A", "Faction B", "Faction C"}
		createdFactions := make([]*world.Faction, 0, len(factionNames))

		for _, name := range factionNames {
			faction, err := world.NewFaction(testTenant.ID, testWorld.ID, name, nil)
			if err != nil {
				t.Fatalf("failed to create faction: %v", err)
			}
			err = factionRepo.Create(ctx, faction)
			if err != nil {
				t.Fatalf("failed to create faction: %v", err)
			}
			createdFactions = append(createdFactions, faction)
			time.Sleep(10 * time.Millisecond)
		}

		factions, err := factionRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(factions) != len(factionNames) {
			t.Errorf("expected %d factions, got %d", len(factionNames), len(factions))
		}

		// Verify ordering (by hierarchy level, then name)
		for i := 0; i < len(factions)-1; i++ {
			if factions[i].HierarchyLevel > factions[i+1].HierarchyLevel {
				t.Error("expected factions to be ordered by hierarchy level ascending")
			}
		}
	})
}

func TestFactionRepository_GetChildren(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)

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

	t.Run("get direct children", func(t *testing.T) {
		// Create parent
		parent, err := world.NewFaction(testTenant.ID, testWorld.ID, "Parent", nil)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}
		err = factionRepo.Create(ctx, parent)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}

		// Create children
		child1, _ := world.NewFaction(testTenant.ID, testWorld.ID, "Child 1", &parent.ID)
		child2, _ := world.NewFaction(testTenant.ID, testWorld.ID, "Child 2", &parent.ID)
		factionRepo.Create(ctx, child1)
		factionRepo.Create(ctx, child2)

		// Get children
		children, err := factionRepo.GetChildren(ctx, testTenant.ID, parent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(children) != 2 {
			t.Errorf("expected 2 children, got %d", len(children))
		}

		for _, child := range children {
			if child.ParentID == nil || *child.ParentID != parent.ID {
				t.Errorf("expected parent_id to be %s, got %v", parent.ID, child.ParentID)
			}
		}
	})

	t.Run("no children", func(t *testing.T) {
		parent, err := world.NewFaction(testTenant.ID, testWorld.ID, "Parent No Children", nil)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}
		err = factionRepo.Create(ctx, parent)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}

		children, err := factionRepo.GetChildren(ctx, testTenant.ID, parent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(children) != 0 {
			t.Errorf("expected no children, got %d", len(children))
		}
	})
}

func TestFactionRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)

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
		faction, err := world.NewFaction(testTenant.ID, testWorld.ID, "Update Faction", nil)
		if err != nil {
			t.Fatalf("failed to create faction: %v", err)
		}

		err = factionRepo.Create(ctx, faction)
		if err != nil {
			t.Fatalf("failed to create faction: %v", err)
		}

		// Update name
		err = faction.UpdateName("Updated Name")
		if err != nil {
			t.Fatalf("failed to update name: %v", err)
		}

		// Update type
		factionType := "Updated Type"
		faction.UpdateType(&factionType)

		// Update description
		faction.UpdateDescription("Updated Description")

		// Update beliefs
		faction.UpdateBeliefs("Updated Beliefs")

		// Update structure
		faction.UpdateStructure("Updated Structure")

		// Update symbols
		faction.UpdateSymbols("Updated Symbols")

		err = factionRepo.Update(ctx, faction)
		if err != nil {
			t.Fatalf("failed to update faction: %v", err)
		}

		// Verify update
		retrieved, err := factionRepo.GetByID(ctx, testTenant.ID, faction.ID)
		if err != nil {
			t.Fatalf("failed to get faction: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("expected name to be 'Updated Name', got '%s'", retrieved.Name)
		}

		if retrieved.Type == nil || *retrieved.Type != factionType {
			t.Errorf("expected type to be '%s', got %v", factionType, retrieved.Type)
		}

		if retrieved.Description != "Updated Description" {
			t.Errorf("expected description to be 'Updated Description', got '%s'", retrieved.Description)
		}

		if retrieved.Beliefs != "Updated Beliefs" {
			t.Errorf("expected beliefs to be 'Updated Beliefs', got '%s'", retrieved.Beliefs)
		}

		if retrieved.Structure != "Updated Structure" {
			t.Errorf("expected structure to be 'Updated Structure', got '%s'", retrieved.Structure)
		}

		if retrieved.Symbols != "Updated Symbols" {
			t.Errorf("expected symbols to be 'Updated Symbols', got '%s'", retrieved.Symbols)
		}
	})

	t.Run("update parent", func(t *testing.T) {
		// Create two root factions
		root1, _ := world.NewFaction(testTenant.ID, testWorld.ID, "Root 1", nil)
		root2, _ := world.NewFaction(testTenant.ID, testWorld.ID, "Root 2", nil)
		factionRepo.Create(ctx, root1)
		factionRepo.Create(ctx, root2)

		// Create child of root1
		child, _ := world.NewFaction(testTenant.ID, testWorld.ID, "Child", &root1.ID)
		factionRepo.Create(ctx, child)

		// Move child to root2
		root2Retrieved, _ := factionRepo.GetByID(ctx, testTenant.ID, root2.ID)
		child.SetParent(&root2.ID, root2Retrieved.HierarchyLevel)
		err = factionRepo.Update(ctx, child)
		if err != nil {
			t.Fatalf("failed to update faction: %v", err)
		}

		// Verify update
		retrieved, err := factionRepo.GetByID(ctx, testTenant.ID, child.ID)
		if err != nil {
			t.Fatalf("failed to get faction: %v", err)
		}

		if retrieved.ParentID == nil || *retrieved.ParentID != root2.ID {
			t.Errorf("expected parent_id to be %s, got %v", root2.ID, retrieved.ParentID)
		}

		if retrieved.HierarchyLevel != root2Retrieved.HierarchyLevel+1 {
			t.Errorf("expected hierarchy level to be %d, got %d", root2Retrieved.HierarchyLevel+1, retrieved.HierarchyLevel)
		}
	})
}

func TestFactionRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	factionRepo := NewFactionRepository(db)

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
		faction, err := world.NewFaction(testTenant.ID, testWorld.ID, "Delete Faction", nil)
		if err != nil {
			t.Fatalf("failed to create faction: %v", err)
		}

		err = factionRepo.Create(ctx, faction)
		if err != nil {
			t.Fatalf("failed to create faction: %v", err)
		}

		err = factionRepo.Delete(ctx, testTenant.ID, faction.ID)
		if err != nil {
			t.Fatalf("failed to delete faction: %v", err)
		}

		// Verify faction is deleted
		_, err = factionRepo.GetByID(ctx, testTenant.ID, faction.ID)
		if err == nil {
			t.Fatal("expected error for deleted faction")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "faction" {
			t.Errorf("expected resource to be 'faction', got '%s'", notFoundErr.Resource)
		}
	})
}

