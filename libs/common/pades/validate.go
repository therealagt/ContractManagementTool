package pades

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/digitorus/pdfsign/verify"
)

type ValidationResult struct {
	Valid      bool     `json:"valid"`
	Errors     []string `json:"errors,omitempty"`
	SignerCN   string   `json:"signer_cn,omitempty"`
	SignedAt   string   `json:"signed_at,omitempty"`
	CertIssuer string   `json:"cert_issuer,omitempty"`
	SHA256     string   `json:"sha256"`
}

func Validate(pdf []byte) (*ValidationResult, error) {
	result := &ValidationResult{SHA256: hashBytes(pdf)}
	if len(pdf) == 0 {
		result.Valid = false
		result.Errors = []string{"empty PDF"}
		return result, nil
	}

	reader := bytes.NewReader(pdf)
	opts := verify.DefaultVerifyOptions()
	opts.AllowUntrustedRoots = true // dev/embedded certs; prod PKI via PADES_TRUSTED_CERTS_PATH later

	resp, err := verify.VerifyWithOptions(reader, int64(len(pdf)), opts)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("verify PDF: %v", err))
		return result, nil
	}
	if resp.Error != "" {
		result.Valid = false
		result.Errors = append(result.Errors, resp.Error)
		return result, nil
	}
	if len(resp.Signers) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "no digital signature found")
		return result, nil
	}

	sig := resp.Signers[len(resp.Signers)-1]
	result.SignerCN = sig.Name
	if sig.SignatureTime != nil {
		result.SignedAt = sig.SignatureTime.UTC().Format(time.RFC3339)
	} else if sig.TimeStamp != nil {
		result.SignedAt = sig.TimeStamp.Time.UTC().Format(time.RFC3339)
	}
	if len(sig.Certificates) > 0 && sig.Certificates[0].Certificate != nil {
		result.CertIssuer = sig.Certificates[0].Certificate.Issuer.CommonName
	}

	if !sig.ValidSignature {
		result.Valid = false
		result.Errors = append(result.Errors, "signature cryptographic validation failed")
		if sig.RevokedCertificate {
			result.Errors = append(result.Errors, "signing certificate revoked")
		}
		return result, nil
	}

	result.Valid = true
	return result, nil
}

func (r *ValidationResult) ToJSON() json.RawMessage {
	b, _ := json.Marshal(r)
	return b
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
