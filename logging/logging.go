package logging

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger        atomic.Pointer[zap.Logger]
	currentLevel  atomic.Value // holds zapcore.Level
	defaultConfig zap.Config
)

// Initialize sets up the zap logger
func Initialize() error {
	cfg := viper.Get("config")
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	// Set up default production config
	defaultConfig = zap.NewProductionConfig()
	defaultConfig.OutputPaths = []string{"stdout"}
	defaultConfig.ErrorOutputPaths = []string{"stderr"}
	defaultConfig.Encoding = "json" // Explicitly set JSON encoder

	// Get log level from config
	logLevel := strings.ToLower(viper.GetString("logging.level"))
	level := zapcore.InfoLevel
	if logLevel == "debug" {
		level = zapcore.DebugLevel
	}

	return updateLogger(level)
}

// updateLogger creates a new logger with the specified level
func updateLogger(level zapcore.Level) error {
	defaultConfig.Level = zap.NewAtomicLevelAt(level)
	newLogger, err := defaultConfig.Build(zap.AddCaller())
	if err != nil {
		return fmt.Errorf("error building logger: %v", err)
	}

	// Store the new level
	currentLevel.Store(level)

	// Sync the old logger if it exists
	if old := logger.Load(); old != nil {
		if err := old.Sync(); err != nil {
			fmt.Printf("Error syncing old logger: %v\n", err)
		}
	}

	// Store the new logger
	logger.Store(newLogger)
	return nil
}

// GetLogger returns the configured zap logger
func GetLogger() *zap.Logger {
	return logger.Load()
}

// GetLogLevel returns the current logging level
func GetLogLevel() zapcore.Level {
	return currentLevel.Load().(zapcore.Level)
}

// SetLogLevel changes the logging level
func SetLogLevel(level zapcore.Level) error {
	return updateLogger(level)
}

// SetDebugMode enables or disables debug mode
func SetDebugMode(debug bool) error {
	if debug {
		return SetLogLevel(zapcore.DebugLevel)
	}
	return SetLogLevel(zapcore.InfoLevel)
}

// Close flushes any buffered log entries
func Close() error {
	if l := logger.Load(); l != nil {
		return l.Sync()
	}
	return nil
}

// IsDebugMode returns true if debug logging is enabled
func IsDebugMode() bool {
	return GetLogLevel() == zapcore.DebugLevel
}

// Middleware creates a logging middleware that logs HTTP requests
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture the status code
		rw := &responseWriter{w, http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Log the request details
		if l := logger.Load(); l != nil {
			l.Info("HTTP Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Int("status", rw.status),
				zap.Duration("latency", time.Since(start)),
				zap.String("user_agent", r.UserAgent()),
			)
		}
	})
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status code before writing it
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
