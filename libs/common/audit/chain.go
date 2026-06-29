package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type ChainValidation struct {
	Valid       bool
	TotalEvents int
	BrokenAt    *string
	Error       string
}

func ValidateChain(ctx context.Context, db *sql.DB, contractID *string) (*ChainValidation, error) {
	var query string
	var args []any
	if contractID != nil {
		query = `SELECT id, contract_id, actor, action, payload_json, prev_event_hash, event_hash, created_at
		         FROM audit_events WHERE contract_id = $1 ORDER BY created_at ASC`
		args = []any{*contractID}
	} else {
		query = `SELECT id, contract_id, actor, action, payload_json, prev_event_hash, event_hash, created_at
		         FROM audit_events WHERE contract_id IS NULL ORDER BY created_at ASC`
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &ChainValidation{Valid: true}
	var prevHash *string

	for rows.Next() {
		var event AuditEvent
		if err := rows.Scan(
			&event.ID, &event.ContractID, &event.Actor, &event.Action,
			&event.PayloadJSON, &event.PrevEventHash, &event.EventHash, &event.CreatedAt,
		); err != nil {
			return nil, err
		}
		result.TotalEvents++

		var payload map[string]any
		if len(event.PayloadJSON) > 0 {
			_ = json.Unmarshal(event.PayloadJSON, &payload)
		}
		var cid *string
		if event.ContractID.Valid {
			c := event.ContractID.String
			cid = &c
		}

		expected := hashPayload(map[string]any{
			"actor":           event.Actor,
			"action":          event.Action,
			"contract_id":     cid,
			"payload":         payload,
			"prev_event_hash": prevHash,
		})

		if !event.EventHash.Valid || event.EventHash.String != expected {
			result.Valid = false
			result.BrokenAt = &event.ID
			result.Error = fmt.Sprintf("event hash mismatch at %s", event.ID)
			break
		}

		if prevHash == nil {
			if event.PrevEventHash.Valid {
				result.Valid = false
				result.BrokenAt = &event.ID
				result.Error = "first event has unexpected prev_event_hash"
				break
			}
		} else if !event.PrevEventHash.Valid || event.PrevEventHash.String != *prevHash {
			result.Valid = false
			result.BrokenAt = &event.ID
			result.Error = "prev_event_hash does not match chain"
			break
		}

		if event.EventHash.Valid {
			h := event.EventHash.String
			prevHash = &h
		}
	}
	return result, rows.Err()
}

func ListAuditEvents(ctx context.Context, db *sql.DB, contractID *string, limit int) ([]AuditEvent, error) {
	if limit <= 0 {
		limit = 500
	}
	var query string
	var args []any
	if contractID != nil {
		query = `SELECT id, contract_id, actor, action, payload_json, prev_event_hash, event_hash, created_at
		         FROM audit_events WHERE contract_id = $1 ORDER BY created_at DESC LIMIT $2`
		args = []any{*contractID, limit}
	} else {
		query = `SELECT id, contract_id, actor, action, payload_json, prev_event_hash, event_hash, created_at
		         FROM audit_events ORDER BY created_at DESC LIMIT $1`
		args = []any{limit}
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var event AuditEvent
		if err := rows.Scan(
			&event.ID, &event.ContractID, &event.Actor, &event.Action,
			&event.PayloadJSON, &event.PrevEventHash, &event.EventHash, &event.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}
