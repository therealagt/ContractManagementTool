package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/therealagt/ContractManagementTool/services/api/internal/config"
)

const migrationSQL = `
CREATE TABLE IF NOT EXISTS access_events (
    id VARCHAR(36) PRIMARY KEY,
    actor VARCHAR(320) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    resource_id VARCHAR(64),
    action VARCHAR(64) NOT NULL,
    ip VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_events (
    id VARCHAR(36) PRIMARY KEY,
    contract_id VARCHAR(64),
    actor VARCHAR(320) NOT NULL,
    action VARCHAR(64) NOT NULL,
    payload_json JSONB,
    prev_event_hash VARCHAR(64),
    event_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(32) PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_access_events_actor ON access_events(actor);
CREATE INDEX IF NOT EXISTS idx_audit_events_contract ON audit_events(contract_id);
`

const sqliteMigrationSQL = `
CREATE TABLE IF NOT EXISTS access_events (
    id TEXT PRIMARY KEY,
    actor TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    action TEXT NOT NULL,
    ip TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_events (
    id TEXT PRIMARY KEY,
    contract_id TEXT,
    actor TEXT NOT NULL,
    action TEXT NOT NULL,
    payload_json TEXT,
    prev_event_hash TEXT,
    event_hash TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_access_events_actor ON access_events(actor);
CREATE INDEX IF NOT EXISTS idx_audit_events_contract ON audit_events(contract_id);
`

func Open(settings *config.Settings) (*sql.DB, error) {
	driver := "pgx"
	dsn := settings.DatabaseDSN()
	if settings.IsSQLite() {
		driver = "sqlite"
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

func RunMigrations(ctx context.Context, db *sql.DB, settings *config.Settings) error {
	sqlText := migrationSQL
	if settings.IsSQLite() {
		sqlText = sqliteMigrationSQL
	}

	for _, statement := range strings.Split(sqlText, ";") {
		stmt := strings.TrimSpace(statement)
		if stmt == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("run migration: %w", err)
		}
	}

	if settings.IsSQLite() {
		return nil
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		"001_initial", time.Now().UTC(),
	)
	return err
}
