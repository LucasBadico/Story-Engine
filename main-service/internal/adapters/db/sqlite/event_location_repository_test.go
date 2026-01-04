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

func TestEventLocationRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	locationRepo := NewLocationRepository(db)
	eventLocationRepo := NewEventLocationRepository(db)

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

	// Create event and location
	testEvent, err := world.NewEvent(testTenant.ID, testWorld.ID, "Test Event")
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	err = eventRepo.Create(ctx, testEvent)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	testLocation, err := world.NewLocation(testTenant.ID, testWorld.ID, "Test Location", nil)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}
	err = locationRepo.Create(ctx, testLocation)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		el := world.NewEventLocation(testEvent.ID, testLocation.ID, nil)

		err = eventLocationRepo.Create(ctx, el)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify event-location can be retrieved
		retrieved, err := eventLocationRepo.GetByID(ctx, testTenant.ID, el.ID)
		if err != nil {
			t.Fatalf("failed to retrieve event-location: %v", err)
		}

		if retrieved.EventID != testEvent.ID {
			t.Errorf("expected event_id to be %s, got %s", testEvent.ID, retrieved.EventID)
		}

		if retrieved.LocationID != testLocation.ID {
			t.Errorf("expected location_id to be %s, got %s", testLocation.ID, retrieved.LocationID)
		}
	})

	t.Run("successful creation with significance", func(t *testing.T) {
		significance := "Important location for the event"
		el := world.NewEventLocation(testEvent.ID, testLocation.ID, &significance)

		err = eventLocationRepo.Create(ctx, el)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := eventLocationRepo.GetByID(ctx, testTenant.ID, el.ID)
		if err != nil {
			t.Fatalf("failed to retrieve event-location: %v", err)
		}

		if retrieved.Significance == nil {
			t.Fatal("expected significance to be set, got nil")
		}

		if *retrieved.Significance != significance {
			t.Errorf("expected significance to be '%s', got '%s'", significance, *retrieved.Significance)
		}
	})
}

func TestEventLocationRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	locationRepo := NewLocationRepository(db)
	eventLocationRepo := NewEventLocationRepository(db)

	// Create tenant, world, event, and location
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

	testLocation, err := world.NewLocation(testTenant.ID, testWorld.ID, "Test Location", nil)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}
	err = locationRepo.Create(ctx, testLocation)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}

	t.Run("existing event-location", func(t *testing.T) {
		el := world.NewEventLocation(testEvent.ID, testLocation.ID, nil)
		err = eventLocationRepo.Create(ctx, el)
		if err != nil {
			t.Fatalf("failed to create event-location: %v", err)
		}

		retrieved, err := eventLocationRepo.GetByID(ctx, testTenant.ID, el.ID)
		if err != nil {
			t.Fatalf("failed to get event-location: %v", err)
		}

		if retrieved.ID != el.ID {
			t.Errorf("expected ID to be %s, got %s", el.ID, retrieved.ID)
		}

		if retrieved.EventID != testEvent.ID {
			t.Errorf("expected event_id to be %s, got %s", testEvent.ID, retrieved.EventID)
		}
	})

	t.Run("non-existent event-location", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := eventLocationRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent event-location")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "event_location" {
			t.Errorf("expected resource to be 'event_location', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestEventLocationRepository_ListByEvent(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	locationRepo := NewLocationRepository(db)
	eventLocationRepo := NewEventLocationRepository(db)

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
		eventLocations, err := eventLocationRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventLocations) != 0 {
			t.Errorf("expected empty list, got %d event-locations", len(eventLocations))
		}
	})

	t.Run("list with event-locations", func(t *testing.T) {
		// Create multiple locations
		loc1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 1", nil)
		loc2, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 2", nil)
		loc3, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 3", nil)
		locationRepo.Create(ctx, loc1)
		locationRepo.Create(ctx, loc2)
		locationRepo.Create(ctx, loc3)

		// Create event-locations
		el1 := world.NewEventLocation(testEvent.ID, loc1.ID, nil)
		el2 := world.NewEventLocation(testEvent.ID, loc2.ID, nil)
		el3 := world.NewEventLocation(testEvent.ID, loc3.ID, nil)
		eventLocationRepo.Create(ctx, el1)
		time.Sleep(10 * time.Millisecond)
		eventLocationRepo.Create(ctx, el2)
		time.Sleep(10 * time.Millisecond)
		eventLocationRepo.Create(ctx, el3)

		eventLocations, err := eventLocationRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventLocations) != 3 {
			t.Errorf("expected 3 event-locations, got %d", len(eventLocations))
		}

		// Verify all belong to the event
		for _, el := range eventLocations {
			if el.EventID != testEvent.ID {
				t.Errorf("expected event_id to be %s, got %s", testEvent.ID, el.EventID)
			}
		}
	})
}

