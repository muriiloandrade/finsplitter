package config

import (
	"os"
	"testing"
	"time"
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
		t.Errorf("expected otel service name finsplitter, got %s", cfg.OTel.ServiceName)
	}
	if cfg.OTel.Insecure != true {
		t.Errorf("expected otel insecure true, got %v", cfg.OTel.Insecure)
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

func TestMissingRequiredEnvVars(t *testing.T) {
	// Required variables are set to empty values so the loader must fail.
	t.Setenv("PG_USER", "")
	t.Setenv("PG_PASS", "")
	t.Setenv("PG_HOST", "")
	t.Setenv("PG_DB", "")

	cfg := LoadEnv("test", "abc123", "2024-01-01")

	if cfg != nil {
		t.Fatal("expected nil config when required env vars are missing, got non-nil")
	}
}

func TestEnvApplicationMapping(t *testing.T) {
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")

	tests := []struct {
		name     string
		envVar   string
		envValue string
		assert   func(t *testing.T, cfg *Config)
	}{
		{
			name:     "app port from environment",
			envVar:   "APP_PORT",
			envValue: "7070",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.App.Port != 7070 {
					t.Errorf("expected app.port 7070, got %d", cfg.App.Port)
				}
			},
		},
		{
			name:     "app name from environment",
			envVar:   "APP_NAME",
			envValue: "mapped-app",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.App.Name != "mapped-app" {
					t.Errorf("expected app.name mapped-app, got %s", cfg.App.Name)
				}
			},
		},
		{
			name:     "app version from environment",
			envVar:   "APP_VERSION",
			envValue: "v9.9.9",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.App.Version != "v9.9.9" {
					t.Errorf("expected app.version v9.9.9, got %s", cfg.App.Version)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Setenv(tt.envVar, tt.envValue)

		cfg := LoadEnv("test", "abc123", "2024-01-01")
		if cfg == nil {
			t.Fatalf("expected config to load for env var %s", tt.envVar)
		}
		tt.assert(t, cfg)
	}
}

func TestEnvEnvironmentMapping(t *testing.T) {
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")

	tests := []struct {
		name     string
		envVar   string
		envValue string
		assert   func(t *testing.T, cfg *Config)
	}{
		{
			name:     "environment name from environment",
			envVar:   "ENV_NAME",
			envValue: "production",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.Env.Name != "production" {
					t.Errorf("expected env.name production, got %s", cfg.Env.Name)
				}
			},
		},
		{
			name:     "log format from environment",
			envVar:   "LOG_FORMAT",
			envValue: "json",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.Env.LogFormat != "json" {
					t.Errorf("expected env.log_format json, got %s", cfg.Env.LogFormat)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Setenv(tt.envVar, tt.envValue)

		cfg := LoadEnv("test", "abc123", "2024-01-01")
		if cfg == nil {
			t.Fatalf("expected config to load for env var %s", tt.envVar)
		}
		tt.assert(t, cfg)
	}
}

func TestEnvDatabaseMapping(t *testing.T) {
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")

	tests := []struct {
		name     string
		envVar   string
		envValue string
		assert   func(t *testing.T, cfg *Config)
	}{
		{
			name:     "database user from environment",
			envVar:   "PG_USER",
			envValue: "mapuser",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.User != "mapuser" {
					t.Errorf("expected db.user mapuser, got %s", cfg.DB.User)
				}
			},
		},
		{
			name:     "database password from environment",
			envVar:   "PG_PASS",
			envValue: "mapsecret",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.Password != "mapsecret" {
					t.Errorf("expected db.password mapsecret, got %s", cfg.DB.Password)
				}
			},
		},
		{
			name:     "database host from environment",
			envVar:   "PG_HOST",
			envValue: "maphost",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.Host != "maphost" {
					t.Errorf("expected db.host maphost, got %s", cfg.DB.Host)
				}
			},
		},
		{
			name:     "database name from environment",
			envVar:   "PG_DB",
			envValue: "mapdb",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.DBName != "mapdb" {
					t.Errorf("expected db.db_name mapdb, got %s", cfg.DB.DBName)
				}
			},
		},
		{
			name:     "database port from environment",
			envVar:   "PG_PORT",
			envValue: "6543",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.Port != 6543 {
					t.Errorf("expected db.port 6543, got %d", cfg.DB.Port)
				}
			},
		},
		{
			name:     "database ssl mode from environment",
			envVar:   "PG_SSL_MODE",
			envValue: "disable",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.SSLMode != "disable" {
					t.Errorf("expected db.ssl_mode disable, got %s", cfg.DB.SSLMode)
				}
			},
		},
		{
			name:     "database schema from environment",
			envVar:   "PG_SCHEMA",
			envValue: "otherschema",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.Schema != "otherschema" {
					t.Errorf("expected db.schema otherschema, got %s", cfg.DB.Schema)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Setenv(tt.envVar, tt.envValue)

		cfg := LoadEnv("test", "abc123", "2024-01-01")
		if cfg == nil {
			t.Fatalf("expected config to load for env var %s", tt.envVar)
		}
		tt.assert(t, cfg)
	}
}

