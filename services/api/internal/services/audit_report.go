package services

import (
	"context"
	"database/sql"

	"github.com/therealagt/ContractManagementTool/libs/common/report"
)

type AuditReportService struct {
	generator *report.Generator
	archive   report.ArchiveDownloader
}

func NewAuditReportService(db *sql.DB, archive report.ArchiveDownloader, reviewSLADays int) *AuditReportService {
	return &AuditReportService{
		generator: report.NewGenerator(db, reviewSLADays),
		archive:   archive,
	}
}

func (s *AuditReportService) ExecutiveSummary(ctx context.Context) (*report.ExecutiveSummary, error) {
	return s.generator.ExecutiveSummary(ctx)
}

func (s *AuditReportService) ControlMatrix(ctx context.Context) ([]report.Control, error) {
	return s.generator.ControlMatrix(ctx)
}

func (s *AuditReportService) AuditTrail(ctx context.Context, contractID *string, limit int) ([]map[string]any, error) {
	return s.generator.AuditTrail(ctx, contractID, limit)
}

func (s *AuditReportService) EvidencePackage(ctx context.Context, contractIDs []string) ([]byte, error) {
	return s.generator.EvidencePackage(ctx, s.archive, contractIDs)
}
