package config

import (
	"errors"
	"log/slog"
	"time"

	"github.com/ardanlabs/conf/v3"
)

type Config struct {
	App Application
	Env Environment
	DB  Database
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
	DbName   string `conf:"env:PG_DB,required"`
	SSLMode  bool   `conf:"env:PG_SSL_MODE,default:true"`
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
