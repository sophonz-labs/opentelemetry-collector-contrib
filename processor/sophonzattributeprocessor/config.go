// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"errors"
	"os"
	"time"

	"go.opentelemetry.io/collector/component"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/crypto"
)

const defaultDatabaseName = "default"

var (
	ErrEmptyClickhouseHost     = errors.New("sophonzattributeprocessor: clickhouse host is empty")
	ErrEmptyClickhousePort     = errors.New("sophonzattributeprocessor: clickhouse port is empty")
	ErrEmptyClickhouseUserName = errors.New("sophonzattributeprocessor: clickhouse username is empty")
	ErrInvalidFetchInterval    = errors.New("sophonzattributeprocessor: invalid fetch_interval")
	ErrFetchIntervalTooShort   = errors.New("sophonzattributeprocessor: fetch_interval must be at least 60 seconds")
	ErrEmptyDecryptKeyKINFA    = errors.New("sophonzattributeprocessor: required KINFA decryption key not provided. To fix, provide the key or set 'encrypted_user_id_kinfa' to false")
)

type Config struct {
	Enabled              bool             `mapstructure:"enabled"`
	FetchInterval        string           `mapstructure:"fetch_interval"`
	Clickhouse           ClickhouseConfig `mapstructure:"clickhouse"`
	EncryptedUserIDKINFA bool             `mapstructure:"encrypted_user_id_kinfa"`
	ServiceKeyCheck      bool             `mapstructure:"service_key_check"`
	Interval             time.Duration
}

type ClickhouseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	UserName string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

var _ component.Config = (*Config)(nil)

func (cfg *Config) Validate() error {
	if cfg.Clickhouse.Host == "" {
		return ErrEmptyClickhouseHost
	}
	if cfg.Clickhouse.Port == 0 {
		return ErrEmptyClickhousePort
	}
	if cfg.Clickhouse.UserName == "" {
		return ErrEmptyClickhouseUserName
	}
	if cfg.Clickhouse.Database == "" {
		cfg.Clickhouse.Database = defaultDatabaseName
	}
	interval, err := time.ParseDuration(cfg.FetchInterval)
	if err != nil {
		return ErrInvalidFetchInterval
	}
	if interval < 60*time.Second {
		return ErrFetchIntervalTooShort
	}
	cfg.Interval = interval

	if cfg.EncryptedUserIDKINFA {
		key := os.Getenv("AES_DECRYPT_KEY_KINFA")
		if key == "" {
			return ErrEmptyDecryptKeyKINFA
		}
		crypto.InitializeKey(key)
	}
	return nil
}
