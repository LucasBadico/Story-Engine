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

func TestArchetypeTraitRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	archetypeRepo := NewArchetypeRepository(db)
	traitRepo := NewTraitRepository(db)
	archetypeTraitRepo := NewArchetypeTraitRepository(db)

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
		// Create archetype and trait
		testArchetype, err := world.NewArchetype(testTenant.ID, "Test Archetype")
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}
		err = archetypeRepo.Create(ctx, testArchetype)
		if err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}

		testTrait, err := world.NewTrait(testTenant.ID, "Test Trait")
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}
		err = traitRepo.Create(ctx, testTrait)
		if err != nil {
			t.Fatalf("failed to create trait: %v", err)
		}

		// Create archetype-trait relationship
		archetypeTrait := world.NewArchetypeTrait(testArchetype.ID, testTrait.ID, "10")

		err = archetypeTraitRepo.Create(ctx, archetypeTrait)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify relationship can be retrieved
		traits, err := archetypeTraitRepo.GetByArchetype(ctx, testTenant.ID, testArchetype.ID)
		if err != nil {
			t.Fatalf("failed to retrieve archetype traits: %v", err)
		}

		if len(traits) != 1 {
			t.Fatalf("expected 1 trait, got %d", len(traits))
		}

		if traits[0].TraitID != testTrait.ID {
			t.Errorf("expected trait_id to be %s, got %s", testTrait.ID, traits[0].TraitID)
		}

		if traits[0].DefaultValue != "10" {
			t.Errorf("expected default_value to be '10', got '%s'", traits[0].DefaultValue)
		}
	})
}

