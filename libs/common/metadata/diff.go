package metadata

import (
	"encoding/json"
	"fmt"
)

type FieldDiff struct {
	Field     string `json:"field"`
	Draft     any    `json:"draft,omitempty"`
	Confirmed any    `json:"confirmed,omitempty"`
}

func DiffFromDraft(draftJSON, confirmedJSON json.RawMessage) (json.RawMessage, error) {
	var draft, confirmed map[string]any
	if err := json.Unmarshal(draftJSON, &draft); err != nil {
		return nil, fmt.Errorf("unmarshal draft: %w", err)
	}
	if err := json.Unmarshal(confirmedJSON, &confirmed); err != nil {
		return nil, fmt.Errorf("unmarshal confirmed: %w", err)
	}

	keys := make(map[string]struct{})
	for k := range draft {
		keys[k] = struct{}{}
	}
	for k := range confirmed {
		keys[k] = struct{}{}
	}

	var diffs []FieldDiff
	for k := range keys {
		dv, dok := draft[k]
		cv, cok := confirmed[k]
		if dok && cok && jsonEqual(dv, cv) {
			continue
		}
		diff := FieldDiff{Field: k}
		if dok {
			diff.Draft = dv
		}
		if cok {
			diff.Confirmed = cv
		}
		diffs = append(diffs, diff)
	}

	if len(diffs) == 0 {
		return json.Marshal([]FieldDiff{})
	}
	out, err := json.Marshal(diffs)
	if err != nil {
		return nil, fmt.Errorf("marshal diff: %w", err)
	}
	return out, nil
}

func jsonEqual(a, b any) bool {
	ab, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(ab) == string(bb)
}
