// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhousetracesexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousetracesexporter"

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/schema"
)

const (
	defaultDatasource string = "tcp://127.0.0.1:9000/?database=sophonz_traces"
	// DefaultTraceDatabase is the trace database name. Defined via schema so the
	// migrator and exporter stay in sync.
	DefaultTraceDatabase   string = schema.DatabaseTraces
	defaultMigrations      string = "/migrations"
	defaultOperationsTable string = schema.TracesOperationsTable
	// DefaultIndexTable is the distributed index table.
	DefaultIndexTable string = schema.TracesIndexTable
	// LocalIndexTable is the underlying (non-distributed) index table.
	LocalIndexTable                 string   = schema.TracesLocalIndexTable
	defaultErrorTable               string   = schema.TracesErrorIndexTable
	defaultSpansTable               string   = schema.TracesSpansTable
	defaultAttributeTable           string   = schema.TracesAttributesTable
	defaultAttributeKeyTable        string   = schema.TracesAttributeKeysTable
	DefaultDurationSortTable        string   = "durationSort"
	DefaultDurationSortMVTable      string   = "durationSortMV"
	defaultArchiveSpansTable        string   = "sophonz_archive_spans"
	defaultClusterName              string   = "cluster"
	defaultDependencyGraphTable     string   = "dependency_graph_minutes"
	defaultDependencyGraphServiceMV string   = "dependency_graph_minutes_service_calls_mv"
	defaultDependencyGraphDbMV      string   = "dependency_graph_minutes_db_calls_mv"
	DependencyGraphMessagingMV      string   = "dependency_graph_minutes_messaging_calls_mv"
	defaultEncoding                 Encoding = EncodingJSON
)

// namespaceConfig is ClickHouse's internal configuration data.
type namespaceConfig struct {
	namespace                  string
	Enabled                    bool
	Datasource                 string
	Migrations                 string
	TraceDatabase              string
	OperationsTable            string
	IndexTable                 string
	LocalIndexTable            string
	SpansTable                 string
	ErrorTable                 string
	AttributeTable             string
	AttributeKeyTable          string
	Cluster                    string
	DurationSortTable          string
	DurationSortMVTable        string
	DependencyGraphServiceMV   string
	DependencyGraphDbMV        string
	DependencyGraphMessagingMV string
	DependencyGraphTable       string
	DockerMultiNodeCluster     bool
	NumConsumers               int
	Encoding                   Encoding
	Connector                  Connector
	ExporterId                 uuid.UUID
}

// Connector defines how to connect to the database.
type Connector func(cfg *namespaceConfig) (clickhouse.Conn, error)

func defaultConnector(cfg *namespaceConfig) (clickhouse.Conn, error) {
	ctx := context.Background()
	options, err := clickhouse.ParseDSN(cfg.Datasource)
	if err != nil {
		return nil, err
	}

	// setting maxOpenIdleConnections = numConsumers + 1 to avoid `prepareBatch:clickhouse: acquire conn timeout`
	// error when using multiple consumers along with usage exporter
	maxIdleConnections := cfg.NumConsumers + 1

	if options.MaxIdleConns < maxIdleConnections {
		options.MaxIdleConns = maxIdleConnections
		options.MaxOpenConns = maxIdleConnections + 5
	}
	db, err := clickhouse.Open(options)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s ON CLUSTER %s`, options.Auth.Database, cfg.Cluster)
	if err := db.Exec(ctx, query); err != nil {
		return nil, err
	}
	return db, nil
}

// Options stores storage plugin related configs.
type Options struct {
	primary *namespaceConfig

	others map[string]*namespaceConfig
}

// NewOptions creates a new Options struct.
func NewOptions(exporterId uuid.UUID, migrations string, datasource string, dockerMultiNodeCluster bool, numConsumers int, primaryNamespace string, otherNamespaces ...string) *Options {
	if datasource == "" {
		datasource = defaultDatasource
	}
	if migrations == "" {
		migrations = defaultMigrations
	}

	options := &Options{
		primary: &namespaceConfig{
			namespace:                  primaryNamespace,
			Enabled:                    true,
			Datasource:                 datasource,
			Migrations:                 migrations,
			TraceDatabase:              DefaultTraceDatabase,
			OperationsTable:            defaultOperationsTable,
			IndexTable:                 DefaultIndexTable,
			LocalIndexTable:            LocalIndexTable,
			ErrorTable:                 defaultErrorTable,
			SpansTable:                 defaultSpansTable,
			AttributeTable:             defaultAttributeTable,
			AttributeKeyTable:          defaultAttributeKeyTable,
			DurationSortTable:          DefaultDurationSortTable,
			DurationSortMVTable:        DefaultDurationSortMVTable,
			Cluster:                    defaultClusterName,
			DependencyGraphTable:       defaultDependencyGraphTable,
			DependencyGraphServiceMV:   defaultDependencyGraphServiceMV,
			DependencyGraphDbMV:        defaultDependencyGraphDbMV,
			DependencyGraphMessagingMV: DependencyGraphMessagingMV,
			DockerMultiNodeCluster:     dockerMultiNodeCluster,
			NumConsumers:               numConsumers,
			Encoding:                   defaultEncoding,
			Connector:                  defaultConnector,
			ExporterId:                 exporterId,
		},
		others: make(map[string]*namespaceConfig, len(otherNamespaces)),
	}

	for _, namespace := range otherNamespaces {
		if namespace == archiveNamespace {
			options.others[namespace] = &namespaceConfig{
				namespace:       namespace,
				Datasource:      datasource,
				Migrations:      migrations,
				OperationsTable: "",
				IndexTable:      "",
				SpansTable:      defaultArchiveSpansTable,
				Encoding:        defaultEncoding,
				Connector:       defaultConnector,
				ExporterId:      exporterId,
			}
		} else {
			options.others[namespace] = &namespaceConfig{namespace: namespace}
		}
	}

	return options
}

// getPrimary returns the primary namespace configuration.
func (opt *Options) getPrimary() *namespaceConfig {
	return opt.primary
}
