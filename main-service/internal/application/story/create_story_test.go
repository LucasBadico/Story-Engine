//go:build integration

package story

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestCreateStoryUseCase_Execute(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := postgres.TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	// Create a tenant first
	tenantRepo := postgres.NewTenantRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	log := logger.New()

	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	tenantOutput, err := createTenantUseCase.Execute(ctx, tenant.CreateTenantInput{
		Name:      "Test Tenant",
		CreatedBy: nil,
	})
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	storyRepo := postgres.NewStoryRepository(db)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	createStoryUseCase := NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, nil, log)

	t.Run("successful creation", func(t *testing.T) {
		input := CreateStoryInput{
			TenantID:       tenantOutput.Tenant.ID,
			Title:          "Test Story",
			CreatedByUserID: nil,
		}

		output, err := createStoryUseCase.Execute(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if output.Story == nil {
			t.Fatal("expected story to be created")
		}

		if output.Story.Title != "Test Story" {
			t.Errorf("expected story title to be 'Test Story', got '%s'", output.Story.Title)
		}

		if output.Story.VersionNumber != 1 {
			t.Errorf("expected version number to be 1, got %d", output.Story.VersionNumber)
		}

		if output.Story.RootStoryID != output.Story.ID {
			t.Errorf("expected root_story_id to equal story id")
		}

		if output.Story.PreviousStoryID != nil {
			t.Errorf("expected previous_story_id to be nil for first version")
		}

		// Verify story can be retrieved
		retrieved, err := storyRepo.GetByID(ctx, tenantOutput.Tenant.ID, output.Story.ID)
		if err != nil {
			t.Fatalf("failed to retrieve story: %v", err)
		}

		if retrieved.Title != "Test Story" {
			t.Errorf("expected retrieved story title to be 'Test Story', got '%s'", retrieved.Title)
		}
	})

	t.Run("multi-tenant isolation", func(t *testing.T) {
		// Create another tenant
		tenant2Output, err := createTenantUseCase.Execute(ctx, tenant.CreateTenantInput{
			Name:      "Test Tenant 2",
			CreatedBy: nil,
		})
		if err != nil {
			t.Fatalf("failed to create second tenant: %v", err)
		}

		// Create story for second tenant
		input := CreateStoryInput{
			TenantID:       tenant2Output.Tenant.ID,
			Title:          "Test Story 2",
			CreatedByUserID: nil,
		}

		_, err = createStoryUseCase.Execute(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify stories are isolated by tenant
		stories1, err := storyRepo.ListByTenant(ctx, tenantOutput.Tenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("failed to list stories for tenant 1: %v", err)
		}

		if len(stories1) != 1 {
			t.Errorf("expected 1 story for tenant 1, got %d", len(stories1))
		}

		stories2, err := storyRepo.ListByTenant(ctx, tenant2Output.Tenant.ID, 10, 0)
		if err != nil {
			t.Fatalf("failed to list stories for tenant 2: %v", err)
		}

		if len(stories2) != 1 {
			t.Errorf("expected 1 story for tenant 2, got %d", len(stories2))
		}

		if stories1[0].ID == stories2[0].ID {
			t.Error("stories should have different IDs")
		}
	})

	t.Run("invalid tenant", func(t *testing.T) {
		invalidTenantID := uuid.New()
		input := CreateStoryInput{
			TenantID:       invalidTenantID,
			Title:          "Test Story",
			CreatedByUserID: nil,
		}

		_, err := createStoryUseCase.Execute(ctx, input)
		if err == nil {
			t.Fatal("expected error for invalid tenant")
		}

		if !platformerrors.IsNotFound(err) {
			t.Errorf("expected NotFoundError, got %T", err)
		}
	})

	t.Run("empty title validation", func(t *testing.T) {
		input := CreateStoryInput{
			TenantID:       tenantOutput.Tenant.ID,
			Title:          "",
			CreatedByUserID: nil,
		}

		_, err := createStoryUseCase.Execute(ctx, input)
		if err == nil {
			t.Fatal("expected error for empty title")
		}

		if !platformerrors.IsValidation(err) {
			t.Errorf("expected ValidationError, got %T", err)
		}
	})
}

