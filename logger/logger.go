package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Log is the global structured logger instance.
// Use logger.Log.Info().Msg("...") throughout the application.
var Log zerolog.Logger

// Init initializes the global logger.
// In development, it uses a pretty console writer.
// In production (ENV=production), it uses JSON output for structured log aggregation.
func Init() {
	env := os.Getenv("ENV")

	if env == "production" {
		// Production: JSON output for log aggregation (ELK, Loki, etc.)
		Log = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Caller().
			Logger()
	} else {
		// Development: pretty colored console output
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		Log = zerolog.New(output).
			With().
			Timestamp().
			Caller().
			Logger()
	}
}

// Info returns an info-level event logger.
func Info() *zerolog.Event {
	return Log.Info()
}

// Error returns an error-level event logger.
func Error() *zerolog.Event {
	return Log.Error()
}

// Warn returns a warn-level event logger.
func Warn() *zerolog.Event {
	return Log.Warn()
}

// Debug returns a debug-level event logger.
func Debug() *zerolog.Event {
	return Log.Debug()
}

// Fatal returns a fatal-level event logger. This will call os.Exit(1) after logging.
func Fatal() *zerolog.Event {
	return Log.Fatal()
}