func TestArchetypeTraitRepository_GetByArchetype(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	archetypeRepo := NewArchetypeRepository(db)
	traitRepo := NewTraitRepository(db)
	archetypeTraitRepo := NewArchetypeTraitRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("get traits for archetype", func(t *testing.T) {
		testArchetype, _ := world.NewArchetype(testTenant.ID, "Test Archetype")
		archetypeRepo.Create(ctx, testArchetype)

		// Create multiple traits
		trait1, _ := world.NewTrait(testTenant.ID, "Trait 1")
		trait2, _ := world.NewTrait(testTenant.ID, "Trait 2")
		traitRepo.Create(ctx, trait1)
		traitRepo.Create(ctx, trait2)

		// Create relationships
		at1 := world.NewArchetypeTrait(testArchetype.ID, trait1.ID, "5")
		at2 := world.NewArchetypeTrait(testArchetype.ID, trait2.ID, "10")
		archetypeTraitRepo.Create(ctx, at1)
		archetypeTraitRepo.Create(ctx, at2)

		traits, err := archetypeTraitRepo.GetByArchetype(ctx, testTenant.ID, testArchetype.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(traits) != 2 {
			t.Errorf("expected 2 traits, got %d", len(traits))
		}

		for _, at := range traits {
			if at.ArchetypeID != testArchetype.ID {
				t.Errorf("expected archetype_id to be %s, got %s", testArchetype.ID, at.ArchetypeID)
			}
		}
	})

	t.Run("empty list", func(t *testing.T) {
		testArchetype, _ := world.NewArchetype(testTenant.ID, "Empty Archetype")
		archetypeRepo.Create(ctx, testArchetype)

		traits, err := archetypeTraitRepo.GetByArchetype(ctx, testTenant.ID, testArchetype.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(traits) != 0 {
			t.Errorf("expected empty list, got %d traits", len(traits))
		}
	})
}

func TestArchetypeTraitRepository_GetByTrait(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	archetypeRepo := NewArchetypeRepository(db)
	traitRepo := NewTraitRepository(db)
	archetypeTraitRepo := NewArchetypeTraitRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("get archetypes for trait", func(t *testing.T) {
		testTrait, _ := world.NewTrait(testTenant.ID, "Test Trait")
		traitRepo.Create(ctx, testTrait)

		// Create multiple archetypes
		archetype1, _ := world.NewArchetype(testTenant.ID, "Archetype 1")
		archetype2, _ := world.NewArchetype(testTenant.ID, "Archetype 2")
		archetypeRepo.Create(ctx, archetype1)
		archetypeRepo.Create(ctx, archetype2)

		// Create relationships
		at1 := world.NewArchetypeTrait(archetype1.ID, testTrait.ID, "5")
		at2 := world.NewArchetypeTrait(archetype2.ID, testTrait.ID, "10")
		archetypeTraitRepo.Create(ctx, at1)
		archetypeTraitRepo.Create(ctx, at2)

		archetypes, err := archetypeTraitRepo.GetByTrait(ctx, testTenant.ID, testTrait.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(archetypes) != 2 {
			t.Errorf("expected 2 archetypes, got %d", len(archetypes))
		}

		for _, at := range archetypes {
			if at.TraitID != testTrait.ID {
				t.Errorf("expected trait_id to be %s, got %s", testTrait.ID, at.TraitID)
			}
		}
	})
}

func TestArchetypeTraitRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	archetypeRepo := NewArchetypeRepository(db)
	traitRepo := NewTraitRepository(db)
	archetypeTraitRepo := NewArchetypeTraitRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("delete specific relationship", func(t *testing.T) {
		testArchetype, _ := world.NewArchetype(testTenant.ID, "Test Archetype")
		archetypeRepo.Create(ctx, testArchetype)

		testTrait, _ := world.NewTrait(testTenant.ID, "Test Trait")
		traitRepo.Create(ctx, testTrait)

		archetypeTrait := world.NewArchetypeTrait(testArchetype.ID, testTrait.ID, "10")
		archetypeTraitRepo.Create(ctx, archetypeTrait)

		// Verify it exists
		traits, err := archetypeTraitRepo.GetByArchetype(ctx, testTenant.ID, testArchetype.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(traits) != 1 {
			t.Fatalf("expected 1 trait, got %d", len(traits))
		}

		// Delete
		err = archetypeTraitRepo.Delete(ctx, testTenant.ID, testArchetype.ID, testTrait.ID)
		if err != nil {
			t.Fatalf("failed to delete: %v", err)
		}

		// Verify deletion
		traits, err = archetypeTraitRepo.GetByArchetype(ctx, testTenant.ID, testArchetype.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(traits) != 0 {
			t.Errorf("expected no traits, got %d", len(traits))
		}
	})
}

func TestArchetypeTraitRepository_DeleteByArchetype(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	archetypeRepo := NewArchetypeRepository(db)
	traitRepo := NewTraitRepository(db)
	archetypeTraitRepo := NewArchetypeTraitRepository(db)

	// Create tenant first
	testTenant, err := tenant.NewTenant("test-tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	t.Run("delete all relationships for archetype", func(t *testing.T) {
		testArchetype, _ := world.NewArchetype(testTenant.ID, "Test Archetype")
		archetypeRepo.Create(ctx, testArchetype)

		// Create multiple traits
		trait1, _ := world.NewTrait(testTenant.ID, "Trait 1")
		trait2, _ := world.NewTrait(testTenant.ID, "Trait 2")
		traitRepo.Create(ctx, trait1)
		traitRepo.Create(ctx, trait2)

		// Create relationships
		at1 := world.NewArchetypeTrait(testArchetype.ID, trait1.ID, "5")
		at2 := world.NewArchetypeTrait(testArchetype.ID, trait2.ID, "10")
		archetypeTraitRepo.Create(ctx, at1)
		archetypeTraitRepo.Create(ctx, at2)

		// Verify they exist
		traits, err := archetypeTraitRepo.GetByArchetype(ctx, testTenant.ID, testArchetype.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(traits) != 2 {
			t.Fatalf("expected 2 traits, got %d", len(traits))
		}

		// Delete all
		err = archetypeTraitRepo.DeleteByArchetype(ctx, testTenant.ID, testArchetype.ID)
		if err != nil {
			t.Fatalf("failed to delete by archetype: %v", err)
		}

		// Verify all deleted
		traits, err = archetypeTraitRepo.GetByArchetype(ctx, testTenant.ID, testArchetype.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(traits) != 0 {
			t.Errorf("expected no traits, got %d", len(traits))
		}
	})
}

