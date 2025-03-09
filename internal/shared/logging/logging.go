package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"
)

// Initialize sets up the global logger settings
func Initialize(logLevel string, logFile string) error {
	// Parse log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	// Format timestamps to be human-readable
	zerolog.TimeFieldFormat = time.RFC3339

	// Default logger output to console
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	// If log file is specified, also write to file
	if logFile != "" {
		// Create directory if it doesn't exist
		dir := filepath.Dir(logFile)
		if dir != "." && dir != "/" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create log directory: %w", err)
			}
		}

		// Open file for writing/appending
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		// Use MultiWriter to write to both file and console
		multi := zerolog.MultiLevelWriter(output, file)
		log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	} else {
		// Console output only
		log.Logger = zerolog.New(output).With().Timestamp().Logger()
	}

	// Initialize the global logger
	globalLogger = &Logger{
		logger: log.Logger.With().Str("component", "global").Logger(),
	}

	return nil
}

// Logger is a simple wrapper around zerolog.Logger
type Logger struct {
	logger zerolog.Logger
}

// NewLogger creates a new logger with a given context name
func NewLogger(component string) *Logger {
	return &Logger{
		logger: log.Logger.With().Str("component", component).Logger(),
	}
}

// SetOutput sets a custom output writer for the logger
func (l *Logger) SetOutput(w io.Writer) {
	l.logger = l.logger.Output(w)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.logger.Debug().Msgf(msg, args...)
	} else {
		l.logger.Debug().Msg(msg)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.logger.Info().Msgf(msg, args...)
	} else {
		l.logger.Info().Msg(msg)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.logger.Warn().Msgf(msg, args...)
	} else {
		l.logger.Warn().Msg(msg)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, args ...interface{}) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}

	if len(args) > 0 {
		event.Msgf(msg, args...)
	} else {
		event.Msg(msg)
	}
}

// Fatal logs a fatal message and exits the application
func (l *Logger) Fatal(msg string, err error, args ...interface{}) {
	event := l.logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}

	if len(args) > 0 {
		event.Msgf(msg, args...)
	} else {
		event.Msg(msg)
	}
}

// WithField adds a field to the logger context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// WithFields adds multiple fields to the logger context
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	contextLogger := l.logger.With()
	for k, v := range fields {
		contextLogger = contextLogger.Interface(k, v)
	}
	return &Logger{
		logger: contextLogger.Logger(),
	}
}

// Global logger for package-level logging
var globalLogger *Logger

// Debug logs a debug message to the global logger
func Debug(msg string, args ...interface{}) {
	if globalLogger == nil {
		globalLogger = NewLogger("global")
	}
	globalLogger.Debug(msg, args...)
}

// Info logs an info message to the global logger
func Info(msg string, args ...interface{}) {
	if globalLogger == nil {
		globalLogger = NewLogger("global")
	}
	globalLogger.Info(msg, args...)
}

// Warn logs a warning message to the global logger
func Warn(msg string, args ...interface{}) {
	if globalLogger == nil {
		globalLogger = NewLogger("global")
	}
	globalLogger.Warn(msg, args...)
}

// Error logs an error message to the global logger
func Error(msg string, err error, args ...interface{}) {
	if globalLogger == nil {
		globalLogger = NewLogger("global")
	}
	globalLogger.Error(msg, err, args...)
}

// Fatal logs a fatal message to the global logger and exits
func Fatal(msg string, err error, args ...interface{}) {
	if globalLogger == nil {
		globalLogger = NewLogger("global")
	}
	globalLogger.Fatal(msg, err, args...)
}

// WithComponent creates a new logger with the component field set
func WithComponent(component string) *Logger {
	return NewLogger(component)
}

// GormLogger returns a GORM logger implementation
func (l *Logger) GormLogger() logger.Interface {
	return &gormLogger{
		logger:        l,
		SlowThreshold: 200 * time.Millisecond,
	}
}

// gormLogger implements the gorm.logger.Interface
type gormLogger struct {
	logger        *Logger
	SlowThreshold time.Duration
}

// LogMode implementation of logger.Interface
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info implementation of logger.Interface
func (l *gormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Warn implementation of logger.Interface
func (l *gormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Error implementation of logger.Interface
func (l *gormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if len(args) > 0 {
		if err, ok := args[0].(error); ok {
			l.logger.Error(msg, err)
			return
		}
	}
	l.logger.Error(msg, nil, args...)
}

// Trace implementation of logger.Interface
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := map[string]interface{}{
		"elapsed": elapsed,
		"rows":    rows,
		"sql":     sql,
	}

	logEvent := l.logger.WithFields(fields)

	if err != nil {
		logEvent.Error("GORM error", err)
		return
	}

	if elapsed > l.SlowThreshold {
		logEvent.Warn("GORM slow query")
		return
	}

	logEvent.Debug("GORM query")
}
