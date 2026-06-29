package pipeline_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/therealagt/ContractManagementTool/libs/common/alerts"
	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/gcs"
	"github.com/therealagt/ContractManagementTool/libs/common/migrate"
	"github.com/therealagt/ContractManagementTool/libs/common/monitoring"
	"github.com/therealagt/ContractManagementTool/services/integrity-cron/internal/pipeline"

	_ "modernc.org/sqlite"
	"database/sql"
)

type memArchive struct {
	objects map[string][]byte
}

func (m *memArchive) Download(_ context.Context, objectPath string) ([]byte, error) {
	data, ok := m.objects[objectPath]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func TestIntegrityCheckDetectsMismatch(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	repo := contracts.NewRepository(db)

	pdf := []byte("%PDF-1.4 integrity test")
	sum := sha256.Sum256(pdf)
	hash := hex.EncodeToString(sum[:])

	c, err := repo.CreateContract(ctx, contracts.TypeNDA, contracts.StatusArchived, nil, hash, "uploader@example.com")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	gcsPath := "gs://bucket/archive/" + c.ID + ".pdf"
	if err := repo.SaveArchiveRecord(ctx, &contracts.ArchiveRecord{
		ContractID:         c.ID,
		GCSPath:            gcsPath,
		SHA256:             hash,
		RetentionExpiresAt: time.Now().UTC().AddDate(10, 0, 0),
		ArchivedAt:         time.Now().UTC(),
	}); err != nil {
		t.Fatalf("archive record: %v", err)
	}

	archive := &memArchive{objects: map[string][]byte{
		gcs.ArchiveObjectPath(c.ID): []byte("tampered"),
	}}

	rec := alerts.NewRecorder(db, nil)
	p := pipeline.New(db, archive, rec, monitoring.NoopPublisher{}, 7)
	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.FailedCount != 1 {
		t.Fatalf("expected 1 failure, got %d", result.FailedCount)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", "file:"+dsn+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := migrate.Run(context.Background(), db, true); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
