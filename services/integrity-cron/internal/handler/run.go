package handler

import (
	"encoding/json"
	"net/http"

	"github.com/therealagt/ContractManagementTool/services/integrity-cron/internal/pipeline"
)

type Handler struct {
	pipeline *pipeline.Pipeline
}

func New(p *pipeline.Pipeline) *Handler {
	return &Handler{pipeline: p}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := h.pipeline.Run(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"checked_count": result.CheckedCount,
		"failed_count":  result.FailedCount,
		"chain_valid":   result.ChainValid,
	})
}
