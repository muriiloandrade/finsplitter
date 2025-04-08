package config

import (
	"errors"
	"log/slog"

	"github.com/ardanlabs/conf/v3"
)

type Config struct {
	App Application
	Env Environment
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
