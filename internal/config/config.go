// Package config loads the application runtime configuration from default
// values and environment variables.
package config

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

// defaultAppPort is the fallback HTTP port used when APP_PORT is not set.
const defaultAppPort = 3033

// defaultDBPort is the fallback PostgreSQL port used when PG_PORT is not set.
const defaultDBPort = 5432

// defaultPoolMaxConns is the fallback maximum database connections used when
// PG_MAX_CONNS is not set.
const defaultPoolMaxConns = 10

// defaultPoolMinConns is the fallback minimum database connections used when
// PG_MIN_CONNS is not set.
const defaultPoolMinConns = 1

// Config is the root application configuration.
type Config struct {
	App   Application   `koanf:"app"`
	Env   Environment   `koanf:"env"`
	DB    Database      `koanf:"db"`
	Redis Redis         `koanf:"redis"`
	OTel  OpenTelemetry `koanf:"otel"`
	Logto Logto         `koanf:"logto"`
}

// Application holds the application level settings.
type Application struct {
	BuildTime   string
	BuildTag    string
	BuildCommit string
	Port        int    `koanf:"port"`
	Name        string `koanf:"name"`
	Version     string `koanf:"version"`
}

// Environment holds the environment name and log format settings.
type Environment struct {
	Name      string `koanf:"name"`
	LogFormat string `koanf:"log_format"`
}

// Database holds the database connection settings and pool configuration.
type Database struct {
	User     string     `koanf:"user"`
	Password string     `koanf:"password"`
	Host     string     `koanf:"host"`
	Port     int        `koanf:"port"`
	DBName   string     `koanf:"db_name"`
	SSLMode  string     `koanf:"ssl_mode"`
	Schema   string     `koanf:"schema"`
	Pool     PoolConfig `koanf:"pool"`
}

// PoolConfig holds the database connection pool settings.
type PoolConfig struct {
	MaxConns          int32         `koanf:"max_conns"`
	MinConns          int32         `koanf:"min_conns"`
	MaxConnLifetime   time.Duration `koanf:"max_conn_lifetime"`
	MaxConnIdleTime   time.Duration `koanf:"max_conn_idle_time"`
	HealthCheckPeriod time.Duration `koanf:"health_check_period"`
	ConnectTimeout    time.Duration `koanf:"connect_timeout"`
}

// Redis holds the cache client settings.
type Redis struct {
	URL string `koanf:"url"`
}

// Logto holds the Logto OIDC and management API client settings.
type Logto struct {
	OIDCEndpoint      string `koanf:"oidc_endpoint"`
	OIDCIssuer        string `koanf:"issuer"`
	ManagementBaseURL string `koanf:"endpoint"`
	MgmtClientID      string `koanf:"mgmt_client_id"`
	MgmtClientSecret  string `koanf:"mgmt_client_secret"`
	MgmtAPIResource   string `koanf:"mgmt_api_resource"`
	AppClientID       string `koanf:"app_client_id"`
	AppClientSecret   string `koanf:"app_client_secret"`
	DeviceAppClientID string `koanf:"device_app_client_id"`
}

// OpenTelemetry holds the OpenTelemetry provider settings.
type OpenTelemetry struct {
	Enabled         bool          `koanf:"enabled"`
	ServiceName     string        `koanf:"service_name"`
	ExporterURL     string        `koanf:"exporter_otlp_endpoint"`
	Insecure        bool          `koanf:"exporter_insecure"`
	EnableTraces    bool          `koanf:"enable_traces"`
	EnableMetrics   bool          `koanf:"enable_metrics"`
	EnableLogs      bool          `koanf:"enable_logs"`
	SamplerRatio    float64       `koanf:"sampler_ratio"`
	ExporterTimeout time.Duration `koanf:"exporter_timeout"`
	ExportInterval  time.Duration `koanf:"export_interval"`
}

// LoadEnv loads the application configuration from defaults and overrides
// provided via environment variables. It returns nil when the configuration
// cannot be loaded or fails validation.
func LoadEnv(buildTag, buildCommit, buildTime string) *Config {
	k := koanf.New(".")

	for key, value := range defaultConfig() {
		if err := k.Set(key, value); err != nil {
			slog.Default().Error("fail to set default configurations", slog.Any("error", err))
			return nil
		}
	}

	if err := k.Load(env.Provider("", ".", envKeyToKoanfKey), nil); err != nil {
		slog.Default().Error("fail to load environment configurations", slog.Any("error", err))
		return nil
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		slog.Default().Error("fail to unmarshal configurations", slog.Any("error", err))
		return nil
	}

	cfg.App.BuildCommit = buildCommit
	cfg.App.BuildTag = buildTag
	cfg.App.BuildTime = buildTime

	if err := cfg.validate(); err != nil {
		slog.Default().Error("fail to validate configurations", slog.Any("error", err))
		return nil
	}

	return &cfg
}

