//go:build integration

package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/relation"
	"github.com/story-engine/main-service/internal/core/tenant"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

func TestEntityRelationRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	relationRepo := NewEntityRelationRepository(db)

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

	// Create characters
	char1, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, char1)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	char2, err := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	err = characterRepo.Create(ctx, char2)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		rel, err := relation.NewEntityRelation(
			testTenant.ID,
			testWorld.ID,
			"character",
			char1.ID,
			"character",
			char2.ID,
			"friend",
		)
		if err != nil {
			t.Fatalf("failed to create relation: %v", err)
		}

		err = relationRepo.Create(ctx, rel)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify relation can be retrieved
		retrieved, err := relationRepo.GetByID(ctx, testTenant.ID, rel.ID)
		if err != nil {
			t.Fatalf("failed to retrieve relation: %v", err)
		}

		if retrieved.SourceType != "character" {
			t.Errorf("expected source_type to be 'character', got '%s'", retrieved.SourceType)
		}

		if retrieved.TargetType != "character" {
			t.Errorf("expected target_type to be 'character', got '%s'", retrieved.TargetType)
		}

		if retrieved.RelationType != "friend" {
			t.Errorf("expected relation_type to be 'friend', got '%s'", retrieved.RelationType)
		}
	})

	t.Run("create with attributes", func(t *testing.T) {
		rel, err := relation.NewEntityRelation(
			testTenant.ID,
			testWorld.ID,
			"character",
			char1.ID,
			"character",
			char2.ID,
			"ally",
		)
		if err != nil {
			t.Fatalf("failed to create relation: %v", err)
		}
		rel.Attributes = map[string]interface{}{
			"description": "They are allies",
			"strength":    5,
		}

		err = relationRepo.Create(ctx, rel)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := relationRepo.GetByID(ctx, testTenant.ID, rel.ID)
		if err != nil {
			t.Fatalf("failed to retrieve relation: %v", err)
		}

		if retrieved.Attributes == nil {
			t.Error("expected attributes to be set")
		} else {
			if desc, ok := retrieved.Attributes["description"].(string); !ok || desc != "They are allies" {
				t.Errorf("expected description 'They are allies', got %v", retrieved.Attributes["description"])
			}
		}
	})
}

func TestEntityRelationRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	relationRepo := NewEntityRelationRepository(db)

	// Setup
	testTenant, _ := tenant.NewTenant("test-tenant", nil)
	tenantRepo.Create(ctx, testTenant)
	testWorld, _ := world.NewWorld(testTenant.ID, "Test World", false)
	worldRepo.Create(ctx, testWorld)
	char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
	characterRepo.Create(ctx, char1)
	char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	characterRepo.Create(ctx, char2)

	rel, _ := relation.NewEntityRelation(
		testTenant.ID,
		testWorld.ID,
		"character",
		char1.ID,
		"character",
		char2.ID,
		"friend",
	)
	relationRepo.Create(ctx, rel)

	t.Run("existing relation", func(t *testing.T) {
		retrieved, err := relationRepo.GetByID(ctx, testTenant.ID, rel.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if retrieved.ID != rel.ID {
			t.Errorf("expected ID %s, got %s", rel.ID, retrieved.ID)
		}
	})

	t.Run("non-existing relation", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := relationRepo.GetByID(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Error("expected error for non-existing relation")
		}
		if _, ok := err.(*platformerrors.NotFoundError); !ok {
			t.Errorf("expected NotFoundError, got %T", err)
		}
	})
}

