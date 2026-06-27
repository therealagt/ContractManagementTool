package schemas

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
)

func Load(contractType contracts.Type) ([]byte, error) {
	dir := os.Getenv("SCHEMAS_DIR")
	if dir == "" {
		dir = "schemas"
	}
	path := filepath.Join(dir, string(contractType), "v1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load schema %s: %w", path, err)
	}
	return data, nil
}

func Version(contractType contracts.Type) string {
	return contractType.SchemaVersion()
}
