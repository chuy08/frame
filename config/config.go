package config

import (
	"fmt"
	"strings"
	"sync"

	"frame/logging"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Logging  LoggingConfig
}

type LoggingConfig struct {
	Level string // "debug" or "info"
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ServerConfig struct {
	Port int
}

// ConfigCallback is a function that will be called when configuration changes
type ConfigCallback func(*Config)

var (
	callbacks []ConfigCallback
	mu        sync.RWMutex
)

// RegisterCallback registers a function to be called when configuration changes
func RegisterCallback(callback ConfigCallback) {
	mu.Lock()
	defer mu.Unlock()
	callbacks = append(callbacks, callback)
}

// notifyCallbacks calls all registered callbacks with the new configuration
func notifyCallbacks(config *Config) {
	mu.RLock()
	defer mu.RUnlock()
	for _, callback := range callbacks {
		callback(config)
	}
}

// WatchConfig starts watching the configuration file for changes
func WatchConfig() {
	// Create a logger instance specifically for config watching
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Error creating config logger: %v\n", err)
		return
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Error syncing config logger: %v\n", err)
		}
	}()

	viper.OnConfigChange(func(e fsnotify.Event) {
		// Get a new logger for each event to ensure we have the latest configuration
		eventLogger := logging.GetLogger()
		if eventLogger == nil {
			// Fallback to our local logger if the global one isn't ready
			eventLogger = logger
		}

		eventLogger.Info("Config file changed",
			zap.String("file", e.Name),
			zap.String("operation", e.Op.String()))

		config, err := Load()
		if err != nil {
			eventLogger.Error("Error reloading config",
				zap.Error(err))
			return
		}

		notifyCallbacks(config)
	})

	viper.WatchConfig()
}

// Load initializes configuration from various sources in the following order:
// 1. Default values
// 2. Configuration file
// 3. Environment variables
// 4. Command line flags
func Load() (*Config, error) {
	setDefaults()

	// Enable environment variable support
	viper.SetEnvPrefix("FRAME")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file if it exists
	viper.SetConfigName("config")       // name of config file (without extension)
	viper.SetConfigType("yaml")         // type of config file
	viper.AddConfigPath(".")            // look for config in current directory
	viper.AddConfigPath("$HOME/.frame") // look for config in home directory
	viper.AddConfigPath("/etc/frame/")  // look for config in /etc/frame/

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %v", err)
		}
		// Config file not found, using defaults and env vars
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	return &config, nil
}

// setDefaults sets default values for all configuration options
func setDefaults() {
	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 15432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.name", "postgres")
	viper.SetDefault("database.sslmode", "disable")

	// Server defaults
	viper.SetDefault("server.port", 8080)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
}
