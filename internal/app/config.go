package app

import (
	"errors"
	"time"
)

type ServerAuthConfig struct {
	APIKey  string            // default, required
	APIKeys map[string]string // additional, optional
}

type ServerConfig struct {
	Addr string `default:"0.0.0.0:9000"`
}

type DatabaseDDBConfig struct {
	Table string
}

type DatabaseConfig struct {
	DDB DatabaseDDBConfig
}

type APIConfig struct {
	Enabled bool `default:"true"`
	Auth    ServerAuthConfig
}

type SchedulerConfig struct {
	Enabled      bool          `default:"true"`
	PollInterval time.Duration `default:"5s"`
}

type ExecutorConfig struct {
	Enabled      bool          `default:"true"`
	PollInterval time.Duration `default:"5s"`
	MaxLogSize   int           `default:"4096"`
}

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	API       APIConfig
	Scheduler SchedulerConfig
	Executor  ExecutorConfig

	Now func() time.Time
}

func (cfg Config) Validate() error {
	if cfg.API.Auth.APIKey == "" {
		return errors.New("api.auth.apikey is required")
	}

	for k := range cfg.API.Auth.APIKeys {
		if k == "default" {
			return errors.New("api.auth.apikeys cannot have 'default' key")
		}
	}

	if cfg.Database.DDB.Table == "" {
		return errors.New("database.ddb.table is required")
	}

	if cfg.Scheduler.Enabled {
		if cfg.Scheduler.PollInterval <= time.Second {
			return errors.New("scheduler.poll_interval is too short")
		}
	}

	if cfg.Executor.Enabled {
		if cfg.Executor.PollInterval <= time.Second {
			return errors.New("executor.poll_interval is too short")
		}
	}

	return nil
}
