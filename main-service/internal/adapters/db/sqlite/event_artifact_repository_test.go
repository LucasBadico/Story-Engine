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

func TestEventArtifactRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	artifactRepo := NewArtifactRepository(db)
	eventArtifactRepo := NewEventArtifactRepository(db)

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

	// Create event and artifact
	testEvent, err := world.NewEvent(testTenant.ID, testWorld.ID, "Test Event")
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	err = eventRepo.Create(ctx, testEvent)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	err = artifactRepo.Create(ctx, testArtifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		ea := world.NewEventArtifact(testEvent.ID, testArtifact.ID, nil)

		err = eventArtifactRepo.Create(ctx, ea)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify event-artifact can be retrieved
		retrieved, err := eventArtifactRepo.GetByID(ctx, testTenant.ID, ea.ID)
		if err != nil {
			t.Fatalf("failed to retrieve event-artifact: %v", err)
		}

		if retrieved.EventID != testEvent.ID {
			t.Errorf("expected event_id to be %s, got %s", testEvent.ID, retrieved.EventID)
		}

		if retrieved.ArtifactID != testArtifact.ID {
			t.Errorf("expected artifact_id to be %s, got %s", testArtifact.ID, retrieved.ArtifactID)
		}
	})

	t.Run("successful creation with role", func(t *testing.T) {
		// Create a new artifact to avoid UNIQUE constraint violation
		artifact2, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact With Role")
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}
		err = artifactRepo.Create(ctx, artifact2)
		if err != nil {
			t.Fatalf("failed to create artifact: %v", err)
		}

		role := "Key Artifact"
		ea := world.NewEventArtifact(testEvent.ID, artifact2.ID, &role)

		err = eventArtifactRepo.Create(ctx, ea)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := eventArtifactRepo.GetByID(ctx, testTenant.ID, ea.ID)
		if err != nil {
			t.Fatalf("failed to retrieve event-artifact: %v", err)
		}

		if retrieved.Role == nil {
			t.Fatal("expected role to be set, got nil")
		}

		if *retrieved.Role != role {
			t.Errorf("expected role to be '%s', got '%s'", role, *retrieved.Role)
		}
	})
}

func TestEventArtifactRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	artifactRepo := NewArtifactRepository(db)
	eventArtifactRepo := NewEventArtifactRepository(db)

	// Create tenant, world, event, and artifact
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

	testEvent, err := world.NewEvent(testTenant.ID, testWorld.ID, "Test Event")
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	err = eventRepo.Create(ctx, testEvent)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	err = artifactRepo.Create(ctx, testArtifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	t.Run("existing event-artifact", func(t *testing.T) {
		ea := world.NewEventArtifact(testEvent.ID, testArtifact.ID, nil)
		err = eventArtifactRepo.Create(ctx, ea)
		if err != nil {
			t.Fatalf("failed to create event-artifact: %v", err)
		}

		retrieved, err := eventArtifactRepo.GetByID(ctx, testTenant.ID, ea.ID)
		if err != nil {
			t.Fatalf("failed to get event-artifact: %v", err)
		}

		if retrieved.ID != ea.ID {
			t.Errorf("expected ID to be %s, got %s", ea.ID, retrieved.ID)
		}

		if retrieved.EventID != testEvent.ID {
			t.Errorf("expected event_id to be %s, got %s", testEvent.ID, retrieved.EventID)
		}
	})

	t.Run("non-existent event-artifact", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := eventArtifactRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent event-artifact")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "event_artifact" {
			t.Errorf("expected resource to be 'event_artifact', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestEventArtifactRepository_ListByEvent(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	artifactRepo := NewArtifactRepository(db)
	eventArtifactRepo := NewEventArtifactRepository(db)

	// Create tenant, world, and event
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

	testEvent, err := world.NewEvent(testTenant.ID, testWorld.ID, "Test Event")
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	err = eventRepo.Create(ctx, testEvent)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	t.Run("empty list", func(t *testing.T) {
		eventArtifacts, err := eventArtifactRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventArtifacts) != 0 {
			t.Errorf("expected empty list, got %d event-artifacts", len(eventArtifacts))
		}
	})

	t.Run("list with event-artifacts", func(t *testing.T) {
		// Create multiple artifacts
		artifact1, _ := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact 1")
		artifact2, _ := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact 2")
		artifact3, _ := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact 3")
		artifactRepo.Create(ctx, artifact1)
		artifactRepo.Create(ctx, artifact2)
		artifactRepo.Create(ctx, artifact3)

		// Create event-artifacts
		ea1 := world.NewEventArtifact(testEvent.ID, artifact1.ID, nil)
		ea2 := world.NewEventArtifact(testEvent.ID, artifact2.ID, nil)
		ea3 := world.NewEventArtifact(testEvent.ID, artifact3.ID, nil)
		eventArtifactRepo.Create(ctx, ea1)
		time.Sleep(10 * time.Millisecond)
		eventArtifactRepo.Create(ctx, ea2)
		time.Sleep(10 * time.Millisecond)
		eventArtifactRepo.Create(ctx, ea3)

		eventArtifacts, err := eventArtifactRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventArtifacts) != 3 {
			t.Errorf("expected 3 event-artifacts, got %d", len(eventArtifacts))
		}

		// Verify all belong to the event
		for _, ea := range eventArtifacts {
			if ea.EventID != testEvent.ID {
				t.Errorf("expected event_id to be %s, got %s", testEvent.ID, ea.EventID)
			}
		}
	})
}

