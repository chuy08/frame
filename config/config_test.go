package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Save original environment
	originalEnv := os.Environ()

	// Clean up after each test
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			parts := strings.SplitN(env, "=", 2)
			if err := os.Setenv(parts[0], parts[1]); err != nil {
				t.Errorf("Failed to restore environment variable %s: %v", parts[0], err)
			}
		}
		viper.Reset()
	}()

	tests := []struct {
		name      string
		envVars   map[string]string
		configStr string
		want      *Config
		wantErr   bool
	}{
		{
			name: "default values",
			want: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     15432,
					User:     "postgres",
					Password: "postgres",
					Name:     "postgres",
					SSLMode:  "disable",
				},
				Server: ServerConfig{
					Port: 8080,
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
		},
		{
			name: "environment variables override defaults",
			envVars: map[string]string{
				"FRAME_DATABASE_HOST":     "db.example.com",
				"FRAME_DATABASE_PORT":     "5432",
				"FRAME_DATABASE_USER":     "admin",
				"FRAME_DATABASE_PASSWORD": "secret",
				"FRAME_DATABASE_NAME":     "myapp",
				"FRAME_SERVER_PORT":       "3000",
				"FRAME_LOGGING_LEVEL":     "debug",
			},
			want: &Config{
				Database: DatabaseConfig{
					Host:     "db.example.com",
					Port:     5432,
					User:     "admin",
					Password: "secret",
					Name:     "myapp",
					SSLMode:  "disable",
				},
				Server: ServerConfig{
					Port: 3000,
				},
				Logging: LoggingConfig{
					Level: "debug",
				},
			},
		},
		{
			name: "config file values",
			configStr: `
database:
  host: "confighost"
  port: 6543
  user: "configuser"
  password: "configpass"
  name: "configdb"
  sslmode: "verify-full"
server:
  port: 9090
logging:
  level: "debug"
`,
			want: &Config{
				Database: DatabaseConfig{
					Host:     "confighost",
					Port:     6543,
					User:     "configuser",
					Password: "configpass",
					Name:     "configdb",
					SSLMode:  "verify-full",
				},
				Server: ServerConfig{
					Port: 9090,
				},
				Logging: LoggingConfig{
					Level: "debug",
				},
			},
		},
		{
			name: "environment variables override config file",
			configStr: `
database:
  host: "confighost"
  port: 6543
  user: "configuser"
  password: "configpass"
  name: "configdb"
server:
  port: 9090
logging:
  level: "debug"
`,
			envVars: map[string]string{
				"FRAME_DATABASE_HOST": "envhost",
				"FRAME_SERVER_PORT":   "1234",
			},
			want: &Config{
				Database: DatabaseConfig{
					Host:     "envhost",
					Port:     6543,
					User:     "configuser",
					Password: "configpass",
					Name:     "configdb",
					SSLMode:  "disable",
				},
				Server: ServerConfig{
					Port: 1234,
				},
				Logging: LoggingConfig{
					Level: "debug",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment and viper for each test
			os.Clearenv()
			viper.Reset()

			// Set environment variables if any
			for k, v := range tt.envVars {
				if err := os.Setenv(k, v); err != nil {
					t.Fatalf("Failed to set environment variable %s: %v", k, err)
				}
			}

			// Create config file if content provided
			if tt.configStr != "" {
				tmpfile, err := os.CreateTemp("", "config.*.yaml")
				require.NoError(t, err)
				defer func() {
					if err := os.Remove(tmpfile.Name()); err != nil {
						t.Errorf("Failed to remove temporary config file: %v", err)
					}
				}()

				_, err = tmpfile.Write([]byte(tt.configStr))
				require.NoError(t, err)
				require.NoError(t, tmpfile.Close())

				viper.SetConfigFile(tmpfile.Name())
			}

			// Call Load after setting up config
			got, err := Load()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfigCallbacks(t *testing.T) {
	// Reset callback slice
	callbacks = nil

	var callCount int
	var lastConfig *Config

	callback := func(cfg *Config) {
		callCount++
		lastConfig = cfg
	}

	// Test registration
	RegisterCallback(callback)
	assert.Equal(t, 1, len(callbacks))

	// Test notification
	config := &Config{
		Database: DatabaseConfig{Host: "testhost"},
		Server:   ServerConfig{Port: 1234},
		Logging:  LoggingConfig{Level: "debug"},
	}

	notifyCallbacks(config)
	assert.Equal(t, 1, callCount)
	assert.Equal(t, config, lastConfig)

	// Test multiple callbacks
	var secondCallCount int
	var secondLastConfig *Config

	RegisterCallback(func(cfg *Config) {
		secondCallCount++
		secondLastConfig = cfg
	})

	notifyCallbacks(config)
	assert.Equal(t, 2, callCount)
	assert.Equal(t, 1, secondCallCount)
	assert.Equal(t, config, lastConfig)
	assert.Equal(t, config, secondLastConfig)
}

func TestWatchConfig(t *testing.T) {
	// Create a temporary config file
	tmpfile, err := os.CreateTemp("", "config.*.yaml")
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Errorf("Failed to remove temporary config file: %v", err)
		}
	}()

	initialConfig := `
database:
  host: "initial"
server:
  port: 8080
logging:
  level: "info"
`
	_, err = tmpfile.Write([]byte(initialConfig))
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	viper.Reset()
	viper.SetConfigFile(tmpfile.Name())

	// Start watching config
	var configChanged bool
	RegisterCallback(func(cfg *Config) {
		configChanged = true
		assert.Equal(t, "modified", cfg.Database.Host)
	})

	WatchConfig()

	// Modify the config file
	modifiedConfig := `
database:
  host: "modified"
server:
  port: 8080
logging:
  level: "info"
`
	err = os.WriteFile(tmpfile.Name(), []byte(modifiedConfig), 0644)
	require.NoError(t, err)

	// Wait for the file system event to be processed
	time.Sleep(100 * time.Millisecond)

	assert.True(t, configChanged)
}
