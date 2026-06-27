-- Phase 2 schema: contracts, signature validation, extraction drafts
-- Canonical migrations live in libs/common/migrate/migrations/

CREATE TABLE IF NOT EXISTS contracts (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(16) NOT NULL,
    status VARCHAR(32) NOT NULL,
    partner_id VARCHAR(128),
    gcs_staging_path VARCHAR(512),
    sha256 VARCHAR(64) NOT NULL,
    uploaded_by VARCHAR(320) NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS signature_validation (
    contract_id VARCHAR(36) PRIMARY KEY REFERENCES contracts(id),
    is_valid BOOLEAN NOT NULL,
    signer_cn VARCHAR(320),
    signed_at TIMESTAMPTZ,
    cert_issuer VARCHAR(512),
    validation_result_json JSONB,
    validated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS extraction_drafts (
    contract_id VARCHAR(36) PRIMARY KEY REFERENCES contracts(id),
    extracted_json JSONB NOT NULL,
    gemini_model VARCHAR(64),
    prompt_version VARCHAR(32),
    schema_version VARCHAR(32) NOT NULL,
    confidence_flags JSONB,
    extracted_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_contracts_status ON contracts(status);
CREATE INDEX IF NOT EXISTS idx_contracts_uploaded_by ON contracts(uploaded_by);
