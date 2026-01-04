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

func TestLoreRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)

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

	t.Run("successful creation root lore", func(t *testing.T) {
		lore, err := world.NewLore(testTenant.ID, testWorld.ID, "Root Lore", nil)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}

		err = loreRepo.Create(ctx, lore)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify lore can be retrieved
		retrieved, err := loreRepo.GetByID(ctx, testTenant.ID, lore.ID)
		if err != nil {
			t.Fatalf("failed to retrieve lore: %v", err)
		}

		if retrieved.Name != "Root Lore" {
			t.Errorf("expected name to be 'Root Lore', got '%s'", retrieved.Name)
		}

		if retrieved.HierarchyLevel != 0 {
			t.Errorf("expected hierarchy level to be 0, got %d", retrieved.HierarchyLevel)
		}

		if retrieved.ParentID != nil {
			t.Error("expected parent_id to be nil, got non-nil")
		}
	})

	t.Run("successful creation with parent", func(t *testing.T) {
		// Create parent lore first
		parent, err := world.NewLore(testTenant.ID, testWorld.ID, "Parent Lore", nil)
		if err != nil {
			t.Fatalf("failed to create parent lore: %v", err)
		}
		err = loreRepo.Create(ctx, parent)
		if err != nil {
			t.Fatalf("failed to create parent lore: %v", err)
		}

		// Create child lore
		child, err := world.NewLore(testTenant.ID, testWorld.ID, "Child Lore", &parent.ID)
		if err != nil {
			t.Fatalf("failed to create child lore: %v", err)
		}

		err = loreRepo.Create(ctx, child)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify child lore
		retrieved, err := loreRepo.GetByID(ctx, testTenant.ID, child.ID)
		if err != nil {
			t.Fatalf("failed to retrieve lore: %v", err)
		}

		if retrieved.Name != "Child Lore" {
			t.Errorf("expected name to be 'Child Lore', got '%s'", retrieved.Name)
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

	t.Run("creation with category and fields", func(t *testing.T) {
		category := "Magic"
		lore, err := world.NewLore(testTenant.ID, testWorld.ID, "Typed Lore", nil)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}
		lore.UpdateCategory(&category)
		lore.UpdateDescription("A test lore description")
		lore.UpdateRules("Test rules")
		lore.UpdateLimitations("Test limitations")
		lore.UpdateRequirements("Test requirements")

		err = loreRepo.Create(ctx, lore)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := loreRepo.GetByID(ctx, testTenant.ID, lore.ID)
		if err != nil {
			t.Fatalf("failed to retrieve lore: %v", err)
		}

		if retrieved.Category == nil || *retrieved.Category != category {
			t.Errorf("expected category to be '%s', got %v", category, retrieved.Category)
		}

		if retrieved.Description != "A test lore description" {
			t.Errorf("expected description to be 'A test lore description', got '%s'", retrieved.Description)
		}

		if retrieved.Rules != "Test rules" {
			t.Errorf("expected rules to be 'Test rules', got '%s'", retrieved.Rules)
		}

		if retrieved.Limitations != "Test limitations" {
			t.Errorf("expected limitations to be 'Test limitations', got '%s'", retrieved.Limitations)
		}

		if retrieved.Requirements != "Test requirements" {
			t.Errorf("expected requirements to be 'Test requirements', got '%s'", retrieved.Requirements)
		}
	})
}

func TestLoreRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)

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

	t.Run("existing lore", func(t *testing.T) {
		lore, err := world.NewLore(testTenant.ID, testWorld.ID, "GetByID Lore", nil)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}

		err = loreRepo.Create(ctx, lore)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}

		retrieved, err := loreRepo.GetByID(ctx, testTenant.ID, lore.ID)
		if err != nil {
			t.Fatalf("failed to get lore: %v", err)
		}

		if retrieved.ID != lore.ID {
			t.Errorf("expected ID to be %s, got %s", lore.ID, retrieved.ID)
		}

		if retrieved.Name != "GetByID Lore" {
			t.Errorf("expected name to be 'GetByID Lore', got '%s'", retrieved.Name)
		}
	})

	t.Run("non-existent lore", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := loreRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent lore")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "lore" {
			t.Errorf("expected resource to be 'lore', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestLoreRepository_ListByWorld(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)

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
		lores, err := loreRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(lores) != 0 {
			t.Errorf("expected empty list, got %d lores", len(lores))
		}
	})

	t.Run("list with lores", func(t *testing.T) {
		// Create multiple lores
		loreNames := []string{"Lore A", "Lore B", "Lore C"}
		createdLores := make([]*world.Lore, 0, len(loreNames))

		for _, name := range loreNames {
			lore, err := world.NewLore(testTenant.ID, testWorld.ID, name, nil)
			if err != nil {
				t.Fatalf("failed to create lore: %v", err)
			}
			err = loreRepo.Create(ctx, lore)
			if err != nil {
				t.Fatalf("failed to create lore: %v", err)
			}
			createdLores = append(createdLores, lore)
			time.Sleep(10 * time.Millisecond)
		}

		lores, err := loreRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(lores) != len(loreNames) {
			t.Errorf("expected %d lores, got %d", len(loreNames), len(lores))
		}

		// Verify ordering (by hierarchy level, then name)
		for i := 0; i < len(lores)-1; i++ {
			if lores[i].HierarchyLevel > lores[i+1].HierarchyLevel {
				t.Error("expected lores to be ordered by hierarchy level ascending")
			}
		}
	})
}

