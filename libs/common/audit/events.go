package audit

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrAuditChain = errors.New("prev_event_hash does not match audit chain tip")

type AccessEvent struct {
	ID           string
	Actor        string
	ResourceType string
	ResourceID   sql.NullString
	Action       string
	IP           sql.NullString
	CreatedAt    time.Time
}

type AuditEvent struct {
	ID            string
	ContractID    sql.NullString
	Actor         string
	Action        string
	PayloadJSON   []byte
	PrevEventHash sql.NullString
	EventHash     sql.NullString
	CreatedAt     time.Time
}

func hashPayload(payload map[string]any) string {
	encoded := sortedJSON(payload)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func sortedJSON(v any) []byte {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sortStrings(keys)
		buf := []byte{'{'}
		for i, k := range keys {
			if i > 0 {
				buf = append(buf, ',')
			}
			key, _ := json.Marshal(k)
			val := sortedJSON(t[k])
			buf = append(buf, key...)
			buf = append(buf, ':')
			buf = append(buf, val...)
		}
		buf = append(buf, '}')
		return buf
	case []any:
		buf := []byte{'['}
		for i, item := range t {
			if i > 0 {
				buf = append(buf, ',')
			}
			buf = append(buf, sortedJSON(item)...)
		}
		buf = append(buf, ']')
		return buf
	default:
		encoded, _ := json.Marshal(t)
		return encoded
	}
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func latestAuditHash(ctx context.Context, db *sql.DB, contractID *string) (*string, error) {
	var query string
	var args []any

	if contractID != nil {
		query = `SELECT event_hash FROM audit_events WHERE contract_id = $1 ORDER BY created_at DESC LIMIT 1`
		args = []any{*contractID}
	} else {
		query = `SELECT event_hash FROM audit_events WHERE contract_id IS NULL ORDER BY created_at DESC LIMIT 1`
	}

	var hash sql.NullString
	err := db.QueryRowContext(ctx, query, args...).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !hash.Valid {
		return nil, nil
	}
	v := hash.String
	return &v, nil
}

func RecordAccessEvent(
	ctx context.Context,
	db *sql.DB,
	actor, resourceType, action string,
	resourceID, ip *string,
) (*AccessEvent, error) {
	event := &AccessEvent{
		ID:           uuid.NewString(),
		Actor:        actor,
		ResourceType: resourceType,
		Action:       action,
		CreatedAt:    time.Now().UTC(),
	}
	if resourceID != nil {
		event.ResourceID = sql.NullString{String: *resourceID, Valid: true}
	}
	if ip != nil {
		event.IP = sql.NullString{String: *ip, Valid: true}
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO access_events (id, actor, resource_type, resource_id, action, ip, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		event.ID, event.Actor, event.ResourceType, event.ResourceID, event.Action, event.IP, event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func RecordAuditEvent(
	ctx context.Context,
	db *sql.DB,
	actor, action string,
	contractID *string,
	payload map[string]any,
	prevEventHash *string,
) (*AuditEvent, error) {
	chainPrev, err := latestAuditHash(ctx, db, contractID)
	if err != nil {
		return nil, err
	}
	if prevEventHash != nil && (chainPrev == nil || *prevEventHash != *chainPrev) {
		return nil, ErrAuditChain
	}

	body := payload
	if body == nil {
		body = map[string]any{}
	}
	payloadBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal audit payload: %w", err)
	}

	eventHash := hashPayload(map[string]any{
		"actor":           actor,
		"action":          action,
		"contract_id":     contractID,
		"payload":         body,
		"prev_event_hash": chainPrev,
	})

	event := &AuditEvent{
		ID:          uuid.NewString(),
		Actor:       actor,
		Action:      action,
		PayloadJSON: payloadBytes,
		CreatedAt:   time.Now().UTC(),
	}
	if contractID != nil {
		event.ContractID = sql.NullString{String: *contractID, Valid: true}
	}
	if chainPrev != nil {
		event.PrevEventHash = sql.NullString{String: *chainPrev, Valid: true}
	}
	event.EventHash = sql.NullString{String: eventHash, Valid: true}

	_, err = db.ExecContext(ctx,
		`INSERT INTO audit_events (id, contract_id, actor, action, payload_json, prev_event_hash, event_hash, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		event.ID, event.ContractID, event.Actor, event.Action, event.PayloadJSON,
		event.PrevEventHash, event.EventHash, event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return event, nil
}
