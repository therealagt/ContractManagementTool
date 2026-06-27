package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/therealagt/ContractManagementTool/libs/common/migrate"
	"github.com/therealagt/ContractManagementTool/services/api/internal/config"
)

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
	return migrate.Run(ctx, db, settings.IsSQLite())
}
