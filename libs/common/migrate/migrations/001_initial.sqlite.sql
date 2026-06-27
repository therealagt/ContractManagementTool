CREATE TABLE IF NOT EXISTS access_events (
    id TEXT PRIMARY KEY,
    actor TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    action TEXT NOT NULL,
    ip TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_events (
    id TEXT PRIMARY KEY,
    contract_id TEXT,
    actor TEXT NOT NULL,
    action TEXT NOT NULL,
    payload_json TEXT,
    prev_event_hash TEXT,
    event_hash TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_access_events_actor ON access_events(actor);
CREATE INDEX IF NOT EXISTS idx_audit_events_contract ON audit_events(contract_id);
