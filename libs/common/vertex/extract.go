package vertex

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/libs/common/schemas"
)

type ExtractionResult struct {
	ExtractedJSON   json.RawMessage
	ConfidenceFlags map[string]any
	Model           string
	PromptVersion   string
	SchemaVersion   string
}

type Extractor struct {
	client        *genai.Client
	model         string
	promptVersion string
}

func NewExtractor(ctx context.Context, projectID, region, model, promptVersion string) (*Extractor, error) {
	if model == "" {
		model = "gemini-2.0-flash"
	}
	if promptVersion == "" {
		promptVersion = "v1"
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  projectID,
		Location: region,
	})
	if err != nil {
		return nil, fmt.Errorf("vertex client: %w", err)
	}
	return &Extractor{client: client, model: model, promptVersion: promptVersion}, nil
}

func (e *Extractor) Extract(ctx context.Context, contractType contracts.Type, pdf []byte) (*ExtractionResult, error) {
	schemaBytes, err := schemas.Load(contractType)
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`Extract contract metadata from the attached signed PDF.
Return ONLY valid JSON matching this schema (draft for human review, not legal truth):
%s
Add a top-level object field "_confidence" mapping field names to "high", "medium", or "low".`, string(schemaBytes))

	contents := []*genai.Content{{
		Role: genai.RoleUser,
		Parts: []*genai.Part{
			{Text: prompt},
			{InlineData: &genai.Blob{MIMEType: "application/pdf", Data: pdf}},
		},
	}}

	resp, err := e.client.Models.GenerateContent(ctx, e.model, contents, &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		Temperature:      genai.Ptr(float32(0.1)),
	})
	if err != nil {
		return nil, fmt.Errorf("gemini extract: %w", err)
	}

	text := strings.TrimSpace(resp.Text())
	if text == "" {
		return nil, fmt.Errorf("empty gemini response")
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini json: %w", err)
	}

	confidence := map[string]any{}
	if c, ok := parsed["_confidence"].(map[string]any); ok {
		confidence = c
		delete(parsed, "_confidence")
	}

	extracted, err := json.Marshal(parsed)
	if err != nil {
		return nil, err
	}

	return &ExtractionResult{
		ExtractedJSON:   extracted,
		ConfidenceFlags: confidence,
		Model:           e.model,
		PromptVersion:   e.promptVersion,
		SchemaVersion:   schemas.Version(contractType),
	}, nil
}

// ExtractWithConfidence returns extraction result and confidence JSON for DB storage.
func (e *Extractor) ExtractWithConfidence(ctx context.Context, contractType contracts.Type, pdf []byte) (*ExtractionResult, []byte, error) {
	res, err := e.Extract(ctx, contractType, pdf)
	if err != nil {
		return nil, nil, err
	}
	flags, _ := json.Marshal(res.ConfidenceFlags)
	return res, flags, nil
}

func (e *Extractor) Close() error {
	return nil
}
