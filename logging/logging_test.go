package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestLoggingInitialization(t *testing.T) {
	// Save and restore viper config
	oldConfig := viper.GetViper()
	defer func() {
		viper.Reset()
		*viper.GetViper() = *oldConfig
	}()

	tests := []struct {
		name      string
		logLevel  string
		wantLevel zapcore.Level
	}{
		{
			name:      "default info level",
			logLevel:  "info",
			wantLevel: zapcore.InfoLevel,
		},
		{
			name:      "debug level",
			logLevel:  "debug",
			wantLevel: zapcore.DebugLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("config", map[string]interface{}{
				"logging": map[string]interface{}{
					"level": tt.logLevel,
				},
			})
			viper.Set("logging.level", tt.logLevel)

			err := Initialize()
			require.NoError(t, err)

			assert.Equal(t, tt.wantLevel, GetLogLevel())
		})
	}
}

func TestSetLogLevel(t *testing.T) {
	// Initialize first
	viper.Set("config", map[string]interface{}{
		"logging": map[string]interface{}{
			"level": "info",
		},
	})
	err := Initialize()
	require.NoError(t, err)

	tests := []struct {
		name      string
		level     zapcore.Level
		wantLevel zapcore.Level
	}{
		{
			name:      "set to debug",
			level:     zapcore.DebugLevel,
			wantLevel: zapcore.DebugLevel,
		},
		{
			name:      "set to info",
			level:     zapcore.InfoLevel,
			wantLevel: zapcore.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetLogLevel(tt.level)
			require.NoError(t, err)
			assert.Equal(t, tt.wantLevel, GetLogLevel())
		})
	}
}

func TestMiddleware(t *testing.T) {
	// Initialize logger first
	viper.Set("config", map[string]interface{}{
		"logging": map[string]interface{}{
			"level": "info",
		},
	})
	err := Initialize()
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// Wrap handler with logging middleware
	loggingHandler := Middleware(handler)
	loggingHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}
