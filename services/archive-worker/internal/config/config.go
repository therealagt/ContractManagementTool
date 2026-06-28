package config

import (
	"net/url"
	"os"
	"strconv"
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
	GCSArchiveBucket       string
	RetentionYears         int
}

func Load() *Settings {
	years, _ := strconv.Atoi(envOr("RETENTION_YEARS", "10"))
	return &Settings{
		Environment:            envOr("ENVIRONMENT", "dev"),
		GCPProjectID:             os.Getenv("GCP_PROJECT_ID"),
		GCPRegion:                envOr("GCP_REGION", "europe-west3"),
		CloudSQLConnectionName:   os.Getenv("CLOUD_SQL_CONNECTION_NAME"),
		DBName:                   envOr("DB_NAME", "contracts"),
		DBUser:                   envOr("DB_USER", "contract_api"),
		DBPassword:               os.Getenv("DB_PASSWORD"),
		DatabaseURL:              os.Getenv("DATABASE_URL"),
		GCSStagingBucket:         os.Getenv("GCS_STAGING_BUCKET"),
		GCSArchiveBucket:         os.Getenv("GCS_ARCHIVE_BUCKET"),
		RetentionYears:           years,
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