func TestEventLocationRepository_ListByLocation(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	locationRepo := NewLocationRepository(db)
	eventLocationRepo := NewEventLocationRepository(db)

	// Create tenant, world, and location
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

	testLocation, err := world.NewLocation(testTenant.ID, testWorld.ID, "Test Location", nil)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}
	err = locationRepo.Create(ctx, testLocation)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}

	t.Run("empty list", func(t *testing.T) {
		eventLocations, err := eventLocationRepo.ListByLocation(ctx, testTenant.ID, testLocation.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventLocations) != 0 {
			t.Errorf("expected empty list, got %d event-locations", len(eventLocations))
		}
	})

	t.Run("list with event-locations", func(t *testing.T) {
		// Create multiple events
		event1, _ := world.NewEvent(testTenant.ID, testWorld.ID, "Event 1")
		event2, _ := world.NewEvent(testTenant.ID, testWorld.ID, "Event 2")
		event3, _ := world.NewEvent(testTenant.ID, testWorld.ID, "Event 3")
		eventRepo.Create(ctx, event1)
		eventRepo.Create(ctx, event2)
		eventRepo.Create(ctx, event3)

		// Create event-locations
		el1 := world.NewEventLocation(event1.ID, testLocation.ID, nil)
		el2 := world.NewEventLocation(event2.ID, testLocation.ID, nil)
		el3 := world.NewEventLocation(event3.ID, testLocation.ID, nil)
		eventLocationRepo.Create(ctx, el1)
		eventLocationRepo.Create(ctx, el2)
		eventLocationRepo.Create(ctx, el3)

		eventLocations, err := eventLocationRepo.ListByLocation(ctx, testTenant.ID, testLocation.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventLocations) != 3 {
			t.Errorf("expected 3 event-locations, got %d", len(eventLocations))
		}

		// Verify all belong to the location
		for _, el := range eventLocations {
			if el.LocationID != testLocation.ID {
				t.Errorf("expected location_id to be %s, got %s", testLocation.ID, el.LocationID)
			}
		}
	})
}

func TestEventLocationRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	locationRepo := NewLocationRepository(db)
	eventLocationRepo := NewEventLocationRepository(db)

	// Create tenant, world, event, and location
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

	testLocation, err := world.NewLocation(testTenant.ID, testWorld.ID, "Test Location", nil)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}
	err = locationRepo.Create(ctx, testLocation)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		el := world.NewEventLocation(testEvent.ID, testLocation.ID, nil)
		err = eventLocationRepo.Create(ctx, el)
		if err != nil {
			t.Fatalf("failed to create event-location: %v", err)
		}

		err = eventLocationRepo.Delete(ctx, testTenant.ID, el.ID)
		if err != nil {
			t.Fatalf("failed to delete event-location: %v", err)
		}

		// Verify event-location is deleted
		_, err = eventLocationRepo.GetByID(ctx, testTenant.ID, el.ID)
		if err == nil {
			t.Fatal("expected error for deleted event-location")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "event_location" {
			t.Errorf("expected resource to be 'event_location', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestEventLocationRepository_DeleteByEvent(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	locationRepo := NewLocationRepository(db)
	eventLocationRepo := NewEventLocationRepository(db)

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

	t.Run("delete all event-locations for event", func(t *testing.T) {
		// Create multiple locations
		loc1, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 1", nil)
		loc2, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 2", nil)
		loc3, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 3", nil)
		locationRepo.Create(ctx, loc1)
		locationRepo.Create(ctx, loc2)
		locationRepo.Create(ctx, loc3)

		// Create event-locations
		el1 := world.NewEventLocation(testEvent.ID, loc1.ID, nil)
		el2 := world.NewEventLocation(testEvent.ID, loc2.ID, nil)
		el3 := world.NewEventLocation(testEvent.ID, loc3.ID, nil)
		eventLocationRepo.Create(ctx, el1)
		eventLocationRepo.Create(ctx, el2)
		eventLocationRepo.Create(ctx, el3)

		// Verify they exist
		eventLocations, err := eventLocationRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(eventLocations) != 3 {
			t.Errorf("expected 3 event-locations, got %d", len(eventLocations))
		}

		// Delete all by event
		err = eventLocationRepo.DeleteByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("failed to delete event-locations: %v", err)
		}

		// Verify all are deleted
		eventLocations, err = eventLocationRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventLocations) != 0 {
			t.Errorf("expected no event-locations, got %d", len(eventLocations))
		}
	})
}

func TestEventLocationRepository_DeleteByEventAndLocation(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	locationRepo := NewLocationRepository(db)
	eventLocationRepo := NewEventLocationRepository(db)

	// Create tenant, world, event, and location
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

	testLocation, err := world.NewLocation(testTenant.ID, testWorld.ID, "Test Location", nil)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}
	err = locationRepo.Create(ctx, testLocation)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}

	t.Run("delete specific event-location", func(t *testing.T) {
		// Create multiple locations
		loc2, _ := world.NewLocation(testTenant.ID, testWorld.ID, "Location 2", nil)
		locationRepo.Create(ctx, loc2)

		// Create event-locations
		el1 := world.NewEventLocation(testEvent.ID, testLocation.ID, nil)
		el2 := world.NewEventLocation(testEvent.ID, loc2.ID, nil)
		eventLocationRepo.Create(ctx, el1)
		eventLocationRepo.Create(ctx, el2)

		// Verify they exist
		eventLocations, err := eventLocationRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(eventLocations) != 2 {
			t.Errorf("expected 2 event-locations, got %d", len(eventLocations))
		}

		// Delete specific event-location
		err = eventLocationRepo.DeleteByEventAndLocation(ctx, testTenant.ID, testEvent.ID, testLocation.ID)
		if err != nil {
			t.Fatalf("failed to delete event-location: %v", err)
		}

		// Verify only one is deleted
		eventLocations, err = eventLocationRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventLocations) != 1 {
			t.Errorf("expected 1 event-location, got %d", len(eventLocations))
		}

		if eventLocations[0].LocationID != loc2.ID {
			t.Errorf("expected remaining event-location to have location_id %s, got %s", loc2.ID, eventLocations[0].LocationID)
		}
	})
}

