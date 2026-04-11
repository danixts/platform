// Package logger is a thin wrapper around rs/zerolog that provides a
// pre-configured process-wide logger and a small set of convenience helpers.
//
// Services call Init() once at startup with the desired Config; subsequent
// calls to Get(), Info(), Warn(), etc. return the configured logger.
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Config controls how Init sets up the global logger.
type Config struct {
	// Level is a zerolog level string: trace, debug, info, warn, error, fatal.
	// Invalid values default to info.
	Level string
	// Pretty enables the colourised console writer (dev only). Disabled by
	// default so prod emits JSON.
	Pretty bool
	// Service is attached as the "service" field on every log line.
	Service string
	// Output overrides the sink. Default is os.Stderr.
	Output io.Writer
}

var log zerolog.Logger

// Init configures the global logger. It is safe to call multiple times in
// tests, but should generally only be called once at process startup.
func Init(cfg Config) {
	lvl, err := zerolog.ParseLevel(cfg.Level)
	if err != nil || cfg.Level == "" {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	zerolog.TimeFieldFormat = time.RFC3339

	var out io.Writer = os.Stderr
	if cfg.Output != nil {
		out = cfg.Output
	}
	if cfg.Pretty {
		out = zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
	}

	ctx := zerolog.New(out).With().Timestamp()
	if cfg.Service != "" {
		ctx = ctx.Str("service", cfg.Service)
	}
	log = ctx.Logger()
}

// Get returns the configured global logger. Services may clone it with
// .With() to attach request-scoped fields.
func Get() *zerolog.Logger { return &log }

// Event shortcuts — safe before Init (zerolog's zero value is a no-op).
func Trace() *zerolog.Event { return log.Trace() }
func Debug() *zerolog.Event { return log.Debug() }
func Info() *zerolog.Event  { return log.Info() }
func Warn() *zerolog.Event  { return log.Warn() }
func Error() *zerolog.Event { return log.Error() }
func Fatal() *zerolog.Event { return log.Fatal() }
