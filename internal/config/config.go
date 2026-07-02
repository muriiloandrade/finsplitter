package config

import (
	"errors"
	"log/slog"
	"time"

	"github.com/ardanlabs/conf/v3"
)

type Config struct {
	App   Application
	Env   Environment
	DB    Database
	Redis Redis
	OTel  OpenTelemetry
	Logto Logto
}

type Application struct {
	BuildTime   string
	BuildTag    string
	BuildCommit string
	Port        int    `conf:"env:APP_PORT,default:3033"`
	Name        string `conf:"env:APP_NAME,default:finsplitter"`
	Version     string `conf:"env:APP_VERSION,default:dev"`
}

type Environment struct {
	Name      string `conf:"env:ENV_NAME,default:local"`
	LogFormat string `conf:"env:LOG_FORMAT,default:text"`
}

type Database struct {
	User     string `conf:"env:PG_USER,required"`
	Password string `conf:"env:PG_PASS,required"`
	Host     string `conf:"env:PG_HOST,required"`
	Port     int    `conf:"env:PG_PORT,default:5432"`
	DBName   string `conf:"env:PG_DB,required"`
	SSLMode  string `conf:"env:PG_SSL_MODE,default:require"`
	Schema   string `conf:"env:PG_SCHEMA,default:public"`
	Pool     PoolConfig
}

type PoolConfig struct {
	MaxConns          int32         `conf:"env:PG_MAX_CONNS,default:10"`
	MinConns          int32         `conf:"env:PG_MIN_CONNS,default:1"`
	MaxConnLifetime   time.Duration `conf:"env:PG_MAX_CONN_LIFETIME,default:1h"`
	MaxConnIdleTime   time.Duration `conf:"env:PG_MAX_CONN_IDLE_TIME,default:10m"`
	HealthCheckPeriod time.Duration `conf:"env:PG_HEALTH_CHECK_PERIOD,default:1m"`
	ConnectTimeout    time.Duration `conf:"env:PG_CONNECT_TIMEOUT,default:15s"`
}

type Redis struct {
	URL string `conf:"env:REDIS_URL,default:redis://localhost:6379/0"`
}

type Logto struct {
	OIDCEndpoint      string `conf:"env:LOGTO_OIDC_ENDPOINT,default:http://localhost:3001/oidc"`
	OIDCIssuer        string `conf:"env:LOGTO_ISSUER,default:http://localhost:3001/oidc"`
	ManagementBaseURL string `conf:"env:LOGTO_ENDPOINT,default:http://localhost:3001"`
	MgmtClientID      string `conf:"env:LOGTO_MGMT_CLIENT_ID"`
	MgmtClientSecret  string `conf:"env:LOGTO_MGMT_CLIENT_SECRET"`
	MgmtAPIResource   string `conf:"env:LOGTO_MGMT_API_RESOURCE"`
	AppClientID       string `conf:"env:LOGTO_APP_CLIENT_ID"`
	// AppClientSecret is reserved for future use (e.g. token refresh via
	// confidential client grant). It is not currently consumed by any code.
	AppClientSecret string `conf:"env:LOGTO_APP_CLIENT_SECRET"`
}

type OpenTelemetry struct {
	Enabled         bool          `conf:"env:OTEL_ENABLED,default:false"`
	ServiceName     string        `conf:"env:OTEL_SERVICE_NAME,default:finsplitter"`
	ExporterURL     string        `conf:"env:OTEL_EXPORTER_OTLP_ENDPOINT,default:http://localhost:4318"`
	Insecure        bool          `conf:"env:OTEL_EXPORTER_INSECURE,default:true"`
	EnableTraces    bool          `conf:"env:OTEL_ENABLE_TRACES,default:true"`
	EnableMetrics   bool          `conf:"env:OTEL_ENABLE_METRICS,default:true"`
	EnableLogs      bool          `conf:"env:OTEL_ENABLE_LOGS,default:true"`
	SamplerRatio    float64       `conf:"env:OTEL_SAMPLER_RATIO,default:1.0"`
	ExporterTimeout time.Duration `conf:"env:OTEL_EXPORTER_TIMEOUT,default:30s"`
	ExportInterval  time.Duration `conf:"env:OTEL_EXPORT_INTERVAL,default:5s"`
}

func LoadEnv(buildTag, buildCommit, buildTime string) *Config {
	var cfg Config
	if _, err := conf.Parse("", &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return nil
		}
		slog.Default().Error("fail to load configurations", slog.Any("error", err))
		return nil
	}

	cfg.App.BuildCommit = buildCommit
	cfg.App.BuildTag = buildTag
	cfg.App.BuildTime = buildTime

	return &cfg
}
