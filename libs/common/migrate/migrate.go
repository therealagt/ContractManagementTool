package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"
	"time"
)

//go:embed migrations/*
var migrationFS embed.FS

var migrationVersions = []string{
	"001_initial",
	"002_contracts",
}

func Run(ctx context.Context, db *sql.DB, sqlite bool) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	for _, version := range migrationVersions {
		applied, err := isApplied(ctx, db, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		filename := version + ".sql"
		if sqlite {
			filename = version + ".sqlite.sql"
		}
		content, err := migrationFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}

		if err := execStatements(ctx, db, string(content)); err != nil {
			return fmt.Errorf("migration %s: %w", version, err)
		}

		if err := markApplied(ctx, db, version); err != nil {
			return err
		}
	}
	return nil
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `SELECT 1 FROM schema_migrations LIMIT 1`)
	if err == nil {
		return nil
	}
	// First boot before 001: Run all migrations from scratch via version loop.
	return nil
}

func isApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM schema_migrations WHERE version = $1`, version,
	).Scan(&count)
	if err != nil {
		if strings.Contains(err.Error(), "no such table") ||
			strings.Contains(err.Error(), "does not exist") {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func markApplied(ctx context.Context, db *sql.DB, version string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		version, time.Now().UTC(),
	)
	return err
}

func execStatements(ctx context.Context, db *sql.DB, sqlText string) error {
	for _, statement := range strings.Split(sqlText, ";") {
		stmt := strings.TrimSpace(stripSQLComments(statement))
		if stmt == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func stripSQLComments(sql string) string {
	lines := strings.Split(sql, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}
