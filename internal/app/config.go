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
	Addr string
	Auth ServerAuthConfig
}

type DatabaseDDBConfig struct {
	Table string
}

type DatabaseConfig struct {
	DDB DatabaseDDBConfig
}

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig

	Now func() time.Time
}

func (cfg Config) Validate() error {
	if cfg.Server.Auth.APIKey == "" {
		return errors.New("server.auth.apikey is required")
	}

	for k := range cfg.Server.Auth.APIKeys {
		if k == "default" {
			return errors.New("server.auth.apikeys cannot have 'default' key")
		}
	}

	if cfg.Database.DDB.Table == "" {
		return errors.New("database.ddb.table is required")
	}

	return nil
}
