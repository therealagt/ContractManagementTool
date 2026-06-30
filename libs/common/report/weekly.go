package report

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/therealagt/ContractManagementTool/libs/common/alerts"
)

type WeeklyStats struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	Uploaded    int
	Archived    int
	Rejected    int
	P1Alerts    int
	P2Alerts    int
}

type WeeklyStatus struct {
	GeneratedAt   time.Time
	TrafficLight  string
	Summary       *ExecutiveSummary
	Week          WeeklyStats
	ControlGaps   []string
}

func (g *Generator) WeeklyStatus(ctx context.Context) (*WeeklyStatus, error) {
	summary, err := g.ExecutiveSummary(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	weekStart := now.AddDate(0, 0, -7)

	uploaded, err := g.repo.CountUploadedSince(ctx, weekStart)
	if err != nil {
		return nil, err
	}
	archived, err := g.repo.CountArchivedSince(ctx, weekStart)
	if err != nil {
		return nil, err
	}
	rejected, err := g.repo.CountRejectedSince(ctx, weekStart)
	if err != nil {
		return nil, err
	}

	alertCounts, err := alerts.CountBySeveritySince(ctx, g.db, weekStart)
	if err != nil {
		return nil, err
	}

	status := &WeeklyStatus{
		GeneratedAt:  now,
		Summary:      summary,
		TrafficLight: trafficLight(summary),
		Week: WeeklyStats{
			PeriodStart: weekStart,
			PeriodEnd:   now,
			Uploaded:    uploaded,
			Archived:    archived,
			Rejected:    rejected,
			P1Alerts:    alertCounts[alerts.SeverityP1],
			P2Alerts:    alertCounts[alerts.SeverityP2],
		},
		ControlGaps: controlGaps(summary),
	}
	return status, nil
}

func trafficLight(summary *ExecutiveSummary) string {
	if summary.IntegrityFailures > 0 || !summary.AuditChainValid || summary.PendingBeyondSLA > 0 {
		return "red"
	}
	if summary.PendingReview > 0 {
		return "yellow"
	}
	return "green"
}

func controlGaps(summary *ExecutiveSummary) []string {
	var gaps []string
	for _, risk := range summary.TopRisks {
		gaps = append(gaps, risk)
	}
	return gaps
}

func (w *WeeklyStatus) HTML() string {
	light := strings.ToUpper(w.TrafficLight)
	period := fmt.Sprintf("%s — %s",
		w.Week.PeriodStart.Format("2006-01-02"),
		w.Week.PeriodEnd.Format("2006-01-02"),
	)

	integrity := "no run yet"
	if w.Summary.IntegrityLastRun != nil {
		integrity = fmt.Sprintf("%s — %d failure(s), chain valid: %t",
			w.Summary.IntegrityLastRun.Format(time.RFC3339),
			w.Summary.IntegrityFailures,
			w.Summary.AuditChainValid,
		)
	}

	var gaps string
	if len(w.ControlGaps) == 0 {
		gaps = "<li>None</li>"
	} else {
		for _, g := range w.ControlGaps {
			gaps += fmt.Sprintf("<li>%s</li>", html.EscapeString(g))
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family: sans-serif; color: #222;">
<h2>Contract Management — Weekly Status</h2>
<p><strong>Status:</strong> %s &nbsp;|&nbsp; <strong>Period:</strong> %s</p>
<h3>Pipeline (this week)</h3>
<ul>
<li>Uploads: %d</li>
<li>Archived: %d</li>
<li>Rejected: %d</li>
<li>Pending review (current): %d</li>
</ul>
<h3>Integrity</h3>
<p>%s</p>
<h3>Alerts (this week)</h3>
<ul>
<li>P1: %d</li>
<li>P2: %d</li>
</ul>
<h3>Open compliance gaps</h3>
<ul>%s</ul>
<p style="color:#666;font-size:12px;">Aggregated metrics only — no contract text or PII. Formal audit reports via IAP dashboard.</p>
</body></html>`,
		html.EscapeString(light),
		html.EscapeString(period),
		w.Week.Uploaded,
		w.Week.Archived,
		w.Week.Rejected,
		w.Summary.PendingReview,
		html.EscapeString(integrity),
		w.Week.P1Alerts,
		w.Week.P2Alerts,
		gaps,
	)
}

func (w *WeeklyStatus) Subject(environment string) string {
	return fmt.Sprintf("[%s] Contract Management weekly status — %s",
		environment,
		strings.ToUpper(w.TrafficLight),
	)
}
