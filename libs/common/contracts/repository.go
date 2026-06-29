package contracts

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateContract(
	ctx context.Context,
	contractType Type,
	status Status,
	partnerID *string,
	sha256, uploadedBy string,
) (*Contract, error) {
	c := &Contract{
		ID:         uuid.NewString(),
		Type:       contractType,
		Status:     status,
		SHA256:     sha256,
		UploadedBy: uploadedBy,
		UploadedAt: time.Now().UTC(),
	}
	if partnerID != nil {
		c.PartnerID = sql.NullString{String: *partnerID, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO contracts (id, type, status, partner_id, gcs_staging_path, sha256, uploaded_by, uploaded_at, confirmed_by, confirmed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		c.ID, c.Type, c.Status, c.PartnerID, c.GCSStagingPath, c.SHA256, c.UploadedBy, c.UploadedAt, c.ConfirmedBy, c.ConfirmedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert contract: %w", err)
	}
	return c, nil
}

func (r *Repository) SaveSignatureValidation(ctx context.Context, v *SignatureValidation) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO signature_validation (contract_id, is_valid, signer_cn, signed_at, cert_issuer, validation_result_json, validated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		v.ContractID, v.IsValid, v.SignerCN, v.SignedAt, v.CertIssuer, v.ValidationResultJSON, v.ValidatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert signature_validation: %w", err)
	}
	return nil
}

func (r *Repository) SetGCSPathAndStatus(ctx context.Context, contractID, gcsPath string, status Status) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE contracts SET gcs_staging_path = $1, status = $2 WHERE id = $3`,
		gcsPath, status, contractID,
	)
	if err != nil {
		return fmt.Errorf("update contract path/status: %w", err)
	}
	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, contractID string, status Status) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE contracts SET status = $1 WHERE id = $2`,
		status, contractID,
	)
	if err != nil {
		return fmt.Errorf("update contract status: %w", err)
	}
	return nil
}

func (r *Repository) SaveExtractionDraft(ctx context.Context, d *ExtractionDraft) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO extraction_drafts (contract_id, extracted_json, gemini_model, prompt_version, schema_version, confidence_flags, extracted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (contract_id) DO UPDATE SET
		   extracted_json = EXCLUDED.extracted_json,
		   gemini_model = EXCLUDED.gemini_model,
		   prompt_version = EXCLUDED.prompt_version,
		   schema_version = EXCLUDED.schema_version,
		   confidence_flags = EXCLUDED.confidence_flags,
		   extracted_at = EXCLUDED.extracted_at`,
		d.ContractID, d.ExtractedJSON, d.GeminiModel, d.PromptVersion, d.SchemaVersion, d.ConfidenceFlags, d.ExtractedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert extraction_draft: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, contractID string) (*ContractDetail, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, type, status, partner_id, gcs_staging_path, sha256, uploaded_by, uploaded_at, confirmed_by, confirmed_at
		 FROM contracts WHERE id = $1`, contractID,
	)
	c, err := scanContract(row)
	if err != nil {
		return nil, err
	}

	detail := &ContractDetail{Contract: *c}
	detail.Signature, _ = r.getSignature(ctx, contractID)
	detail.Draft, _ = r.getDraft(ctx, contractID)
	detail.Confirmed, _ = r.getConfirmed(ctx, contractID)
	detail.Archive, _ = r.getArchive(ctx, contractID)
	detail.LegalHold, _ = r.getActiveLegalHold(ctx, contractID)
	return detail, nil
}

type ReviewQueueItem struct {
	Contract
	HasDraft bool
}

func (r *Repository) ListPendingReview(ctx context.Context, excludeUploader string, limit int) ([]ReviewQueueItem, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT c.id, c.type, c.status, c.partner_id, c.gcs_staging_path, c.sha256, c.uploaded_by, c.uploaded_at,
		        c.confirmed_by, c.confirmed_at,
		        CASE WHEN d.contract_id IS NOT NULL THEN 1 ELSE 0 END AS has_draft
		 FROM contracts c
		 LEFT JOIN extraction_drafts d ON d.contract_id = c.id
		 WHERE c.status = $1 AND c.uploaded_by <> $2
		 ORDER BY c.uploaded_at ASC
		 LIMIT $3`,
		StatusPendingReview, excludeUploader, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list pending review: %w", err)
	}
	defer rows.Close()

	var items []ReviewQueueItem
	for rows.Next() {
		var item ReviewQueueItem
		var contractType, status string
		var uploadedAt any
		var hasDraft int
		if err := rows.Scan(
			&item.ID, &contractType, &status, &item.PartnerID, &item.GCSStagingPath, &item.SHA256,
			&item.UploadedBy, &uploadedAt, &item.ConfirmedBy, &item.ConfirmedAt, &hasDraft,
		); err != nil {
			return nil, fmt.Errorf("scan review queue: %w", err)
		}
		item.Type = Type(contractType)
		item.Status = Status(status)
		item.UploadedAt = scanTime(uploadedAt)
		item.HasDraft = hasDraft == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ConfirmContract(
	ctx context.Context,
	contractID, confirmedBy string,
	metadataJSON, diffJSON json.RawMessage,
) error {
	now := time.Now().UTC()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`UPDATE contracts SET status = $1, confirmed_by = $2, confirmed_at = $3
		 WHERE id = $4 AND status = $5`,
		StatusConfirmed, confirmedBy, now, contractID, StatusPendingReview,
	)
	if err != nil {
		return fmt.Errorf("update contract confirmed: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO confirmed_metadata (contract_id, metadata_json, confirmed_by, confirmed_at, diff_from_draft)
		 VALUES ($1, $2, $3, $4, $5)`,
		contractID, metadataJSON, confirmedBy, now, diffJSON,
	)
	if err != nil {
		return fmt.Errorf("insert confirmed_metadata: %w", err)
	}
	return tx.Commit()
}

