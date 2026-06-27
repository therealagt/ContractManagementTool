package config

import (
	"strings"
	"testing"
)

func TestLoadDevDefaultsDisableIAPWithoutAudience(t *testing.T) {
	t.Setenv("ENVIRONMENT", "dev")
	t.Setenv("IAP_AUDIENCE", "")
	t.Setenv("IAP_JWT_VALIDATION_DISABLED", "")

	s := Load()
	if err := s.ValidateSecurity(); err != nil {
		t.Fatalf("ValidateSecurity() = %v", err)
	}
	if !s.IAPJWTValidationDisabled {
		t.Fatal("expected IAP JWT validation disabled in dev without audience")
	}
}

func TestLoadProdRequiresIAPAudience(t *testing.T) {
	t.Setenv("ENVIRONMENT", "prod")
	t.Setenv("IAP_AUDIENCE", "")
	t.Setenv("IAP_JWT_VALIDATION_DISABLED", "false")
	t.Setenv("ALLOWED_EMAIL_DOMAINS", "example.com")

	s := Load()
	if err := s.ValidateSecurity(); err == nil {
		t.Fatal("expected error when prod has no IAP audience")
	}
}

func TestDatabaseDSNEncodesPassword(t *testing.T) {
	s := &Settings{
		DBUser:                 "contract_api",
		DBPassword:             "p@ss:word",
		DBName:                 "contracts",
		CloudSQLConnectionName: "proj:europe-west3:db",
	}
	dsn := s.DatabaseDSN()
	if !strings.Contains(dsn, "%40") {
		t.Fatalf("expected URL-encoded @ in DSN, got %q", dsn)
	}
}