func TestEntityRelationRepository_ListBySource(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	relationRepo := NewEntityRelationRepository(db)

	// Setup
	testTenant, _ := tenant.NewTenant("test-tenant", nil)
	tenantRepo.Create(ctx, testTenant)
	testWorld, _ := world.NewWorld(testTenant.ID, "Test World", false)
	worldRepo.Create(ctx, testWorld)
	char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
	characterRepo.Create(ctx, char1)
	char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	characterRepo.Create(ctx, char2)
	char3, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 3")
	characterRepo.Create(ctx, char3)

	// Create relations
	rel1, _ := relation.NewEntityRelation(
		testTenant.ID,
		testWorld.ID,
		"character",
		char1.ID,
		"character",
		char2.ID,
		"friend",
	)
	relationRepo.Create(ctx, rel1)

	rel2, _ := relation.NewEntityRelation(
		testTenant.ID,
		testWorld.ID,
		"character",
		char1.ID,
		"character",
		char3.ID,
		"ally",
	)
	relationRepo.Create(ctx, rel2)

	t.Run("list by source", func(t *testing.T) {
		result, err := relationRepo.ListBySource(ctx, testTenant.ID, "character", char1.ID, repositories.ListOptions{
			Limit: 100,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Items) != 2 {
			t.Errorf("expected 2 relations, got %d", len(result.Items))
		}
	})

	t.Run("list by source with filter", func(t *testing.T) {
		relationType := "friend"
		result, err := relationRepo.ListBySource(ctx, testTenant.ID, "character", char1.ID, repositories.ListOptions{
			Limit:        100,
			RelationType: &relationType,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Items) != 1 {
			t.Errorf("expected 1 relation, got %d", len(result.Items))
		}

		if result.Items[0].RelationType != "friend" {
			t.Errorf("expected relation_type 'friend', got '%s'", result.Items[0].RelationType)
		}
	})
}

func TestEntityRelationRepository_ListByTarget(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	relationRepo := NewEntityRelationRepository(db)

	// Setup
	testTenant, _ := tenant.NewTenant("test-tenant", nil)
	tenantRepo.Create(ctx, testTenant)
	testWorld, _ := world.NewWorld(testTenant.ID, "Test World", false)
	worldRepo.Create(ctx, testWorld)
	char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
	characterRepo.Create(ctx, char1)
	char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	characterRepo.Create(ctx, char2)

	rel, _ := relation.NewEntityRelation(
		testTenant.ID,
		testWorld.ID,
		"character",
		char1.ID,
		"character",
		char2.ID,
		"friend",
	)
	relationRepo.Create(ctx, rel)

	t.Run("list by target", func(t *testing.T) {
		result, err := relationRepo.ListByTarget(ctx, testTenant.ID, "character", char2.ID, repositories.ListOptions{
			Limit: 100,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Items) != 1 {
			t.Errorf("expected 1 relation, got %d", len(result.Items))
		}

		if result.Items[0].TargetID != char2.ID {
			t.Errorf("expected target_id to be %s, got %v", char2.ID, result.Items[0].TargetID)
		}
	})
}

func TestEntityRelationRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	relationRepo := NewEntityRelationRepository(db)

	// Setup
	testTenant, _ := tenant.NewTenant("test-tenant", nil)
	tenantRepo.Create(ctx, testTenant)
	testWorld, _ := world.NewWorld(testTenant.ID, "Test World", false)
	worldRepo.Create(ctx, testWorld)
	char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
	characterRepo.Create(ctx, char1)
	char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	characterRepo.Create(ctx, char2)

	rel, _ := relation.NewEntityRelation(
		testTenant.ID,
		testWorld.ID,
		"character",
		char1.ID,
		"character",
		char2.ID,
		"friend",
	)
	relationRepo.Create(ctx, rel)

	t.Run("update relation type", func(t *testing.T) {
		rel.RelationType = "ally"
		rel.UpdatedAt = time.Now()

		err := relationRepo.Update(ctx, rel)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := relationRepo.GetByID(ctx, testTenant.ID, rel.ID)
		if err != nil {
			t.Fatalf("failed to retrieve relation: %v", err)
		}

		if retrieved.RelationType != "ally" {
			t.Errorf("expected relation_type 'ally', got '%s'", retrieved.RelationType)
		}
	})
}

func TestEntityRelationRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	relationRepo := NewEntityRelationRepository(db)

	// Setup
	testTenant, _ := tenant.NewTenant("test-tenant", nil)
	tenantRepo.Create(ctx, testTenant)
	testWorld, _ := world.NewWorld(testTenant.ID, "Test World", false)
	worldRepo.Create(ctx, testWorld)
	char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
	characterRepo.Create(ctx, char1)
	char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	characterRepo.Create(ctx, char2)

	rel, _ := relation.NewEntityRelation(
		testTenant.ID,
		testWorld.ID,
		"character",
		char1.ID,
		"character",
		char2.ID,
		"friend",
	)
	relationRepo.Create(ctx, rel)

	t.Run("delete existing relation", func(t *testing.T) {
		err := relationRepo.Delete(ctx, testTenant.ID, rel.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify relation is deleted
		_, err = relationRepo.GetByID(ctx, testTenant.ID, rel.ID)
		if err == nil {
			t.Error("expected error when getting deleted relation")
		}
		if _, ok := err.(*platformerrors.NotFoundError); !ok {
			t.Errorf("expected NotFoundError, got %T", err)
		}
	})

	t.Run("delete non-existing relation", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := relationRepo.Delete(ctx, testTenant.ID, nonExistentID)
		if err == nil {
			t.Error("expected error for non-existing relation")
		}
		if _, ok := err.(*platformerrors.NotFoundError); !ok {
			t.Errorf("expected NotFoundError, got %T", err)
		}
	})
}

func TestEntityRelationRepository_CreateWithMirror(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	tenantRepo := NewTenantRepository(db)
	worldRepo := NewWorldRepository(db)
	characterRepo := NewCharacterRepository(db)
	relationRepo := NewEntityRelationRepository(db)

	// Setup
	testTenant, _ := tenant.NewTenant("test-tenant", nil)
	tenantRepo.Create(ctx, testTenant)
	testWorld, _ := world.NewWorld(testTenant.ID, "Test World", false)
	worldRepo.Create(ctx, testWorld)
	char1, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 1")
	characterRepo.Create(ctx, char1)
	char2, _ := world.NewCharacter(testTenant.ID, testWorld.ID, "Character 2")
	characterRepo.Create(ctx, char2)

	t.Run("create with mirror", func(t *testing.T) {
		rel, err := relation.NewEntityRelation(
			testTenant.ID,
			testWorld.ID,
			"character",
			char1.ID,
			"character",
			char2.ID,
			"parent_of",
		)
		if err != nil {
			t.Fatalf("failed to create relation: %v", err)
		}

		mirror, err := relationRepo.CreateWithMirror(ctx, rel)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mirror == nil {
			t.Error("expected mirror relation to be created")
		}

		// Verify mirror relation
		if mirror.RelationType != "child_of" {
			t.Errorf("expected mirror relation_type 'child_of', got '%s'", mirror.RelationType)
		}

		if mirror.SourceID != char2.ID {
			t.Errorf("expected mirror source_id to be %s, got %v", char2.ID, mirror.SourceID)
		}

		if mirror.TargetID != char1.ID {
			t.Errorf("expected mirror target_id to be %s, got %v", char1.ID, mirror.TargetID)
		}

		// Verify original relation has mirror_id
		retrieved, err := relationRepo.GetByID(ctx, testTenant.ID, rel.ID)
		if err != nil {
			t.Fatalf("failed to retrieve relation: %v", err)
		}

		if retrieved.MirrorID == nil || *retrieved.MirrorID != mirror.ID {
			t.Errorf("expected mirror_id to be %s, got %v", mirror.ID, retrieved.MirrorID)
		}
	})
}
