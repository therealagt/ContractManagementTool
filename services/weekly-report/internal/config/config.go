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
	CloudSQLConnectionName string
	DBName                 string
	DBUser                 string
	DBPassword             string
	DatabaseURL            string
	ReviewSLADays          int
	ReportEmailOps         string
	ReportEmailAudit       string
	EmailFrom              string
	SMTPHost               string
	SMTPPort               int
	SMTPUser               string
	SMTPPassword           string
}

func Load() *Settings {
	sla, _ := strconv.Atoi(envOr("REVIEW_SLA_DAYS", "7"))
	port, _ := strconv.Atoi(envOr("SMTP_PORT", "587"))
	return &Settings{
		Environment:            envOr("ENVIRONMENT", "dev"),
		GCPProjectID:           os.Getenv("GCP_PROJECT_ID"),
		CloudSQLConnectionName: os.Getenv("CLOUD_SQL_CONNECTION_NAME"),
		DBName:                 envOr("DB_NAME", "contracts"),
		DBUser:                 envOr("DB_USER", "contract_api"),
		DBPassword:             os.Getenv("DB_PASSWORD"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		ReviewSLADays:          sla,
		ReportEmailOps:         os.Getenv("REPORT_EMAIL_OPS"),
		ReportEmailAudit:       os.Getenv("REPORT_EMAIL_AUDIT"),
		EmailFrom:              os.Getenv("EMAIL_FROM"),
		SMTPHost:               os.Getenv("SMTP_HOST"),
		SMTPPort:               port,
		SMTPUser:               os.Getenv("SMTP_USER"),
		SMTPPassword:           os.Getenv("SMTP_PASSWORD"),
	}
}

func (s *Settings) DatabaseDSN() string {
	if s.DatabaseURL != "" {
		return s.DatabaseURL
	}
	if s.CloudSQLConnectionName == "" {
		return "file:.local-data/weekly-report.db?_pragma=foreign_keys(1)"
	}
	socket := fmt.Sprintf("/cloudsql/%s", s.CloudSQLConnectionName)
	user := url.UserPassword(s.DBUser, s.DBPassword)
	return fmt.Sprintf("postgres://%s@/%s?host=%s", user.String(), s.DBName, socket)
}

func (s *Settings) IsSQLite() bool {
	return strings.HasPrefix(s.DatabaseDSN(), "file:")
}

func (s *Settings) Recipients() []string {
	return []string{s.ReportEmailOps, s.ReportEmailAudit}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
