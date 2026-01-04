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

func TestArtifactRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	artifactRepo := NewArtifactRepository(db)

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
		artifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		err = artifactRepo.Create(ctx, artifact)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify artifact can be retrieved
		retrieved, err := artifactRepo.GetByID(ctx, testTenant.ID, artifact.ID)
		if err != nil {
			t.Fatalf("failed to retrieve artifact: %v", err)
		}

		if retrieved.Name != "Test Artifact" {
			t.Errorf("expected name to be 'Test Artifact', got '%s'", retrieved.Name)
		}

		if retrieved.TenantID != testTenant.ID {
			t.Errorf("expected tenant_id to be %s, got %s", testTenant.ID, retrieved.TenantID)
		}
	})

	t.Run("successful creation with description", func(t *testing.T) {
		artifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact With Description")
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}
		artifact.UpdateDescription("A test artifact description")

		err = artifactRepo.Create(ctx, artifact)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := artifactRepo.GetByID(ctx, testTenant.ID, artifact.ID)
		if err != nil {
			t.Fatalf("failed to retrieve artifact: %v", err)
		}

		if retrieved.Description != "A test artifact description" {
			t.Errorf("expected description to be 'A test artifact description', got '%s'", retrieved.Description)
		}
	})

	t.Run("successful creation with rarity", func(t *testing.T) {
		artifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Rare Artifact")
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}
		artifact.UpdateRarity("Legendary")

		err = artifactRepo.Create(ctx, artifact)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := artifactRepo.GetByID(ctx, testTenant.ID, artifact.ID)
		if err != nil {
			t.Fatalf("failed to retrieve artifact: %v", err)
		}

		if retrieved.Rarity != "Legendary" {
			t.Errorf("expected rarity to be 'Legendary', got '%s'", retrieved.Rarity)
		}
	})
}

func TestArtifactRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	artifactRepo := NewArtifactRepository(db)

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

	t.Run("existing artifact", func(t *testing.T) {
		artifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "GetByID Artifact")
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		err = artifactRepo.Create(ctx, artifact)
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		retrieved, err := artifactRepo.GetByID(ctx, testTenant.ID, artifact.ID)
		if err != nil {
			t.Fatalf("failed to get artifact: %v", err)
		}

		if retrieved.ID != artifact.ID {
			t.Errorf("expected ID to be %s, got %s", artifact.ID, retrieved.ID)
		}

		if retrieved.Name != "GetByID Artifact" {
			t.Errorf("expected name to be 'GetByID Artifact', got '%s'", retrieved.Name)
		}
	})

	t.Run("non-existent artifact", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := artifactRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent artifact")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "artifact" {
			t.Errorf("expected resource to be 'artifact', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestArtifactRepository_ListByWorld(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	artifactRepo := NewArtifactRepository(db)

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
		artifacts, err := artifactRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(artifacts) != 0 {
			t.Errorf("expected empty list, got %d artifacts", len(artifacts))
		}
	})

	t.Run("list with artifacts", func(t *testing.T) {
		// Create multiple artifacts
		artifactNames := []string{"Artifact A", "Artifact B", "Artifact C"}
		createdArtifacts := make([]*world.Artifact, 0, len(artifactNames))

		for _, name := range artifactNames {
			artifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, name)
			if err != nil {
				t.Fatalf("failed to create artifact: %v", err)
			}
			err = artifactRepo.Create(ctx, artifact)
			if err != nil {
				t.Fatalf("failed to create artifact: %v", err)
			}
			createdArtifacts = append(createdArtifacts, artifact)
			time.Sleep(10 * time.Millisecond)
		}

		artifacts, err := artifactRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(artifacts) != len(artifactNames) {
			t.Errorf("expected %d artifacts, got %d", len(artifactNames), len(artifacts))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		// Create 5 artifacts
		for i := 1; i <= 5; i++ {
			artifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Pagination Artifact")
			if err != nil {
				t.Fatalf("failed to create artifact: %v", err)
			}
			err = artifactRepo.Create(ctx, artifact)
			if err != nil {
				t.Fatalf("failed to create artifact: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}

		// Get first page
		artifacts, err := artifactRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(artifacts) != 2 {
			t.Errorf("expected 2 artifacts, got %d", len(artifacts))
		}

		// Get second page
		artifacts, err = artifactRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID, 2, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(artifacts) != 2 {
			t.Errorf("expected 2 artifacts, got %d", len(artifacts))
		}
	})
}

func TestArtifactRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	artifactRepo := NewArtifactRepository(db)

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
		artifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Update Artifact")
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		err = artifactRepo.Create(ctx, artifact)
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		// Update name
		err = artifact.UpdateName("Updated Name")
		if err != nil {
			t.Fatalf("failed to update name: %v", err)
		}

		// Update description
		artifact.UpdateDescription("Updated Description")

		// Update rarity
		artifact.UpdateRarity("Epic")

		err = artifactRepo.Update(ctx, artifact)
		if err != nil {
			t.Fatalf("failed to update artifact: %v", err)
		}

		// Verify update
		retrieved, err := artifactRepo.GetByID(ctx, testTenant.ID, artifact.ID)
		if err != nil {
			t.Fatalf("failed to get artifact: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("expected name to be 'Updated Name', got '%s'", retrieved.Name)
		}

		if retrieved.Description != "Updated Description" {
			t.Errorf("expected description to be 'Updated Description', got '%s'", retrieved.Description)
		}

		if retrieved.Rarity != "Epic" {
			t.Errorf("expected rarity to be 'Epic', got '%s'", retrieved.Rarity)
		}
	})
}

func TestArtifactRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	artifactRepo := NewArtifactRepository(db)

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
		artifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Delete Artifact")
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		err = artifactRepo.Create(ctx, artifact)
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		err = artifactRepo.Delete(ctx, testTenant.ID, artifact.ID)
		if err != nil {
			t.Fatalf("failed to delete artifact: %v", err)
		}

		// Verify artifact is deleted
		_, err = artifactRepo.GetByID(ctx, testTenant.ID, artifact.ID)
		if err == nil {
			t.Fatal("expected error for deleted artifact")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "artifact" {
			t.Errorf("expected resource to be 'artifact', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestArtifactRepository_CountByWorld(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	artifactRepo := NewArtifactRepository(db)

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
		count, err := artifactRepo.CountByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 0 {
			t.Errorf("expected count to be 0, got %d", count)
		}
	})

	t.Run("count with artifacts", func(t *testing.T) {
		// Create 3 artifacts
		for i := 1; i <= 3; i++ {
			artifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Count Artifact")
			if err != nil {
				t.Fatalf("failed to create artifact: %v", err)
			}
			err = artifactRepo.Create(ctx, artifact)
			if err != nil {
				t.Fatalf("failed to create artifact: %v", err)
			}
		}

		count, err := artifactRepo.CountByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 3 {
			t.Errorf("expected count to be 3, got %d", count)
		}
	})
}

