package handler

import (
	"encoding/json"
	"net/http"

	"github.com/therealagt/ContractManagementTool/services/weekly-report/internal/pipeline"
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

	if err := h.pipeline.Run(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
