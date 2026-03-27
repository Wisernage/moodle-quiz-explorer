package db

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"timo/internal/model"
)

// Open opens a SQLite database and runs AutoMigrate for all models.
func Open(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		dsn = "file:timo.db?cache=shared&_fk=1"
	}
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := gdb.AutoMigrate(model.AllModels()...); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return gdb, nil
}

// OpenMemory opens an in-memory SQLite database (for tests).
func OpenMemory() (*gorm.DB, error) {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := gdb.AutoMigrate(model.AllModels()...); err != nil {
		return nil, err
	}
	return gdb, nil
}

// EnsureDir ensures the directory for a file DSN exists (e.g. ./data/app.db).
func EnsureDir(dsn string) error {
	if dsn == "" || len(dsn) < 6 || dsn[:5] != "file:" {
		return nil
	}
	rest := dsn[5:]
	path := rest
	for i := 0; i < len(rest); i++ {
		if rest[i] == '?' {
			path = rest[:i]
			break
		}
	}
	if path == "" || path == ":memory:" {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
