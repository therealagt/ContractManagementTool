package alerts

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	SeverityP1 = "P1"
	SeverityP2 = "P2"
	SeverityP3 = "P3"
)

type Event struct {
	ID         string
	Severity   string
	Source     string
	Payload    map[string]any
	IncidentID *string
	CreatedAt  time.Time
}

type Recorder struct {
	db *sql.DB
	bq *BigQuerySink
}

func NewRecorder(db *sql.DB, bq *BigQuerySink) *Recorder {
	return &Recorder{db: db, bq: bq}
}

func (r *Recorder) Record(ctx context.Context, severity, source string, payload map[string]any, incidentID *string) (*Event, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	event := &Event{
		ID:        uuid.NewString(),
		Severity:  severity,
		Source:    source,
		Payload:   payload,
		IncidentID: incidentID,
		CreatedAt: time.Now().UTC(),
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO alert_events (id, severity, source, payload_json, incident_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		event.ID, event.Severity, event.Source, payloadBytes, event.IncidentID, event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if r.bq != nil {
		_ = r.bq.Insert(ctx, event)
	}
	return event, nil
}
