package routes

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"

	"github.com/therealagt/ContractManagementTool/services/api/internal/auth"
	"github.com/therealagt/ContractManagementTool/services/api/internal/config"
)

func Mount(r chi.Router, settings *config.Settings, validator *auth.IAPValidator, dbHandler func(*http.Request) *auth.AccessLogger) {
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.RequireIAP(validator))
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			user, _ := auth.UserFromContext(r.Context())
			logger := dbHandler(r)
			_ = logger.Log(r.Context(), "api", "status", nil)

			roles := auth.RolesForUser(settings, user)
			roleValues := make([]string, 0, len(roles))
			for _, role := range roles {
				roleValues = append(roleValues, string(role))
			}
			slices.Sort(roleValues)

			writeJSON(w, http.StatusOK, map[string]any{
				"status":      "ok",
				"environment": settings.Environment,
				"actor":       user.Email,
				"roles":       roleValues,
			})
		})

		r.With(auth.RequireRoles(settings, auth.RoleAdmin)).Get("/status/admin", func(w http.ResponseWriter, r *http.Request) {
			user, _ := auth.UserFromContext(r.Context())
			logger := dbHandler(r)
			_ = logger.Log(r.Context(), "api", "status_admin", nil)
			writeJSON(w, http.StatusOK, map[string]string{
				"status": "ok",
				"actor":  user.Email,
			})
		})
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
