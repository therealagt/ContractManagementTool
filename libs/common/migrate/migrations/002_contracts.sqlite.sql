CREATE TABLE IF NOT EXISTS contracts (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    partner_id TEXT,
    gcs_staging_path TEXT,
    sha256 TEXT NOT NULL,
    uploaded_by TEXT NOT NULL,
    uploaded_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS signature_validation (
    contract_id TEXT PRIMARY KEY REFERENCES contracts(id),
    is_valid INTEGER NOT NULL,
    signer_cn TEXT,
    signed_at TEXT,
    cert_issuer TEXT,
    validation_result_json TEXT,
    validated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS extraction_drafts (
    contract_id TEXT PRIMARY KEY REFERENCES contracts(id),
    extracted_json TEXT NOT NULL,
    gemini_model TEXT,
    prompt_version TEXT,
    schema_version TEXT NOT NULL,
    confidence_flags TEXT,
    extracted_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_contracts_status ON contracts(status);
CREATE INDEX IF NOT EXISTS idx_contracts_uploaded_by ON contracts(uploaded_by);
