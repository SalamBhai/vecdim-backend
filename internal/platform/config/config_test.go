package config

import (
	"os"
	"testing"
)

func TestConfigLoadDevelopment(t *testing.T) {
	// Ensure APP_ENV is development
	os.Setenv("APP_ENV", "development")
	os.Unsetenv("DB_PASSWORD")

	cfg := Load()

	if cfg.Port == "" {
		t.Errorf("expected non-empty Port, got %q", cfg.Port)
	}
	if cfg.DBPassword != "vecdim_secret_2026" {
		t.Errorf("expected default dev password 'vecdim_secret_2026', got %q", cfg.DBPassword)
	}
	if cfg.DSN() == "" {
		t.Errorf("expected valid DSN string, got empty")
	}
	if cfg.CronInterval() <= 0 {
		t.Errorf("expected positive CronInterval, got %v", cfg.CronInterval())
	}
}
