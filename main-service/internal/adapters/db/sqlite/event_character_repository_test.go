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

func TestEventCharacterRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	characterRepo := NewCharacterRepository(db)
	eventCharacterRepo := NewEventCharacterRepository(db)

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

	// Create event and character
	testEvent, err := world.NewEvent(testTenant.ID, testWorld.ID, "Test Event")
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	err = eventRepo.Create(ctx, testEvent)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		ec := world.NewEventCharacter(testEvent.ID, testCharacter.ID, nil)

		err = eventCharacterRepo.Create(ctx, ec)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify event-character can be retrieved
		retrieved, err := eventCharacterRepo.GetByID(ctx, testTenant.ID, ec.ID)
		if err != nil {
			t.Fatalf("failed to retrieve event-character: %v", err)
		}

		if retrieved.EventID != testEvent.ID {
			t.Errorf("expected event_id to be %s, got %s", testEvent.ID, retrieved.EventID)
		}

		if retrieved.CharacterID != testCharacter.ID {
			t.Errorf("expected character_id to be %s, got %s", testCharacter.ID, retrieved.CharacterID)
		}
	})

	t.Run("successful creation with role", func(t *testing.T) {
		role := "Protagonist"
		ec := world.NewEventCharacter(testEvent.ID, testCharacter.ID, &role)

		err = eventCharacterRepo.Create(ctx, ec)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := eventCharacterRepo.GetByID(ctx, testTenant.ID, ec.ID)
		if err != nil {
			t.Fatalf("failed to retrieve event-character: %v", err)
		}

		if retrieved.Role == nil {
			t.Fatal("expected role to be set, got nil")
		}

		if *retrieved.Role != role {
			t.Errorf("expected role to be '%s', got '%s'", role, *retrieved.Role)
		}
	})
}

func TestEventCharacterRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	characterRepo := NewCharacterRepository(db)
	eventCharacterRepo := NewEventCharacterRepository(db)

	// Create tenant, world, event, and character
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

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("existing event-character", func(t *testing.T) {
		ec := world.NewEventCharacter(testEvent.ID, testCharacter.ID, nil)
		err = eventCharacterRepo.Create(ctx, ec)
		if err != nil {
			t.Fatalf("failed to create event-character: %v", err)
		}

		retrieved, err := eventCharacterRepo.GetByID(ctx, testTenant.ID, ec.ID)
		if err != nil {
			t.Fatalf("failed to get event-character: %v", err)
		}

		if retrieved.ID != ec.ID {
			t.Errorf("expected ID to be %s, got %s", ec.ID, retrieved.ID)
		}

		if retrieved.EventID != testEvent.ID {
			t.Errorf("expected event_id to be %s, got %s", testEvent.ID, retrieved.EventID)
		}
	})

	t.Run("non-existent event-character", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := eventCharacterRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent event-character")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "event_character" {
			t.Errorf("expected resource to be 'event_character', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestEventCharacterRepository_ListByEvent(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	characterRepo := NewCharacterRepository(db)
	eventCharacterRepo := NewEventCharacterRepository(db)

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
		eventCharacters, err := eventCharacterRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventCharacters) != 0 {
			t.Errorf("expected empty list, got %d event-characters", len(eventCharacters))
		}
	})

	t.Run("list with event-characters", func(t *testing.T) {
		// Create multiple characters
		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		char3, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 3")
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)
		characterRepo.Create(ctx, char3)

		// Create event-characters
		ec1 := world.NewEventCharacter(testEvent.ID, char1.ID, nil)
		ec2 := world.NewEventCharacter(testEvent.ID, char2.ID, nil)
		ec3 := world.NewEventCharacter(testEvent.ID, char3.ID, nil)
		eventCharacterRepo.Create(ctx, ec1)
		time.Sleep(10 * time.Millisecond)
		eventCharacterRepo.Create(ctx, ec2)
		time.Sleep(10 * time.Millisecond)
		eventCharacterRepo.Create(ctx, ec3)

		eventCharacters, err := eventCharacterRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventCharacters) != 3 {
			t.Errorf("expected 3 event-characters, got %d", len(eventCharacters))
		}

		// Verify all belong to the event
		for _, ec := range eventCharacters {
			if ec.EventID != testEvent.ID {
				t.Errorf("expected event_id to be %s, got %s", testEvent.ID, ec.EventID)
			}
		}
	})
}

