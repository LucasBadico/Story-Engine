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

func TestWorldRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)

	// Create a tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		w, err := world.NewWorld(testTenant.ID, "Test World", false)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		err = worldRepo.Create(ctx, w)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify world can be retrieved
		retrieved, err := worldRepo.GetByID(ctx, testTenant.ID, w.ID)
		if err != nil {
			t.Fatalf("failed to retrieve world: %v", err)
		}

		if retrieved.Name != "Test World" {
			t.Errorf("expected name to be 'Test World', got '%s'", retrieved.Name)
		}

		if retrieved.TenantID != testTenant.ID {
			t.Errorf("expected tenant_id to be %s, got %s", testTenant.ID, retrieved.TenantID)
		}

		if retrieved.IsImplicit {
			t.Errorf("expected IsImplicit to be false, got true")
		}
	})

	t.Run("successful creation with RPG system", func(t *testing.T) {
		rpgSystemID := uuid.New()
		w, err := world.NewWorld(testTenant.ID, "RPG World", false)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}
		w.SetRPGSystem(&rpgSystemID)

		err = worldRepo.Create(ctx, w)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify world can be retrieved
		retrieved, err := worldRepo.GetByID(ctx, testTenant.ID, w.ID)
		if err != nil {
			t.Fatalf("failed to retrieve world: %v", err)
		}

		if retrieved.RPGSystemID == nil {
			t.Fatal("expected RPGSystemID to be set, got nil")
		}

		if *retrieved.RPGSystemID != rpgSystemID {
			t.Errorf("expected RPGSystemID to be %s, got %s", rpgSystemID, *retrieved.RPGSystemID)
		}
	})

	t.Run("successful creation with implicit flag", func(t *testing.T) {
		w, err := world.NewWorld(testTenant.ID, "Implicit World", true)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		err = worldRepo.Create(ctx, w)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify world can be retrieved
		retrieved, err := worldRepo.GetByID(ctx, testTenant.ID, w.ID)
		if err != nil {
			t.Fatalf("failed to retrieve world: %v", err)
		}

		if !retrieved.IsImplicit {
			t.Errorf("expected IsImplicit to be true, got false")
		}
	})
}

func TestWorldRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)

	// Create a tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("existing world", func(t *testing.T) {
		w, err := world.NewWorld(testTenant.ID, "GetByID World", false)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		err = worldRepo.Create(ctx, w)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		retrieved, err := worldRepo.GetByID(ctx, testTenant.ID, w.ID)
		if err != nil {
			t.Fatalf("failed to get world: %v", err)
		}

		if retrieved.ID != w.ID {
			t.Errorf("expected ID to be %s, got %s", w.ID, retrieved.ID)
		}

		if retrieved.Name != "GetByID World" {
			t.Errorf("expected name to be 'GetByID World', got '%s'", retrieved.Name)
		}
	})

	t.Run("non-existent world", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := worldRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent world")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "world" {
			t.Errorf("expected resource to be 'world', got '%s'", notFoundErr.Resource)
		}
	})

	t.Run("wrong tenant", func(t *testing.T) {
		// Create another tenant
		otherTenant, err := tenant.NewTenant("other-tenant", nil)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}
		err = tenantRepo.Create(ctx, otherTenant)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		// Create world for first tenant
		w, err := world.NewWorld(testTenant.ID, "First Tenant World", false)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}
		err = worldRepo.Create(ctx, w)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		// Try to get world with wrong tenant
		_, err = worldRepo.GetByID(ctx, otherTenant.ID, w.ID)
		if err == nil {
			t.Fatal("expected error for wrong tenant")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "world" {
			t.Errorf("expected resource to be 'world', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestWorldRepository_ListByTenant(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)

	// Create a tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("empty list", func(t *testing.T) {
		worlds, err := worldRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(worlds) != 0 {
			t.Errorf("expected empty list, got %d worlds", len(worlds))
		}
	})

	t.Run("list with worlds", func(t *testing.T) {
		// Create multiple worlds
		worldNames := []string{"World 1", "World 2", "World 3"}
		createdWorlds := make([]*world.World, 0, len(worldNames))

		for _, name := range worldNames {
			w, err := world.NewWorld(testTenant.ID, name, false)
			if err != nil {
				t.Fatalf("failed to create world: %v", err)
			}
			err = worldRepo.Create(ctx, w)
			if err != nil {
				t.Fatalf("failed to create world: %v", err)
			}
			createdWorlds = append(createdWorlds, w)
			// Small delay to ensure different timestamps
			time.Sleep(10 * time.Millisecond)
		}

		worlds, err := worldRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(worlds) != len(worldNames) {
			t.Errorf("expected %d worlds, got %d", len(worldNames), len(worlds))
		}

		// Verify worlds are returned in descending order by created_at
		for i := 0; i < len(worlds)-1; i++ {
			if worlds[i].CreatedAt.Before(worlds[i+1].CreatedAt) {
				t.Error("expected worlds to be in descending order by created_at")
			}
		}
	})

	t.Run("pagination", func(t *testing.T) {
		// Create 5 worlds
		for i := 1; i <= 5; i++ {
			w, err := world.NewWorld(testTenant.ID, "Pagination World", false)
			if err != nil {
				t.Fatalf("failed to create world: %v", err)
			}
			err = worldRepo.Create(ctx, w)
			if err != nil {
				t.Fatalf("failed to create world: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}

		// Get first page
		worlds, err := worldRepo.ListByTenant(ctx, testTenant.ID, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(worlds) != 2 {
			t.Errorf("expected 2 worlds, got %d", len(worlds))
		}

		// Get second page
		worlds, err = worldRepo.ListByTenant(ctx, testTenant.ID, 2, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(worlds) != 2 {
			t.Errorf("expected 2 worlds, got %d", len(worlds))
		}

		// Get third page
		worlds, err = worldRepo.ListByTenant(ctx, testTenant.ID, 2, 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(worlds) > 2 {
			t.Errorf("expected at most 2 worlds, got %d", len(worlds))
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

		// Create world for other tenant
		otherWorld, err := world.NewWorld(otherTenant.ID, "Other Tenant World", false)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}
		err = worldRepo.Create(ctx, otherWorld)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		// List worlds for first tenant
		worlds, err := worldRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify other tenant's world is not included
		for _, w := range worlds {
			if w.ID == otherWorld.ID || w.TenantID == otherTenant.ID {
				t.Error("found world from other tenant in list")
			}
		}
	})
}

func TestWorldRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)

	// Create a tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		w, err := world.NewWorld(testTenant.ID, "Update World", false)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		err = worldRepo.Create(ctx, w)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		// Update name
		err = w.UpdateName("Updated Name")
		if err != nil {
			t.Fatalf("failed to update world name: %v", err)
		}

		// Update description
		w.UpdateDescription("Updated Description")

		// Update genre
		w.UpdateGenre("Fantasy")

		// Update implicit flag
		w.SetImplicit(true)

		err = worldRepo.Update(ctx, w)
		if err != nil {
			t.Fatalf("failed to update world: %v", err)
		}

		// Verify update
		retrieved, err := worldRepo.GetByID(ctx, testTenant.ID, w.ID)
		if err != nil {
			t.Fatalf("failed to get world: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("expected name to be 'Updated Name', got '%s'", retrieved.Name)
		}

		if retrieved.Description != "Updated Description" {
			t.Errorf("expected description to be 'Updated Description', got '%s'", retrieved.Description)
		}

		if retrieved.Genre != "Fantasy" {
			t.Errorf("expected genre to be 'Fantasy', got '%s'", retrieved.Genre)
		}

		if !retrieved.IsImplicit {
			t.Errorf("expected IsImplicit to be true, got false")
		}
	})

	t.Run("update RPG system", func(t *testing.T) {
		w, err := world.NewWorld(testTenant.ID, "RPG Update World", false)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		err = worldRepo.Create(ctx, w)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		// Set RPG system
		rpgSystemID := uuid.New()
		w.SetRPGSystem(&rpgSystemID)

		err = worldRepo.Update(ctx, w)
		if err != nil {
			t.Fatalf("failed to update world: %v", err)
		}

		// Verify update
		retrieved, err := worldRepo.GetByID(ctx, testTenant.ID, w.ID)
		if err != nil {
			t.Fatalf("failed to get world: %v", err)
		}

		if retrieved.RPGSystemID == nil {
			t.Fatal("expected RPGSystemID to be set, got nil")
		}

		if *retrieved.RPGSystemID != rpgSystemID {
			t.Errorf("expected RPGSystemID to be %s, got %s", rpgSystemID, *retrieved.RPGSystemID)
		}

		// Clear RPG system
		w.SetRPGSystem(nil)

		err = worldRepo.Update(ctx, w)
		if err != nil {
			t.Fatalf("failed to update world: %v", err)
		}

		// Verify update
		retrieved, err = worldRepo.GetByID(ctx, testTenant.ID, w.ID)
		if err != nil {
			t.Fatalf("failed to get world: %v", err)
		}

		if retrieved.RPGSystemID != nil {
			t.Error("expected RPGSystemID to be nil, got non-nil")
		}
	})
}

func TestWorldRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)

	// Create a tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		w, err := world.NewWorld(testTenant.ID, "Delete World", false)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		err = worldRepo.Create(ctx, w)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		err = worldRepo.Delete(ctx, testTenant.ID, w.ID)
		if err != nil {
			t.Fatalf("failed to delete world: %v", err)
		}

		// Verify world is deleted
		_, err = worldRepo.GetByID(ctx, testTenant.ID, w.ID)
		if err == nil {
			t.Fatal("expected error for deleted world")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "world" {
			t.Errorf("expected resource to be 'world', got '%s'", notFoundErr.Resource)
		}
	})

	t.Run("delete non-existent world", func(t *testing.T) {
		nonExistentID := uuid.New()

		err := worldRepo.Delete(ctx, testTenant.ID, nonExistentID)
		if err != nil {
			// Delete can succeed even if world doesn't exist (idempotent)
			// This is acceptable behavior
		}
	})
}

