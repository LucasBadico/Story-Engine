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

func TestTraitRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	traitRepo := NewTraitRepository(db)

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
		trait, err := world.NewTrait(testTenant.ID, "Test Trait")
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		err = traitRepo.Create(ctx, trait)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify trait can be retrieved
		retrieved, err := traitRepo.GetByID(ctx, testTenant.ID, trait.ID)
		if err != nil {
			t.Fatalf("failed to retrieve trait: %v", err)
		}

		if retrieved.Name != "Test Trait" {
			t.Errorf("expected name to be 'Test Trait', got '%s'", retrieved.Name)
		}
	})

	t.Run("creation with category and description", func(t *testing.T) {
		trait, err := world.NewTrait(testTenant.ID, "Typed Trait")
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}
		trait.UpdateCategory("Physical")
		trait.UpdateDescription("A test trait description")

		err = traitRepo.Create(ctx, trait)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := traitRepo.GetByID(ctx, testTenant.ID, trait.ID)
		if err != nil {
			t.Fatalf("failed to retrieve trait: %v", err)
		}

		if retrieved.Category != "Physical" {
			t.Errorf("expected category to be 'Physical', got '%s'", retrieved.Category)
		}

		if retrieved.Description != "A test trait description" {
			t.Errorf("expected description to be 'A test trait description', got '%s'", retrieved.Description)
		}
	})
}

func TestTraitRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	traitRepo := NewTraitRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("existing trait", func(t *testing.T) {
		trait, err := world.NewTrait(testTenant.ID, "GetByID Trait")
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		err = traitRepo.Create(ctx, trait)
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		retrieved, err := traitRepo.GetByID(ctx, testTenant.ID, trait.ID)
		if err != nil {
			t.Fatalf("failed to get trait: %v", err)
		}

		if retrieved.ID != trait.ID {
			t.Errorf("expected ID to be %s, got %s", trait.ID, retrieved.ID)
		}

		if retrieved.Name != "GetByID Trait" {
			t.Errorf("expected name to be 'GetByID Trait', got '%s'", retrieved.Name)
		}
	})

	t.Run("non-existent trait", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := traitRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent trait")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "trait" {
			t.Errorf("expected resource to be 'trait', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestTraitRepository_ListByTenant(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	traitRepo := NewTraitRepository(db)

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
		traits, err := traitRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(traits) != 0 {
			t.Errorf("expected empty list, got %d traits", len(traits))
		}
	})

	t.Run("list with traits", func(t *testing.T) {
		// Create multiple traits
		traitNames := []string{"Trait A", "Trait B", "Trait C"}
		createdTraits := make([]*world.Trait, 0, len(traitNames))

		for _, name := range traitNames {
			trait, err := world.NewTrait(testTenant.ID, name)
			if err != nil {
				t.Fatalf("failed to create trait: %v", err)
			}
			err = traitRepo.Create(ctx, trait)
			if err != nil {
				t.Fatalf("failed to create trait: %v", err)
			}
			createdTraits = append(createdTraits, trait)
		}

		traits, err := traitRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(traits) != len(traitNames) {
			t.Errorf("expected %d traits, got %d", len(traitNames), len(traits))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		// Create more traits
		traitNames := []string{"Trait D", "Trait E"}
		for _, name := range traitNames {
			trait, _ := world.NewTrait(testTenant.ID, name)
			traitRepo.Create(ctx, trait)
		}

		// Test limit
		traits, err := traitRepo.ListByTenant(ctx, testTenant.ID, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(traits) > 2 {
			t.Errorf("expected at most 2 traits, got %d", len(traits))
		}

		// Test offset
		allTraits, err := traitRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(allTraits) < 3 {
			t.Skip("not enough traits to test pagination")
		}

		offsetTraits, err := traitRepo.ListByTenant(ctx, testTenant.ID, 10, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(offsetTraits) >= len(allTraits) {
			t.Error("expected offset to reduce number of traits")
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

		// Create trait for other tenant
		otherTrait, err := world.NewTrait(otherTenant.ID, "Other Trait")
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}
		err = traitRepo.Create(ctx, otherTrait)
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		// List traits for test tenant
		traits, err := traitRepo.ListByTenant(ctx, testTenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify other tenant's trait is not included
		for _, trait := range traits {
			if trait.ID == otherTrait.ID {
				t.Error("expected other tenant's trait to be excluded")
			}
		}
	})
}

func TestTraitRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	traitRepo := NewTraitRepository(db)

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
		trait, err := world.NewTrait(testTenant.ID, "Update Trait")
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		err = traitRepo.Create(ctx, trait)
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		// Update name
		err = trait.UpdateName("Updated Name")
		if err != nil {
			t.Fatalf("failed to update name: %v", err)
		}

		// Update category
		trait.UpdateCategory("Mental")

		// Update description
		trait.UpdateDescription("Updated Description")

		err = traitRepo.Update(ctx, trait)
		if err != nil {
			t.Fatalf("failed to update trait: %v", err)
		}

		// Verify update
		retrieved, err := traitRepo.GetByID(ctx, testTenant.ID, trait.ID)
		if err != nil {
			t.Fatalf("failed to get trait: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("expected name to be 'Updated Name', got '%s'", retrieved.Name)
		}

		if retrieved.Category != "Mental" {
			t.Errorf("expected category to be 'Mental', got '%s'", retrieved.Category)
		}

		if retrieved.Description != "Updated Description" {
			t.Errorf("expected description to be 'Updated Description', got '%s'", retrieved.Description)
		}
	})
}

func TestTraitRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	traitRepo := NewTraitRepository(db)

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
		trait, err := world.NewTrait(testTenant.ID, "Delete Trait")
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		err = traitRepo.Create(ctx, trait)
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		err = traitRepo.Delete(ctx, testTenant.ID, trait.ID)
		if err != nil {
			t.Fatalf("failed to delete trait: %v", err)
		}

		// Verify trait is deleted
		_, err = traitRepo.GetByID(ctx, testTenant.ID, trait.ID)
		if err == nil {
			t.Fatal("expected error for deleted trait")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "trait" {
			t.Errorf("expected resource to be 'trait', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestTraitRepository_CountByTenant(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	traitRepo := NewTraitRepository(db)

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
		count, err := traitRepo.CountByTenant(ctx, testTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 0 {
			t.Errorf("expected count to be 0, got %d", count)
		}
	})

	t.Run("count with traits", func(t *testing.T) {
		// Create multiple traits
		traitNames := []string{"Trait 1", "Trait 2", "Trait 3"}
		for _, name := range traitNames {
			trait, err := world.NewTrait(testTenant.ID, name)
			if err != nil {
				t.Fatalf("failed to create trait: %v", err)
			}
			err = traitRepo.Create(ctx, trait)
			if err != nil {
				t.Fatalf("failed to create trait: %v", err)
			}
		}

		count, err := traitRepo.CountByTenant(ctx, testTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count < len(traitNames) {
			t.Errorf("expected count to be at least %d, got %d", len(traitNames), count)
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

		// Create trait for other tenant
		otherTrait, err := world.NewTrait(otherTenant.ID, "Other Trait")
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}
		err = traitRepo.Create(ctx, otherTrait)
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		// Get count for test tenant
		testCount, err := traitRepo.CountByTenant(ctx, testTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Get count for other tenant
		otherCount, err := traitRepo.CountByTenant(ctx, otherTenant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if otherCount != 1 {
			t.Errorf("expected other tenant count to be 1, got %d", otherCount)
		}

		// Test tenant count should not include other tenant's trait
		if testCount == otherCount && testCount > 0 {
			t.Error("expected counts to be different due to tenant isolation")
		}
	})
}

