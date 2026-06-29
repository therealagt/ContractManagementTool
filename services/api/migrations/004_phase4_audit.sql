-- Phase 4 schema: legal hold, integrity checks, alert events
-- Canonical migrations live in libs/common/migrate/migrations/

CREATE TABLE IF NOT EXISTS legal_holds (
    contract_id VARCHAR(36) PRIMARY KEY REFERENCES contracts(id),
    reason TEXT NOT NULL,
    placed_by VARCHAR(320) NOT NULL,
    placed_at TIMESTAMPTZ NOT NULL,
    released_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS integrity_check_runs (
    id VARCHAR(36) PRIMARY KEY,
    checked_count INT NOT NULL,
    failed_count INT NOT NULL,
    chain_valid BOOLEAN NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS alert_events (
    id VARCHAR(36) PRIMARY KEY,
    severity VARCHAR(8) NOT NULL,
    source VARCHAR(64) NOT NULL,
    payload_json JSONB,
    incident_id VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_legal_holds_active ON legal_holds(released_at);
CREATE INDEX IF NOT EXISTS idx_alert_events_created ON alert_events(created_at);
