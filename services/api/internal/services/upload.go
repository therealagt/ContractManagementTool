package services

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/therealagt/ContractManagementTool/libs/common/audit"
	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/gcs"
	"github.com/therealagt/ContractManagementTool/libs/common/pades"
	"github.com/therealagt/ContractManagementTool/libs/common/pubsub"
)

const maxUploadBytes = 25 << 20 // 25 MiB

type StagingStorage interface {
	Upload(ctx context.Context, objectPath string, r io.Reader, contentType string) error
	FullPath(objectPath string) string
}

type ExtractionPublisher interface {
	PublishExtraction(ctx context.Context, msg pubsub.ExtractionRequested) error
}

type UploadService struct {
	db        *sql.DB
	repo      *contracts.Repository
	staging   StagingStorage
	publisher ExtractionPublisher
}

func NewUploadService(db *sql.DB, staging StagingStorage, publisher ExtractionPublisher) *UploadService {
	return &UploadService{
		db:        db,
		repo:      contracts.NewRepository(db),
		staging:   staging,
		publisher: publisher,
	}
}

type UploadInput struct {
	Actor      string
	Type       contracts.Type
	PartnerID  *string
	PDF        []byte
}

type UploadResult struct {
	ContractID string
	Status     contracts.Status
}

func (s *UploadService) Upload(ctx context.Context, in UploadInput) (*UploadResult, error) {
	if !in.Type.Valid() {
		return nil, fmt.Errorf("invalid contract type")
	}
	if len(in.PDF) == 0 || len(in.PDF) > maxUploadBytes {
		return nil, fmt.Errorf("invalid PDF size")
	}

	validation, err := pades.Validate(in.PDF)
	if err != nil {
		return nil, fmt.Errorf("pades validate: %w", err)
	}

	status := contracts.StatusExtracting
	if !validation.Valid {
		status = contracts.StatusRejected
	}

	contract, err := s.repo.CreateContract(ctx, in.Type, status, in.PartnerID, validation.SHA256, in.Actor)
	if err != nil {
		return nil, err
	}

	sig := &contracts.SignatureValidation{
		ContractID:           contract.ID,
		IsValid:              validation.Valid,
		ValidationResultJSON: validation.ToJSON(),
		ValidatedAt:          time.Now().UTC(),
	}
	if validation.SignerCN != "" {
		sig.SignerCN = sql.NullString{String: validation.SignerCN, Valid: true}
	}
	if validation.CertIssuer != "" {
		sig.CertIssuer = sql.NullString{String: validation.CertIssuer, Valid: true}
	}
	if validation.SignedAt != "" {
		if t, parseErr := time.Parse(time.RFC3339, validation.SignedAt); parseErr == nil {
			sig.SignedAt = sql.NullTime{Time: t, Valid: true}
		}
	}
	if err := s.repo.SaveSignatureValidation(ctx, sig); err != nil {
		return nil, err
	}

	if !validation.Valid {
		_, _ = audit.RecordAuditEvent(ctx, s.db, in.Actor, "upload_rejected", &contract.ID, map[string]any{
			"errors": validation.Errors,
			"sha256": validation.SHA256,
		}, nil)
		return &UploadResult{ContractID: contract.ID, Status: contracts.StatusRejected}, nil
	}

	objectPath := gcs.StagingObjectPath(contract.ID)
	if err := s.staging.Upload(ctx, objectPath, bytes.NewReader(in.PDF), "application/pdf"); err != nil {
		return nil, err
	}
	gcsPath := s.staging.FullPath(objectPath)
	if err := s.repo.SetGCSPathAndStatus(ctx, contract.ID, gcsPath, contracts.StatusExtracting); err != nil {
		return nil, err
	}

	_, _ = audit.RecordAuditEvent(ctx, s.db, in.Actor, "upload_accepted", &contract.ID, map[string]any{
		"sha256":     validation.SHA256,
		"gcs_path":   gcsPath,
		"signer_cn":  validation.SignerCN,
		"signed_at":  validation.SignedAt,
	}, nil)

	if s.publisher != nil {
		if err := s.publisher.PublishExtraction(ctx, pubsub.ExtractionRequested{
			ContractID:    contract.ID,
			Type:          string(in.Type),
			GCSPath:       gcsPath,
			SchemaVersion: in.Type.SchemaVersion(),
		}); err != nil {
			return nil, fmt.Errorf("publish extraction: %w", err)
		}
	}

	return &UploadResult{ContractID: contract.ID, Status: contracts.StatusExtracting}, nil
}

func (s *UploadService) GetContract(ctx context.Context, contractID, actor string, isAdmin, isAuditor, isUploader bool) (*contracts.ContractDetail, error) {
	detail, err := s.repo.GetByID(ctx, contractID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}
	if isAdmin || isAuditor {
		return detail, nil
	}
	if isUploader && detail.UploadedBy == actor {
		return detail, nil
	}
	return nil, fmt.Errorf("forbidden")
}
