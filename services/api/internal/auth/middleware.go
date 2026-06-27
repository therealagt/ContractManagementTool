package auth

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"strings"

	"github.com/therealagt/ContractManagementTool/libs/common/audit"
	"github.com/therealagt/ContractManagementTool/services/api/internal/config"
)

type contextKey string

const userContextKey contextKey = "iapUser"

func UserFromContext(ctx context.Context) (*IAPUser, bool) {
	user, ok := ctx.Value(userContextKey).(*IAPUser)
	return user, ok
}

func WithUser(ctx context.Context, user *IAPUser) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

type AccessLogger struct {
	db   *sql.DB
	user *IAPUser
	ip   *string
}

func NewAccessLogger(db *sql.DB, user *IAPUser, ip *string) *AccessLogger {
	return &AccessLogger{db: db, user: user, ip: ip}
}

func (l *AccessLogger) Log(ctx context.Context, resourceType, action string, resourceID *string) error {
	_, err := audit.RecordAccessEvent(ctx, l.db, l.user.Email, resourceType, action, resourceID, l.ip)
	return err
}

func ClientIP(r *http.Request) *string {
	if forwarded := r.Header.Get("x-forwarded-for"); forwarded != "" {
		ip := strings.TrimSpace(strings.Split(forwarded, ",")[0])
		return &ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip := r.RemoteAddr
		return &ip
	}
	return &host
}

func RequireIAP(validator *IAPValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get(IAPHeader)
			if token == "" {
				if devBypass := validator.devBypassUser(); devBypass != nil {
					next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), devBypass)))
					return
				}
				WriteUnauthorized(w, "Missing IAP JWT")
				return
			}

			user, err := validator.Validate(r.Context(), token)
			if err != nil {
				WriteUnauthorized(w, err.Error())
				return
			}

			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
		})
	}
}

func RequireRoles(settings *config.Settings, required ...Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				WriteUnauthorized(w, "Missing IAP JWT")
				return
			}
			granted := RolesForUser(settings, user)
			if !HasAnyRole(granted, required...) {
				WriteForbidden(w, "Insufficient permissions for this action")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
