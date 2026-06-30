package schemas_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/schemas"
)

func TestValidateMetadataNDA(t *testing.T) {
	dir, err := filepath.Abs("../../../schemas")
	if err != nil {
		t.Fatalf("schemas path: %v", err)
	}
	t.Setenv("SCHEMAS_DIR", dir)

	valid := json.RawMessage(`{
		"disclosing_party": "A GmbH",
		"receiving_party": "B GmbH",
		"effective_date": "2024-01-01",
		"confidentiality_term_months": 12,
		"governing_law": "Germany"
	}`)
	if err := schemas.ValidateMetadata(contracts.TypeNDA, valid); err != nil {
		t.Fatalf("valid metadata: %v", err)
	}

	invalid := json.RawMessage(`{"disclosing_party": "only one field"}`)
	if err := schemas.ValidateMetadata(contracts.TypeNDA, invalid); err == nil {
		t.Fatal("expected schema validation error")
	}
}
