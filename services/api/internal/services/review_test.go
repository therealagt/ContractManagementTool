package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/migrate"
	"github.com/therealagt/ContractManagementTool/libs/common/pubsub"

	_ "modernc.org/sqlite"
)

func validNDAMetadata(t *testing.T, overrides map[string]any) json.RawMessage {
	t.Helper()
	base := map[string]any{
		"disclosing_party":            "Party A",
		"receiving_party":             "Party B",
		"effective_date":              "2024-01-01",
		"confidentiality_term_months": 12,
		"governing_law":               "Germany",
	}
	for k, v := range overrides {
		base[k] = v
	}
	raw, err := json.Marshal(base)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}
	return raw
}

type stubArchivePublisher struct {
	last pubsub.ArchiveRequested
}

func (s *stubArchivePublisher) PublishArchive(_ context.Context, msg pubsub.ArchiveRequested) error {
	s.last = msg
	return nil
}

func TestReviewConfirmEnforcesSOD(t *testing.T) {
	db := openReviewTestDB(t)
	ctx := context.Background()
	repo := contracts.NewRepository(db)

	c, err := repo.CreateContract(ctx, contracts.TypeNDA, contracts.StatusPendingReview, nil, "abc", "uploader@example.com")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	draft := &contracts.ExtractionDraft{
		ContractID:    c.ID,
		ExtractedJSON: json.RawMessage(`{"disclosing_party":"A"}`),
		SchemaVersion: "1",
		ExtractedAt:   time.Now().UTC(),
	}
	if err := repo.SaveExtractionDraft(ctx, draft); err != nil {
		t.Fatalf("draft: %v", err)
	}

	svc := NewReviewService(db, &stubArchivePublisher{})
	_, err = svc.Confirm(ctx, ConfirmInput{
		Actor:        "uploader@example.com",
		ContractID:   c.ID,
		MetadataJSON: validNDAMetadata(t, map[string]any{"disclosing_party": "A"}),
	})
	if !IsSODViolation(err) {
		t.Fatalf("expected SoD violation, got %v", err)
	}
}

func TestReviewConfirmPublishesArchive(t *testing.T) {
	db := openReviewTestDB(t)
	ctx := context.Background()
	repo := contracts.NewRepository(db)

	c, err := repo.CreateContract(ctx, contracts.TypeNDA, contracts.StatusPendingReview, nil, "abc", "uploader@example.com")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.SetGCSPathAndStatus(ctx, c.ID, "gs://bucket/staging/x.pdf", contracts.StatusPendingReview); err != nil {
		t.Fatalf("path: %v", err)
	}
	if err := repo.SaveExtractionDraft(ctx, &contracts.ExtractionDraft{
		ContractID:    c.ID,
		ExtractedJSON: json.RawMessage(`{"disclosing_party":"A"}`),
		SchemaVersion: "1",
		ExtractedAt:   time.Now().UTC(),
	}); err != nil {
		t.Fatalf("draft: %v", err)
	}

	pub := &stubArchivePublisher{}
	svc := NewReviewService(db, pub)
	result, err := svc.Confirm(ctx, ConfirmInput{
		Actor:        "reviewer@example.com",
		ContractID:   c.ID,
		MetadataJSON: validNDAMetadata(t, map[string]any{"disclosing_party": "B"}),
	})
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if result.Status != contracts.StatusConfirmed {
		t.Fatalf("status = %s", result.Status)
	}
	if pub.last.ContractID != c.ID {
		t.Fatalf("archive publish contract_id = %s", pub.last.ContractID)
	}
}

func openReviewTestDB(t *testing.T) *sql.DB {
	t.Helper()
	schemasDir, err := filepath.Abs("../../../../schemas")
	if err != nil {
		t.Fatalf("schemas path: %v", err)
	}
	t.Setenv("SCHEMAS_DIR", schemasDir)

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