func TestLoreRepository_GetChildren(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)

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
		parent, err := world.NewLore(testTenant.ID, testWorld.ID, "Parent", nil)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}
		err = loreRepo.Create(ctx, parent)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}

		// Create children
		child1, _ := world.NewLore(testTenant.ID, testWorld.ID, "Child 1", &parent.ID)
		child2, _ := world.NewLore(testTenant.ID, testWorld.ID, "Child 2", &parent.ID)
		loreRepo.Create(ctx, child1)
		loreRepo.Create(ctx, child2)

		// Get children
		children, err := loreRepo.GetChildren(ctx, testTenant.ID, parent.ID)
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
		parent, err := world.NewLore(testTenant.ID, testWorld.ID, "Parent No Children", nil)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}
		err = loreRepo.Create(ctx, parent)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}

		children, err := loreRepo.GetChildren(ctx, testTenant.ID, parent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(children) != 0 {
			t.Errorf("expected no children, got %d", len(children))
		}
	})
}

func TestLoreRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)

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
		lore, err := world.NewLore(testTenant.ID, testWorld.ID, "Update Lore", nil)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}

		err = loreRepo.Create(ctx, lore)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}

		// Update name
		err = lore.UpdateName("Updated Name")
		if err != nil {
			t.Fatalf("failed to update name: %v", err)
		}

		// Update category
		category := "Updated Category"
		lore.UpdateCategory(&category)

		// Update description
		lore.UpdateDescription("Updated Description")

		// Update rules
		lore.UpdateRules("Updated Rules")

		// Update limitations
		lore.UpdateLimitations("Updated Limitations")

		// Update requirements
		lore.UpdateRequirements("Updated Requirements")

		err = loreRepo.Update(ctx, lore)
		if err != nil {
			t.Fatalf("failed to update lore: %v", err)
		}

		// Verify update
		retrieved, err := loreRepo.GetByID(ctx, testTenant.ID, lore.ID)
		if err != nil {
			t.Fatalf("failed to get lore: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("expected name to be 'Updated Name', got '%s'", retrieved.Name)
		}

		if retrieved.Category == nil || *retrieved.Category != category {
			t.Errorf("expected category to be '%s', got %v", category, retrieved.Category)
		}

		if retrieved.Description != "Updated Description" {
			t.Errorf("expected description to be 'Updated Description', got '%s'", retrieved.Description)
		}

		if retrieved.Rules != "Updated Rules" {
			t.Errorf("expected rules to be 'Updated Rules', got '%s'", retrieved.Rules)
		}

		if retrieved.Limitations != "Updated Limitations" {
			t.Errorf("expected limitations to be 'Updated Limitations', got '%s'", retrieved.Limitations)
		}

		if retrieved.Requirements != "Updated Requirements" {
			t.Errorf("expected requirements to be 'Updated Requirements', got '%s'", retrieved.Requirements)
		}
	})

	t.Run("update parent", func(t *testing.T) {
		// Create two root lores
		root1, _ := world.NewLore(testTenant.ID, testWorld.ID, "Root 1", nil)
		root2, _ := world.NewLore(testTenant.ID, testWorld.ID, "Root 2", nil)
		loreRepo.Create(ctx, root1)
		loreRepo.Create(ctx, root2)

		// Create child of root1
		child, _ := world.NewLore(testTenant.ID, testWorld.ID, "Child", &root1.ID)
		loreRepo.Create(ctx, child)

		// Move child to root2
		root2Retrieved, _ := loreRepo.GetByID(ctx, testTenant.ID, root2.ID)
		child.SetParent(&root2.ID, root2Retrieved.HierarchyLevel)
		err = loreRepo.Update(ctx, child)
		if err != nil {
			t.Fatalf("failed to update lore: %v", err)
		}

		// Verify update
		retrieved, err := loreRepo.GetByID(ctx, testTenant.ID, child.ID)
		if err != nil {
			t.Fatalf("failed to get lore: %v", err)
		}

		if retrieved.ParentID == nil || *retrieved.ParentID != root2.ID {
			t.Errorf("expected parent_id to be %s, got %v", root2.ID, retrieved.ParentID)
		}

		if retrieved.HierarchyLevel != root2Retrieved.HierarchyLevel+1 {
			t.Errorf("expected hierarchy level to be %d, got %d", root2Retrieved.HierarchyLevel+1, retrieved.HierarchyLevel)
		}
	})
}

func TestLoreRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	loreRepo := NewLoreRepository(db)

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
		lore, err := world.NewLore(testTenant.ID, testWorld.ID, "Delete Lore", nil)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}

		err = loreRepo.Create(ctx, lore)
		if err != nil {
			t.Fatalf("failed to create lore: %v", err)
		}

		err = loreRepo.Delete(ctx, testTenant.ID, lore.ID)
		if err != nil {
			t.Fatalf("failed to delete lore: %v", err)
		}

		// Verify lore is deleted
		_, err = loreRepo.GetByID(ctx, testTenant.ID, lore.ID)
		if err == nil {
			t.Fatal("expected error for deleted lore")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "lore" {
			t.Errorf("expected resource to be 'lore', got '%s'", notFoundErr.Resource)
		}
	})
}

