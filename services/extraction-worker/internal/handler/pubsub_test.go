package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/migrate"
	"github.com/therealagt/ContractManagementTool/libs/common/pubsub"
	"github.com/therealagt/ContractManagementTool/libs/common/vertex"
	"github.com/therealagt/ContractManagementTool/services/extraction-worker/internal/pipeline"
)

func TestPubSubPushDecoding(t *testing.T) {
	db := openTestDB(t)
	loader := &fakeLoader{data: []byte("%PDF")}
	extractor := &fakeExtractor{}
	p := pipeline.New(db, loader, extractor)

	contract, err := contracts.NewRepository(db).CreateContract(
		context.Background(), contracts.TypeNDA, contracts.StatusExtracting, nil, "hash", "u@example.com",
	)
	if err != nil {
		t.Fatalf("create contract: %v", err)
	}
	stagingPath := "gs://bucket/staging/" + contract.ID + ".pdf"
	if err := contracts.NewRepository(db).SetGCSPathAndStatus(
		context.Background(), contract.ID, stagingPath, contracts.StatusExtracting,
	); err != nil {
		t.Fatalf("set path: %v", err)
	}

	msg := pubsub.ExtractionRequested{
		ContractID:    contract.ID,
		Type:          "nda",
		GCSPath:       stagingPath,
		SchemaVersion: "nda/v1",
	}
	payload, _ := json.Marshal(msg)
	envelope := map[string]any{
		"message": map[string]any{
			"data": base64.StdEncoding.EncodeToString(payload),
		},
	}
	body, _ := json.Marshal(envelope)

	h := New(p)
	req := httptest.NewRequest(http.MethodPost, "/pubsub/extraction", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	detail, err := contracts.NewRepository(db).GetByID(context.Background(), contract.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if detail.Status != contracts.StatusPendingReview {
		t.Fatalf("status = %s", detail.Status)
	}
}

type fakeLoader struct{ data []byte }

func (f *fakeLoader) Download(ctx context.Context, objectPath string) ([]byte, error) {
	_ = ctx
	_ = objectPath
	return f.data, nil
}

type fakeExtractor struct{}

func (f *fakeExtractor) ExtractWithConfidence(ctx context.Context, contractType contracts.Type, pdf []byte) (*vertex.ExtractionResult, []byte, error) {
	_ = ctx
	_ = contractType
	_ = pdf
	return &vertex.ExtractionResult{
		ExtractedJSON: []byte(`{"disclosing_party":"A","receiving_party":"B","effective_date":"2024-01-01","confidentiality_term_months":12,"governing_law":"DE"}`),
		Model:         "test",
		PromptVersion: "v1",
		SchemaVersion: "nda/v1",
	}, []byte(`{"disclosing_party":"high"}`), nil
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
