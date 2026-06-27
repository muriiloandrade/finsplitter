package config

import (
	"os"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	// Set required environment variables
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")
	defer os.Unsetenv("PG_USER")
	defer os.Unsetenv("PG_PASS")
	defer os.Unsetenv("PG_HOST")
	defer os.Unsetenv("PG_DB")

	cfg := LoadEnv("test", "abc123", "2024-01-01")

	if cfg == nil {
		t.Fatal("expected config to load, got nil")
		return
	}

	// Application defaults
	if cfg.App.Port != 3033 {
		t.Errorf("expected default port 3033, got %d", cfg.App.Port)
	}
	if cfg.App.Name != "finsplitter" {
		t.Errorf("expected default name finsplitter, got %s", cfg.App.Name)
	}
	if cfg.App.Version != "dev" {
		t.Errorf("expected default version dev, got %s", cfg.App.Version)
	}

	// Environment defaults
	if cfg.Env.Name != "local" {
		t.Errorf("expected default env name local, got %s", cfg.Env.Name)
	}
	if cfg.Env.LogFormat != "text" {
		t.Errorf("expected default log format text, got %s", cfg.Env.LogFormat)
	}

	// Database defaults
	if cfg.DB.Port != 5432 {
		t.Errorf("expected default db port 5432, got %d", cfg.DB.Port)
	}
	if cfg.DB.SSLMode != "require" {
		t.Errorf("expected default ssl mode require, got %s", cfg.DB.SSLMode)
	}
	if cfg.DB.Schema != "public" {
		t.Errorf("expected default schema public, got %s", cfg.DB.Schema)
	}

	// Pool defaults
	if cfg.DB.Pool.MaxConns != 10 {
		t.Errorf("expected default max conns 10, got %d", cfg.DB.Pool.MaxConns)
	}
	if cfg.DB.Pool.MinConns != 1 {
		t.Errorf("expected default min conns 1, got %d", cfg.DB.Pool.MinConns)
	}

	// OpenTelemetry defaults
	if cfg.OTel.Enabled != false {
		t.Errorf("expected default otel enabled false, got %v", cfg.OTel.Enabled)
	}
	if cfg.OTel.ServiceName != "finsplitter" {
		t.Errorf("expected default otel service name finsplitter, got %s", cfg.OTel.ServiceName)
	}
	if cfg.OTel.Insecure != true {
		t.Errorf("expected default otel insecure true, got %v", cfg.OTel.Insecure)
	}
}

func TestConfigFromEnvironment(t *testing.T) {
	// Set custom environment variables
	t.Setenv("PG_USER", "customuser")
	t.Setenv("PG_PASS", "custompass")
	t.Setenv("PG_HOST", "customhost")
	t.Setenv("PG_DB", "customdb")
	t.Setenv("APP_PORT", "8080")
	t.Setenv("APP_NAME", "custom-app")
	t.Setenv("ENV_NAME", "production")
	defer os.Unsetenv("PG_USER")
	defer os.Unsetenv("PG_PASS")
	defer os.Unsetenv("PG_HOST")
	defer os.Unsetenv("PG_DB")
	defer os.Unsetenv("APP_PORT")
	defer os.Unsetenv("APP_NAME")
	defer os.Unsetenv("ENV_NAME")

	cfg := LoadEnv("test", "abc123", "2024-01-01")

	if cfg == nil {
		t.Fatal("expected config to load, got nil")
		return
	}

	// Verify environment overrides
	if cfg.DB.User != "customuser" {
		t.Errorf("expected user customuser, got %s", cfg.DB.User)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.App.Port)
	}
	if cfg.App.Name != "custom-app" {
		t.Errorf("expected name custom-app, got %s", cfg.App.Name)
	}
	if cfg.Env.Name != "production" {
		t.Errorf("expected env production, got %s", cfg.Env.Name)
	}
}

func TestOpenTelemetryConfig(t *testing.T) {
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "my-service")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel:4318")
	defer os.Unsetenv("PG_USER")
	defer os.Unsetenv("PG_PASS")
	defer os.Unsetenv("PG_HOST")
	defer os.Unsetenv("PG_DB")
	defer os.Unsetenv("OTEL_ENABLED")
	defer os.Unsetenv("OTEL_SERVICE_NAME")
	defer os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	cfg := LoadEnv("test", "abc123", "2024-01-01")

	if cfg == nil {
		t.Fatal("expected config to load, got nil")
		return
	}

	if !cfg.OTel.Enabled {
		t.Error("expected otel enabled true")
	}
	if cfg.OTel.ServiceName != "my-service" {
		t.Errorf("expected otel service name my-service, got %s", cfg.OTel.ServiceName)
	}
	if cfg.OTel.ExporterURL != "http://otel:4318" {
		t.Errorf("expected otel endpoint http://otel:4318, got %s", cfg.OTel.ExporterURL)
	}
}

func TestBuildInfo(t *testing.T) {
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")
	defer os.Unsetenv("PG_USER")
	defer os.Unsetenv("PG_PASS")
	defer os.Unsetenv("PG_HOST")
	defer os.Unsetenv("PG_DB")

	cfg := LoadEnv("v1.0.0", "abc123def", "2024-06-15T10:30:00Z")

	if cfg == nil {
		t.Fatal("expected config to load, got nil")
		return
	}

	if cfg.App.BuildTag != "v1.0.0" {
		t.Errorf("expected build tag v1.0.0, got %s", cfg.App.BuildTag)
	}
	if cfg.App.BuildCommit != "abc123def" {
		t.Errorf("expected build commit abc123def, got %s", cfg.App.BuildCommit)
	}
	if cfg.App.BuildTime != "2024-06-15T10:30:00Z" {
		t.Errorf("expected build time 2024-06-15T10:30:00Z, got %s", cfg.App.BuildTime)
	}
}
