package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/therealagt/ContractManagementTool/services/api/internal/auth"
	"github.com/therealagt/ContractManagementTool/services/api/internal/config"
	"github.com/therealagt/ContractManagementTool/services/api/internal/services"
	"github.com/therealagt/ContractManagementTool/services/api/internal/webui"
)

func MountReview(
	r chi.Router,
	settings *config.Settings,
	validator *auth.IAPValidator,
	review *services.ReviewService,
	accessLogger func(*http.Request) *auth.AccessLogger,
) {
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireIAP(validator))

		r.With(auth.RequireRoles(settings, auth.RoleReviewer, auth.RoleAdmin)).Get("/contracts/review-queue", func(w http.ResponseWriter, r *http.Request) {
			user, _ := auth.UserFromContext(r.Context())
			logger := accessLogger(r)
			_ = logger.Log(r.Context(), "contract", "review_queue", nil)

			items, err := review.ListQueue(r.Context(), user.Email)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			out := make([]map[string]any, 0, len(items))
			for _, item := range items {
				entry := map[string]any{
					"id":          item.ID,
					"type":        item.Type,
					"status":      item.Status,
					"uploaded_by": item.UploadedBy,
					"uploaded_at": item.UploadedAt,
					"has_draft":   item.HasDraft,
				}
				if item.PartnerID.Valid {
					entry["partner_id"] = item.PartnerID.String
				}
				out = append(out, entry)
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": out})
		})

		r.With(auth.RequireRoles(settings, auth.RoleReviewer, auth.RoleAdmin)).Post("/contracts/{id}/confirm", func(w http.ResponseWriter, r *http.Request) {
			user, _ := auth.UserFromContext(r.Context())
			contractID := chi.URLParam(r, "id")
			logger := accessLogger(r)
			_ = logger.Log(r.Context(), "contract", "confirm", &contractID)

			body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			var req struct {
				Metadata json.RawMessage `json:"metadata"`
			}
			if err := json.Unmarshal(body, &req); err != nil || len(req.Metadata) == 0 {
				http.Error(w, "metadata is required", http.StatusBadRequest)
				return
			}

			result, err := review.Confirm(r.Context(), services.ConfirmInput{
				Actor:        user.Email,
				ContractID:   contractID,
				MetadataJSON: req.Metadata,
			})
			if err != nil {
				if services.IsSODViolation(err) {
					auth.WriteForbidden(w, err.Error())
					return
				}
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "contract not found", http.StatusNotFound)
					return
				}
				if err.Error() == "contract not pending review" {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"contract_id": result.ContractID,
				"status":      result.Status,
			})
		})

		r.With(auth.RequireRoles(settings, auth.RoleReviewer, auth.RoleAdmin)).Post("/contracts/{id}/reject", func(w http.ResponseWriter, r *http.Request) {
			user, _ := auth.UserFromContext(r.Context())
			contractID := chi.URLParam(r, "id")
			logger := accessLogger(r)
			_ = logger.Log(r.Context(), "contract", "reject", &contractID)

			var req struct {
				Reason string `json:"reason"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)

			err := review.Reject(r.Context(), services.RejectInput{
				Actor:      user.Email,
				ContractID: contractID,
				Reason:     req.Reason,
			})
			if err != nil {
				if services.IsSODViolation(err) {
					auth.WriteForbidden(w, err.Error())
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"contract_id": contractID,
				"status":      "rejected",
			})
		})

		r.With(auth.RequireRoles(settings, auth.RoleReviewer, auth.RoleAdmin)).Get("/review", func(w http.ResponseWriter, r *http.Request) {
			_ = accessLogger(r).Log(r.Context(), "contract", "review_ui", nil)
			webui.ReviewPage(w, r)
		})
	})
}
