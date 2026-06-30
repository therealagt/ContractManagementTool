-- Worker least-privilege grants (login roles provisioned via Terraform)

GRANT USAGE ON SCHEMA public TO contract_ingestion;
GRANT SELECT, UPDATE ON contracts TO contract_ingestion;
GRANT SELECT ON signature_validation TO contract_ingestion;
GRANT INSERT, UPDATE ON extraction_drafts TO contract_ingestion;
GRANT INSERT ON audit_events TO contract_ingestion;

GRANT USAGE ON SCHEMA public TO contract_archive;
GRANT SELECT, UPDATE ON contracts TO contract_archive;
GRANT SELECT ON signature_validation TO contract_archive;
GRANT SELECT ON extraction_drafts TO contract_archive;
GRANT SELECT ON confirmed_metadata TO contract_archive;
GRANT SELECT ON archive_records TO contract_archive;
GRANT SELECT ON legal_holds TO contract_archive;
GRANT INSERT ON archive_records TO contract_archive;
GRANT INSERT ON audit_events TO contract_archive;

GRANT USAGE ON SCHEMA public TO contract_integrity;
GRANT SELECT ON contracts TO contract_integrity;
GRANT SELECT ON archive_records TO contract_integrity;
GRANT SELECT ON audit_events TO contract_integrity;
GRANT INSERT ON audit_events TO contract_integrity;
GRANT INSERT ON integrity_check_runs TO contract_integrity;
GRANT INSERT ON alert_events TO contract_integrity;

GRANT USAGE ON SCHEMA public TO contract_report;
GRANT SELECT ON contracts TO contract_report;
GRANT SELECT ON signature_validation TO contract_report;
GRANT SELECT ON extraction_drafts TO contract_report;
GRANT SELECT ON confirmed_metadata TO contract_report;
GRANT SELECT ON archive_records TO contract_report;
GRANT SELECT ON legal_holds TO contract_report;
GRANT SELECT ON integrity_check_runs TO contract_report;
GRANT SELECT ON alert_events TO contract_report;
GRANT SELECT ON audit_events TO contract_report;
