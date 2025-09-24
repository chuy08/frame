package db

import (
	"context"
	"fmt"

	"frame/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// queryer represents a subset of pgxpool.Pool methods needed by UserRepository
type queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// UserRepository handles all user-related database operations
type UserRepository struct {
	pool queryer
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(pool queryer) *UserRepository {
	return &UserRepository{pool: pool}
}

// Exists checks if a user with the given first name and last name already exists
// Returns (nil, nil) if user doesn't exist, (uuid.UUID, nil) if user exists, and (nil, error) if there's an error
func (r *UserRepository) Exists(ctx context.Context, firstName, lastName string) (*uuid.UUID, error) {
	query := `
		SELECT id
		FROM users 
		WHERE first_name = $1 AND last_name = $2
		LIMIT 1`

	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, firstName, lastName).Scan(&id)

	if err == pgx.ErrNoRows {
		return nil, nil // User doesn't exist
	}
	if err != nil {
		return nil, fmt.Errorf("error checking if user exists: %v", err)
	}

	return &id, nil
}

// Create inserts a new user into the database or returns existing user
// Returns (user, isNewUser, error) where isNewUser indicates if the user was created or found
func (r *UserRepository) Create(ctx context.Context, firstName, lastName string) (*models.User, bool, error) {
	// Check if user already exists
	existingID, err := r.Exists(ctx, firstName, lastName)
	if err != nil {
		return nil, false, err
	}
	if existingID != nil {
		return &models.User{ID: *existingID}, false, nil
	}

	user := &models.User{
		ID:        uuid.New(),
		FirstName: firstName,
		LastName:  lastName,
	}

	query := `
		INSERT INTO users (id, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, first_name, last_name, created_at, updated_at`

	err = r.pool.QueryRow(ctx, query, user.ID, user.FirstName, user.LastName).
		Scan(&user.ID, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, false, fmt.Errorf("error creating user: %v", err)
	}

	return user, true, nil
}

// GetByID retrieves a user by their ID
// Returns (nil, error) if user not found or there's an error
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, first_name, last_name, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1`

	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user by ID: %v", err)
	}

	return user, nil
}
