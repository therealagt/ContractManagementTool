package auth

import (
	"net/http"
	"slices"
	"strings"

	"github.com/therealagt/ContractManagementTool/services/api/internal/config"
)

type Role string

const (
	RoleUploader Role = "uploader"
	RoleReviewer Role = "reviewer"
	RoleAuditor  Role = "auditor"
	RoleAdmin    Role = "admin"
)

var allRoles = []Role{RoleUploader, RoleReviewer, RoleAuditor, RoleAdmin}

func RolesForUser(settings *config.Settings, user *IAPUser) []Role {
	email := strings.ToLower(user.Email)
	if _, ok := settings.AdminEmails()[email]; ok {
		return slices.Clone(allRoles)
	}

	var roles []Role
	if _, ok := settings.UploaderEmails()[email]; ok {
		roles = append(roles, RoleUploader)
	}
	if _, ok := settings.ReviewerEmails()[email]; ok {
		roles = append(roles, RoleReviewer)
	}
	if _, ok := settings.AuditorEmails()[email]; ok {
		roles = append(roles, RoleAuditor)
	}
	return roles
}

func HasAnyRole(granted []Role, required ...Role) bool {
	for _, role := range required {
		if slices.Contains(granted, role) {
			return true
		}
	}
	return false
}

func AssertSeparationOfDuty(actor, other, action string) error {
	if strings.EqualFold(actor, other) {
		return &SODViolation{Action: action}
	}
	return nil
}

type SODViolation struct {
	Action string
}

func (e *SODViolation) Error() string {
	return "separation of duties violation: same actor cannot " + e.Action
}

func WriteForbidden(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusForbidden)
}

func WriteUnauthorized(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusUnauthorized)
}