func TestEnvDatabasePoolMapping(t *testing.T) {
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")

	tests := []struct {
		name     string
		envVar   string
		envValue string
		assert   func(t *testing.T, cfg *Config)
	}{
		{
			name:     "database min conns from environment",
			envVar:   "PG_MIN_CONNS",
			envValue: "5",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.Pool.MinConns != 5 {
					t.Errorf("expected db.pool.min_conns 5, got %d", cfg.DB.Pool.MinConns)
				}
			},
		},
		{
			name:     "database max conns from environment",
			envVar:   "PG_MAX_CONNS",
			envValue: "25",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.Pool.MaxConns != 25 {
					t.Errorf("expected db.pool.max_conns 25, got %d", cfg.DB.Pool.MaxConns)
				}
			},
		},
		{
			name:     "database max conn lifetime from environment",
			envVar:   "PG_MAX_CONN_LIFETIME",
			envValue: "30m",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.Pool.MaxConnLifetime != 30*time.Minute {
					t.Errorf("expected db.pool.max_conn_lifetime 30m, got %s", cfg.DB.Pool.MaxConnLifetime)
				}
			},
		},
		{
			name:     "database max conn idle time from environment",
			envVar:   "PG_MAX_CONN_IDLE_TIME",
			envValue: "20m",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.Pool.MaxConnIdleTime != 20*time.Minute {
					t.Errorf("expected db.pool.max_conn_idle_time 20m, got %s", cfg.DB.Pool.MaxConnIdleTime)
				}
			},
		},
		{
			name:     "database health check period from environment",
			envVar:   "PG_HEALTH_CHECK_PERIOD",
			envValue: "2m",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.Pool.HealthCheckPeriod != 2*time.Minute {
					t.Errorf("expected db.pool.health_check_period 2m, got %s", cfg.DB.Pool.HealthCheckPeriod)
				}
			},
		},
		{
			name:     "database connect timeout from environment",
			envVar:   "PG_CONNECT_TIMEOUT",
			envValue: "10s",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.DB.Pool.ConnectTimeout != 10*time.Second {
					t.Errorf("expected db.pool.connect_timeout 10s, got %s", cfg.DB.Pool.ConnectTimeout)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Setenv(tt.envVar, tt.envValue)

		cfg := LoadEnv("test", "abc123", "2024-01-01")
		if cfg == nil {
			t.Fatalf("expected config to load for env var %s", tt.envVar)
		}
		tt.assert(t, cfg)
	}
}

