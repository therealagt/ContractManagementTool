package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/therealagt/ContractManagementTool/libs/common/audit"
	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/gcs"
)

type HoldStorage interface {
	SetEventHold(ctx context.Context, objectPath string, hold bool) error
}

type LegalHoldService struct {
	db      *sql.DB
	repo    *contracts.Repository
	storage HoldStorage
}

func NewLegalHoldService(db *sql.DB, storage HoldStorage) *LegalHoldService {
	return &LegalHoldService{
		db:      db,
		repo:    contracts.NewRepository(db),
		storage: storage,
	}
}

func (s *LegalHoldService) Place(ctx context.Context, actor, contractID, reason string) error {
	detail, err := s.repo.GetByID(ctx, contractID)
	if err != nil {
		return err
	}
	if detail.Archive == nil {
		return fmt.Errorf("contract not archived")
	}

	if err := s.repo.PlaceLegalHold(ctx, contractID, reason, actor); err != nil {
		return err
	}

	if s.storage != nil {
		_, objectPath, err := gcs.ParseFullPath(detail.Archive.GCSPath)
		if err != nil {
			return err
		}
		if err := s.storage.SetEventHold(ctx, objectPath, true); err != nil {
			return err
		}
	}

	_, _ = audit.RecordAuditEvent(ctx, s.db, actor, "legal_hold_placed", &contractID, map[string]any{
		"reason": reason,
	}, nil)
	return nil
}

func (s *LegalHoldService) Release(ctx context.Context, actor, contractID string) error {
	detail, err := s.repo.GetByID(ctx, contractID)
	if err != nil {
		return err
	}

	if err := s.repo.ReleaseLegalHold(ctx, contractID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no active legal hold")
		}
		return err
	}

	if s.storage != nil && detail.Archive != nil {
		_, objectPath, err := gcs.ParseFullPath(detail.Archive.GCSPath)
		if err != nil {
			return err
		}
		if err := s.storage.SetEventHold(ctx, objectPath, false); err != nil {
			return err
		}
	}

	_, _ = audit.RecordAuditEvent(ctx, s.db, actor, "legal_hold_released", &contractID, nil, nil)
	return nil
}
