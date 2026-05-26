package repository

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestMigrationFilesAreContiguous(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(projectRoot(t), "migrations", "*.sql"))
	if err != nil {
		t.Fatalf("glob migrations failed: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no migrations found")
	}
	sort.Strings(files)
	seen := map[string]string{}
	for i, file := range files {
		version, name := migrationVersionName(file)
		if version == "" {
			t.Fatalf("migration without version: %s", file)
		}
		if previous, ok := seen[version]; ok {
			t.Fatalf("duplicate migration version %s: %s and %s", version, previous, filepath.Base(file))
		}
		seen[version] = filepath.Base(file)
		got, err := strconv.Atoi(version)
		if err != nil {
			t.Fatalf("migration version is not numeric in %s: %v", filepath.Base(file), err)
		}
		want := i + 1
		if got != want {
			t.Fatalf("migration %s has version %03d, want %03d", filepath.Base(file), got, want)
		}
		if !strings.HasPrefix(name, fmt.Sprintf("%03d_", want)) {
			t.Fatalf("migration %s does not use expected prefix %03d_", filepath.Base(file), want)
		}
	}
}

func TestPostgresMigrationsApplyWhenURLProvided(t *testing.T) {
	postgresURL := strings.TrimSpace(os.Getenv("POSTGRES_TEST_URL"))
	if postgresURL == "" {
		t.Skip("POSTGRES_TEST_URL not set; skipping live PostgreSQL migration validation")
	}

	schema := "codex_migration_test_" + time.Now().Format("20060102150405")
	schema = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(schema, "_")
	adminDB, err := sql.Open("postgres", postgresURL)
	if err != nil {
		t.Fatalf("open postgres admin connection failed: %v", err)
	}
	defer adminDB.Close()
	if _, err := adminDB.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatalf("create schema failed: %v", err)
	}
	defer adminDB.Exec(`DROP SCHEMA ` + schema + ` CASCADE`)

	repo, err := Open("postgres", postgresURLWithSearchPath(t, postgresURL, schema), filepath.Join(projectRoot(t), "migrations"))
	if err != nil {
		t.Fatalf("postgres migrations failed: %v", err)
	}
	defer repo.Close()

	var count int
	if err := repo.DB().QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("schema_migrations count failed: %v", err)
	}
	if count != 20 {
		t.Fatalf("applied migration count = %d, want 20", count)
	}
}

func postgresURLWithSearchPath(t *testing.T, rawURL, schema string) string {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse POSTGRES_TEST_URL failed: %v", err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func projectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate project root")
		}
		dir = parent
	}
}
