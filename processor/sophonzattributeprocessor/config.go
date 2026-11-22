// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"errors"
	"os"
	"strings"
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
	ErrEmptyDecryptKey         = errors.New("sophonzattributeprocessor: required decryption key not provided. To fix, provide the key or set 'encrypted_user_id' to false")
	ErrInvalidServiceKeyMode   = errors.New("sophonzattributeprocessor: service_key_mode must be one of off, shadow, enforce")
)

// ServiceKeyMode selects what the processor does with the SDK app key
// (service.key) on each resource.
type ServiceKeyMode string

const (
	// ServiceKeyModeOff leaves telemetry untouched -- no lookup, no tenant
	// stamp, no drops.
	ServiceKeyModeOff ServiceKeyMode = "off"
	// ServiceKeyModeShadow resolves the tenant and stamps it, but never drops.
	// The counters it emits are what tells you whether enforce would be safe.
	ServiceKeyModeShadow ServiceKeyMode = "shadow"
	// ServiceKeyModeEnforce additionally drops resources whose key is missing
	// or unknown to the metadata cache.
	ServiceKeyModeEnforce ServiceKeyMode = "enforce"
)

type Config struct {
	Enabled         bool             `mapstructure:"enabled"`
	FetchInterval   string           `mapstructure:"fetch_interval"`
	Clickhouse      ClickhouseConfig `mapstructure:"clickhouse"`
	EncryptedUserID bool             `mapstructure:"encrypted_user_id"`
	ServiceKeyMode  string           `mapstructure:"service_key_mode"`
	// ServiceKeyCheck is the pre-mode boolean, kept so deployed configs keep
	// working: true means service_key_mode: enforce. Ignored once
	// service_key_mode is set explicitly.
	ServiceKeyCheck bool `mapstructure:"service_key_check"`
	Interval        time.Duration
	KeyMode         ServiceKeyMode
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

	switch mode := ServiceKeyMode(strings.ToLower(strings.TrimSpace(cfg.ServiceKeyMode))); mode {
	case "":
		if cfg.ServiceKeyCheck {
			cfg.KeyMode = ServiceKeyModeEnforce
		} else {
			cfg.KeyMode = ServiceKeyModeOff
		}
	case ServiceKeyModeOff, ServiceKeyModeShadow, ServiceKeyModeEnforce:
		cfg.KeyMode = mode
	default:
		return ErrInvalidServiceKeyMode
	}

	if cfg.EncryptedUserID {
		key := os.Getenv("AES_DECRYPT_KEY")
		if key == "" {
			return ErrEmptyDecryptKey
		}
		crypto.InitializeKey(key)
	}
	return nil
}