func TestWorldRepository_CountByTenant(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)

	// Create a tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("empty count", func(t *testing.T) {
		count, err := worldRepo.CountByTenant(ctx, testTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 0 {
			t.Errorf("expected count to be 0, got %d", count)
		}
	})

	t.Run("count with worlds", func(t *testing.T) {
		// Create 3 worlds
		for i := 1; i <= 3; i++ {
			w, err := world.NewWorld(testTenant.ID, "Count World", false)
			if err != nil {
				t.Fatalf("failed to create world: %v", err)
			}
			err = worldRepo.Create(ctx, w)
			if err != nil {
				t.Fatalf("failed to create world: %v", err)
			}
		}

		count, err := worldRepo.CountByTenant(ctx, testTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 3 {
			t.Errorf("expected count to be 3, got %d", count)
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

		// Create world for other tenant
		w, err := world.NewWorld(otherTenant.ID, "Other Tenant World", false)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}
		err = worldRepo.Create(ctx, w)
		if err != nil {
			t.Fatalf("failed to create world: %v", err)
		}

		// Count worlds for first tenant (should still be 3)
		count, err := worldRepo.CountByTenant(ctx, testTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 3 {
			t.Errorf("expected count to be 3, got %d", count)
		}

		// Count worlds for other tenant (should be 1)
		count, err = worldRepo.CountByTenant(ctx, otherTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 1 {
			t.Errorf("expected count to be 1, got %d", count)
		}
	})
}