func (r *Repository) RejectContract(ctx context.Context, contractID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE contracts SET status = $1 WHERE id = $2 AND status = $3`,
		StatusRejected, contractID, StatusPendingReview,
	)
	if err != nil {
		return fmt.Errorf("reject contract: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) SaveArchiveRecord(ctx context.Context, rec *ArchiveRecord) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO archive_records (contract_id, gcs_path, sha256, retention_expires_at, archived_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (contract_id) DO UPDATE SET
		   gcs_path = EXCLUDED.gcs_path,
		   sha256 = EXCLUDED.sha256,
		   retention_expires_at = EXCLUDED.retention_expires_at,
		   archived_at = EXCLUDED.archived_at`,
		rec.ContractID, rec.GCSPath, rec.SHA256, rec.RetentionExpiresAt, rec.ArchivedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert archive_record: %w", err)
	}
	return nil
}

func (r *Repository) getSignature(ctx context.Context, contractID string) (*SignatureValidation, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT contract_id, is_valid, signer_cn, signed_at, cert_issuer, validation_result_json, validated_at
		 FROM signature_validation WHERE contract_id = $1`, contractID,
	)
	var v SignatureValidation
	var result []byte
	err := row.Scan(&v.ContractID, &v.IsValid, &v.SignerCN, &v.SignedAt, &v.CertIssuer, &result, &v.ValidatedAt)
	if err != nil {
		return nil, err
	}
	v.ValidationResultJSON = result
	return &v, nil
}

func (r *Repository) getDraft(ctx context.Context, contractID string) (*ExtractionDraft, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT contract_id, extracted_json, gemini_model, prompt_version, schema_version, confidence_flags, extracted_at
		 FROM extraction_drafts WHERE contract_id = $1`, contractID,
	)
	var d ExtractionDraft
	var extracted, flags []byte
	var extractedAt any
	err := row.Scan(&d.ContractID, &extracted, &d.GeminiModel, &d.PromptVersion, &d.SchemaVersion, &flags, &extractedAt)
	if err != nil {
		return nil, err
	}
	d.ExtractedJSON = extracted
	d.ConfidenceFlags = flags
	d.ExtractedAt = scanTime(extractedAt)
	return &d, nil
}

func (r *Repository) getConfirmed(ctx context.Context, contractID string) (*ConfirmedMetadata, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT contract_id, metadata_json, confirmed_by, confirmed_at, diff_from_draft
		 FROM confirmed_metadata WHERE contract_id = $1`, contractID,
	)
	var m ConfirmedMetadata
	var metadata, diff []byte
	var confirmedAt any
	err := row.Scan(&m.ContractID, &metadata, &m.ConfirmedBy, &confirmedAt, &diff)
	if err != nil {
		return nil, err
	}
	m.MetadataJSON = metadata
	m.ConfirmedAt = scanTime(confirmedAt)
	if len(diff) > 0 {
		m.DiffFromDraft = diff
	}
	return &m, nil
}

func (r *Repository) getArchive(ctx context.Context, contractID string) (*ArchiveRecord, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT contract_id, gcs_path, sha256, retention_expires_at, archived_at
		 FROM archive_records WHERE contract_id = $1`, contractID,
	)
	var rec ArchiveRecord
	var retentionAt, archivedAt any
	err := row.Scan(&rec.ContractID, &rec.GCSPath, &rec.SHA256, &retentionAt, &archivedAt)
	if err != nil {
		return nil, err
	}
	rec.RetentionExpiresAt = scanTime(retentionAt)
	rec.ArchivedAt = scanTime(archivedAt)
	return &rec, nil
}

func scanContract(row *sql.Row) (*Contract, error) {
	var c Contract
	var contractType, status string
	var uploadedAt any
	var confirmedAt any
	err := row.Scan(
		&c.ID, &contractType, &status, &c.PartnerID, &c.GCSStagingPath, &c.SHA256,
		&c.UploadedBy, &uploadedAt, &c.ConfirmedBy, &confirmedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("scan contract: %w", err)
	}
	c.Type = Type(contractType)
	c.Status = Status(status)
	c.UploadedAt = scanTime(uploadedAt)
	if t := scanTime(confirmedAt); !t.IsZero() {
		c.ConfirmedAt = sql.NullTime{Time: t, Valid: true}
	}
	return &c, nil
}

