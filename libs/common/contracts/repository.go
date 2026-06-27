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
		`INSERT INTO contracts (id, type, status, partner_id, gcs_staging_path, sha256, uploaded_by, uploaded_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.Type, c.Status, c.PartnerID, c.GCSStagingPath, c.SHA256, c.UploadedBy, c.UploadedAt,
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
		`SELECT id, type, status, partner_id, gcs_staging_path, sha256, uploaded_by, uploaded_at
		 FROM contracts WHERE id = $1`, contractID,
	)
	c, err := scanContract(row)
	if err != nil {
		return nil, err
	}

	detail := &ContractDetail{Contract: *c}
	detail.Signature, _ = r.getSignature(ctx, contractID)
	detail.Draft, _ = r.getDraft(ctx, contractID)
	return detail, nil
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
	err := row.Scan(&d.ContractID, &extracted, &d.GeminiModel, &d.PromptVersion, &d.SchemaVersion, &flags, &d.ExtractedAt)
	if err != nil {
		return nil, err
	}
	d.ExtractedJSON = extracted
	d.ConfidenceFlags = flags
	return &d, nil
}

func scanContract(row *sql.Row) (*Contract, error) {
	var c Contract
	var contractType, status string
	var uploadedAt any
	err := row.Scan(&c.ID, &contractType, &status, &c.PartnerID, &c.GCSStagingPath, &c.SHA256, &c.UploadedBy, &uploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("scan contract: %w", err)
	}
	c.Type = Type(contractType)
	c.Status = Status(status)
	c.UploadedAt = scanTime(uploadedAt)
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
