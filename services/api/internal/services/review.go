package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/therealagt/ContractManagementTool/libs/common/audit"
	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/metadata"
	"github.com/therealagt/ContractManagementTool/libs/common/pubsub"
	"github.com/therealagt/ContractManagementTool/services/api/internal/auth"
)

type ArchivePublisher interface {
	PublishArchive(ctx context.Context, msg pubsub.ArchiveRequested) error
}

type ReviewService struct {
	db        *sql.DB
	repo      *contracts.Repository
	publisher ArchivePublisher
}

func NewReviewService(db *sql.DB, publisher ArchivePublisher) *ReviewService {
	return &ReviewService{
		db:        db,
		repo:      contracts.NewRepository(db),
		publisher: publisher,
	}
}

func (s *ReviewService) ListQueue(ctx context.Context, actor string) ([]contracts.ReviewQueueItem, error) {
	return s.repo.ListPendingReview(ctx, actor, 100)
}

type ConfirmInput struct {
	Actor        string
	ContractID   string
	MetadataJSON json.RawMessage
}

type ConfirmResult struct {
	ContractID string
	Status     contracts.Status
}

func (s *ReviewService) Confirm(ctx context.Context, in ConfirmInput) (*ConfirmResult, error) {
	detail, err := s.repo.GetByID(ctx, in.ContractID)
	if err != nil {
		return nil, err
	}
	if detail.Status != contracts.StatusPendingReview {
		return nil, fmt.Errorf("contract not pending review")
	}
	if err := auth.AssertSeparationOfDuty(in.Actor, detail.UploadedBy, "confirm contract"); err != nil {
		return nil, err
	}
	if detail.Draft == nil {
		return nil, fmt.Errorf("extraction draft missing")
	}
	if !json.Valid(in.MetadataJSON) {
		return nil, fmt.Errorf("invalid metadata json")
	}

	diff, err := metadata.DiffFromDraft(detail.Draft.ExtractedJSON, in.MetadataJSON)
	if err != nil {
		return nil, err
	}

	if err := s.repo.ConfirmContract(ctx, in.ContractID, in.Actor, in.MetadataJSON, diff); err != nil {
		return nil, err
	}

	var diffPayload any
	_ = json.Unmarshal(diff, &diffPayload)
	_, _ = audit.RecordAuditEvent(ctx, s.db, in.Actor, "review_confirmed", &in.ContractID, map[string]any{
		"diff_from_draft": diffPayload,
		"confirmed_by":    in.Actor,
	}, nil)

	if s.publisher != nil && detail.GCSStagingPath.Valid {
		if err := s.publisher.PublishArchive(ctx, pubsub.ArchiveRequested{
			ContractID:     in.ContractID,
			GCSStagingPath: detail.GCSStagingPath.String,
			SHA256:         detail.SHA256,
		}); err != nil {
			return nil, fmt.Errorf("publish archive: %w", err)
		}
	}

	return &ConfirmResult{ContractID: in.ContractID, Status: contracts.StatusConfirmed}, nil
}

type RejectInput struct {
	Actor      string
	ContractID string
	Reason     string
}

func (s *ReviewService) Reject(ctx context.Context, in RejectInput) error {
	detail, err := s.repo.GetByID(ctx, in.ContractID)
	if err != nil {
		return err
	}
	if detail.Status != contracts.StatusPendingReview {
		return fmt.Errorf("contract not pending review")
	}
	if err := auth.AssertSeparationOfDuty(in.Actor, detail.UploadedBy, "reject contract"); err != nil {
		return err
	}

	if err := s.repo.RejectContract(ctx, in.ContractID); err != nil {
		return err
	}

	payload := map[string]any{"reason": in.Reason}
	_, _ = audit.RecordAuditEvent(ctx, s.db, in.Actor, "review_rejected", &in.ContractID, payload, nil)
	return nil
}

func IsSODViolation(err error) bool {
	var sod *auth.SODViolation
	return errors.As(err, &sod)
}
