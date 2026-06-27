package pipeline

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/therealagt/ContractManagementTool/libs/common/audit"
	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/gcs"
	"github.com/therealagt/ContractManagementTool/libs/common/pubsub"
	"github.com/therealagt/ContractManagementTool/libs/common/vertex"
)

type PDFLoader interface {
	Download(ctx context.Context, objectPath string) ([]byte, error)
}

type GeminiExtractor interface {
	ExtractWithConfidence(ctx context.Context, contractType contracts.Type, pdf []byte) (*vertex.ExtractionResult, []byte, error)
}

type Pipeline struct {
	db        *sql.DB
	repo      *contracts.Repository
	loader    PDFLoader
	extractor GeminiExtractor
}

func New(db *sql.DB, loader PDFLoader, extractor GeminiExtractor) *Pipeline {
	return &Pipeline{
		db:        db,
		repo:      contracts.NewRepository(db),
		loader:    loader,
		extractor: extractor,
	}
}

func (p *Pipeline) Run(ctx context.Context, msg pubsub.ExtractionRequested) error {
	detail, err := p.repo.GetByID(ctx, msg.ContractID)
	if err != nil {
		return err
	}

	switch detail.Status {
	case contracts.StatusPendingReview, contracts.StatusRejected:
		return nil
	}
	if detail.Status != contracts.StatusExtracting {
		return fmt.Errorf("unexpected contract status %s", detail.Status)
	}

	objectPath := objectPathFromGCS(msg.GCSPath, msg.ContractID)
	pdf, err := p.loader.Download(ctx, objectPath)
	if err != nil {
		p.logFailure(ctx, msg.ContractID, err)
		return err
	}

	contractType := contracts.Type(msg.Type)
	result, confidenceJSON, err := p.extractor.ExtractWithConfidence(ctx, contractType, pdf)
	if err != nil {
		p.logFailure(ctx, msg.ContractID, err)
		return err
	}

	draft := &contracts.ExtractionDraft{
		ContractID:      msg.ContractID,
		ExtractedJSON:   result.ExtractedJSON,
		SchemaVersion:   result.SchemaVersion,
		ConfidenceFlags: confidenceJSON,
		ExtractedAt:     time.Now().UTC(),
	}
	draft.GeminiModel = sql.NullString{String: result.Model, Valid: true}
	draft.PromptVersion = sql.NullString{String: result.PromptVersion, Valid: true}

	if err := p.repo.SaveExtractionDraft(ctx, draft); err != nil {
		return err
	}
	if err := p.repo.UpdateStatus(ctx, msg.ContractID, contracts.StatusPendingReview); err != nil {
		return err
	}

	_, _ = audit.RecordAuditEvent(ctx, p.db, "extraction-worker", "extraction_completed", &msg.ContractID, map[string]any{
		"gemini_model":    result.Model,
		"prompt_version":  result.PromptVersion,
		"schema_version":  result.SchemaVersion,
	}, nil)
	return nil
}

func (p *Pipeline) logFailure(ctx context.Context, contractID string, err error) {
	_, _ = audit.RecordAuditEvent(ctx, p.db, "extraction-worker", "extraction_failed", &contractID, map[string]any{
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
