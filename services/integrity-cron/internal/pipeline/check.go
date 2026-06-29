package pipeline

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/therealagt/ContractManagementTool/libs/common/alerts"
	"github.com/therealagt/ContractManagementTool/libs/common/audit"
	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/gcs"
)

type ArchiveReader interface {
	Download(ctx context.Context, objectPath string) ([]byte, error)
}

type MetricPublisher interface {
	PublishCount(ctx context.Context, metricName string, labels map[string]string, value int64) error
}

type Pipeline struct {
	db            *sql.DB
	repo          *contracts.Repository
	archive       ArchiveReader
	alerts        *alerts.Recorder
	metrics       MetricPublisher
	reviewSLADays int
}

func New(db *sql.DB, archive ArchiveReader, alerts *alerts.Recorder, metrics MetricPublisher, reviewSLADays int) *Pipeline {
	if reviewSLADays <= 0 {
		reviewSLADays = 7
	}
	return &Pipeline{
		db:            db,
		repo:          contracts.NewRepository(db),
		archive:       archive,
		alerts:        alerts,
		metrics:       metrics,
		reviewSLADays: reviewSLADays,
	}
}

type Result struct {
	CheckedCount int
	FailedCount  int
	ChainValid   bool
}

func (p *Pipeline) Run(ctx context.Context) (*Result, error) {
	started := time.Now().UTC()
	result := &Result{ChainValid: true}

	records, err := p.repo.ListArchiveRecords(ctx)
	if err != nil {
		return nil, err
	}

	for _, rec := range records {
		result.CheckedCount++
		if err := p.checkRecord(ctx, rec); err != nil {
			result.FailedCount++
		}
	}

	chain, err := audit.ValidateChain(ctx, p.db, nil)
	if err != nil {
		return nil, err
	}
	result.ChainValid = chain.Valid
	if !chain.Valid {
		_, _ = p.alerts.Record(ctx, alerts.SeverityP1, "integrity-cron", map[string]any{
			"error":      chain.Error,
			"broken_at":  chain.BrokenAt,
			"check_type": "audit_chain",
		}, nil)
		_ = p.metrics.PublishCount(ctx, "audit_chain_broken", nil, 1)
	}

	beyondSLA, err := p.repo.CountPendingReviewBeyondSLA(ctx, p.reviewSLADays)
	if err != nil {
		return nil, err
	}
	if beyondSLA > 0 {
		_, _ = p.alerts.Record(ctx, alerts.SeverityP2, "integrity-cron", map[string]any{
			"count":      beyondSLA,
			"sla_days":   p.reviewSLADays,
			"check_type": "review_sla",
		}, nil)
		_ = p.metrics.PublishCount(ctx, "pending_review_sla_exceeded", nil, int64(beyondSLA))
	}

	completed := time.Now().UTC()
	_ = p.repo.SaveIntegrityCheckRun(ctx, &contracts.IntegrityCheckRun{
		ID:            uuid.NewString(),
		CheckedCount:  result.CheckedCount,
		FailedCount:   result.FailedCount,
		ChainValid:    result.ChainValid,
		StartedAt:     started,
		CompletedAt:   completed,
	})

	if result.FailedCount > 0 {
		_ = p.metrics.PublishCount(ctx, "integrity_check_failed", nil, int64(result.FailedCount))
	} else {
		_ = p.metrics.PublishCount(ctx, "integrity_check_ok", nil, 1)
	}

	_, _ = audit.RecordAuditEvent(ctx, p.db, "integrity-cron", "integrity_check_completed", nil, map[string]any{
		"checked_count": result.CheckedCount,
		"failed_count":  result.FailedCount,
		"chain_valid":   result.ChainValid,
	}, nil)

	return result, nil
}

func (p *Pipeline) checkRecord(ctx context.Context, rec contracts.ArchiveRecord) error {
	_, objectPath, err := gcs.ParseFullPath(rec.GCSPath)
	if err != nil {
		return err
	}

	data, err := p.archive.Download(ctx, objectPath)
	if err != nil {
		_, _ = p.alerts.Record(ctx, alerts.SeverityP1, "integrity-cron", map[string]any{
			"contract_id": rec.ContractID,
			"gcs_path":    rec.GCSPath,
			"error":       err.Error(),
			"check_type":  "missing_object",
		}, nil)
		return err
	}

	sum := sha256.Sum256(data)
	actual := hex.EncodeToString(sum[:])
	if actual != rec.SHA256 {
		_, _ = p.alerts.Record(ctx, alerts.SeverityP1, "integrity-cron", map[string]any{
			"contract_id":   rec.ContractID,
			"gcs_path":      rec.GCSPath,
			"expected_hash": rec.SHA256,
			"actual_hash":   actual,
			"check_type":    "hash_mismatch",
		}, nil)
		return fmt.Errorf("hash mismatch for %s", rec.ContractID)
	}
	return nil
}
