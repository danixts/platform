package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Level   string
	Pretty  bool
	Service string
	Output  io.Writer
}

var log zerolog.Logger

func Init(cfg Config) {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil || cfg.Level == "" {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339

	var out io.Writer = os.Stderr
	if cfg.Output != nil {
		out = cfg.Output
	}
	if cfg.Pretty {
		out = zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
	}

	builder := zerolog.New(out).With().Timestamp()
	if cfg.Service != "" {
		builder = builder.Str("service", cfg.Service)
	}
	log = builder.Logger()
}

func Get() *zerolog.Logger { return &log }

func Trace() *zerolog.Event { return log.Trace() }
func Debug() *zerolog.Event { return log.Debug() }
func Info() *zerolog.Event  { return log.Info() }
func Warn() *zerolog.Event  { return log.Warn() }
func Error() *zerolog.Event { return log.Error() }
func Fatal() *zerolog.Event { return log.Fatal() }
