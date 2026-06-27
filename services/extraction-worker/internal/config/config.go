package config

import (
	"net/url"
	"os"
	"strings"
)

type Settings struct {
	Environment            string
	GCPProjectID           string
	GCPRegion              string
	CloudSQLConnectionName string
	DBName                 string
	DBUser                 string
	DBPassword             string
	DatabaseURL            string
	GCSStagingBucket       string
	GeminiModel            string
	PromptVersion          string
}

func Load() *Settings {
	return &Settings{
		Environment:            envOr("ENVIRONMENT", "dev"),
		GCPProjectID:           os.Getenv("GCP_PROJECT_ID"),
		GCPRegion:              envOr("GCP_REGION", "europe-west3"),
		CloudSQLConnectionName: os.Getenv("CLOUD_SQL_CONNECTION_NAME"),
		DBName:                 envOr("DB_NAME", "contracts"),
		DBUser:                 envOr("DB_USER", "contract_api"),
		DBPassword:             os.Getenv("DB_PASSWORD"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		GCSStagingBucket:       os.Getenv("GCS_STAGING_BUCKET"),
		GeminiModel:            envOr("GEMINI_MODEL", "gemini-2.0-flash"),
		PromptVersion:          envOr("PROMPT_VERSION", "v1"),
	}
}

func (s *Settings) DatabaseDSN() string {
	if s.DatabaseURL != "" {
		return s.DatabaseURL
	}
	if s.CloudSQLConnectionName == "" {
		return "file:local.db?_pragma=foreign_keys(1)"
	}
	socket := "/cloudsql/" + s.CloudSQLConnectionName
	user := url.UserPassword(s.DBUser, s.DBPassword)
	return "postgres://" + user.String() + "@/" + s.DBName + "?host=" + socket
}

func (s *Settings) IsSQLite() bool {
	return strings.HasPrefix(s.DatabaseDSN(), "file:")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
