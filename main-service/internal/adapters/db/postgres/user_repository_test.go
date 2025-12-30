//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/auth"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
)

func TestUserRepository_Create(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	repo := NewUserRepository(db)

	t.Run("successful creation", func(t *testing.T) {
		user, err := auth.NewUser("test@example.com", "Test User")
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		err = repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify user can be retrieved
		retrieved, err := repo.GetByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %v", err)
		}

		if retrieved.Email != "test@example.com" {
			t.Errorf("expected email to be 'test@example.com', got '%s'", retrieved.Email)
		}

		if retrieved.Name != "Test User" {
			t.Errorf("expected name to be 'Test User', got '%s'", retrieved.Name)
		}

		if retrieved.Status != auth.UserStatusActive {
			t.Errorf("expected status to be 'active', got '%s'", retrieved.Status)
		}
	})

	t.Run("duplicate email constraint", func(t *testing.T) {
		user, err := auth.NewUser("duplicate@example.com", "Duplicate User")
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		err = repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Try to create another user with the same email
		duplicateUser, err := auth.NewUser("duplicate@example.com", "Another User")
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		err = repo.Create(ctx, duplicateUser)
		if err == nil {
			t.Fatal("expected error for duplicate email")
		}
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	repo := NewUserRepository(db)

	t.Run("existing user", func(t *testing.T) {
		user, err := auth.NewUser("getbyid@example.com", "Get By ID User")
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		err = repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if retrieved.ID != user.ID {
			t.Errorf("expected ID to be %s, got %s", user.ID, retrieved.ID)
		}
	})

	t.Run("non-existing user", func(t *testing.T) {
		nonExistentID := uuid.New()

		_, err := repo.GetByID(ctx, nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existing user")
		}

		if !platformerrors.IsNotFound(err) {
			t.Errorf("expected NotFoundError, got %T", err)
		}
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	repo := NewUserRepository(db)

	t.Run("existing user", func(t *testing.T) {
		user, err := auth.NewUser("getbyemail@example.com", "Get By Email User")
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		err = repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved, err := repo.GetByEmail(ctx, "getbyemail@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if retrieved.ID != user.ID {
			t.Errorf("expected ID to be %s, got %s", user.ID, retrieved.ID)
		}

		if retrieved.Email != "getbyemail@example.com" {
			t.Errorf("expected email to be 'getbyemail@example.com', got '%s'", retrieved.Email)
		}
	})

	t.Run("non-existing email", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		if err == nil {
			t.Fatal("expected error for non-existing email")
		}

		if !platformerrors.IsNotFound(err) {
			t.Errorf("expected NotFoundError, got %T", err)
		}
	})
}

func TestUserRepository_Update(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	repo := NewUserRepository(db)

	user, err := auth.NewUser("update@example.com", "Update User")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Update user
	user.Name = "Updated Name"
	user.Email = "updated@example.com"
	user.Suspend()

	err = repo.Update(ctx, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to retrieve user: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("expected name to be 'Updated Name', got '%s'", retrieved.Name)
	}

	if retrieved.Email != "updated@example.com" {
		t.Errorf("expected email to be 'updated@example.com', got '%s'", retrieved.Email)
	}

	if retrieved.Status != auth.UserStatusSuspended {
		t.Errorf("expected status to be 'suspended', got '%s'", retrieved.Status)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	repo := NewUserRepository(db)

	user, err := auth.NewUser("delete@example.com", "Delete User")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Delete user
	err = repo.Delete(ctx, user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, user.ID)
	if err == nil {
		t.Fatal("expected error for deleted user")
	}

	if !platformerrors.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestUserRepository_List(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Clean up tables before test
	if err := TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	repo := NewUserRepository(db)

	// Create multiple users
	users := []*auth.User{}
	for i := 0; i < 5; i++ {
		user, err := auth.NewUser(
			fmt.Sprintf("list%d@example.com", i),
			fmt.Sprintf("List User %d", i),
		)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		err = repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		users = append(users, user)
	}

	// Test pagination
	t.Run("list with limit", func(t *testing.T) {
		list, err := repo.List(ctx, 3, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(list) != 3 {
			t.Errorf("expected 3 users, got %d", len(list))
		}
	})

	t.Run("list with offset", func(t *testing.T) {
		list, err := repo.List(ctx, 2, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(list) != 2 {
			t.Errorf("expected 2 users, got %d", len(list))
		}
	})

	t.Run("list all", func(t *testing.T) {
		list, err := repo.List(ctx, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(list) < 5 {
			t.Errorf("expected at least 5 users, got %d", len(list))
		}
	})
}

