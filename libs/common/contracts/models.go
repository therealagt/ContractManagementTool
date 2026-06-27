package contracts

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Contract struct {
	ID             string
	Type           Type
	Status         Status
	PartnerID      sql.NullString
	GCSStagingPath sql.NullString
	SHA256         string
	UploadedBy     string
	UploadedAt     time.Time
}

type SignatureValidation struct {
	ContractID           string
	IsValid              bool
	SignerCN             sql.NullString
	SignedAt             sql.NullTime
	CertIssuer           sql.NullString
	ValidationResultJSON json.RawMessage
	ValidatedAt          time.Time
}

type ExtractionDraft struct {
	ContractID      string
	ExtractedJSON   json.RawMessage
	GeminiModel     sql.NullString
	PromptVersion   sql.NullString
	SchemaVersion   string
	ConfidenceFlags json.RawMessage
	ExtractedAt     time.Time
}

type ContractDetail struct {
	Contract
	Signature *SignatureValidation
	Draft     *ExtractionDraft
}
