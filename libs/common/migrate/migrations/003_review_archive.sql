-- Phase 3 schema: HITL confirmation and archive records

ALTER TABLE contracts ADD COLUMN IF NOT EXISTS confirmed_by VARCHAR(320);
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS confirmed_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS confirmed_metadata (
    contract_id VARCHAR(36) PRIMARY KEY REFERENCES contracts(id),
    metadata_json JSONB NOT NULL,
    confirmed_by VARCHAR(320) NOT NULL,
    confirmed_at TIMESTAMPTZ NOT NULL,
    diff_from_draft JSONB
);

CREATE TABLE IF NOT EXISTS archive_records (
    contract_id VARCHAR(36) PRIMARY KEY REFERENCES contracts(id),
    gcs_path VARCHAR(512) NOT NULL,
    sha256 VARCHAR(64) NOT NULL,
    retention_expires_at TIMESTAMPTZ NOT NULL,
    archived_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_contracts_confirmed_by ON contracts(confirmed_by);
