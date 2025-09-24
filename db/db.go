package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"frame/config"
	"frame/logging"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	pool           *pgxpool.Pool
	mu             sync.RWMutex
	ctx            context.Context
	pingCancelFunc context.CancelFunc
)

// Initialize creates a connection pool to the PostgreSQL database
func Initialize(initCtx context.Context) error {
	// Create a new context with cancel for the ping routine
	ctx, pingCancelFunc = context.WithCancel(initCtx)

	if err := connect(); err != nil {
		return err
	}

	// Start the periodic ping routine
	go startPingRoutine()

	// Register callback for config changes
	config.RegisterCallback(func(cfg *config.Config) {
		logger := logging.GetLogger()
		logger.Info("Reconnecting to database due to configuration change")

		mu.Lock()
		defer mu.Unlock()

		// Close existing connection
		if pool != nil {
			pool.Close()
		}

		// Reconnect with new configuration
		if err := connect(); err != nil {
			logger.Error("Failed to reconnect to database",
				zap.Error(err))
		}
	})

	return nil
}

// connect establishes a new database connection with current configuration
func connect() error {
	cfg := viper.Get("config").(*config.Config)
	dbConfig := cfg.Database

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
		dbConfig.SSLMode,
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("error parsing database config: %v", err)
	}

	newPool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}

	// Test the connection
	if err := newPool.Ping(ctx); err != nil {
		newPool.Close() // Clean up on failure
		return fmt.Errorf("error pinging database: %v", err)
	}

	pool = newPool
	return nil
}

// GetPool returns the database connection pool
func GetPool() *pgxpool.Pool {
	mu.RLock()
	defer mu.RUnlock()
	return pool
}

// startPingRoutine starts a goroutine that pings the database every minute
func startPingRoutine() {
	logger := logging.GetLogger()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping database ping routine")
			return
		case <-ticker.C:
			mu.RLock()
			if pool != nil {
				if err := pool.Ping(ctx); err != nil {
					logger.Error("Database ping failed",
						zap.Error(err))
				} else {
					logger.Debug("Database ping successful")
				}
			}
			mu.RUnlock()
		}
	}
}

// Close closes the database connection pool and stops the ping routine
func Close() {
	if pingCancelFunc != nil {
		pingCancelFunc() // Stop the ping routine
	}

	mu.Lock()
	defer mu.Unlock()
	if pool != nil {
		pool.Close()
		pool = nil
	}
}
