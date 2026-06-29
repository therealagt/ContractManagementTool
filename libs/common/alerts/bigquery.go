package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
)

type BigQuerySink struct {
	inserter *bigquery.Inserter
}

func NewBigQuerySink(ctx context.Context, projectID, datasetID string) (*BigQuerySink, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery client: %w", err)
	}
	return &BigQuerySink{
		inserter: client.Dataset(datasetID).Table("alert_events").Inserter(),
	}, nil
}

type bqRow struct {
	ID         string          `bigquery:"id"`
	Severity   string          `bigquery:"severity"`
	Source     string          `bigquery:"source"`
	Payload    json.RawMessage `bigquery:"payload_json"`
	IncidentID string          `bigquery:"incident_id"`
	CreatedAt  time.Time       `bigquery:"created_at"`
}

func (s *BigQuerySink) Insert(ctx context.Context, event *Event) error {
	payload, _ := json.Marshal(event.Payload)
	incidentID := ""
	if event.IncidentID != nil {
		incidentID = *event.IncidentID
	}
	row := &bqRow{
		ID:         event.ID,
		Severity:   event.Severity,
		Source:     event.Source,
		Payload:    payload,
		IncidentID: incidentID,
		CreatedAt:  event.CreatedAt,
	}
	return s.inserter.Put(ctx, row)
}
