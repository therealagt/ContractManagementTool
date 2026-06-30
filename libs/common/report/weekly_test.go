package report

import (
	"strings"
	"testing"
	"time"
)

func TestWeeklyStatusHTML(t *testing.T) {
	now := time.Now().UTC()
	w := &WeeklyStatus{
		GeneratedAt:  now,
		TrafficLight: "green",
		Summary: &ExecutiveSummary{
			PendingReview:   2,
			IntegrityFailures: 0,
			AuditChainValid: true,
		},
		Week: WeeklyStats{
			PeriodStart: now.AddDate(0, 0, -7),
			PeriodEnd:   now,
			Uploaded:    5,
			Archived:    3,
			Rejected:    1,
			P1Alerts:    0,
			P2Alerts:    1,
		},
	}

	html := w.HTML()
	if !strings.Contains(html, "Weekly Status") {
		t.Fatal("expected title in html")
	}
	if !strings.Contains(html, "Uploads: 5") {
		t.Fatal("expected upload count")
	}

	subject := w.Subject("prod")
	if !strings.Contains(subject, "prod") || !strings.Contains(subject, "GREEN") {
		t.Fatalf("unexpected subject: %s", subject)
	}
}

func TestTrafficLight(t *testing.T) {
	if trafficLight(&ExecutiveSummary{AuditChainValid: true}) != "green" {
		t.Fatal("expected green")
	}
	if trafficLight(&ExecutiveSummary{AuditChainValid: false}) != "red" {
		t.Fatal("expected red for broken chain")
	}
}
