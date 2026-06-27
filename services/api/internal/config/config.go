package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

var ErrSecurityConfig = errors.New("security configuration error")

type Settings struct {
	Environment              string
	GCPProjectID             string
	GCPRegion                string
	CloudSQLConnectionName   string
	DBName                   string
	DBUser                   string
	DBPassword               string
	DatabaseURL              string
	GCSStagingBucket         string
	GCSArchiveBucket         string
	IAPAudience              string
	IAPJWTValidationDisabled bool
	AllowedEmailDomains      string
	AuthUploaderEmails       string
	AuthReviewerEmails       string
	AuthAuditorEmails        string
	AuthAdminEmails          string
	PubSubExtractionTopic    string
}

func Load() *Settings {
	s := &Settings{
		Environment:              envOr("ENVIRONMENT", "dev"),
		GCPProjectID:             os.Getenv("GCP_PROJECT_ID"),
		GCPRegion:                envOr("GCP_REGION", "europe-west3"),
		CloudSQLConnectionName:   os.Getenv("CLOUD_SQL_CONNECTION_NAME"),
		DBName:                   envOr("DB_NAME", "contracts"),
		DBUser:                   envOr("DB_USER", "contract_api"),
		DBPassword:               os.Getenv("DB_PASSWORD"),
		DatabaseURL:              os.Getenv("DATABASE_URL"),
		GCSStagingBucket:         os.Getenv("GCS_STAGING_BUCKET"),
		GCSArchiveBucket:         os.Getenv("GCS_ARCHIVE_BUCKET"),
		IAPAudience:              os.Getenv("IAP_AUDIENCE"),
		IAPJWTValidationDisabled: envBool("IAP_JWT_VALIDATION_DISABLED"),
		AllowedEmailDomains:      os.Getenv("ALLOWED_EMAIL_DOMAINS"),
		AuthUploaderEmails:       os.Getenv("AUTH_UPLOADER_EMAILS"),
		AuthReviewerEmails:       os.Getenv("AUTH_REVIEWER_EMAILS"),
		AuthAuditorEmails:        os.Getenv("AUTH_AUDITOR_EMAILS"),
		AuthAdminEmails:          os.Getenv("AUTH_ADMIN_EMAILS"),
		PubSubExtractionTopic:    os.Getenv("PUBSUB_EXTRACTION_TOPIC"),
	}
	s.applyDevDefaults()
	return s
}

// applyDevDefaults mirrors Terraform dev (enable_iap=false): skip JWT validation when no audience is set.
func (s *Settings) applyDevDefaults() {
	if s.Environment == "prod" {
		return
	}
	if s.IAPAudience != "" {
		return
	}
	if os.Getenv("IAP_JWT_VALIDATION_DISABLED") != "" {
		return
	}
	s.IAPJWTValidationDisabled = true
}

func (s *Settings) DatabaseDSN() string {
	if s.DatabaseURL != "" {
		return s.DatabaseURL
	}
	if s.CloudSQLConnectionName == "" {
		return "file:local.db?_pragma=foreign_keys(1)"
	}
	socket := fmt.Sprintf("/cloudsql/%s", s.CloudSQLConnectionName)
	user := url.UserPassword(s.DBUser, s.DBPassword)
	return fmt.Sprintf(
		"postgres://%s@/%s?host=%s",
		user.String(), s.DBName, socket,
	)
}

func (s *Settings) IsSQLite() bool {
	dsn := s.DatabaseDSN()
	return strings.HasPrefix(dsn, "file:")
}

func (s *Settings) ParsedAllowedEmailDomains() []string {
	return parseCSV(s.AllowedEmailDomains)
}

func (s *Settings) UploaderEmails() map[string]struct{} {
	return parseEmailSet(s.AuthUploaderEmails)
}

func (s *Settings) ReviewerEmails() map[string]struct{} {
	return parseEmailSet(s.AuthReviewerEmails)
}

func (s *Settings) AuditorEmails() map[string]struct{} {
	return parseEmailSet(s.AuthAuditorEmails)
}

func (s *Settings) AdminEmails() map[string]struct{} {
	return parseEmailSet(s.AuthAdminEmails)
}

func (s *Settings) ValidateSecurity() error {
	if s.Environment == "prod" {
		if s.IAPJWTValidationDisabled {
			return fmt.Errorf("%w: IAP JWT validation cannot be disabled in prod", ErrSecurityConfig)
		}
		if s.IAPAudience == "" {
			return fmt.Errorf("%w: IAP audience is required in prod", ErrSecurityConfig)
		}
		if len(s.ParsedAllowedEmailDomains()) == 0 {
			return fmt.Errorf("%w: allowed_email_domains is required in prod", ErrSecurityConfig)
		}
	}
	if !s.IAPJWTValidationDisabled && s.IAPAudience == "" {
		return fmt.Errorf("%w: IAP audience is required when JWT validation is enabled", ErrSecurityConfig)
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes"
}

func parseCSV(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if v := strings.TrimSpace(strings.ToLower(part)); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func parseEmailSet(raw string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		if v := strings.TrimSpace(strings.ToLower(part)); v != "" {
			set[v] = struct{}{}
		}
	}
	return set
}
