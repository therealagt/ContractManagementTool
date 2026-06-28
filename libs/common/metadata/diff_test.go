package metadata

import (
	"encoding/json"
	"testing"
)

func TestDiffFromDraft(t *testing.T) {
	draft := json.RawMessage(`{"a":"1","b":"2"}`)
	confirmed := json.RawMessage(`{"a":"1","b":"3","c":"new"}`)

	diff, err := DiffFromDraft(draft, confirmed)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}

	var fields []FieldDiff
	if err := json.Unmarshal(diff, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(fields))
	}
}

func TestDiffFromDraftIdentical(t *testing.T) {
	raw := json.RawMessage(`{"a":"1"}`)
	diff, err := DiffFromDraft(raw, raw)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	var fields []FieldDiff
	if err := json.Unmarshal(diff, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(fields) != 0 {
		t.Fatalf("expected no diffs, got %d", len(fields))
	}
}
