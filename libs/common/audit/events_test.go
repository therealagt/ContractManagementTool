package audit

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestRecordAuditEventChainsHashes(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	first, err := RecordAuditEvent(ctx, db, "actor@example.com", "upload", nil, map[string]any{"step": 1}, nil)
	if err != nil {
		t.Fatalf("first event: %v", err)
	}
	if !first.EventHash.Valid || first.EventHash.String == "" {
		t.Fatal("expected event hash on first audit event")
	}

	second, err := RecordAuditEvent(ctx, db, "actor@example.com", "confirm", nil, map[string]any{"step": 2}, nil)
	if err != nil {
		t.Fatalf("second event: %v", err)
	}
	if !second.PrevEventHash.Valid || second.PrevEventHash.String != first.EventHash.String {
		t.Fatalf("prev hash = %v, want %s", second.PrevEventHash, first.EventHash.String)
	}
}

func TestRecordAuditEventRejectsStalePrevHash(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	_, err := RecordAuditEvent(ctx, db, "actor@example.com", "upload", nil, nil, nil)
	if err != nil {
		t.Fatalf("seed event: %v", err)
	}

	stale := "deadbeef"
	_, err = RecordAuditEvent(ctx, db, "actor@example.com", "confirm", nil, nil, &stale)
	if err != ErrAuditChain {
		t.Fatalf("err = %v, want ErrAuditChain", err)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+t.TempDir()+"/test.db?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	schema := `
CREATE TABLE access_events (
    id TEXT PRIMARY KEY,
    actor TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    action TEXT NOT NULL,
    ip TEXT,
    created_at TEXT NOT NULL
);
CREATE TABLE audit_events (
    id TEXT PRIMARY KEY,
    contract_id TEXT,
    actor TEXT NOT NULL,
    action TEXT NOT NULL,
    payload_json TEXT,
    prev_event_hash TEXT,
    event_hash TEXT,
    created_at TEXT NOT NULL
);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}
