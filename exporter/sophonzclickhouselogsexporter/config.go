// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhouselogsexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhouselogsexporter"

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/schema"
)

// Config defines configuration for the SOPHONZ ClickHouse logs exporter.
type Config struct {
	// collectorVersion is the build version of the collector. This is overridden when an exporter is initialized.
	collectorVersion string
	driverName       string // for testing

	TimeoutSettings           exporterhelper.TimeoutConfig `mapstructure:",squash"`
	configretry.BackOffConfig `mapstructure:"retry_on_failure"`
	// QueueSettings is the sending queue configuration.
	// As of collector v0.150+ this is wrapped in configoptional.Optional.
	QueueSettings configoptional.Optional[exporterhelper.QueueBatchConfig] `mapstructure:"sending_queue"`

	// Host is the clickhouse host.
	Host string `mapstructure:"host"`
	// Port is the clickhouse port.
	Port string `mapstructure:"port"`
	// Username is the authentication username.
	Username string `mapstructure:"username"`
	// Password is the authentication password.
	Password configopaque.String `mapstructure:"password"`
	// Database is the database name to export.
	Database string `mapstructure:"database"`
	// ConnectionParams is the extra connection parameters with map format. for example compression/dial_timeout
	ConnectionParams map[string]string `mapstructure:"connection_params"`
	// LogsTableName is the table name for logs.
	LogsTableName string `mapstructure:"logs_table_name"`
	// TTL is The data time-to-live example 30m, 48h. 0 means no ttl.
	TTL time.Duration `mapstructure:"ttl"`
	// TableEngine is the table engine to use. default is `MergeTree()`.
	TableEngine TableEngine `mapstructure:"table_engine"`
	// ClusterName if set will append `ON CLUSTER` with the provided name when creating tables.
	ClusterName string `mapstructure:"cluster_name"`
	// CreateSchema if set to true will run the DDL for creating the database and tables. default is true.
	CreateSchema bool `mapstructure:"create_schema"`
	// Compress controls the compression algorithm. Valid options: `none` (disabled), `zstd`, `lz4` (default), `gzip`, `deflate`, `br`, `true` (lz4).
	Compress string `mapstructure:"compress"`
	// AsyncInsert if true will enable async inserts. Default is `true`.
	// Ignored if async inserts are configured in the `endpoint` or `connection_params`.
	// Async inserts may still be overridden server-side.
	AsyncInsert bool `mapstructure:"async_insert"`
}

// TableEngine defines the ENGINE string value when creating the table.
type TableEngine struct {
	Name   string `mapstructure:"name"`
	Params string `mapstructure:"params"`
}

const (
	defaultDatabase        = schema.DatabaseLogs
	defaultTableEngineName = "MergeTree"
)

var (
	errConfigNoEndpoint      = errors.New("host and port must be specified")
	errConfigInvalidEndpoint = errors.New("endpoint must be url format")
)

// Validate the ClickHouse server configuration.
func (cfg *Config) Validate() (err error) {
	if cfg.Host == "" || cfg.Port == "" {
		err = errors.Join(err, errConfigNoEndpoint)
	}
	dsn, e := cfg.buildDSN()
	if e != nil {
		err = errors.Join(err, e)
	}

	// Validate DSN with clickhouse driver.
	// Last chance to catch invalid config.
	if _, e := clickhouse.ParseDSN(dsn); e != nil {
		err = errors.Join(err, e)
	}

	return err
}

func (cfg *Config) buildDSN() (string, error) {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	endpoint := fmt.Sprintf("tcp://%s", addr)
	dsnURL, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errConfigInvalidEndpoint, err.Error())
	}

	queryParams := dsnURL.Query()

	// Add connection params to query params.
	for k, v := range cfg.ConnectionParams {
		queryParams.Set(k, v)
	}

	// Enable TLS if scheme is https. This flag is necessary to support https connections.
	if dsnURL.Scheme == "https" {
		queryParams.Set("secure", "true")
	}

	// Use async_insert from config if not specified in DSN.
	if !queryParams.Has("async_insert") {
		queryParams.Set("async_insert", fmt.Sprintf("%t", cfg.AsyncInsert))
	}

	if !queryParams.Has("compress") && (cfg.Compress == "" || cfg.Compress == "true") {
		queryParams.Set("compress", "lz4")
	} else if !queryParams.Has("compress") {
		queryParams.Set("compress", cfg.Compress)
	}

	productInfo := queryParams.Get("client_info_product")
	collectorProductInfo := fmt.Sprintf("%s/%s", "otelcol", cfg.collectorVersion)
	if productInfo == "" {
		productInfo = collectorProductInfo
	} else {
		productInfo = fmt.Sprintf("%s,%s", productInfo, collectorProductInfo)
	}
	queryParams.Set("client_info_product", productInfo)

	// Use database from config if not specified in path, or if config is not default.
	if dsnURL.Path == "" || cfg.Database != defaultDatabase {
		dsnURL.Path = cfg.Database
	}

	// Override username and password if specified in config.
	if cfg.Username != "" {
		dsnURL.User = url.UserPassword(cfg.Username, string(cfg.Password))
	}

	dsnURL.RawQuery = queryParams.Encode()

	return dsnURL.String(), nil
}

func (cfg *Config) buildDB() (*sql.DB, error) {
	dsn, err := cfg.buildDSN()
	if err != nil {
		return nil, err
	}

	// ClickHouse sql driver will read clickhouse settings from the DSN string.
	conn, err := sql.Open(cfg.driverName, dsn)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// shouldCreateSchema returns true if the exporter should run the DDL for creating database/tables.
func (cfg *Config) shouldCreateSchema() bool {
	return cfg.CreateSchema
}

// tableEngineString generates the ENGINE string.
func (cfg *Config) tableEngineString() string {
	engine := cfg.TableEngine.Name
	params := cfg.TableEngine.Params

	if cfg.TableEngine.Name == "" {
		engine = defaultTableEngineName
		params = ""
	}

	return fmt.Sprintf("%s(%s)", engine, params)
}

// clusterString generates the ON CLUSTER string. Returns empty string if not set.
func (cfg *Config) clusterString() string {
	if cfg.ClusterName == "" {
		return ""
	}

	return fmt.Sprintf("ON CLUSTER %s", cfg.ClusterName)
}
