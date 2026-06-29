package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/therealagt/ContractManagementTool/libs/common/contracts"
	"github.com/therealagt/ContractManagementTool/services/api/internal/auth"
	"github.com/therealagt/ContractManagementTool/services/api/internal/config"
	"github.com/therealagt/ContractManagementTool/services/api/internal/services"
)

func MountContracts(
	r chi.Router,
	settings *config.Settings,
	validator *auth.IAPValidator,
	uploads *services.UploadService,
	accessLogger func(*http.Request) *auth.AccessLogger,
) {
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireIAP(validator))

		r.With(auth.RequireRoles(settings, auth.RoleUploader, auth.RoleAdmin)).Post("/contracts", func(w http.ResponseWriter, r *http.Request) {
			user, _ := auth.UserFromContext(r.Context())
			logger := accessLogger(r)
			_ = logger.Log(r.Context(), "contract", "upload", nil)

			if err := r.ParseMultipartForm(26 << 20); err != nil {
				http.Error(w, "invalid multipart form", http.StatusBadRequest)
				return
			}

			contractType := contracts.Type(r.FormValue("type"))
			if !contractType.Valid() {
				http.Error(w, "type must be nda or avv", http.StatusBadRequest)
				return
			}

			file, _, err := r.FormFile("file")
			if err != nil {
				http.Error(w, "file is required", http.StatusBadRequest)
				return
			}
			defer file.Close()

			pdf, err := io.ReadAll(io.LimitReader(file, 26<<20))
			if err != nil {
				http.Error(w, "failed to read file", http.StatusBadRequest)
				return
			}

			var partnerID *string
			if p := r.FormValue("partner_id"); p != "" {
				partnerID = &p
			}

			result, err := uploads.Upload(r.Context(), services.UploadInput{
				Actor:     user.Email,
				Type:      contractType,
				PartnerID: partnerID,
				PDF:       pdf,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			status := http.StatusAccepted
			if result.Status == contracts.StatusRejected {
				status = http.StatusUnprocessableEntity
			}
			writeJSON(w, status, map[string]any{
				"contract_id": result.ContractID,
				"status":      result.Status,
			})
		})

		r.Get("/contracts/{id}", func(w http.ResponseWriter, r *http.Request) {
			user, _ := auth.UserFromContext(r.Context())
			contractID := chi.URLParam(r, "id")
			logger := accessLogger(r)
			_ = logger.Log(r.Context(), "contract", "read", &contractID)

			roles := auth.RolesForUser(settings, user)
			isAdmin := auth.HasAnyRole(roles, auth.RoleAdmin)
			isAuditor := auth.HasAnyRole(roles, auth.RoleAuditor)
			isUploader := auth.HasAnyRole(roles, auth.RoleUploader)

			if !isAdmin && !isAuditor && !isUploader {
				auth.WriteForbidden(w, "Insufficient permissions for this action")
				return
			}

			detail, err := uploads.GetContract(r.Context(), contractID, user.Email, isAdmin, isAuditor, isUploader)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "contract not found", http.StatusNotFound)
					return
				}
				if err.Error() == "forbidden" {
					auth.WriteForbidden(w, "Insufficient permissions for this action")
					return
				}
				http.Error(w, "contract not found", http.StatusNotFound)
				return
			}

			writeJSON(w, http.StatusOK, contractDetailResponse(detail))
		})
	})
}

func contractDetailResponse(d *contracts.ContractDetail) map[string]any {
	out := map[string]any{
		"id":          d.ID,
		"type":        d.Type,
		"status":      d.Status,
		"sha256":      d.SHA256,
		"uploaded_by": d.UploadedBy,
		"uploaded_at": d.UploadedAt,
	}
	if d.PartnerID.Valid {
		out["partner_id"] = d.PartnerID.String
	}
	if d.GCSStagingPath.Valid {
		out["gcs_staging_path"] = d.GCSStagingPath.String
	}
	if d.Signature != nil {
		out["signature"] = map[string]any{
			"is_valid":    d.Signature.IsValid,
			"validated_at": d.Signature.ValidatedAt,
		}
	}
	if d.Draft != nil {
		out["extraction_draft"] = json.RawMessage(d.Draft.ExtractedJSON)
		out["schema_version"] = d.Draft.SchemaVersion
	}
	if d.Confirmed != nil {
		out["confirmed_metadata"] = json.RawMessage(d.Confirmed.MetadataJSON)
		out["confirmed_by"] = d.Confirmed.ConfirmedBy
		out["confirmed_at"] = d.Confirmed.ConfirmedAt
		if len(d.Confirmed.DiffFromDraft) > 0 {
			out["diff_from_draft"] = json.RawMessage(d.Confirmed.DiffFromDraft)
		}
	}
	if d.Archive != nil {
		out["archive"] = map[string]any{
			"gcs_path":             d.Archive.GCSPath,
			"sha256":               d.Archive.SHA256,
			"archived_at":          d.Archive.ArchivedAt,
			"retention_expires_at": d.Archive.RetentionExpiresAt,
		}
	}
	if d.LegalHold != nil {
		out["legal_hold"] = map[string]any{
			"reason":     d.LegalHold.Reason,
			"placed_by":  d.LegalHold.PlacedBy,
			"placed_at":  d.LegalHold.PlacedAt,
		}
	}
	return out
}