func TestEventArtifactRepository_ListByArtifact(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	artifactRepo := NewArtifactRepository(db)
	eventArtifactRepo := NewEventArtifactRepository(db)

	// Create tenant, world, and artifact
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

	testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	err = artifactRepo.Create(ctx, testArtifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	t.Run("empty list", func(t *testing.T) {
		eventArtifacts, err := eventArtifactRepo.ListByArtifact(ctx, testTenant.ID, testArtifact.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventArtifacts) != 0 {
			t.Errorf("expected empty list, got %d event-artifacts", len(eventArtifacts))
		}
	})

	t.Run("list with event-artifacts", func(t *testing.T) {
		// Create multiple events
		event1, _ := world.NewEvent(testTenant.ID, testWorld.ID, "Event 1")
		event2, _ := world.NewEvent(testTenant.ID, testWorld.ID, "Event 2")
		event3, _ := world.NewEvent(testTenant.ID, testWorld.ID, "Event 3")
		eventRepo.Create(ctx, event1)
		eventRepo.Create(ctx, event2)
		eventRepo.Create(ctx, event3)

		// Create event-artifacts
		ea1 := world.NewEventArtifact(event1.ID, testArtifact.ID, nil)
		ea2 := world.NewEventArtifact(event2.ID, testArtifact.ID, nil)
		ea3 := world.NewEventArtifact(event3.ID, testArtifact.ID, nil)
		eventArtifactRepo.Create(ctx, ea1)
		eventArtifactRepo.Create(ctx, ea2)
		eventArtifactRepo.Create(ctx, ea3)

		eventArtifacts, err := eventArtifactRepo.ListByArtifact(ctx, testTenant.ID, testArtifact.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventArtifacts) != 3 {
			t.Errorf("expected 3 event-artifacts, got %d", len(eventArtifacts))
		}

		// Verify all belong to the artifact
		for _, ea := range eventArtifacts {
			if ea.ArtifactID != testArtifact.ID {
				t.Errorf("expected artifact_id to be %s, got %s", testArtifact.ID, ea.ArtifactID)
			}
		}
	})
}

func TestEventArtifactRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	artifactRepo := NewArtifactRepository(db)
	eventArtifactRepo := NewEventArtifactRepository(db)

	// Create tenant, world, event, and artifact
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

	testEvent, err := world.NewEvent(testTenant.ID, testWorld.ID, "Test Event")
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	err = eventRepo.Create(ctx, testEvent)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	err = artifactRepo.Create(ctx, testArtifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		ea := world.NewEventArtifact(testEvent.ID, testArtifact.ID, nil)
		err = eventArtifactRepo.Create(ctx, ea)
		if err != nil {
			t.Fatalf("failed to create event-artifact: %v", err)
		}

		err = eventArtifactRepo.Delete(ctx, testTenant.ID, ea.ID)
		if err != nil {
			t.Fatalf("failed to delete event-artifact: %v", err)
		}

		// Verify event-artifact is deleted
		_, err = eventArtifactRepo.GetByID(ctx, testTenant.ID, ea.ID)
		if err == nil {
			t.Fatal("expected error for deleted event-artifact")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "event_artifact" {
			t.Errorf("expected resource to be 'event_artifact', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestEventArtifactRepository_DeleteByEvent(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	artifactRepo := NewArtifactRepository(db)
	eventArtifactRepo := NewEventArtifactRepository(db)

	// Create tenant, world, and event
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

	testEvent, err := world.NewEvent(testTenant.ID, testWorld.ID, "Test Event")
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	err = eventRepo.Create(ctx, testEvent)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	t.Run("delete all event-artifacts for event", func(t *testing.T) {
		// Create multiple artifacts
		artifact1, _ := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact 1")
		artifact2, _ := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact 2")
		artifact3, _ := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact 3")
		artifactRepo.Create(ctx, artifact1)
		artifactRepo.Create(ctx, artifact2)
		artifactRepo.Create(ctx, artifact3)

		// Create event-artifacts
		ea1 := world.NewEventArtifact(testEvent.ID, artifact1.ID, nil)
		ea2 := world.NewEventArtifact(testEvent.ID, artifact2.ID, nil)
		ea3 := world.NewEventArtifact(testEvent.ID, artifact3.ID, nil)
		eventArtifactRepo.Create(ctx, ea1)
		eventArtifactRepo.Create(ctx, ea2)
		eventArtifactRepo.Create(ctx, ea3)

		// Verify they exist
		eventArtifacts, err := eventArtifactRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(eventArtifacts) != 3 {
			t.Errorf("expected 3 event-artifacts, got %d", len(eventArtifacts))
		}

		// Delete all by event
		err = eventArtifactRepo.DeleteByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("failed to delete event-artifacts: %v", err)
		}

		// Verify all are deleted
		eventArtifacts, err = eventArtifactRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventArtifacts) != 0 {
			t.Errorf("expected no event-artifacts, got %d", len(eventArtifacts))
		}
	})
}

func TestEventArtifactRepository_DeleteByEventAndArtifact(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	artifactRepo := NewArtifactRepository(db)
	eventArtifactRepo := NewEventArtifactRepository(db)

	// Create tenant, world, event, and artifact
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

	testEvent, err := world.NewEvent(testTenant.ID, testWorld.ID, "Test Event")
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	err = eventRepo.Create(ctx, testEvent)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	testArtifact, err := world.NewArtifact(testTenant.ID, testWorld.ID, "Test Artifact")
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	err = artifactRepo.Create(ctx, testArtifact)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	t.Run("delete specific event-artifact", func(t *testing.T) {
		// Create multiple artifacts
		artifact2, _ := world.NewArtifact(testTenant.ID, testWorld.ID, "Artifact 2")
		artifactRepo.Create(ctx, artifact2)

		// Create event-artifacts
		ea1 := world.NewEventArtifact(testEvent.ID, testArtifact.ID, nil)
		ea2 := world.NewEventArtifact(testEvent.ID, artifact2.ID, nil)
		eventArtifactRepo.Create(ctx, ea1)
		eventArtifactRepo.Create(ctx, ea2)

		// Verify they exist
		eventArtifacts, err := eventArtifactRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(eventArtifacts) != 2 {
			t.Errorf("expected 2 event-artifacts, got %d", len(eventArtifacts))
		}

		// Delete specific event-artifact
		err = eventArtifactRepo.DeleteByEventAndArtifact(ctx, testTenant.ID, testEvent.ID, testArtifact.ID)
		if err != nil {
			t.Fatalf("failed to delete event-artifact: %v", err)
		}

		// Verify only one is deleted
		eventArtifacts, err = eventArtifactRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventArtifacts) != 1 {
			t.Errorf("expected 1 event-artifact, got %d", len(eventArtifacts))
		}

		if eventArtifacts[0].ArtifactID != artifact2.ID {
			t.Errorf("expected remaining event-artifact to have artifact_id %s, got %s", artifact2.ID, eventArtifacts[0].ArtifactID)
		}
	})
}

