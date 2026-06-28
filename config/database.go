package config

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	maxRetries        = 3
	retryDelay        = 5 * time.Second
	defaultMaxOpen    = 25
	defaultMaxIdle    = 5
	defaultLifetime   = 30 * time.Minute
	defaultIdleTime   = 5 * time.Minute
	defaultTimeZone   = "UTC"
	gormSlowThreshold = 200 * time.Millisecond
)

func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		logError("could not load .env file: %v (falling back to process env)", err)
	}

	switch os.Getenv("GIN_MODE") {
	case gin.ReleaseMode:
		gin.SetMode(gin.ReleaseMode)
	case gin.TestMode:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	return nil
}

func Connect() (*gorm.DB, error) {
	if err := LoadEnv(); err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		envOr("DB_HOST", "localhost"),
		envOr("DB_PORT", "5432"),
		envOr("DB_USER", "bookstore"),
		envOr("DB_PASSWORD", "bookstore_secret"),
		envOr("DB_NAME", "bookstore_db"),
		envOr("DB_SSLMODE", "disable"),
		envOr("DB_TIMEZONE", defaultTimeZone),
	)

	gormCfg := &gorm.Config{
		Logger: logger.New(ginLogger{gin.DefaultWriter}, logger.Config{
			SlowThreshold:             gormSlowThreshold,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  gin.Mode() != gin.ReleaseMode,
		}),
	}

	var (
		db  *gorm.DB
		err error
	)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logInfo("connecting to postgres (attempt %d/%d)...", attempt, maxRetries)

		db, err = gorm.Open(postgres.Open(dsn), gormCfg)
		if err == nil {
			if sqlDB, e := db.DB(); e == nil {
				if perr := sqlDB.Ping(); perr != nil {
					err = perr
				}
			} else {
				err = e
			}
		}

		if err == nil {
			err = applyPool(db)
		}

		if err == nil {
			logInfo("connected to postgres on attempt %d/%d", attempt, maxRetries)
			return db, nil
		}

		logError("attempt %d/%d failed: %v", attempt, maxRetries, err)
		if attempt < maxRetries {
			logInfo("retrying in %s...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to connect to postgres after %d attempts: %w", maxRetries, err)
}

func applyPool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(envInt("DB_MAX_OPEN_CONNS", defaultMaxOpen))
	sqlDB.SetMaxIdleConns(envInt("DB_MAX_IDLE_CONNS", defaultMaxIdle))
	sqlDB.SetConnMaxLifetime(envDuration("DB_CONN_MAX_LIFETIME", defaultLifetime))
	sqlDB.SetConnMaxIdleTime(envDuration("DB_CONN_MAX_IDLE_TIME", defaultIdleTime))

	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		logError("invalid int for %s=%q, using default %d", key, v, fallback)
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		logError("invalid duration for %s=%q, using default %s", key, v, fallback)
	}
	return fallback
}

func logInfo(format string, args ...any) {
	fmt.Fprintf(gin.DefaultWriter, "[db] "+format+"\n", args...)
}

func logError(format string, args ...any) {
	fmt.Fprintf(gin.DefaultErrorWriter, "[db] "+format+"\n", args...)
}

type ginLogger struct {
	out io.Writer
}

func (g ginLogger) Printf(format string, args ...any) {
	fmt.Fprintf(g.out, format, args...)
}
