package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/therealagt/ContractManagementTool/services/api/internal/auth"
	"github.com/therealagt/ContractManagementTool/services/api/internal/config"
	"github.com/therealagt/ContractManagementTool/services/api/internal/services"
	"github.com/therealagt/ContractManagementTool/services/api/internal/webui"
)

func MountAudit(
	r chi.Router,
	settings *config.Settings,
	validator *auth.IAPValidator,
	reports *services.AuditReportService,
	legalHold *services.LegalHoldService,
	accessLogger func(*http.Request) *auth.AccessLogger,
) {
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireIAP(validator))

		r.With(auth.RequireRoles(settings, auth.RoleAuditor, auth.RoleAdmin)).Get("/audit/summary", func(w http.ResponseWriter, r *http.Request) {
			_ = accessLogger(r).Log(r.Context(), "audit", "summary", nil)
			summary, err := reports.ExecutiveSummary(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, summary)
		})

		r.With(auth.RequireRoles(settings, auth.RoleAuditor, auth.RoleAdmin)).Get("/audit/control-matrix", func(w http.ResponseWriter, r *http.Request) {
			_ = accessLogger(r).Log(r.Context(), "audit", "control_matrix", nil)
			matrix, err := reports.ControlMatrix(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"controls": matrix})
		})

		r.With(auth.RequireRoles(settings, auth.RoleAuditor, auth.RoleAdmin)).Get("/audit/trail", func(w http.ResponseWriter, r *http.Request) {
			_ = accessLogger(r).Log(r.Context(), "audit", "trail", nil)
			var contractID *string
			if id := r.URL.Query().Get("contract_id"); id != "" {
				contractID = &id
			}
			trail, err := reports.AuditTrail(r.Context(), contractID, 500)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"events": trail})
		})

		r.With(auth.RequireRoles(settings, auth.RoleAuditor, auth.RoleAdmin)).Post("/audit/evidence", func(w http.ResponseWriter, r *http.Request) {
			_ = accessLogger(r).Log(r.Context(), "audit", "evidence_export", nil)

			var req struct {
				ContractIDs []string `json:"contract_ids"`
			}
			body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			if len(body) > 0 {
				_ = json.Unmarshal(body, &req)
			}

			data, err := reports.EvidencePackage(r.Context(), req.ContractIDs)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", `attachment; filename="evidence-package.zip"`)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
		})

		r.With(auth.RequireRoles(settings, auth.RoleAuditor, auth.RoleAdmin)).Get("/audit", func(w http.ResponseWriter, r *http.Request) {
			_ = accessLogger(r).Log(r.Context(), "audit", "dashboard", nil)
			webui.AuditPage(w, r)
		})

		r.With(auth.RequireRoles(settings, auth.RoleAdmin)).Post("/contracts/{id}/legal-hold", func(w http.ResponseWriter, r *http.Request) {
			user, _ := auth.UserFromContext(r.Context())
			contractID := chi.URLParam(r, "id")
			_ = accessLogger(r).Log(r.Context(), "contract", "legal_hold_place", &contractID)

			var req struct {
				Reason string `json:"reason"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
				http.Error(w, "reason is required", http.StatusBadRequest)
				return
			}

			if err := legalHold.Place(r.Context(), user.Email, contractID, req.Reason); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "contract not found", http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"contract_id": contractID, "legal_hold": true})
		})

		r.With(auth.RequireRoles(settings, auth.RoleAdmin)).Delete("/contracts/{id}/legal-hold", func(w http.ResponseWriter, r *http.Request) {
			user, _ := auth.UserFromContext(r.Context())
			contractID := chi.URLParam(r, "id")
			_ = accessLogger(r).Log(r.Context(), "contract", "legal_hold_release", &contractID)

			if err := legalHold.Release(r.Context(), user.Email, contractID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "contract not found", http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"contract_id": contractID, "legal_hold": false})
		})
	})
}
