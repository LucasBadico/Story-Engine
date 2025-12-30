//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/auth"
	"github.com/story-engine/main-service/internal/core/tenant"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
)

func TestMembershipRepository_Create(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	membershipRepo := NewMembershipRepository(db)

	// Create tenant and user
	testTenant, err := tenant.NewTenant("Test Tenant", nil)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	err = tenantRepo.Create(ctx, testTenant)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	testUser, err := auth.NewUser("membership@example.com", "Membership User")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	err = userRepo.Create(ctx, testUser)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Run("successful creation", func(t *testing.T) {
		membership, err := auth.NewMembership(testTenant.ID, testUser.ID, auth.RoleEditor)
		if err != nil {
			t.Fatalf("failed to create membership: %v", err)
		}

		err = membershipRepo.Create(ctx, membership)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify membership can be retrieved
		retrieved, err := membershipRepo.GetByID(ctx, membership.ID)
		if err != nil {
			t.Fatalf("failed to retrieve membership: %v", err)
		}

		if retrieved.TenantID != testTenant.ID {
			t.Errorf("expected tenant ID to be %s, got %s", testTenant.ID, retrieved.TenantID)
		}

		if retrieved.UserID != testUser.ID {
			t.Errorf("expected user ID to be %s, got %s", testUser.ID, retrieved.UserID)
		}

		if retrieved.Role != auth.RoleEditor {
			t.Errorf("expected role to be 'editor', got '%s'", retrieved.Role)
		}

		if retrieved.Status != auth.MembershipStatusActive {
			t.Errorf("expected status to be 'active', got '%s'", retrieved.Status)
		}
	})

	t.Run("duplicate tenant-user constraint", func(t *testing.T) {
		// Create another user
		anotherUser, err := auth.NewUser("another@example.com", "Another User")
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		err = userRepo.Create(ctx, anotherUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Create first membership
		membership1, err := auth.NewMembership(testTenant.ID, anotherUser.ID, auth.RoleViewer)
		if err != nil {
			t.Fatalf("failed to create membership: %v", err)
		}
		err = membershipRepo.Create(ctx, membership1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Try to create duplicate membership
		membership2, err := auth.NewMembership(testTenant.ID, anotherUser.ID, auth.RoleAdmin)
		if err != nil {
			t.Fatalf("failed to create membership: %v", err)
		}
		err = membershipRepo.Create(ctx, membership2)
		if err == nil {
			t.Fatal("expected error for duplicate tenant-user membership")
		}
	})
}

func TestMembershipRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	membershipRepo := NewMembershipRepository(db)

	// Setup
	testTenant, _ := tenant.NewTenant("Test Tenant", nil)
	tenantRepo.Create(ctx, testTenant)
	testUser, _ := auth.NewUser("getbyid@example.com", "Get By ID User")
	userRepo.Create(ctx, testUser)

	membership, _ := auth.NewMembership(testTenant.ID, testUser.ID, auth.RoleAdmin)
	membershipRepo.Create(ctx, membership)

	t.Run("existing membership", func(t *testing.T) {
		retrieved, err := membershipRepo.GetByID(ctx, membership.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if retrieved.ID != membership.ID {
			t.Errorf("expected ID to be %s, got %s", membership.ID, retrieved.ID)
		}
	})

	t.Run("non-existing membership", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := membershipRepo.GetByID(ctx, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existing membership")
		}

		if !platformerrors.IsNotFound(err) {
			t.Errorf("expected NotFoundError, got %T", err)
		}
	})
}

func TestMembershipRepository_GetByTenantAndUser(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	membershipRepo := NewMembershipRepository(db)

	// Setup
	testTenant, _ := tenant.NewTenant("Test Tenant", nil)
	tenantRepo.Create(ctx, testTenant)
	testUser, _ := auth.NewUser("getbytenantuser@example.com", "Get By Tenant User")
	userRepo.Create(ctx, testUser)

	membership, _ := auth.NewMembership(testTenant.ID, testUser.ID, auth.RoleOwner)
	membershipRepo.Create(ctx, membership)

	t.Run("existing membership", func(t *testing.T) {
		retrieved, err := membershipRepo.GetByTenantAndUser(ctx, testTenant.ID, testUser.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if retrieved.ID != membership.ID {
			t.Errorf("expected ID to be %s, got %s", membership.ID, retrieved.ID)
		}

		if retrieved.Role != auth.RoleOwner {
			t.Errorf("expected role to be 'owner', got '%s'", retrieved.Role)
		}
	})

	t.Run("non-existing membership", func(t *testing.T) {
		nonExistentTenantID := uuid.New()
		nonExistentUserID := uuid.New()

		_, err := membershipRepo.GetByTenantAndUser(ctx, nonExistentTenantID, nonExistentUserID)
		if err == nil {
			t.Fatal("expected error for non-existing membership")
		}

		if !platformerrors.IsNotFound(err) {
			t.Errorf("expected NotFoundError, got %T", err)
		}
	})
}

func TestMembershipRepository_ListByTenant(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	membershipRepo := NewMembershipRepository(db)

	// Create tenant
	testTenant, _ := tenant.NewTenant("Test Tenant", nil)
	tenantRepo.Create(ctx, testTenant)

	// Create multiple users and memberships
	for i := 0; i < 3; i++ {
		user, _ := auth.NewUser(
			fmt.Sprintf("tenantuser%d@example.com", i),
			fmt.Sprintf("Tenant User %d", i),
		)
		userRepo.Create(ctx, user)

		membership, _ := auth.NewMembership(testTenant.ID, user.ID, auth.RoleEditor)
		membershipRepo.Create(ctx, membership)
	}

	// List memberships for tenant
	memberships, err := membershipRepo.ListByTenant(ctx, testTenant.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(memberships) != 3 {
		t.Errorf("expected 3 memberships, got %d", len(memberships))
	}

	for _, m := range memberships {
		if m.TenantID != testTenant.ID {
			t.Errorf("expected tenant ID to be %s, got %s", testTenant.ID, m.TenantID)
		}
	}
}

func TestMembershipRepository_ListByUser(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	membershipRepo := NewMembershipRepository(db)

	// Create user
	testUser, _ := auth.NewUser("multitenant@example.com", "Multi Tenant User")
	userRepo.Create(ctx, testUser)

	// Create multiple tenants and memberships
	for i := 0; i < 3; i++ {
		tenant, _ := tenant.NewTenant(fmt.Sprintf("Tenant %d", i), nil)
		tenantRepo.Create(ctx, tenant)

		membership, _ := auth.NewMembership(tenant.ID, testUser.ID, auth.RoleViewer)
		membershipRepo.Create(ctx, membership)
	}

	// List memberships for user
	memberships, err := membershipRepo.ListByUser(ctx, testUser.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(memberships) != 3 {
		t.Errorf("expected 3 memberships, got %d", len(memberships))
	}

	for _, m := range memberships {
		if m.UserID != testUser.ID {
			t.Errorf("expected user ID to be %s, got %s", testUser.ID, m.UserID)
		}
	}
}

func TestMembershipRepository_Update(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	membershipRepo := NewMembershipRepository(db)

	// Setup
	testTenant, _ := tenant.NewTenant("Test Tenant", nil)
	tenantRepo.Create(ctx, testTenant)
	testUser, _ := auth.NewUser("update@example.com", "Update User")
	userRepo.Create(ctx, testUser)

	membership, _ := auth.NewMembership(testTenant.ID, testUser.ID, auth.RoleViewer)
	membershipRepo.Create(ctx, membership)

	// Update membership
	membership.UpdateRole(auth.RoleAdmin)
	membership.Suspend()

	err := membershipRepo.Update(ctx, membership)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify update
	retrieved, err := membershipRepo.GetByID(ctx, membership.ID)
	if err != nil {
		t.Fatalf("failed to retrieve membership: %v", err)
	}

	if retrieved.Role != auth.RoleAdmin {
		t.Errorf("expected role to be 'admin', got '%s'", retrieved.Role)
	}

	if retrieved.Status != auth.MembershipStatusSuspended {
		t.Errorf("expected status to be 'suspended', got '%s'", retrieved.Status)
	}
}

func TestMembershipRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	membershipRepo := NewMembershipRepository(db)

	// Setup
	testTenant, _ := tenant.NewTenant("Test Tenant", nil)
	tenantRepo.Create(ctx, testTenant)
	testUser, _ := auth.NewUser("delete@example.com", "Delete User")
	userRepo.Create(ctx, testUser)

	membership, _ := auth.NewMembership(testTenant.ID, testUser.ID, auth.RoleEditor)
	membershipRepo.Create(ctx, membership)

	// Delete membership
	err := membershipRepo.Delete(ctx, membership.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify deletion
	_, err = membershipRepo.GetByID(ctx, membership.ID)
	if err == nil {
		t.Fatal("expected error for deleted membership")
	}

	if !platformerrors.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestMembershipRepository_CountOwnersByTenant(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	tenantRepo := NewTenantRepository(db)
	userRepo := NewUserRepository(db)
	membershipRepo := NewMembershipRepository(db)

	// Create tenant
	testTenant, _ := tenant.NewTenant("Test Tenant", nil)
	tenantRepo.Create(ctx, testTenant)

	// Create users with different roles
	owner1, _ := auth.NewUser("owner1@example.com", "Owner 1")
	userRepo.Create(ctx, owner1)
	membership1, _ := auth.NewMembership(testTenant.ID, owner1.ID, auth.RoleOwner)
	membershipRepo.Create(ctx, membership1)

	owner2, _ := auth.NewUser("owner2@example.com", "Owner 2")
	userRepo.Create(ctx, owner2)
	membership2, _ := auth.NewMembership(testTenant.ID, owner2.ID, auth.RoleOwner)
	membershipRepo.Create(ctx, membership2)

	// Create non-owner memberships
	editor, _ := auth.NewUser("editor@example.com", "Editor")
	userRepo.Create(ctx, editor)
	membership3, _ := auth.NewMembership(testTenant.ID, editor.ID, auth.RoleEditor)
	membershipRepo.Create(ctx, membership3)

	// Suspend one owner
	membership2.Suspend()
	membershipRepo.Update(ctx, membership2)

	// Count owners (should only count active owners)
	count, err := membershipRepo.CountOwnersByTenant(ctx, testTenant.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 active owner, got %d", count)
	}
}

