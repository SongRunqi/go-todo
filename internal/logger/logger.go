package logger

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Log zerolog.Logger

// Init initializes the logger with the specified level
func Init(level string) {
	// Set up console writer with colors for better readability
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}

	// Parse log level
	logLevel := parseLevel(level)
	zerolog.SetGlobalLevel(logLevel)

	// Create logger
	Log = zerolog.New(output).With().Timestamp().Logger()

	// Set global logger
	log.Logger = Log
}

// InitWithWriter initializes logger with custom writer (useful for testing)
func InitWithWriter(level string, w io.Writer) {
	output := zerolog.ConsoleWriter{Out: w, TimeFormat: "15:04:05"}
	logLevel := parseLevel(level)
	zerolog.SetGlobalLevel(logLevel)
	Log = zerolog.New(output).With().Timestamp().Logger()
	log.Logger = Log
}

// parseLevel parses log level string to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled", "off":
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

// Debug logs a debug message
func Debug(msg string) {
	Log.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	Log.Debug().Msgf(format, args...)
}

// Info logs an info message
func Info(msg string) {
	Log.Info().Msg(msg)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Log.Info().Msgf(format, args...)
}

// Warn logs a warning message
func Warn(msg string) {
	Log.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Log.Warn().Msgf(format, args...)
}

// Error logs an error message
func Error(msg string) {
	Log.Error().Msg(msg)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Log.Error().Msgf(format, args...)
}

// ErrorWithErr logs an error message with error object
func ErrorWithErr(err error, msg string) {
	Log.Error().Err(err).Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	Log.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	Log.Fatal().Msgf(format, args...)
}

// WithField returns a logger with a field
func WithField(key string, value interface{}) zerolog.Logger {
	return Log.With().Interface(key, value).Logger()
}

// WithFields returns a logger with multiple fields
func WithFields(fields map[string]interface{}) zerolog.Logger {
	ctx := Log.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return ctx.Logger()
}
