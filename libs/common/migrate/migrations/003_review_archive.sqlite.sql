-- Phase 3 schema: HITL confirmation and archive records

ALTER TABLE contracts ADD COLUMN confirmed_by TEXT;
ALTER TABLE contracts ADD COLUMN confirmed_at TEXT;

CREATE TABLE IF NOT EXISTS confirmed_metadata (
    contract_id TEXT PRIMARY KEY REFERENCES contracts(id),
    metadata_json TEXT NOT NULL,
    confirmed_by TEXT NOT NULL,
    confirmed_at TEXT NOT NULL,
    diff_from_draft TEXT
);

CREATE TABLE IF NOT EXISTS archive_records (
    contract_id TEXT PRIMARY KEY REFERENCES contracts(id),
    gcs_path TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    retention_expires_at TEXT NOT NULL,
    archived_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_contracts_confirmed_by ON contracts(confirmed_by);
