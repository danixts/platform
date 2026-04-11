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

func Get() *zerolog.Logger { return &log }

func Trace() *zerolog.Event { return log.Trace() }
func Debug() *zerolog.Event { return log.Debug() }
func Info() *zerolog.Event  { return log.Info() }
func Warn() *zerolog.Event  { return log.Warn() }
func Error() *zerolog.Event { return log.Error() }
func Fatal() *zerolog.Event { return log.Fatal() }
