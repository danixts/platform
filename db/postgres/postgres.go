// Package postgres provides a thin wrapper around GORM + the pgx driver
// preconfigured with the conventions used across XMart Cloud services:
// UTC timestamps, prepared statements, translated errors and tuneable
// connection pools.
package postgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// LogLevel mirrors gorm.io/gorm/logger levels so callers don't need to
// import the gorm logger package directly.
type LogLevel int

const (
	LogSilent LogLevel = 1
	LogError  LogLevel = 2
	LogWarn   LogLevel = 3
	LogInfo   LogLevel = 4
)

// Config describes how to open a Postgres connection pool. Only DSN is
// required; the rest have sensible defaults.
type Config struct {
	// DSN is the full libpq connection string. Callers are expected to
	// build this from their own config layer.
	DSN string

	// MaxOpenConns caps the connection pool. Default: 50.
	MaxOpenConns int
	// MaxIdleConns caps idle connections. Default: 10.
	MaxIdleConns int
	// ConnMaxLifetime recycles connections after this duration. Default: 30m.
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime closes idle connections after this duration. Default: 5m.
	ConnMaxIdleTime time.Duration

	// LogLevel controls GORM's query logger. Default: LogWarn.
	LogLevel LogLevel

	// SlowThreshold logs queries slower than this at Warn. Default: 200ms.
	SlowThreshold time.Duration

	// PrepareStmt toggles GORM's prepared statement cache. Default: true.
	PrepareStmt *bool
}

// New opens a GORM DB against Postgres with the given Config. It pings
// the database before returning so callers get a clear error at startup
// when the DSN is wrong or the server is unreachable.
func New(cfg Config) (*gorm.DB, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("xmart-platform/db/postgres: empty DSN")
	}
	applyDefaults(&cfg)

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		NowFunc:        func() time.Time { return time.Now().UTC() },
		PrepareStmt:    *cfg.PrepareStmt,
		TranslateError: true,
		Logger: gormlogger.Default.LogMode(
			gormlogger.LogLevel(cfg.LogLevel),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

// Close closes the underlying sql.DB. GORM does not expose Close directly
// because the pool is meant to live as long as the process, so callers
// should only use this in tests or graceful shutdown paths.
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func applyDefaults(cfg *Config) {
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 50
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 30 * time.Minute
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = 5 * time.Minute
	}
	if cfg.LogLevel == 0 {
		cfg.LogLevel = LogWarn
	}
	if cfg.SlowThreshold == 0 {
		cfg.SlowThreshold = 200 * time.Millisecond
	}
	if cfg.PrepareStmt == nil {
		t := true
		cfg.PrepareStmt = &t
	}
}