func TestEnvLogtoMapping(t *testing.T) {
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")

	tests := []struct {
		name     string
		envVar   string
		envValue string
		assert   func(t *testing.T, cfg *Config)
	}{
		{
			name:     "logto oidc endpoint from environment",
			envVar:   "LOGTO_OIDC_ENDPOINT",
			envValue: "https://oidc.example.com",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.Logto.OIDCEndpoint != "https://oidc.example.com" {
					t.Errorf("expected logto.oidc_endpoint https://oidc.example.com, got %s", cfg.Logto.OIDCEndpoint)
				}
			},
		},
		{
			name:     "logto issuer from environment",
			envVar:   "LOGTO_ISSUER",
			envValue: "https://issuer.example.com",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.Logto.OIDCIssuer != "https://issuer.example.com" {
					t.Errorf("expected logto.issuer https://issuer.example.com, got %s", cfg.Logto.OIDCIssuer)
				}
			},
		},
		{
			name:     "logto management base url from environment",
			envVar:   "LOGTO_ENDPOINT",
			envValue: "https://logto.example.com",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.Logto.ManagementBaseURL != "https://logto.example.com" {
					t.Errorf("expected logto.endpoint https://logto.example.com, got %s", cfg.Logto.ManagementBaseURL)
				}
			},
		},
		{
			name:     "logto management api resource from environment",
			envVar:   "LOGTO_MGMT_API_RESOURCE",
			envValue: "https://api-resource",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.Logto.MgmtAPIResource != "https://api-resource" {
					t.Errorf("expected logto.mgmt_api_resource https://api-resource, got %s", cfg.Logto.MgmtAPIResource)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Setenv(tt.envVar, tt.envValue)

		cfg := LoadEnv("test", "abc123", "2024-01-01")
		if cfg == nil {
			t.Fatalf("expected config to load for env var %s", tt.envVar)
		}
		tt.assert(t, cfg)
	}
}

func TestEnvLogtoClientsMapping(t *testing.T) {
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")

	tests := []struct {
		name     string
		envVar   string
		envValue string
		assert   func(t *testing.T, cfg *Config)
	}{
		{
			name:     "logto management client id from environment",
			envVar:   "LOGTO_MGMT_CLIENT_ID",
			envValue: "mgmt-client-id",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.Logto.MgmtClientID != "mgmt-client-id" {
					t.Errorf("expected logto.mgmt_client_id mgmt-client-id, got %s", cfg.Logto.MgmtClientID)
				}
			},
		},
		{
			name:     "logto management client secret from environment",
			envVar:   "LOGTO_MGMT_CLIENT_SECRET",
			envValue: "mgmt-client-secret",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.Logto.MgmtClientSecret != "mgmt-client-secret" {
					t.Errorf("expected logto.mgmt_client_secret mgmt-client-secret, got %s", cfg.Logto.MgmtClientSecret)
				}
			},
		},
		{
			name:     "logto app client id from environment",
			envVar:   "LOGTO_APP_CLIENT_ID",
			envValue: "app-client-id",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.Logto.AppClientID != "app-client-id" {
					t.Errorf("expected logto.app_client_id app-client-id, got %s", cfg.Logto.AppClientID)
				}
			},
		},
		{
			name:     "logto app client secret from environment",
			envVar:   "LOGTO_APP_CLIENT_SECRET",
			envValue: "app-client-secret",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.Logto.AppClientSecret != "app-client-secret" {
					t.Errorf("expected logto.app_client_secret app-client-secret, got %s", cfg.Logto.AppClientSecret)
				}
			},
		},
		{
			name:     "logto device app client id from environment",
			envVar:   "LOGTO_DEVICE_APP_CLIENT_ID",
			envValue: "device-client-id",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.Logto.DeviceAppClientID != "device-client-id" {
					t.Errorf(
						"expected logto.device_app_client_id device-client-id, got %s",
						cfg.Logto.DeviceAppClientID,
					)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Setenv(tt.envVar, tt.envValue)

		cfg := LoadEnv("test", "abc123", "2024-01-01")
		if cfg == nil {
			t.Fatalf("expected config to load for env var %s", tt.envVar)
		}
		tt.assert(t, cfg)
	}
}

