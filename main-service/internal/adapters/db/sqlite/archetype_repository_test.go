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

func TestArchetypeRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	archetypeRepo := NewArchetypeRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		archetype, err := world.NewArchetype(testTenant.ID, "Test Archetype")
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		err = archetypeRepo.Create(ctx, archetype)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify archetype can be retrieved
		retrieved, err := archetypeRepo.GetByID(ctx, testTenant.ID, archetype.ID)
		if err != nil {
			t.Fatalf("failed to retrieve archetype: %v", err)
		}

		if retrieved.Name != "Test Archetype" {
			t.Errorf("expected name to be 'Test Archetype', got '%s'", retrieved.Name)
		}
	})

	t.Run("creation with description", func(t *testing.T) {
		archetype, err := world.NewArchetype(testTenant.ID, "Typed Archetype")
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}
		archetype.UpdateDescription("A test archetype description")

		err = archetypeRepo.Create(ctx, archetype)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := archetypeRepo.GetByID(ctx, testTenant.ID, archetype.ID)
		if err != nil {
			t.Fatalf("failed to retrieve archetype: %v", err)
		}

		if retrieved.Description != "A test archetype description" {
			t.Errorf("expected description to be 'A test archetype description', got '%s'", retrieved.Description)
		}
	})
}

func TestArchetypeRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	archetypeRepo := NewArchetypeRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("existing archetype", func(t *testing.T) {
		archetype, err := world.NewArchetype(testTenant.ID, "GetByID Archetype")
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		err = archetypeRepo.Create(ctx, archetype)
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		retrieved, err := archetypeRepo.GetByID(ctx, testTenant.ID, archetype.ID)
		if err != nil {
			t.Fatalf("failed to get archetype: %v", err)
		}

		if retrieved.ID != archetype.ID {
			t.Errorf("expected ID to be %s, got %s", archetype.ID, retrieved.ID)
		}

		if retrieved.Name != "GetByID Archetype" {
			t.Errorf("expected name to be 'GetByID Archetype', got '%s'", retrieved.Name)
		}
	})

	t.Run("non-existent archetype", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := archetypeRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent archetype")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "archetype" {
			t.Errorf("expected resource to be 'archetype', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestArchetypeRepository_ListByTenant(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	archetypeRepo := NewArchetypeRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("empty list", func(t *testing.T) {
		archetypes, err := archetypeRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(archetypes) != 0 {
			t.Errorf("expected empty list, got %d archetypes", len(archetypes))
		}
	})

	t.Run("list with archetypes", func(t *testing.T) {
		// Create multiple archetypes
		archetypeNames := []string{"Archetype A", "Archetype B", "Archetype C"}
		createdArchetypes := make([]*world.Archetype, 0, len(archetypeNames))

		for _, name := range archetypeNames {
			archetype, err := world.NewArchetype(testTenant.ID, name)
			if err != nil {
				t.Fatalf("failed to create archetype: %v", err)
			}
			err = archetypeRepo.Create(ctx, archetype)
			if err != nil {
				t.Fatalf("failed to create archetype: %v", err)
			}
			createdArchetypes = append(createdArchetypes, archetype)
		}

		archetypes, err := archetypeRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(archetypes) != len(archetypeNames) {
			t.Errorf("expected %d archetypes, got %d", len(archetypeNames), len(archetypes))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		// Create more archetypes
		archetypeNames := []string{"Archetype D", "Archetype E"}
		for _, name := range archetypeNames {
			archetype, _ := world.NewArchetype(testTenant.ID, name)
			archetypeRepo.Create(ctx, archetype)
		}

		// Test limit
		archetypes, err := archetypeRepo.ListByTenant(ctx, testTenant.ID, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(archetypes) > 2 {
			t.Errorf("expected at most 2 archetypes, got %d", len(archetypes))
		}

		// Test offset
		allArchetypes, err := archetypeRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(allArchetypes) < 3 {
			t.Skip("not enough archetypes to test pagination")
		}

		offsetArchetypes, err := archetypeRepo.ListByTenant(ctx, testTenant.ID, 10, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(offsetArchetypes) >= len(allArchetypes) {
			t.Error("expected offset to reduce number of archetypes")
		}
	})

	t.Run("tenant isolation", func(t *testing.T) {
		// Create another tenant
		otherTenant, err := tenant.NewTenant("other-tenant", nil)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}
		err = tenantRepo.Create(ctx, otherTenant)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		// Create archetype for other tenant
		otherArchetype, err := world.NewArchetype(otherTenant.ID, "Other Archetype")
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}
		err = archetypeRepo.Create(ctx, otherArchetype)
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		// List archetypes for test tenant
		archetypes, err := archetypeRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify other tenant's archetype is not included
		for _, archetype := range archetypes {
			if archetype.ID == otherArchetype.ID {
				t.Error("expected other tenant's archetype to be excluded")
			}
		}
	})
}

func TestArchetypeRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	archetypeRepo := NewArchetypeRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		archetype, err := world.NewArchetype(testTenant.ID, "Update Archetype")
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		err = archetypeRepo.Create(ctx, archetype)
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		// Update name
		err = archetype.UpdateName("Updated Name")
		if err != nil {
			t.Fatalf("failed to update name: %v", err)
		}

		// Update description
		archetype.UpdateDescription("Updated Description")

		err = archetypeRepo.Update(ctx, archetype)
		if err != nil {
			t.Fatalf("failed to update archetype: %v", err)
		}

		// Verify update
		retrieved, err := archetypeRepo.GetByID(ctx, testTenant.ID, archetype.ID)
		if err != nil {
			t.Fatalf("failed to get archetype: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("expected name to be 'Updated Name', got '%s'", retrieved.Name)
		}

		if retrieved.Description != "Updated Description" {
			t.Errorf("expected description to be 'Updated Description', got '%s'", retrieved.Description)
		}
	})
}

func TestArchetypeRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	archetypeRepo := NewArchetypeRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		archetype, err := world.NewArchetype(testTenant.ID, "Delete Archetype")
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		err = archetypeRepo.Create(ctx, archetype)
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		err = archetypeRepo.Delete(ctx, testTenant.ID, archetype.ID)
		if err != nil {
			t.Fatalf("failed to delete archetype: %v", err)
		}

		// Verify archetype is deleted
		_, err = archetypeRepo.GetByID(ctx, testTenant.ID, archetype.ID)
		if err == nil {
			t.Fatal("expected error for deleted archetype")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "archetype" {
			t.Errorf("expected resource to be 'archetype', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestArchetypeRepository_CountByTenant(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	archetypeRepo := NewArchetypeRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("empty count", func(t *testing.T) {
		count, err := archetypeRepo.CountByTenant(ctx, testTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 0 {
			t.Errorf("expected count to be 0, got %d", count)
		}
	})

	t.Run("count with archetypes", func(t *testing.T) {
		// Create multiple archetypes
		archetypeNames := []string{"Archetype 1", "Archetype 2", "Archetype 3"}
		for _, name := range archetypeNames {
			archetype, err := world.NewArchetype(testTenant.ID, name)
			if err != nil {
				t.Fatalf("failed to create archetype: %v", err)
			}
			err = archetypeRepo.Create(ctx, archetype)
			if err != nil {
				t.Fatalf("failed to create archetype: %v", err)
			}
		}

		count, err := archetypeRepo.CountByTenant(ctx, testTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count < len(archetypeNames) {
			t.Errorf("expected count to be at least %d, got %d", len(archetypeNames), count)
		}
	})

	t.Run("tenant isolation", func(t *testing.T) {
		// Create another tenant
		otherTenant, err := tenant.NewTenant("other-tenant-count", nil)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}
		err = tenantRepo.Create(ctx, otherTenant)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		// Create archetype for other tenant
		otherArchetype, err := world.NewArchetype(otherTenant.ID, "Other Archetype")
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}
		err = archetypeRepo.Create(ctx, otherArchetype)
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		// Get count for test tenant
		testCount, err := archetypeRepo.CountByTenant(ctx, testTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Get count for other tenant
		otherCount, err := archetypeRepo.CountByTenant(ctx, otherTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if otherCount != 1 {
			t.Errorf("expected other tenant count to be 1, got %d", otherCount)
		}

		// Test tenant count should not include other tenant's archetype
		if testCount == otherCount && testCount > 0 {
			t.Error("expected counts to be different due to tenant isolation")
		}
	})
}

