package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"inhumas-em-foco/internal/repository"

	_ "modernc.org/sqlite"
)

var orderedTables = []string{
	"users",
	"password_reset_tokens",
	"categories",
	"tags",
	"posts",
	"post_tags",
	"post_revisions",
	"media_assets",
	"portal_settings",
	"automation_sources",
	"automation_runs",
	"ai_usage_logs",
	"slug_redirects",
	"neighborhoods",
	"stores",
	"promotions",
	"events",
	"classifieds",
	"influencers",
	"banners",
	"jobs",
	"dead_jobs",
	"metrics",
	"audit_logs",
	"edit_locks",
	"login_attempts",
}

var booleanColumns = map[string]map[string]bool{
	"users":                 {"active": true},
	"categories":            {"active": true, "requires_editorial_notes": true},
	"tags":                  {"active": true},
	"posts":                 {"is_sponsored": true, "is_featured": true, "is_pinned": true},
	"media_assets":          {},
	"portal_settings":       {"automation_enabled": true},
	"automation_sources":    {"active": true},
	"stores":                {"is_sponsored": true, "is_featured": true, "active": true},
	"promotions":            {"is_sponsored": true},
	"events":                {"is_featured": true, "is_sponsored": true},
	"classifieds":           {"is_featured": true, "is_sponsored": true},
	"influencers":           {"is_featured": true, "is_sponsored": true, "active": true},
	"banners":               {"active": true},
	"metrics":               {},
	"login_attempts":        {"success": true},
	"password_reset_tokens": {},
}

func main() {
	sqliteURL := firstNonEmpty(os.Getenv("SQLITE_DATABASE_URL"), os.Getenv("SQLITE_DB"), "./inhumas.db")
	postgresURL := firstNonEmpty(os.Getenv("POSTGRES_DATABASE_URL"), os.Getenv("DATABASE_URL"))
	if !strings.HasPrefix(strings.ToLower(postgresURL), "postgres") {
		log.Fatal("configure POSTGRES_DATABASE_URL com a URL do PostgreSQL de destino")
	}
	migrationsDir := firstNonEmpty(os.Getenv("MIGRATIONS_DIR"), "./migrations")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	sqliteDB, err := sql.Open("sqlite", sqliteURL+"?_pragma=query_only(1)")
	if err != nil {
		log.Fatal(err)
	}
	defer sqliteDB.Close()

	pgRepo, err := repository.Open("postgres", postgresURL, migrationsDir)
	if err != nil {
		log.Fatal(err)
	}
	defer pgRepo.Close()
	pgDB := pgRepo.DB()

	for _, table := range orderedTables {
		if err := copyTable(ctx, sqliteDB, pgDB, table); err != nil {
			log.Fatalf("falha ao migrar %s: %v", table, err)
		}
		if err := resetSequence(ctx, pgDB, table); err != nil {
			log.Fatalf("falha ao ajustar sequence de %s: %v", table, err)
		}
	}
	if err := validateCounts(ctx, sqliteDB, pgDB); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Migracao SQLite -> PostgreSQL concluida e validada.")
}

func copyTable(ctx context.Context, source, target *sql.DB, table string) error {
	cols, err := sqliteColumns(ctx, source, table)
	if err != nil {
		return err
	}
	if len(cols) == 0 {
		return nil
	}
	rows, err := source.QueryContext(ctx, `SELECT `+joinQuoted(cols)+` FROM `+quoteIdent(table))
	if err != nil {
		return err
	}
	defer rows.Close()

	tx, err := target.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertSQL := postgresInsertSQL(table, cols)
	count := 0
	for rows.Next() {
		values := make([]any, len(cols))
		dest := make([]any, len(cols))
		for i := range values {
			dest[i] = &values[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return err
		}
		normalizeValues(table, cols, values)
		if _, err := tx.ExecContext(ctx, insertSQL, values...); err != nil {
			return err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Printf("%s: %d registro(s)\n", table, count)
	return nil
}

func sqliteColumns(ctx context.Context, db *sql.DB, table string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(`+quoteIdent(table)+`)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		cols = append(cols, name)
	}
	return cols, rows.Err()
}

func postgresInsertSQL(table string, cols []string) string {
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	return `INSERT INTO ` + quoteIdent(table) + ` (` + joinQuoted(cols) + `) VALUES (` + strings.Join(placeholders, ",") + `) ON CONFLICT DO NOTHING`
}

func validateCounts(ctx context.Context, source, target *sql.DB) error {
	for _, table := range orderedTables {
		var sourceCount, targetCount int
		if err := source.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+quoteIdent(table)).Scan(&sourceCount); err != nil {
			return err
		}
		if err := target.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+quoteIdent(table)).Scan(&targetCount); err != nil {
			return err
		}
		if sourceCount != targetCount {
			return fmt.Errorf("contagem divergente em %s: sqlite=%d postgres=%d", table, sourceCount, targetCount)
		}
	}
	return nil
}

func resetSequence(ctx context.Context, db *sql.DB, table string) error {
	if !tableHasID(table) {
		return nil
	}
	query := fmt.Sprintf(`SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE((SELECT MAX(id) FROM %s), 1), true)`, table, quoteIdent(table))
	_, err := db.ExecContext(ctx, query)
	return err
}

func normalizeValues(table string, cols []string, values []any) {
	boolSet := booleanColumns[table]
	for i, col := range cols {
		if boolSet[col] {
			values[i] = sqliteBool(values[i])
		}
		if raw, ok := values[i].([]byte); ok {
			values[i] = string(raw)
		}
	}
}

func sqliteBool(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case int:
		return v != 0
	case []byte:
		return string(v) == "1" || strings.EqualFold(string(v), "true")
	case string:
		return v == "1" || strings.EqualFold(v, "true")
	default:
		return false
	}
}

func tableHasID(table string) bool {
	switch table {
	case "post_tags", "portal_settings", "schema_migrations", "edit_locks", "slug_redirects":
		return false
	default:
		return true
	}
}

func quoteIdent(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func joinQuoted(cols []string) string {
	quoted := make([]string, len(cols))
	for i, col := range cols {
		quoted[i] = quoteIdent(col)
	}
	return strings.Join(quoted, ",")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func init() {
	if wd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(wd, "migrations")); err == nil && os.Getenv("MIGRATIONS_DIR") == "" {
			_ = os.Setenv("MIGRATIONS_DIR", filepath.Join(wd, "migrations"))
		}
	}
}
