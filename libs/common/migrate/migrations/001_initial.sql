-- Phase 1 schema: audit and access event tables

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
