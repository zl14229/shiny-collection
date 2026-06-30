package database

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"shiny-collection/internal/config"
	"shiny-collection/internal/model"
)

var DB *gorm.DB

func Init(cfg *config.DatabaseConfig, logger *zap.Logger) error {
	// ensure data directory exists
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	logLevel := gormlogger.Warn
	if cfg.Path == ":memory:" {
		logLevel = gormlogger.Silent
	}

	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return err
	}

	// connection pool settings for SQLite
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(1) // SQLite only supports one writer at a time

	// auto migrate all models
	if err := db.AutoMigrate(
		&model.Pokemon{},
		&model.Game{},
		&model.Method{},
		&model.Tag{},
		&model.Record{},
	); err != nil {
		return err
	}

	DB = db
	logger.Info("database initialized", zap.String("path", cfg.Path))
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
