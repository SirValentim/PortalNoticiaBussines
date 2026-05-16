package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (r *Repository) runPostgresMigrations(migrationsDir string) error {
	if strings.TrimSpace(migrationsDir) == "" {
		migrationsDir = "migrations"
	}
	if _, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(32) PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`); err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("nenhuma migration encontrada em %s", migrationsDir)
	}
	for _, file := range files {
		version, name := migrationVersionName(file)
		if version == "" {
			continue
		}
		applied, err := r.migrationApplied(context.Background(), version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		tx, err := r.db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s falhou: %w", filepath.Base(file), err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, $3)`, version, name, time.Now()); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) migrationApplied(ctx context.Context, version string) (bool, error) {
	var current string
	err := r.db.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE version = $1`, version).Scan(&current)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) markSQLiteSchemaVersion() error {
	if _, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
		return err
	}
	_, err := r.db.Exec(`INSERT OR IGNORE INTO schema_migrations (version, name) VALUES ($1, $2)`, "sqlite_auto", "sqlite automatic schema")
	return err
}

func migrationVersionName(file string) (string, string) {
	base := filepath.Base(file)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	parts := strings.SplitN(name, "_", 2)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "", name
	}
	return parts[0], name
}
