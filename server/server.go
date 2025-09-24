package server

import (
	"context"
	"fmt"
	"frame/api"
	"frame/config"
	"frame/db"
	"frame/logging"
	"net/http"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func Start() error {
	// Initialize logger
	if err := logging.Initialize(); err != nil {
		return err
	}
	defer func() {
		if err := logging.Close(); err != nil {
			fmt.Printf("Error closing logger: %v\n", err)
		}
	}()

	// Initialize database connection
	ctx := context.Background()
	if err := db.Initialize(ctx); err != nil {
		return err
	}
	defer db.Close()

	// Create a new mux for routing
	mux := http.NewServeMux()
	mux.HandleFunc("/user", api.UserHandler)

	// Wrap the mux with our logging middleware
	handler := logging.Middleware(mux)

	cfg := viper.Get("config").(*config.Config)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	// Log startup information
	logger := logging.GetLogger()
	logLevel := "INFO"
	if logging.IsDebugMode() {
		logLevel = "DEBUG"
	}
	logger.Info("Starting server",
		zap.String("address", addr),
		zap.String("log_level", logLevel),
		zap.String("database_host", cfg.Database.Host),
		zap.Int("database_port", cfg.Database.Port),
		zap.String("database_name", cfg.Database.Name),
	)

	return http.ListenAndServe(addr, handler)
}
