package schemas

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
)

func ValidateMetadata(contractType contracts.Type, metadata json.RawMessage) error {
	schemaBytes, err := Load(contractType)
	if err != nil {
		return err
	}

	compiler := jsonschema.NewCompiler()
	schemaDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
	if err != nil {
		return fmt.Errorf("parse schema: %w", err)
	}
	schemaURI := fmt.Sprintf("contract-schema://%s/v1.json", contractType)
	if err := compiler.AddResource(schemaURI, schemaDoc); err != nil {
		return fmt.Errorf("register schema: %w", err)
	}
	schema, err := compiler.Compile(schemaURI)
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}

	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(metadata))
	if err != nil {
		return fmt.Errorf("parse metadata: %w", err)
	}
	if err := schema.Validate(instance); err != nil {
		return fmt.Errorf("metadata does not match %s schema: %w", contractType, err)
	}
	return nil
}
