package pipeline

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/therealagt/ContractManagementTool/libs/common/audit"
	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/gcs"
	"github.com/therealagt/ContractManagementTool/libs/common/pubsub"
)

type StagingReader interface {
	Download(ctx context.Context, objectPath string) ([]byte, error)
	Delete(ctx context.Context, objectPath string) error
}

type ArchiveWriter interface {
	Upload(ctx context.Context, objectPath string, r io.Reader, contentType string) error
	FullPath(objectPath string) string
}

type Pipeline struct {
	db              *sql.DB
	repo            *contracts.Repository
	staging         StagingReader
	archive         ArchiveWriter
	retentionYears  int
}

func New(db *sql.DB, staging StagingReader, archive ArchiveWriter, retentionYears int) *Pipeline {
	if retentionYears <= 0 {
		retentionYears = 10
	}
	return &Pipeline{
		db:             db,
		repo:           contracts.NewRepository(db),
		staging:        staging,
		archive:        archive,
		retentionYears: retentionYears,
	}
}

func (p *Pipeline) Run(ctx context.Context, msg pubsub.ArchiveRequested) error {
	detail, err := p.repo.GetByID(ctx, msg.ContractID)
	if err != nil {
		return err
	}

	switch detail.Status {
	case contracts.StatusArchived:
		return nil
	case contracts.StatusConfirmed, contracts.StatusArchiving:
	default:
		return fmt.Errorf("unexpected contract status %s", detail.Status)
	}

	if err := p.repo.UpdateStatus(ctx, msg.ContractID, contracts.StatusArchiving); err != nil {
		return err
	}

	stagingPath := objectPathFromGCS(msg.GCSStagingPath, msg.ContractID)
	pdf, err := p.staging.Download(ctx, stagingPath)
	if err != nil {
		p.logFailure(ctx, msg.ContractID, err)
		return err
	}

	sum := sha256.Sum256(pdf)
	if hex.EncodeToString(sum[:]) != msg.SHA256 {
		err := fmt.Errorf("sha256 mismatch for contract %s", msg.ContractID)
		p.logFailure(ctx, msg.ContractID, err)
		return err
	}

	archivePath := gcs.ArchiveObjectPath(msg.ContractID)
	if err := p.archive.Upload(ctx, archivePath, bytes.NewReader(pdf), "application/pdf"); err != nil {
		p.logFailure(ctx, msg.ContractID, err)
		return err
	}

	now := time.Now().UTC()
	retentionExpires := now.AddDate(p.retentionYears, 0, 0)
	gcsPath := p.archive.FullPath(archivePath)

	if err := p.repo.SaveArchiveRecord(ctx, &contracts.ArchiveRecord{
		ContractID:         msg.ContractID,
		GCSPath:            gcsPath,
		SHA256:             msg.SHA256,
		RetentionExpiresAt: retentionExpires,
		ArchivedAt:         now,
	}); err != nil {
		return err
	}
	if err := p.repo.UpdateStatus(ctx, msg.ContractID, contracts.StatusArchived); err != nil {
		return err
	}

	_ = p.staging.Delete(ctx, stagingPath)

	_, _ = audit.RecordAuditEvent(ctx, p.db, "archive-worker", "archive_completed", &msg.ContractID, map[string]any{
		"gcs_path":             gcsPath,
		"sha256":               msg.SHA256,
		"retention_expires_at": retentionExpires,
	}, nil)
	return nil
}

func (p *Pipeline) logFailure(ctx context.Context, contractID string, err error) {
	_, _ = audit.RecordAuditEvent(ctx, p.db, "archive-worker", "archive_failed", &contractID, map[string]any{
		"error": err.Error(),
	}, nil)
}

func objectPathFromGCS(gcsPath, contractID string) string {
	const marker = "/staging/"
	if idx := strings.Index(gcsPath, marker); idx >= 0 {
		return "staging/" + gcsPath[idx+len(marker):]
	}
	return gcs.StagingObjectPath(contractID)
}
