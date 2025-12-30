//go:build integration

package tenant

import (
	"context"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
)

func TestCreateTenantUseCase_Execute(t *testing.T) {
	db, cleanup := postgres.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := postgres.TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tenantRepo := postgres.NewTenantRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	log := logger.New()

	useCase := NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)

	t.Run("successful creation", func(t *testing.T) {
		input := CreateTenantInput{
			Name:      "Test Tenant",
			CreatedBy: nil,
		}

		output, err := useCase.Execute(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if output.Tenant == nil {
			t.Fatal("expected tenant to be created")
		}

		if output.Tenant.Name != "Test Tenant" {
			t.Errorf("expected tenant name to be 'Test Tenant', got '%s'", output.Tenant.Name)
		}

		if output.Tenant.Status != "active" {
			t.Errorf("expected tenant status to be 'active', got '%s'", output.Tenant.Status)
		}

		// Verify tenant can be retrieved
		retrieved, err := tenantRepo.GetByID(ctx, output.Tenant.ID)
		if err != nil {
			t.Fatalf("failed to retrieve tenant: %v", err)
		}

		if retrieved.Name != "Test Tenant" {
			t.Errorf("expected retrieved tenant name to be 'Test Tenant', got '%s'", retrieved.Name)
		}
	})

	t.Run("duplicate name validation", func(t *testing.T) {
		input := CreateTenantInput{
			Name:      "Test Tenant",
			CreatedBy: nil,
		}

		_, err := useCase.Execute(ctx, input)
		if err == nil {
			t.Fatal("expected error for duplicate tenant name")
		}

		if !platformerrors.IsAlreadyExists(err) {
			t.Errorf("expected AlreadyExistsError, got %T", err)
		}
	})

	t.Run("empty name validation", func(t *testing.T) {
		input := CreateTenantInput{
			Name:      "",
			CreatedBy: nil,
		}

		_, err := useCase.Execute(ctx, input)
		if err == nil {
			t.Fatal("expected error for empty tenant name")
		}

		if !platformerrors.IsValidation(err) {
			t.Errorf("expected ValidationError, got %T", err)
		}
	})
}