func scanTime(v any) time.Time {
	switch t := v.(type) {
	case time.Time:
		return t
	case string:
		parsed, err := time.Parse(time.RFC3339Nano, t)
		if err != nil {
			parsed, _ = time.Parse(time.RFC3339, t)
		}
		return parsed
	case []byte:
		parsed, err := time.Parse(time.RFC3339Nano, string(t))
		if err != nil {
			parsed, _ = time.Parse(time.RFC3339, string(t))
		}
		return parsed
	default:
		return time.Time{}
	}
}

func ValidationResultJSON(result any) json.RawMessage {
	b, _ := json.Marshal(result)
	return b
}

func (r *Repository) ListArchiveRecords(ctx context.Context) ([]ArchiveRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT contract_id, gcs_path, sha256, retention_expires_at, archived_at FROM archive_records`,
	)
	if err != nil {
		return nil, fmt.Errorf("list archive records: %w", err)
	}
	defer rows.Close()

	var records []ArchiveRecord
	for rows.Next() {
		var rec ArchiveRecord
		var retentionAt, archivedAt any
		if err := rows.Scan(&rec.ContractID, &rec.GCSPath, &rec.SHA256, &retentionAt, &archivedAt); err != nil {
			return nil, err
		}
		rec.RetentionExpiresAt = scanTime(retentionAt)
		rec.ArchivedAt = scanTime(archivedAt)
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (r *Repository) ListArchivedContractIDs(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT contract_id FROM archive_records ORDER BY archived_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) CountByStatus(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT status, COUNT(1) FROM contracts GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		out[status] = count
	}
	return out, rows.Err()
}

func (r *Repository) CountPendingReviewBeyondSLA(ctx context.Context, slaDays int) (int, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -slaDays)
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM contracts WHERE status = $1 AND uploaded_at < $2`,
		StatusPendingReview, cutoff,
	).Scan(&count)
	return count, err
}

func (r *Repository) PlaceLegalHold(ctx context.Context, contractID, reason, placedBy string) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO legal_holds (contract_id, reason, placed_by, placed_at, released_at)
		 VALUES ($1, $2, $3, $4, NULL)
		 ON CONFLICT (contract_id) DO UPDATE SET
		   reason = EXCLUDED.reason,
		   placed_by = EXCLUDED.placed_by,
		   placed_at = EXCLUDED.placed_at,
		   released_at = NULL`,
		contractID, reason, placedBy, now,
	)
	if err != nil {
		return fmt.Errorf("place legal hold: %w", err)
	}
	return nil
}

func (r *Repository) ReleaseLegalHold(ctx context.Context, contractID string) error {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx,
		`UPDATE legal_holds SET released_at = $1 WHERE contract_id = $2 AND released_at IS NULL`,
		now, contractID,
	)
	if err != nil {
		return fmt.Errorf("release legal hold: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) CountActiveLegalHolds(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM legal_holds WHERE released_at IS NULL`,
	).Scan(&count)
	return count, err
}

func (r *Repository) SaveIntegrityCheckRun(ctx context.Context, run *IntegrityCheckRun) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO integrity_check_runs (id, checked_count, failed_count, chain_valid, started_at, completed_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		run.ID, run.CheckedCount, run.FailedCount, run.ChainValid, run.StartedAt, run.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("save integrity check run: %w", err)
	}
	return nil
}

func (r *Repository) GetLatestIntegrityCheckRun(ctx context.Context) (*IntegrityCheckRun, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, checked_count, failed_count, chain_valid, started_at, completed_at
		 FROM integrity_check_runs ORDER BY completed_at DESC LIMIT 1`,
	)
	var run IntegrityCheckRun
	var chainValid bool
	var startedAt, completedAt any
	err := row.Scan(&run.ID, &run.CheckedCount, &run.FailedCount, &chainValid, &startedAt, &completedAt)
	if err != nil {
		return nil, err
	}
	run.ChainValid = chainValid
	run.StartedAt = scanTime(startedAt)
	run.CompletedAt = scanTime(completedAt)
	return &run, nil
}

func (r *Repository) getActiveLegalHold(ctx context.Context, contractID string) (*LegalHold, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT contract_id, reason, placed_by, placed_at, released_at
		 FROM legal_holds WHERE contract_id = $1 AND released_at IS NULL`, contractID,
	)
	var h LegalHold
	var placedAt, releasedAt any
	err := row.Scan(&h.ContractID, &h.Reason, &h.PlacedBy, &placedAt, &releasedAt)
	if err != nil {
		return nil, err
	}
	h.PlacedAt = scanTime(placedAt)
	if t := scanTime(releasedAt); !t.IsZero() {
		h.ReleasedAt = sql.NullTime{Time: t, Valid: true}
	}
	return &h, nil
}
