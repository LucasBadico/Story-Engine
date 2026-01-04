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

func TestEventRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)

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
		event, err := world.NewEvent(testTenant.ID, testWorld.ID, "Test Event")
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		err = eventRepo.Create(ctx, event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify event can be retrieved
		retrieved, err := eventRepo.GetByID(ctx, testTenant.ID, event.ID)
		if err != nil {
			t.Fatalf("failed to retrieve event: %v", err)
		}

		if retrieved.Name != "Test Event" {
			t.Errorf("expected name to be 'Test Event', got '%s'", retrieved.Name)
		}

		if retrieved.Importance != 5 {
			t.Errorf("expected importance to be 5, got %d", retrieved.Importance)
		}

		if retrieved.TenantID != testTenant.ID {
			t.Errorf("expected tenant_id to be %s, got %s", testTenant.ID, retrieved.TenantID)
		}
	})

	t.Run("successful creation with type", func(t *testing.T) {
		eventType := "Battle"
		event, err := world.NewEvent(testTenant.ID, testWorld.ID, "Event With Type")
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}
		event.UpdateType(&eventType)

		err = eventRepo.Create(ctx, event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := eventRepo.GetByID(ctx, testTenant.ID, event.ID)
		if err != nil {
			t.Fatalf("failed to retrieve event: %v", err)
		}

		if retrieved.Type == nil {
			t.Fatal("expected type to be set, got nil")
		}

		if *retrieved.Type != eventType {
			t.Errorf("expected type to be '%s', got '%s'", eventType, *retrieved.Type)
		}
	})

	t.Run("successful creation with description", func(t *testing.T) {
		description := "A test event description"
		event, err := world.NewEvent(testTenant.ID, testWorld.ID, "Event With Description")
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}
		event.UpdateDescription(&description)

		err = eventRepo.Create(ctx, event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := eventRepo.GetByID(ctx, testTenant.ID, event.ID)
		if err != nil {
			t.Fatalf("failed to retrieve event: %v", err)
		}

		if retrieved.Description == nil {
			t.Fatal("expected description to be set, got nil")
		}

		if *retrieved.Description != description {
			t.Errorf("expected description to be '%s', got '%s'", description, *retrieved.Description)
		}
	})

	t.Run("successful creation with timeline", func(t *testing.T) {
		timeline := "Year 1000, Spring"
		event, err := world.NewEvent(testTenant.ID, testWorld.ID, "Event With Timeline")
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}
		event.UpdateTimeline(&timeline)

		err = eventRepo.Create(ctx, event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := eventRepo.GetByID(ctx, testTenant.ID, event.ID)
		if err != nil {
			t.Fatalf("failed to retrieve event: %v", err)
		}

		if retrieved.Timeline == nil {
			t.Fatal("expected timeline to be set, got nil")
		}

		if *retrieved.Timeline != timeline {
			t.Errorf("expected timeline to be '%s', got '%s'", timeline, *retrieved.Timeline)
		}
	})

	t.Run("successful creation with importance", func(t *testing.T) {
		event, err := world.NewEvent(testTenant.ID, testWorld.ID, "Event With Importance")
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}
		err = event.UpdateImportance(10)
		if err != nil {
			t.Fatalf("failed to update importance: %v", err)
		}

		err = eventRepo.Create(ctx, event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := eventRepo.GetByID(ctx, testTenant.ID, event.ID)
		if err != nil {
			t.Fatalf("failed to retrieve event: %v", err)
		}

		if retrieved.Importance != 10 {
			t.Errorf("expected importance to be 10, got %d", retrieved.Importance)
		}
	})
}

func TestEventRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)

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

	t.Run("existing event", func(t *testing.T) {
		event, err := world.NewEvent(testTenant.ID, testWorld.ID, "GetByID Event")
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		err = eventRepo.Create(ctx, event)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		retrieved, err := eventRepo.GetByID(ctx, testTenant.ID, event.ID)
		if err != nil {
			t.Fatalf("failed to get event: %v", err)
		}

		if retrieved.ID != event.ID {
			t.Errorf("expected ID to be %s, got %s", event.ID, retrieved.ID)
		}

		if retrieved.Name != "GetByID Event" {
			t.Errorf("expected name to be 'GetByID Event', got '%s'", retrieved.Name)
		}
	})

	t.Run("non-existent event", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := eventRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent event")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "event" {
			t.Errorf("expected resource to be 'event', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestEventRepository_ListByWorld(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)

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
		events, err := eventRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(events) != 0 {
			t.Errorf("expected empty list, got %d events", len(events))
		}
	})

	t.Run("list with events", func(t *testing.T) {
		// Create multiple events
		eventNames := []string{"Event A", "Event B", "Event C"}
		createdEvents := make([]*world.Event, 0, len(eventNames))

		for _, name := range eventNames {
			event, err := world.NewEvent(testTenant.ID, testWorld.ID, name)
			if err != nil {
				t.Fatalf("failed to create event: %v", err)
			}
			err = eventRepo.Create(ctx, event)
			if err != nil {
				t.Fatalf("failed to create event: %v", err)
			}
			createdEvents = append(createdEvents, event)
			time.Sleep(10 * time.Millisecond)
		}

		events, err := eventRepo.ListByWorld(ctx, testTenant.ID, testWorld.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(events) != len(eventNames) {
			t.Errorf("expected %d events, got %d", len(eventNames), len(events))
		}

		// Verify all belong to the world
		for _, event := range events {
			if event.WorldID != testWorld.ID {
				t.Errorf("expected world_id to be %s, got %s", testWorld.ID, event.WorldID)
			}
		}
	})
}

func TestEventRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)

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
		event, err := world.NewEvent(testTenant.ID, testWorld.ID, "Update Event")
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		err = eventRepo.Create(ctx, event)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		// Update name
		err = event.UpdateName("Updated Name")
		if err != nil {
			t.Fatalf("failed to update name: %v", err)
		}

		// Update type
		eventType := "Ceremony"
		event.UpdateType(&eventType)

		// Update description
		description := "Updated Description"
		event.UpdateDescription(&description)

		// Update timeline
		timeline := "Year 2000, Winter"
		event.UpdateTimeline(&timeline)

		// Update importance
		err = event.UpdateImportance(8)
		if err != nil {
			t.Fatalf("failed to update importance: %v", err)
		}

		err = eventRepo.Update(ctx, event)
		if err != nil {
			t.Fatalf("failed to update event: %v", err)
		}

		// Verify update
		retrieved, err := eventRepo.GetByID(ctx, testTenant.ID, event.ID)
		if err != nil {
			t.Fatalf("failed to get event: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("expected name to be 'Updated Name', got '%s'", retrieved.Name)
		}

		if retrieved.Type == nil || *retrieved.Type != eventType {
			t.Errorf("expected type to be '%s', got %v", eventType, retrieved.Type)
		}

		if retrieved.Description == nil || *retrieved.Description != description {
			t.Errorf("expected description to be '%s', got %v", description, retrieved.Description)
		}

		if retrieved.Timeline == nil || *retrieved.Timeline != timeline {
			t.Errorf("expected timeline to be '%s', got %v", timeline, retrieved.Timeline)
		}

		if retrieved.Importance != 8 {
			t.Errorf("expected importance to be 8, got %d", retrieved.Importance)
		}
	})

	t.Run("update nullable fields to nil", func(t *testing.T) {
		eventType := "Initial Type"
		description := "Initial Description"
		timeline := "Initial Timeline"
		event, err := world.NewEvent(testTenant.ID, testWorld.ID, "Update Nullable Event")
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}
		event.UpdateType(&eventType)
		event.UpdateDescription(&description)
		event.UpdateTimeline(&timeline)
		err = eventRepo.Create(ctx, event)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		// Clear nullable fields
		event.UpdateType(nil)
		event.UpdateDescription(nil)
		event.UpdateTimeline(nil)

		err = eventRepo.Update(ctx, event)
		if err != nil {
			t.Fatalf("failed to update event: %v", err)
		}

		retrieved, err := eventRepo.GetByID(ctx, testTenant.ID, event.ID)
		if err != nil {
			t.Fatalf("failed to get event: %v", err)
		}

		if retrieved.Type != nil {
			t.Error("expected type to be nil, got non-nil")
		}

		if retrieved.Description != nil {
			t.Error("expected description to be nil, got non-nil")
		}

		if retrieved.Timeline != nil {
			t.Error("expected timeline to be nil, got non-nil")
		}
	})
}

func TestEventRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)

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
		event, err := world.NewEvent(testTenant.ID, testWorld.ID, "Delete Event")
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		err = eventRepo.Create(ctx, event)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		err = eventRepo.Delete(ctx, testTenant.ID, event.ID)
		if err != nil {
			t.Fatalf("failed to delete event: %v", err)
		}

		// Verify event is deleted
		_, err = eventRepo.GetByID(ctx, testTenant.ID, event.ID)
		if err == nil {
			t.Fatal("expected error for deleted event")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "event" {
			t.Errorf("expected resource to be 'event', got '%s'", notFoundErr.Resource)
		}
	})
}

