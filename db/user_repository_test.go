package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a connection to the test database and cleans up the users table
func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	config, err := pgxpool.ParseConfig("postgres://postgres:postgres@localhost:15432/framework?sslmode=require")
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err)

	// Clean up the users table
	//_, err = pool.Exec(ctx, "TRUNCATE users")
	//require.NoError(t, err)

	return pool
}

func TestUserRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewUserRepository(pool)
	ctx := context.Background()

	t.Run("Create new user", func(t *testing.T) {
		// Test creating a new user
		user, isNew, err := repo.Create(ctx, "John", "Doe")
		require.NoError(t, err)
		assert.True(t, isNew)
		assert.NotNil(t, user)
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())

		// Test getting the same user
		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.FirstName, found.FirstName)
		assert.Equal(t, user.LastName, found.LastName)
		assert.Equal(t, user.CreatedAt, found.CreatedAt)
		assert.Equal(t, user.UpdatedAt, found.UpdatedAt)
	})

	t.Run("Create duplicate user", func(t *testing.T) {
		// Create first user
		first, isNew, err := repo.Create(ctx, "Jane", "Smith")
		require.NoError(t, err)
		assert.True(t, isNew)
		assert.NotNil(t, first)

		// Try to create the same user again
		second, isNew, err := repo.Create(ctx, "Jane", "Smith")
		require.NoError(t, err)
		assert.False(t, isNew)
		assert.Equal(t, first.ID, second.ID)
	})

	t.Run("GetByID non-existent user", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
	})

	t.Run("Exists check", func(t *testing.T) {
		// Check non-existent user
		id, err := repo.Exists(ctx, "Test", "User")
		require.NoError(t, err)
		assert.Nil(t, id)

		// Create a user
		user, _, err := repo.Create(ctx, "Test", "User")
		require.NoError(t, err)

		// Check existing user
		id, err = repo.Exists(ctx, "Test", "User")
		require.NoError(t, err)
		assert.NotNil(t, id)
		assert.Equal(t, user.ID, *id)
	})
}

// Mock implementations for unit tests
type mockRow struct {
	err  error
	vals []interface{}
}

func (m *mockRow) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}
	for i, val := range m.vals {
		switch v := dest[i].(type) {
		case *uuid.UUID:
			*v = val.(uuid.UUID)
		case *string:
			*v = val.(string)
		case *time.Time:
			*v = val.(time.Time)
		}
	}
	return nil
}

type mockPool struct {
	queryRowFunc func(context.Context, string, ...interface{}) pgx.Row
}

func (m *mockPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return m.queryRowFunc(ctx, sql, args...)
}

func TestUserRepository_Unit(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	now := time.Now().UTC()

	t.Run("GetByID success", func(t *testing.T) {
		mock := &mockPool{
			queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
				assert.Equal(t, testID, args[0])
				return &mockRow{
					vals: []interface{}{
						testID,
						"John",
						"Doe",
						now,
						now,
					},
				}
			},
		}

		repo := NewUserRepository(mock)
		user, err := repo.GetByID(ctx, testID)
		require.NoError(t, err)
		assert.Equal(t, testID, user.ID)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, now, user.CreatedAt)
		assert.Equal(t, now, user.UpdatedAt)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		mock := &mockPool{
			queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
				return &mockRow{err: pgx.ErrNoRows}
			},
		}

		repo := NewUserRepository(mock)
		_, err := repo.GetByID(ctx, testID)
		assert.Error(t, err)
	})

	t.Run("Exists found", func(t *testing.T) {
		mock := &mockPool{
			queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
				assert.Equal(t, "John", args[0])
				assert.Equal(t, "Doe", args[1])
				return &mockRow{vals: []interface{}{testID}}
			},
		}

		repo := NewUserRepository(mock)
		id, err := repo.Exists(ctx, "John", "Doe")
		require.NoError(t, err)
		assert.NotNil(t, id)
		assert.Equal(t, testID, *id)
	})

	t.Run("Exists not found", func(t *testing.T) {
		mock := &mockPool{
			queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
				return &mockRow{err: pgx.ErrNoRows}
			},
		}

		repo := NewUserRepository(mock)
		id, err := repo.Exists(ctx, "John", "Doe")
		require.NoError(t, err)
		assert.Nil(t, id)
	})

	t.Run("Create new user", func(t *testing.T) {
		existsCount := 0
		createCount := 0
		mock := &mockPool{
			queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
				if sql == "\n\t\tSELECT id\n\t\tFROM users \n\t\tWHERE first_name = $1 AND last_name = $2\n\t\tLIMIT 1" {
					existsCount++
					return &mockRow{err: pgx.ErrNoRows}
				}
				createCount++
				return &mockRow{vals: []interface{}{testID, "John", "Doe", now, now}}
			},
		}

		repo := NewUserRepository(mock)
		user, isNew, err := repo.Create(ctx, "John", "Doe")
		require.NoError(t, err)
		assert.True(t, isNew)
		assert.Equal(t, testID, user.ID)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, now, user.CreatedAt)
		assert.Equal(t, now, user.UpdatedAt)
		assert.Equal(t, 1, existsCount)
		assert.Equal(t, 1, createCount)
	})

	t.Run("Create existing user", func(t *testing.T) {
		mock := &mockPool{
			queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
				return &mockRow{vals: []interface{}{testID}}
			},
		}

		repo := NewUserRepository(mock)
		user, isNew, err := repo.Create(ctx, "John", "Doe")
		require.NoError(t, err)
		assert.False(t, isNew)
		assert.Equal(t, testID, user.ID)
	})
}
