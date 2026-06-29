package config

import (
	"fmt"
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
	GCSArchiveBucket       string
	BigQueryDataset        string
	ReviewSLADays          int
}

func Load() *Settings {
	sla, _ := strconv.Atoi(envOr("REVIEW_SLA_DAYS", "7"))
	return &Settings{
		Environment:            envOr("ENVIRONMENT", "dev"),
		GCPProjectID:           os.Getenv("GCP_PROJECT_ID"),
		GCPRegion:              envOr("GCP_REGION", "europe-west3"),
		CloudSQLConnectionName: os.Getenv("CLOUD_SQL_CONNECTION_NAME"),
		DBName:                 envOr("DB_NAME", "contracts"),
		DBUser:                 envOr("DB_USER", "contract_api"),
		DBPassword:             os.Getenv("DB_PASSWORD"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		GCSArchiveBucket:       os.Getenv("GCS_ARCHIVE_BUCKET"),
		BigQueryDataset:        os.Getenv("BIGQUERY_DATASET"),
		ReviewSLADays:          sla,
	}
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
	return fmt.Sprintf("postgres://%s@/%s?host=%s", user.String(), s.DBName, socket)
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
