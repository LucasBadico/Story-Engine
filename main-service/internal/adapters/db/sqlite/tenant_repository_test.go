//go:build integration

package sqlite

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/tenant"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
)

func TestTenantRepository_Create(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewTenantRepository(db)

	t.Run("successful creation", func(t *testing.T) {
		tenant, err := tenant.NewTenant("test-tenant", nil)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		err = repo.Create(ctx, tenant)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify tenant can be retrieved
		retrieved, err := repo.GetByID(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("failed to retrieve tenant: %v", err)
		}

		if retrieved.Name != "test-tenant" {
			t.Errorf("expected name to be 'test-tenant', got '%s'", retrieved.Name)
		}

		if retrieved.Status != "active" {
			t.Errorf("expected status to be 'active', got '%s'", retrieved.Status)
		}
	})

	t.Run("duplicate name constraint", func(t *testing.T) {
		tenant1, err := tenant.NewTenant("duplicate-tenant", nil)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		err = repo.Create(ctx, tenant1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Try to create another tenant with the same name
		tenant2, err := tenant.NewTenant("duplicate-tenant", nil)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		err = repo.Create(ctx, tenant2)
		if err == nil {
			t.Fatal("expected error for duplicate name")
		}

		// Check if it's the right error type
		if _, ok := err.(*platformerrors.AlreadyExistsError); !ok {
			t.Errorf("expected AlreadyExistsError, got %T: %v", err, err)
		}
	})
}

func TestTenantRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewTenantRepository(db)

	t.Run("existing tenant", func(t *testing.T) {
		tenant, err := tenant.NewTenant("getbyid-tenant", nil)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		err = repo.Create(ctx, tenant)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("failed to get tenant: %v", err)
		}

		if retrieved.ID != tenant.ID {
			t.Errorf("expected ID to be %s, got %s", tenant.ID, retrieved.ID)
		}

		if retrieved.Name != "getbyid-tenant" {
			t.Errorf("expected name to be 'getbyid-tenant', got '%s'", retrieved.Name)
		}
	})

	t.Run("non-existent tenant", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := repo.GetByID(ctx, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent tenant")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "tenant" {
			t.Errorf("expected resource to be 'tenant', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestTenantRepository_GetByName(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewTenantRepository(db)

	t.Run("existing tenant", func(t *testing.T) {
		tenant, err := tenant.NewTenant("getbyname-tenant", nil)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		err = repo.Create(ctx, tenant)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		retrieved, err := repo.GetByName(ctx, "getbyname-tenant")
		if err != nil {
			t.Fatalf("failed to get tenant: %v", err)
		}

		if retrieved.ID != tenant.ID {
			t.Errorf("expected ID to be %s, got %s", tenant.ID, retrieved.ID)
		}

		if retrieved.Name != "getbyname-tenant" {
			t.Errorf("expected name to be 'getbyname-tenant', got '%s'", retrieved.Name)
		}
	})

	t.Run("non-existent tenant", func(t *testing.T) {
		_, err := repo.GetByName(ctx, "non-existent-tenant")
		if err == nil {
			t.Fatal("expected error for non-existent tenant")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "tenant" {
			t.Errorf("expected resource to be 'tenant', got '%s'", notFoundErr.Resource)
		}
	})
}

func TestTenantRepository_Update(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewTenantRepository(db)

	t.Run("successful update", func(t *testing.T) {
		tenant, err := tenant.NewTenant("update-tenant", nil)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		err = repo.Create(ctx, tenant)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		// Update name
		err = tenant.UpdateName("updated-name")
		if err != nil {
			t.Fatalf("failed to update tenant name: %v", err)
		}

		err = repo.Update(ctx, tenant)
		if err != nil {
			t.Fatalf("failed to update tenant: %v", err)
		}

		// Verify update
		retrieved, err := repo.GetByID(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("failed to get tenant: %v", err)
		}

		if retrieved.Name != "updated-name" {
			t.Errorf("expected name to be 'updated-name', got '%s'", retrieved.Name)
		}
	})
}

func TestTenantRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestSQLiteDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewTenantRepository(db)

	t.Run("successful delete", func(t *testing.T) {
		tenant, err := tenant.NewTenant("delete-tenant", nil)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		err = repo.Create(ctx, tenant)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		err = repo.Delete(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("failed to delete tenant: %v", err)
		}

		// Verify tenant is deleted
		_, err = repo.GetByID(ctx, tenant.ID)
		if err == nil {
			t.Fatal("expected error for deleted tenant")
		}

		notFoundErr, ok := err.(*platformerrors.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}

		if notFoundErr.Resource != "tenant" {
			t.Errorf("expected resource to be 'tenant', got '%s'", notFoundErr.Resource)
		}
	})
}

