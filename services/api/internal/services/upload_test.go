package services

import (
	"context"
	"database/sql"
	"io"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/migrate"
	"github.com/therealagt/ContractManagementTool/libs/common/pubsub"
)

type fakeStaging struct {
	objects map[string][]byte
}

func (f *fakeStaging) Upload(ctx context.Context, objectPath string, r io.Reader, _ string) error {
	_ = ctx
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	f.objects[objectPath] = data
	return nil
}

func (f *fakeStaging) FullPath(objectPath string) string {
	return "gs://test-bucket/" + objectPath
}

type fakePublisher struct {
	msgs []pubsub.ExtractionRequested
}

func (f *fakePublisher) PublishExtraction(ctx context.Context, msg pubsub.ExtractionRequested) error {
	_ = ctx
	f.msgs = append(f.msgs, msg)
	return nil
}

func TestUploadRejectsUnsignedPDF(t *testing.T) {
	db := openTestDB(t)
	staging := &fakeStaging{objects: map[string][]byte{}}
	publisher := &fakePublisher{}
	svc := NewUploadService(db, staging, publisher)

	pdf := []byte("%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")
	result, err := svc.Upload(context.Background(), UploadInput{
		Actor: "uploader@example.com",
		Type:  contracts.TypeNDA,
		PDF:   pdf,
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if result.Status != contracts.StatusRejected {
		t.Fatalf("status = %s", result.Status)
	}
	if len(publisher.msgs) != 0 {
		t.Fatal("expected no pubsub message for rejected upload")
	}
	if len(staging.objects) != 0 {
		t.Fatal("expected no gcs upload for rejected contract")
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
