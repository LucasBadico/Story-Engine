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

func TestLocationRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	locationRepo := NewLocationRepository(db)

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

	t.Run("successful creation root location", func(t *testing.T) {
		loc, err := world.NewLocation(testTenant.ID, testWorld.ID, "Root Location", nil)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		err = locationRepo.Create(ctx, loc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify location can be retrieved
		retrieved, err := locationRepo.GetByID(ctx, testTenant.ID, loc.ID)
		if err != nil {
			t.Fatalf("failed to retrieve location: %v", err)
		}

		if retrieved.Name != "Root Location" {
			t.Errorf("expected name to be 'Root Location', got '%s'", retrieved.Name)
		}

		if retrieved.HierarchyLevel != 0 {
			t.Errorf("expected hierarchy level to be 0, got %d", retrieved.HierarchyLevel)
		}

		if retrieved.ParentID != nil {
			t.Error("expected parent_id to be nil, got non-nil")
		}
	})

	t.Run("successful creation with parent", func(t *testing.T) {
		// Create parent location first
		parent, err := world.NewLocation(testTenant.ID, testWorld.ID, "Parent Location", nil)
		if err != nil {
			t.Fatalf("failed to create parent location: %v", err)
		}
		err = locationRepo.Create(ctx, parent)
		if err != nil {
			t.Fatalf("failed to create parent location: %v", err)
		}

		// Create child location
		child, err := world.NewLocation(testTenant.ID, testWorld.ID, "Child Location", &parent.ID)
		if err != nil {
			t.Fatalf("failed to create child location: %v", err)
		}

		err = locationRepo.Create(ctx, child)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify child location
		retrieved, err := locationRepo.GetByID(ctx, testTenant.ID, child.ID)
		if err != nil {
			t.Fatalf("failed to retrieve location: %v", err)
		}

		if retrieved.Name != "Child Location" {
			t.Errorf("expected name to be 'Child Location', got '%s'", retrieved.Name)
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

	t.Run("creation with type and description", func(t *testing.T) {
		loc, err := world.NewLocation(testTenant.ID, testWorld.ID, "Typed Location", nil)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}
		loc.UpdateType("City")
		loc.UpdateDescription("A test city")

		err = locationRepo.Create(ctx, loc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := locationRepo.GetByID(ctx, testTenant.ID, loc.ID)
		if err != nil {
			t.Fatalf("failed to retrieve location: %v", err)
		}

		if retrieved.Type != "City" {
			t.Errorf("expected type to be 'City', got '%s'", retrieved.Type)
		}

		if retrieved.Description != "A test city" {
			t.Errorf("expected description to be 'A test city', got '%s'", retrieved.Description)
		}
	})
}

func TestLocationRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	locationRepo := NewLocationRepository(db)

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

	t.Run("existing location", func(t *testing.T) {
		loc, err := world.NewLocation(testTenant.ID, testWorld.ID, "GetByID Location", nil)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		err = locationRepo.Create(ctx, loc)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		retrieved, err := locationRepo.GetByID(ctx, testTenant.ID, loc.ID)
		if err != nil {
			t.Fatalf("failed to get location: %v", err)
		}

		if retrieved.ID != loc.ID {
			t.Errorf("expected ID to be %s, got %s", loc.ID, retrieved.ID)
		}

		if retrieved.Name != "GetByID Location" {
			t.Errorf("expected name to be 'GetByID Location', got '%s'", retrieved.Name)
		}
	})

	t.Run("non-existent location", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := locationRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent location")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "location" {
			t.Errorf("expected resource to be 'location', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestLocationRepository_ListByWorld(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	locationRepo := NewLocationRepository(db)

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
		locations, err := locationRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(locations) != 0 {
			t.Errorf("expected empty list, got %d locations", len(locations))
		}
	})

	t.Run("list with locations", func(t *testing.T) {
		// Create multiple locations
		locationNames := []string{"Location A", "Location B", "Location C"}
		createdLocations := make([]*world.Location, 0, len(locationNames))

		for _, name := range locationNames {
			loc, err := world.NewLocation(testTenant.ID, testWorld.ID, name, nil)
			if err != nil {
				t.Fatalf("failed to create location: %v", err)
			}
			err = locationRepo.Create(ctx, loc)
			if err != nil {
				t.Fatalf("failed to create location: %v", err)
			}
			createdLocations = append(createdLocations, loc)
			time.Sleep(10 * time.Millisecond)
		}

		locations, err := locationRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(locations) != len(locationNames) {
			t.Errorf("expected %d locations, got %d", len(locationNames), len(locations))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		// Create 5 locations
		for i := 1; i <= 5; i++ {
			loc, err := world.NewLocation(testTenant.ID, testWorld.ID, "Pagination Location", nil)
			if err != nil {
				t.Fatalf("failed to create location: %v", err)
			}
			err = locationRepo.Create(ctx, loc)
			if err != nil {
				t.Fatalf("failed to create location: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}

		// Get first page
		locations, err := locationRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(locations) != 2 {
			t.Errorf("expected 2 locations, got %d", len(locations))
		}
	})
}

func TestLocationRepository_ListByWorldTree(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	locationRepo := NewLocationRepository(db)

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

	t.Run("tree structure", func(t *testing.T) {
		// Create root locations
		root1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Root 1", nil)
		root2, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Root 2", nil)
		locationRepo.Create(ctx, root1)
		locationRepo.Create(ctx, root2)

		// Create children
		child1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Child 1", &root1.ID)
		child2, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Child 2", &root1.ID)
		locationRepo.Create(ctx, child1)
		locationRepo.Create(ctx, child2)

		locations, err := locationRepo.ListByWorldTree(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(locations) != 4 {
			t.Errorf("expected 4 locations, got %d", len(locations))
		}

		// Verify ordering (by hierarchy level, then name)
		for i := 0; i < len(locations)-1; i++ {
			if locations[i].HierarchyLevel > locations[i+1].HierarchyLevel {
				t.Error("expected locations to be ordered by hierarchy level ascending")
			}
		}
	})
}

func TestLocationRepository_GetChildren(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	locationRepo := NewLocationRepository(db)

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
		parent, err := world.NewLocation(testTenant.ID, testWorld.ID, "Parent", nil)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}
		err = locationRepo.Create(ctx, parent)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}

		// Create children
		child1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Child 1", &parent.ID)
		child2, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Child 2", &parent.ID)
		locationRepo.Create(ctx, child1)
		locationRepo.Create(ctx, child2)

		// Get children
		children, err := locationRepo.GetChildren(ctx, testTenant.ID, parent.ID)
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
		parent, err := world.NewLocation(testTenant.ID, testWorld.ID, "Parent No Children", nil)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}
		err = locationRepo.Create(ctx, parent)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}

		children, err := locationRepo.GetChildren(ctx, testTenant.ID, parent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(children) != 0 {
			t.Errorf("expected no children, got %d", len(children))
		}
	})
}

func TestLocationRepository_GetAncestors(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	locationRepo := NewLocationRepository(db)

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

	t.Run("get ancestors", func(t *testing.T) {
		// Create hierarchy: root -> level1 -> level2 -> level3
		root, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Root", nil)
		locationRepo.Create(ctx, root)

		level1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Level 1", &root.ID)
		locationRepo.Create(ctx, level1)

		level2, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Level 2", &level1.ID)
		locationRepo.Create(ctx, level2)

		level3, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Level 3", &level2.ID)
		locationRepo.Create(ctx, level3)

		// Get ancestors of level3
		ancestors, err := locationRepo.GetAncestors(ctx, testTenant.ID, level3.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(ancestors) != 3 {
			t.Errorf("expected 3 ancestors, got %d", len(ancestors))
		}

		// Verify ancestors are ordered by hierarchy level (ascending)
		for i := 0; i < len(ancestors)-1; i++ {
			if ancestors[i].HierarchyLevel > ancestors[i+1].HierarchyLevel {
				t.Error("expected ancestors to be ordered by hierarchy level ascending")
			}
		}

		// Verify ancestors are correct
		ancestorIDs := make(map[uuid.UUID]bool)
		for _, a := range ancestors {
			ancestorIDs[a.ID] = true
		}

		if !ancestorIDs[root.ID] || !ancestorIDs[level1.ID] || !ancestorIDs[level2.ID] {
			t.Error("expected ancestors to include root, level1, and level2")
		}

		if ancestorIDs[level3.ID] {
			t.Error("expected level3 itself not to be in ancestors")
		}
	})

	t.Run("root has no ancestors", func(t *testing.T) {
		root, err := world.NewLocation(testTenant.ID, testWorld.ID, "Root No Ancestors", nil)
		if err != nil {
			t.Fatalf("failed to create root: %v", err)
		}
		err = locationRepo.Create(ctx, root)
		if err != nil {
			t.Fatalf("failed to create root: %v", err)
		}

		ancestors, err := locationRepo.GetAncestors(ctx, testTenant.ID, root.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(ancestors) != 0 {
			t.Errorf("expected no ancestors, got %d", len(ancestors))
		}
	})
}

func TestLocationRepository_GetDescendants(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	locationRepo := NewLocationRepository(db)

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

	t.Run("get descendants", func(t *testing.T) {
		// Create hierarchy: root -> level1a -> level2a
		//                    root -> level1b
		root, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Root", nil)
		locationRepo.Create(ctx, root)

		level1a, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Level 1A", &root.ID)
		level1b, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Level 1B", &root.ID)
		locationRepo.Create(ctx, level1a)
		locationRepo.Create(ctx, level1b)

		level2a, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Level 2A", &level1a.ID)
		locationRepo.Create(ctx, level2a)

		// Get descendants of root
		descendants, err := locationRepo.GetDescendants(ctx, testTenant.ID, root.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(descendants) != 3 {
			t.Errorf("expected 3 descendants, got %d", len(descendants))
		}

		// Verify descendants are correct
		descendantIDs := make(map[uuid.UUID]bool)
		for _, d := range descendants {
			descendantIDs[d.ID] = true
		}

		if !descendantIDs[level1a.ID] || !descendantIDs[level1b.ID] || !descendantIDs[level2a.ID] {
			t.Error("expected descendants to include level1a, level1b, and level2a")
		}

		if descendantIDs[root.ID] {
			t.Error("expected root itself not to be in descendants")
		}
	})

	t.Run("no descendants", func(t *testing.T) {
		leaf, err := world.NewLocation(testTenant.ID, testWorld.ID, "Leaf", nil)
		if err != nil {
			t.Fatalf("failed to create leaf: %v", err)
		}
		err = locationRepo.Create(ctx, leaf)
		if err != nil {
			t.Fatalf("failed to create leaf: %v", err)
		}

		descendants, err := locationRepo.GetDescendants(ctx, testTenant.ID, leaf.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(descendants) != 0 {
			t.Errorf("expected no descendants, got %d", len(descendants))
		}
	})
}

func TestLocationRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	locationRepo := NewLocationRepository(db)

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
		loc, err := world.NewLocation(testTenant.ID, testWorld.ID, "Update Location", nil)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		err = locationRepo.Create(ctx, loc)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		// Update name
		err = loc.UpdateName("Updated Name")
		if err != nil {
			t.Fatalf("failed to update name: %v", err)
		}

		// Update type
		loc.UpdateType("City")

		// Update description
		loc.UpdateDescription("Updated Description")

		err = locationRepo.Update(ctx, loc)
		if err != nil {
			t.Fatalf("failed to update location: %v", err)
		}

		// Verify update
		retrieved, err := locationRepo.GetByID(ctx, testTenant.ID, loc.ID)
		if err != nil {
			t.Fatalf("failed to get location: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("expected name to be 'Updated Name', got '%s'", retrieved.Name)
		}

		if retrieved.Type != "City" {
			t.Errorf("expected type to be 'City', got '%s'", retrieved.Type)
		}

		if retrieved.Description != "Updated Description" {
			t.Errorf("expected description to be 'Updated Description', got '%s'", retrieved.Description)
		}
	})

	t.Run("update parent", func(t *testing.T) {
		// Create two root locations
		root1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Root 1", nil)
		root2, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Root 2", nil)
		locationRepo.Create(ctx, root1)
		locationRepo.Create(ctx, root2)

		// Create child of root1
		child, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Child", &root1.ID)
		locationRepo.Create(ctx, child)

		// Move child to root2
		root2Retrieved, _ := locationRepo.GetByID(ctx, testTenant.ID, root2.ID)
		child.SetParent(&root2.ID, root2Retrieved.HierarchyLevel)
		err = locationRepo.Update(ctx, child)
		if err != nil {
			t.Fatalf("failed to update location: %v", err)
		}

		// Verify update
		retrieved, err := locationRepo.GetByID(ctx, testTenant.ID, child.ID)
		if err != nil {
			t.Fatalf("failed to get location: %v", err)
		}

		if retrieved.ParentID == nil || *retrieved.ParentID != root2.ID {
			t.Errorf("expected parent_id to be %s, got %v", root2.ID, retrieved.ParentID)
		}

		if retrieved.HierarchyLevel != root2Retrieved.HierarchyLevel+1 {
			t.Errorf("expected hierarchy level to be %d, got %d", root2Retrieved.HierarchyLevel+1, retrieved.HierarchyLevel)
		}
	})
}

func TestLocationRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	locationRepo := NewLocationRepository(db)

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
		loc, err := world.NewLocation(testTenant.ID, testWorld.ID, "Delete Location", nil)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		err = locationRepo.Create(ctx, loc)
		if err != nil {
			t.Fatalf("failed to create location: %v", err)
		}

		err = locationRepo.Delete(ctx, testTenant.ID, loc.ID)
		if err != nil {
			t.Fatalf("failed to delete location: %v", err)
		}

		// Verify location is deleted
		_, err = locationRepo.GetByID(ctx, testTenant.ID, loc.ID)
		if err == nil {
			t.Fatal("expected error for deleted location")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "location" {
			t.Errorf("expected resource to be 'location', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestLocationRepository_CountByWorld(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	locationRepo := NewLocationRepository(db)

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
		count, err := locationRepo.CountByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 0 {
			t.Errorf("expected count to be 0, got %d", count)
		}
	})

	t.Run("count with locations", func(t *testing.T) {
		// Create 3 locations
		for i := 1; i <= 3; i++ {
			loc, err := world.NewLocation(testTenant.ID, testWorld.ID, "Count Location", nil)
			if err != nil {
				t.Fatalf("failed to create location: %v", err)
			}
			err = locationRepo.Create(ctx, loc)
			if err != nil {
				t.Fatalf("failed to create location: %v", err)
			}
		}

		count, err := locationRepo.CountByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 3 {
			t.Errorf("expected count to be 3, got %d", count)
		}
	})
}

