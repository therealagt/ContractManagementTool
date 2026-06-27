package contracts

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/therealagt/ContractManagementTool/libs/common/migrate"
)

func TestRepositoryCreateAndStatus(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	repo := NewRepository(db)

	c, err := repo.CreateContract(ctx, TypeNDA, StatusExtracting, nil, "abc123", "uploader@example.com")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.UpdateStatus(ctx, c.ID, StatusPendingReview); err != nil {
		t.Fatalf("update status: %v", err)
	}

	detail, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if detail.Status != StatusPendingReview {
		t.Fatalf("status = %s", detail.Status)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+t.TempDir()+"/test.db?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrate.Run(context.Background(), db, true); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}
