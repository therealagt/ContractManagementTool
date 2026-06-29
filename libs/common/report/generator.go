package report

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/therealagt/ContractManagementTool/libs/common/audit"
	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/gcs"
)

type Control struct {
	ID          string `json:"id"`
	ControlRef  string `json:"control_ref"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	NDA         string `json:"nda"`
	AVV         string `json:"avv"`
	LastChecked string `json:"last_checked"`
	OpenItems   string `json:"open_items"`
}

type ExecutiveSummary struct {
	GeneratedAt           time.Time        `json:"generated_at"`
	TotalContracts        int              `json:"total_contracts"`
	ByStatus              map[string]int   `json:"by_status"`
	ArchivedCount         int              `json:"archived_count"`
	PendingReview         int              `json:"pending_review"`
	PendingBeyondSLA      int              `json:"pending_beyond_sla"`
	RejectedCount         int              `json:"rejected_count"`
	CompliancePercent     float64          `json:"compliance_percent"`
	IntegrityLastRun      *time.Time       `json:"integrity_last_run,omitempty"`
	IntegrityFailures     int              `json:"integrity_failures"`
	AuditChainValid       bool             `json:"audit_chain_valid"`
	ActiveLegalHolds      int              `json:"active_legal_holds"`
	TopRisks              []string         `json:"top_risks"`
}

type Generator struct {
	db          *sql.DB
	repo        *contracts.Repository
	reviewSLADays int
}

func NewGenerator(db *sql.DB, reviewSLADays int) *Generator {
	if reviewSLADays <= 0 {
		reviewSLADays = 7
	}
	return &Generator{
		db:            db,
		repo:          contracts.NewRepository(db),
		reviewSLADays: reviewSLADays,
	}
}

func (g *Generator) ExecutiveSummary(ctx context.Context) (*ExecutiveSummary, error) {
	byStatus, err := g.repo.CountByStatus(ctx)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, n := range byStatus {
		total += n
	}

	beyondSLA, err := g.repo.CountPendingReviewBeyondSLA(ctx, g.reviewSLADays)
	if err != nil {
		return nil, err
	}

	run, _ := g.repo.GetLatestIntegrityCheckRun(ctx)
	chain, _ := audit.ValidateChain(ctx, g.db, nil)

	holds, err := g.repo.CountActiveLegalHolds(ctx)
	if err != nil {
		return nil, err
	}

	summary := &ExecutiveSummary{
		GeneratedAt:       time.Now().UTC(),
		TotalContracts:  total,
		ByStatus:        byStatus,
		ArchivedCount:   byStatus[string(contracts.StatusArchived)],
		PendingReview:   byStatus[string(contracts.StatusPendingReview)],
		PendingBeyondSLA: beyondSLA,
		RejectedCount:   byStatus[string(contracts.StatusRejected)],
		ActiveLegalHolds: holds,
		AuditChainValid: chain != nil && chain.Valid,
	}

	if total > 0 {
		compliant := summary.ArchivedCount
		summary.CompliancePercent = float64(compliant) / float64(total) * 100
	}

	if run != nil {
		summary.IntegrityLastRun = &run.CompletedAt
		summary.IntegrityFailures = run.FailedCount
	}

	summary.TopRisks = g.topRisks(summary, chain)
	return summary, nil
}

func (g *Generator) topRisks(summary *ExecutiveSummary, chain *audit.ChainValidation) []string {
	var risks []string
	if summary.PendingBeyondSLA > 0 {
		risks = append(risks, fmt.Sprintf("%d contracts exceed review SLA", summary.PendingBeyondSLA))
	}
	if summary.IntegrityFailures > 0 {
		risks = append(risks, fmt.Sprintf("%d archive integrity failures", summary.IntegrityFailures))
	}
	if chain != nil && !chain.Valid {
		risks = append(risks, "audit hash chain validation failed")
	}
	return risks
}

func (g *Generator) ControlMatrix(ctx context.Context) ([]Control, error) {
	byStatus, err := g.repo.CountByStatus(ctx)
	if err != nil {
		return nil, err
	}
	beyondSLA, _ := g.repo.CountPendingReviewBeyondSLA(ctx, g.reviewSLADays)
	run, _ := g.repo.GetLatestIntegrityCheckRun(ctx)
	chain, _ := audit.ValidateChain(ctx, g.db, nil)
	holds, _ := g.repo.CountActiveLegalHolds(ctx)

	now := time.Now().UTC().Format(time.RFC3339)
	integrityStatus := "pass"
	if run != nil && run.FailedCount > 0 {
		integrityStatus = "fail"
	}
	chainStatus := "pass"
	if chain != nil && !chain.Valid {
		chainStatus = "fail"
	}
	slaStatus := "pass"
	slaOpen := ""
	if beyondSLA > 0 {
		slaStatus = "fail"
		slaOpen = fmt.Sprintf("%d contracts past SLA", beyondSLA)
	}

	return []Control{
		{ID: "CTL-001", ControlRef: "ISO 27001 A.5.18", Name: "IAP access control", Status: "pass", NDA: "yes", AVV: "yes", LastChecked: now, OpenItems: ""},
		{ID: "CTL-002", ControlRef: "ISO 27001 A.5.3", Name: "Separation of duties (upload vs confirm)", Status: "pass", NDA: "yes", AVV: "yes", LastChecked: now, OpenItems: ""},
		{ID: "CTL-003", ControlRef: "eIDAS / PAdES", Name: "Digital signature validation on upload", Status: "pass", NDA: "yes", AVV: "yes", LastChecked: now, OpenItems: ""},
		{ID: "CTL-004", ControlRef: "ISO 27001 A.8.24", Name: "WORM archive after HITL confirmation", Status: statusFromCount(byStatus[string(contracts.StatusArchived)]), NDA: "yes", AVV: "yes", LastChecked: now, OpenItems: ""},
		{ID: "CTL-005", ControlRef: "ISO 27001 A.8.24", Name: "Nightly archive integrity recheck", Status: integrityStatus, NDA: "yes", AVV: "yes", LastChecked: checkedAt(run), OpenItems: openIntegrity(run)},
		{ID: "CTL-006", ControlRef: "ISO 27001 A.8.15", Name: "Hash-chained audit events", Status: chainStatus, NDA: "yes", AVV: "yes", LastChecked: now, OpenItems: openChain(chain)},
		{ID: "CTL-007", ControlRef: "DSGVO Art. 28", Name: "HITL review within SLA", Status: slaStatus, NDA: "yes", AVV: "yes", LastChecked: now, OpenItems: slaOpen},
		{ID: "CTL-008", ControlRef: "ISO 27001 A.5.28", Name: "Legal hold capability", Status: "pass", NDA: "yes", AVV: "yes", LastChecked: now, OpenItems: fmt.Sprintf("%d active holds", holds)},
	}, nil
}

func statusFromCount(archived int) string {
	if archived > 0 {
		return "pass"
	}
	return "pending"
}

func checkedAt(run *contracts.IntegrityCheckRun) string {
	if run == nil {
		return ""
	}
	return run.CompletedAt.Format(time.RFC3339)
}

func openIntegrity(run *contracts.IntegrityCheckRun) string {
	if run == nil || run.FailedCount == 0 {
		return ""
	}
	return fmt.Sprintf("%d hash mismatches", run.FailedCount)
}

func openChain(chain *audit.ChainValidation) string {
	if chain == nil || chain.Valid {
		return ""
	}
	return chain.Error
}

func (g *Generator) AuditTrail(ctx context.Context, contractID *string, limit int) ([]map[string]any, error) {
	events, err := audit.ListAuditEvents(ctx, g.db, contractID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(events))
	for _, e := range events {
		entry := map[string]any{
			"id":         e.ID,
			"actor":      e.Actor,
			"action":     e.Action,
			"created_at": e.CreatedAt,
		}
		if e.ContractID.Valid {
			entry["contract_id"] = e.ContractID.String
		}
		if len(e.PayloadJSON) > 0 {
			entry["payload"] = json.RawMessage(e.PayloadJSON)
		}
		if e.PrevEventHash.Valid {
			entry["prev_event_hash"] = e.PrevEventHash.String
		}
		if e.EventHash.Valid {
			entry["event_hash"] = e.EventHash.String
		}
		out = append(out, entry)
	}
	return out, nil
}

type ArchiveDownloader interface {
	Download(ctx context.Context, objectPath string) ([]byte, error)
}

func (g *Generator) EvidencePackage(
	ctx context.Context,
	downloader ArchiveDownloader,
	contractIDs []string,
) ([]byte, error) {
	summary, err := g.ExecutiveSummary(ctx)
	if err != nil {
		return nil, err
	}
	matrix, err := g.ControlMatrix(ctx)
	if err != nil {
		return nil, err
	}
	trail, err := g.AuditTrail(ctx, nil, 1000)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	writeJSONFile(zw, "executive_summary.json", summary)
	writeJSONFile(zw, "audit_trail.json", trail)

	csvBuf := &bytes.Buffer{}
	w := csv.NewWriter(csvBuf)
	_ = w.Write([]string{"id", "control_ref", "name", "status", "nda", "avv", "last_checked", "open_items"})
	for _, c := range matrix {
		_ = w.Write([]string{c.ID, c.ControlRef, c.Name, c.Status, c.NDA, c.AVV, c.LastChecked, c.OpenItems})
	}
	w.Flush()
	csvFile, _ := zw.Create("control_matrix.csv")
	_, _ = csvFile.Write(csvBuf.Bytes())

	if len(contractIDs) == 0 {
		ids, err := g.repo.ListArchivedContractIDs(ctx)
		if err != nil {
			return nil, err
		}
		contractIDs = ids
	}

	manifest := make([]map[string]any, 0, len(contractIDs))
	for _, id := range contractIDs {
		detail, err := g.repo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		entry := map[string]any{
			"contract_id": id,
			"type":        detail.Type,
			"sha256":      detail.SHA256,
		}
		if detail.Archive != nil {
			entry["archive_path"] = detail.Archive.GCSPath
			entry["archive_sha256"] = detail.Archive.SHA256
		}
		if detail.Signature != nil {
			entry["signature_valid"] = detail.Signature.IsValid
		}
		if detail.Confirmed != nil {
			entry["confirmed_metadata"] = json.RawMessage(detail.Confirmed.MetadataJSON)
		}
		manifest = append(manifest, entry)

		if detail.Archive != nil && downloader != nil {
			_, objectPath, err := gcs.ParseFullPath(detail.Archive.GCSPath)
			if err != nil {
				continue
			}
			pdf, err := downloader.Download(ctx, objectPath)
			if err != nil {
				continue
			}
			pdfFile, _ := zw.Create(fmt.Sprintf("contracts/%s.pdf", id))
			_, _ = pdfFile.Write(pdf)
		}
	}
	writeJSONFile(zw, "manifest.json", manifest)

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeJSONFile(zw *zip.Writer, name string, v any) {
	data, _ := json.MarshalIndent(v, "", "  ")
	f, _ := zw.Create(name)
	_, _ = f.Write(data)
}
