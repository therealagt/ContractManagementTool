package handler

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/therealagt/ContractManagementTool/libs/common/pubsub"
	"github.com/therealagt/ContractManagementTool/services/archive-worker/internal/pipeline"
)

type PubSubPush struct {
	Message struct {
		Data       string            `json:"data"`
		MessageID  string            `json:"messageId"`
		Attributes map[string]string `json:"attributes"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

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

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var push PubSubPush
	if err := json.Unmarshal(body, &push); err != nil {
		http.Error(w, "invalid pubsub envelope", http.StatusBadRequest)
		return
	}

	raw, err := base64.StdEncoding.DecodeString(push.Message.Data)
	if err != nil {
		http.Error(w, "invalid message data", http.StatusBadRequest)
		return
	}

	var msg pubsub.ArchiveRequested
	if err := json.Unmarshal(raw, &msg); err != nil {
		http.Error(w, "invalid archive payload", http.StatusBadRequest)
		return
	}

	if err := h.pipeline.Run(r.Context(), msg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