// defaultConfig returns fallback configuration values used when the matching
// environment variables are not set.
func defaultConfig() map[string]any {
	return map[string]any{
		"app.port":                    defaultAppPort,
		"app.name":                    "finsplitter",
		"app.version":                 "dev",
		"env.name":                    "local",
		"env.log_format":              "text",
		"db.port":                     defaultDBPort,
		"db.ssl_mode":                 "require",
		"db.schema":                   "public",
		"db.pool.max_conns":           defaultPoolMaxConns,
		"db.pool.min_conns":           defaultPoolMinConns,
		"db.pool.max_conn_lifetime":   "1h",
		"db.pool.max_conn_idle_time":  "10m",
		"db.pool.health_check_period": "1m",
		"db.pool.connect_timeout":     "15s",
		"redis.url":                   "redis://localhost:6379/0",
		"logto.oidc_endpoint":         "http://localhost:3001/oidc",
		"logto.issuer":                "http://localhost:3001/oidc",
		"logto.endpoint":              "http://localhost:3001",
		"otel.enabled":                false,
		"otel.service_name":           "finsplitter",
		"otel.exporter_otlp_endpoint": "http://localhost:4318",
		"otel.exporter_insecure":      true,
		"otel.enable_traces":          true,
		"otel.enable_metrics":         true,
		"otel.enable_logs":            true,
		"otel.sampler_ratio":          1.0,
		"otel.exporter_timeout":       "30s",
		"otel.export_interval":        "5s",
	}
}

// envKeyToKoanfKey maps an environment variable name to its koanf key path and
// returns an empty string for environment variables that are not used by the
// application.
func envKeyToKoanfKey(envKey string) string {
	envKeys := map[string]string{ //nolint:gosec // env var names are configuration keys, not credentials
		"APP_PORT":                    "app.port",
		"APP_NAME":                    "app.name",
		"APP_VERSION":                 "app.version",
		"ENV_NAME":                    "env.name",
		"LOG_FORMAT":                  "env.log_format",
		"PG_USER":                     "db.user",
		"PG_PASS":                     "db.password",
		"PG_HOST":                     "db.host",
		"PG_PORT":                     "db.port",
		"PG_DB":                       "db.db_name",
		"PG_SSL_MODE":                 "db.ssl_mode",
		"PG_SCHEMA":                   "db.schema",
		"PG_MAX_CONNS":                "db.pool.max_conns",
		"PG_MIN_CONNS":                "db.pool.min_conns",
		"PG_MAX_CONN_LIFETIME":        "db.pool.max_conn_lifetime",
		"PG_MAX_CONN_IDLE_TIME":       "db.pool.max_conn_idle_time",
		"PG_HEALTH_CHECK_PERIOD":      "db.pool.health_check_period",
		"PG_CONNECT_TIMEOUT":          "db.pool.connect_timeout",
		"REDIS_URL":                   "redis.url",
		"LOGTO_OIDC_ENDPOINT":         "logto.oidc_endpoint",
		"LOGTO_ISSUER":                "logto.issuer",
		"LOGTO_ENDPOINT":              "logto.endpoint",
		"LOGTO_MGMT_CLIENT_ID":        "logto.mgmt_client_id",
		"LOGTO_MGMT_CLIENT_SECRET":    "logto.mgmt_client_secret",
		"LOGTO_MGMT_API_RESOURCE":     "logto.mgmt_api_resource",
		"LOGTO_APP_CLIENT_ID":         "logto.app_client_id",
		"LOGTO_APP_CLIENT_SECRET":     "logto.app_client_secret",
		"LOGTO_DEVICE_APP_CLIENT_ID":  "logto.device_app_client_id",
		"OTEL_ENABLED":                "otel.enabled",
		"OTEL_SERVICE_NAME":           "otel.service_name",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "otel.exporter_otlp_endpoint",
		"OTEL_EXPORTER_INSECURE":      "otel.exporter_insecure",
		"OTEL_ENABLE_TRACES":          "otel.enable_traces",
		"OTEL_ENABLE_METRICS":         "otel.enable_metrics",
		"OTEL_ENABLE_LOGS":            "otel.enable_logs",
		"OTEL_SAMPLER_RATIO":          "otel.sampler_ratio",
		"OTEL_EXPORTER_TIMEOUT":       "otel.exporter_timeout",
		"OTEL_EXPORT_INTERVAL":        "otel.export_interval",
	}

	return envKeys[strings.ToUpper(envKey)]
}

// validate ensures that all required environment variables are configured.
func (c *Config) validate() error {
	missing := make([]string, 0)
	if c.DB.User == "" {
		missing = append(missing, "PG_USER")
	}
	if c.DB.Password == "" {
		missing = append(missing, "PG_PASS")
	}
	if c.DB.Host == "" {
		missing = append(missing, "PG_HOST")
	}
	if c.DB.DBName == "" {
		missing = append(missing, "PG_DB")
	}
	if len(missing) > 0 {
		return errors.New("missing required env vars: " + strings.Join(missing, ", "))
	}

	return nil
}
