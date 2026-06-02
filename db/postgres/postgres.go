package postgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	defaultMaxOpenConns    = 50
	defaultMaxIdleConns    = 10
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
	defaultSlowThreshold   = 200 * time.Millisecond
)

type LogLevel int

const (
	LogSilent LogLevel = 1
	LogError  LogLevel = 2
	LogWarn   LogLevel = 3
	LogInfo   LogLevel = 4
)

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	LogLevel        LogLevel
	SlowThreshold   time.Duration
	PrepareStmt     *bool
}

func New(cfg Config) (*gorm.DB, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("platform/db/postgres: empty DSN")
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
		cfg.MaxOpenConns = defaultMaxOpenConns
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = defaultMaxIdleConns
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = defaultConnMaxLifetime
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = defaultConnMaxIdleTime
	}
	if cfg.LogLevel == 0 {
		cfg.LogLevel = LogWarn
	}
	if cfg.SlowThreshold == 0 {
		cfg.SlowThreshold = defaultSlowThreshold
	}
	if cfg.PrepareStmt == nil {
		// Default off: server-side prepared statements break under PgBouncer
		// transaction mode. Services that pool through pgbouncer-tx are safe by
		// default; set PrepareStmt=&true explicitly to opt back in (session mode).
		f := false
		cfg.PrepareStmt = &f
	}
}
