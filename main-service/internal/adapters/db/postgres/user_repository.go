package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/story-engine/main-service/internal/core/auth"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.UserRepository = (*UserRepository)(nil)

// UserRepository implements the user repository interface
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, u *auth.User) error {
	query := `
		INSERT INTO users (id, email, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		u.ID, u.Email, u.Name, string(u.Status), u.CreatedAt, u.UpdatedAt)
	return err
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	query := `
		SELECT id, email, name, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u auth.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Name, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "user",
				ID:       id.String(),
			}
		}
		return nil, err
	}

	return &u, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
		SELECT id, email, name, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var u auth.User

	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Name, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &platformerrors.NotFoundError{
				Resource: "user",
				ID:       email, // Using email as identifier for GetByEmail
			}
		}
		return nil, err
	}

	return &u, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, u *auth.User) error {
	query := `
		UPDATE users
		SET email = $2, name = $3, status = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, u.ID, u.Email, u.Name, string(u.Status), u.UpdatedAt)
	return err
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// List lists users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*auth.User, error) {
	query := `
		SELECT id, email, name, status, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*auth.User
	for rows.Next() {
		var u auth.User

		err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Status, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}

		users = append(users, &u)
	}

	return users, rows.Err()
}