func TestEnvOpenTelemetryMapping(t *testing.T) {
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")

	tests := []struct {
		name     string
		envVar   string
		envValue string
		assert   func(t *testing.T, cfg *Config)
	}{
		{
			name:     "otel enabled from environment",
			envVar:   "OTEL_ENABLED",
			envValue: "true",
			assert: func(t *testing.T, cfg *Config) {
				if !cfg.OTel.Enabled {
					t.Error("expected otel.enabled true")
				}
			},
		},
		{
			name:     "otel service name from environment",
			envVar:   "OTEL_SERVICE_NAME",
			envValue: "mapped-service",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.OTel.ServiceName != "mapped-service" {
					t.Errorf("expected otel.service_name mapped-service, got %s", cfg.OTel.ServiceName)
				}
			},
		},
		{
			name:     "otel exporter otlp endpoint from environment",
			envVar:   "OTEL_EXPORTER_OTLP_ENDPOINT",
			envValue: "http://otel:4318",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.OTel.ExporterURL != "http://otel:4318" {
					t.Errorf("expected otel.exporter_otlp_endpoint http://otel:4318, got %s", cfg.OTel.ExporterURL)
				}
			},
		},
		{
			name:     "otel exporter insecure from environment",
			envVar:   "OTEL_EXPORTER_INSECURE",
			envValue: "false",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.OTel.Insecure {
					t.Error("expected otel.exporter_insecure false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Setenv(tt.envVar, tt.envValue)

		cfg := LoadEnv("test", "abc123", "2024-01-01")
		if cfg == nil {
			t.Fatalf("expected config to load for env var %s", tt.envVar)
		}
		tt.assert(t, cfg)
	}
}

func TestEnvOpenTelemetryBehaviorMapping(t *testing.T) {
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")

	tests := []struct {
		name     string
		envVar   string
		envValue string
		assert   func(t *testing.T, cfg *Config)
	}{
		{
			name:     "otel enable traces from environment",
			envVar:   "OTEL_ENABLE_TRACES",
			envValue: "false",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.OTel.EnableTraces {
					t.Error("expected otel.enable_traces false")
				}
			},
		},
		{
			name:     "otel enable metrics from environment",
			envVar:   "OTEL_ENABLE_METRICS",
			envValue: "false",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.OTel.EnableMetrics {
					t.Error("expected otel.enable_metrics false")
				}
			},
		},
		{
			name:     "otel enable logs from environment",
			envVar:   "OTEL_ENABLE_LOGS",
			envValue: "false",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.OTel.EnableLogs {
					t.Error("expected otel.enable_logs false")
				}
			},
		},
		{
			name:     "otel sampler ratio from environment",
			envVar:   "OTEL_SAMPLER_RATIO",
			envValue: "0.5",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.OTel.SamplerRatio != 0.5 {
					t.Errorf("expected otel.sampler_ratio 0.5, got %v", cfg.OTel.SamplerRatio)
				}
			},
		},
		{
			name:     "otel exporter timeout from environment",
			envVar:   "OTEL_EXPORTER_TIMEOUT",
			envValue: "10s",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.OTel.ExporterTimeout != 10*time.Second {
					t.Errorf("expected otel.exporter_timeout 10s, got %s", cfg.OTel.ExporterTimeout)
				}
			},
		},
		{
			name:     "otel export interval from environment",
			envVar:   "OTEL_EXPORT_INTERVAL",
			envValue: "2s",
			assert: func(t *testing.T, cfg *Config) {
				if cfg.OTel.ExportInterval != 2*time.Second {
					t.Errorf("expected otel.export_interval 2s, got %s", cfg.OTel.ExportInterval)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Setenv(tt.envVar, tt.envValue)

		cfg := LoadEnv("test", "abc123", "2024-01-01")
		if cfg == nil {
			t.Fatalf("expected config to load for env var %s", tt.envVar)
		}
		tt.assert(t, cfg)
	}
}

func TestEnvRedisMapping(t *testing.T) {
	t.Setenv("PG_USER", "testuser")
	t.Setenv("PG_PASS", "testpass")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_DB", "testdb")

	t.Setenv("REDIS_URL", "redis://redis:6379/1")

	cfg := LoadEnv("test", "abc123", "2024-01-01")

	if cfg == nil {
		t.Fatal("expected config to load, got nil")
		return
	}

	if cfg.Redis.URL != "redis://redis:6379/1" {
		t.Errorf("expected redis.url redis://redis:6379/1, got %s", cfg.Redis.URL)
	}
}
