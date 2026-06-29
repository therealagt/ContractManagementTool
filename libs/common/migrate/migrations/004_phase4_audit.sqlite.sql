-- Phase 4 schema: legal hold, integrity checks, alert events

CREATE TABLE IF NOT EXISTS legal_holds (
    contract_id TEXT PRIMARY KEY REFERENCES contracts(id),
    reason TEXT NOT NULL,
    placed_by TEXT NOT NULL,
    placed_at TEXT NOT NULL,
    released_at TEXT
);

CREATE TABLE IF NOT EXISTS integrity_check_runs (
    id TEXT PRIMARY KEY,
    checked_count INTEGER NOT NULL,
    failed_count INTEGER NOT NULL,
    chain_valid INTEGER NOT NULL,
    started_at TEXT NOT NULL,
    completed_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS alert_events (
    id TEXT PRIMARY KEY,
    severity TEXT NOT NULL,
    source TEXT NOT NULL,
    payload_json TEXT,
    incident_id TEXT,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_legal_holds_active ON legal_holds(released_at);
CREATE INDEX IF NOT EXISTS idx_alert_events_created ON alert_events(created_at);