func TestEventCharacterRepository_ListByCharacter(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	characterRepo := NewCharacterRepository(db)
	eventCharacterRepo := NewEventCharacterRepository(db)

	// Create tenant, world, and character
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

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("empty list", func(t *testing.T) {
		eventCharacters, err := eventCharacterRepo.ListByCharacter(ctx, testTenant.ID, testCharacter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventCharacters) != 0 {
			t.Errorf("expected empty list, got %d event-characters", len(eventCharacters))
		}
	})

	t.Run("list with event-characters", func(t *testing.T) {
		// Create multiple events
		event1, _ := world.NewEvent(testTenant.ID, testWorld.ID, "Event 1")
		event2, _ := world.NewEvent(testTenant.ID, testWorld.ID, "Event 2")
		event3, _ := world.NewEvent(testTenant.ID, testWorld.ID, "Event 3")
		eventRepo.Create(ctx, event1)
		eventRepo.Create(ctx, event2)
		eventRepo.Create(ctx, event3)

		// Create event-characters
		ec1 := world.NewEventCharacter(event1.ID, testCharacter.ID, nil)
		ec2 := world.NewEventCharacter(event2.ID, testCharacter.ID, nil)
		ec3 := world.NewEventCharacter(event3.ID, testCharacter.ID, nil)
		eventCharacterRepo.Create(ctx, ec1)
		eventCharacterRepo.Create(ctx, ec2)
		eventCharacterRepo.Create(ctx, ec3)

		eventCharacters, err := eventCharacterRepo.ListByCharacter(ctx, testTenant.ID, testCharacter.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventCharacters) != 3 {
			t.Errorf("expected 3 event-characters, got %d", len(eventCharacters))
		}

		// Verify all belong to the character
		for _, ec := range eventCharacters {
			if ec.CharacterID != testCharacter.ID {
				t.Errorf("expected character_id to be %s, got %s", testCharacter.ID, ec.CharacterID)
			}
		}
	})
}

func TestEventCharacterRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	characterRepo := NewCharacterRepository(db)
	eventCharacterRepo := NewEventCharacterRepository(db)

	// Create tenant, world, event, and character
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

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		ec := world.NewEventCharacter(testEvent.ID, testCharacter.ID, nil)
		err = eventCharacterRepo.Create(ctx, ec)
		if err != nil {
			t.Fatalf("failed to create event-character: %v", err)
		}

		err = eventCharacterRepo.Delete(ctx, testTenant.ID, ec.ID)
		if err != nil {
			t.Fatalf("failed to delete event-character: %v", err)
		}

		// Verify event-character is deleted
		_, err = eventCharacterRepo.GetByID(ctx, testTenant.ID, ec.ID)
		if err == nil {
			t.Fatal("expected error for deleted event-character")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "event_character" {
			t.Errorf("expected resource to be 'event_character', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestEventCharacterRepository_DeleteByEvent(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	characterRepo := NewCharacterRepository(db)
	eventCharacterRepo := NewEventCharacterRepository(db)

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

	t.Run("delete all event-characters for event", func(t *testing.T) {
		// Create multiple characters
		char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		char3, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 3")
		characterRepo.Create(ctx, char1)
		characterRepo.Create(ctx, char2)
		characterRepo.Create(ctx, char3)

		// Create event-characters
		ec1 := world.NewEventCharacter(testEvent.ID, char1.ID, nil)
		ec2 := world.NewEventCharacter(testEvent.ID, char2.ID, nil)
		ec3 := world.NewEventCharacter(testEvent.ID, char3.ID, nil)
		eventCharacterRepo.Create(ctx, ec1)
		eventCharacterRepo.Create(ctx, ec2)
		eventCharacterRepo.Create(ctx, ec3)

		// Verify they exist
		eventCharacters, err := eventCharacterRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(eventCharacters) != 3 {
			t.Errorf("expected 3 event-characters, got %d", len(eventCharacters))
		}

		// Delete all by event
		err = eventCharacterRepo.DeleteByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("failed to delete event-characters: %v", err)
		}

		// Verify all are deleted
		eventCharacters, err = eventCharacterRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventCharacters) != 0 {
			t.Errorf("expected no event-characters, got %d", len(eventCharacters))
		}
	})
}

func TestEventCharacterRepository_DeleteByEventAndCharacter(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	eventRepo := NewEventRepository(db)
	characterRepo := NewCharacterRepository(db)
	eventCharacterRepo := NewEventCharacterRepository(db)

	// Create tenant, world, event, and character
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

	testCharacter, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Test Character")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, testCharacter)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("delete specific event-character", func(t *testing.T) {
		// Create multiple characters
		char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
		characterRepo.Create(ctx, char2)

		// Create event-characters
		ec1 := world.NewEventCharacter(testEvent.ID, testCharacter.ID, nil)
		ec2 := world.NewEventCharacter(testEvent.ID, char2.ID, nil)
		eventCharacterRepo.Create(ctx, ec1)
		eventCharacterRepo.Create(ctx, ec2)

		// Verify they exist
		eventCharacters, err := eventCharacterRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(eventCharacters) != 2 {
			t.Errorf("expected 2 event-characters, got %d", len(eventCharacters))
		}

		// Delete specific event-character
		err = eventCharacterRepo.DeleteByEventAndCharacter(ctx, testTenant.ID, testEvent.ID, testCharacter.ID)
		if err != nil {
			t.Fatalf("failed to delete event-character: %v", err)
		}

		// Verify only one is deleted
		eventCharacters, err = eventCharacterRepo.ListByEvent(ctx, testTenant.ID, testEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(eventCharacters) != 1 {
			t.Errorf("expected 1 event-character, got %d", len(eventCharacters))
		}

		if eventCharacters[0].CharacterID != char2.ID {
			t.Errorf("expected remaining event-character to have character_id %s, got %s", char2.ID, eventCharacters[0].CharacterID)
		}
	})
}

