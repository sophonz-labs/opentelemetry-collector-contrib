// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhouselogsexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhouselogsexporter"

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2" // For register database driver.
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/idutil"
	sophonzsemconv "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/semconv"
)

type logsExporter struct {
	client    *sql.DB
	insertSQL string
	logger    *zap.Logger
	cfg       *Config
}

func newLogsExporter(logger *zap.Logger, cfg *Config) (*logsExporter, error) {
	client, err := newClickhouseClient(cfg)
	if err != nil {
		return nil, err
	}

	return &logsExporter{
		client:    client,
		insertSQL: renderInsertLogsSQL(cfg),
		logger:    logger,
		cfg:       cfg,
	}, nil
}

func (e *logsExporter) start(ctx context.Context, _ component.Host) error {
	if !e.cfg.shouldCreateSchema() {
		return nil
	}

	if err := createDatabase(ctx, e.cfg); err != nil {
		return err
	}

	return createLogsTable(ctx, e.cfg, e.client)
}

// shutdown will shut down the exporter.
func (e *logsExporter) shutdown(_ context.Context) error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}

func (e *logsExporter) pushLogsData(ctx context.Context, ld plog.Logs) error {
	start := time.Now()
	err := doWithTx(ctx, e.client, func(tx *sql.Tx) error {
		statement, err := tx.PrepareContext(ctx, e.insertSQL)
		if err != nil {
			return fmt.Errorf("PrepareContext:%w", err)
		}
		defer func() {
			_ = statement.Close()
		}()

		for i := 0; i < ld.ResourceLogs().Len(); i++ {
			var serviceNamespace, serviceName, serviceVersion, clientPlatform, webVersion, tenantID string
			logs := ld.ResourceLogs().At(i)
			res := logs.Resource()
			resURL := logs.SchemaUrl()
			resAttr := attributesToMap(res.Attributes())
			if v, ok := res.Attributes().Get(string(semconv.ServiceNamespaceKey)); ok {
				serviceNamespace = v.Str()
			}
			if v, ok := res.Attributes().Get(string(semconv.ServiceNameKey)); ok {
				serviceName = v.Str()
			}
			if v, ok := res.Attributes().Get(string(semconv.ServiceVersionKey)); ok {
				serviceVersion = v.Str()
			}
			if v, ok := res.Attributes().Get(sophonzsemconv.ClientPlatform); ok {
				clientPlatform = v.Str()
			}
			if v, ok := res.Attributes().Get(sophonzsemconv.WebVersion); ok {
				webVersion = v.Str()
			}
			// Set by the sophonz attribute processor from the SDK app key; empty
			// while its service_key_mode is off.
			if v, ok := res.Attributes().Get(sophonzsemconv.TenantID); ok {
				tenantID = v.Str()
			}
			for j := 0; j < logs.ScopeLogs().Len(); j++ {
				rs := logs.ScopeLogs().At(j).LogRecords()
				scopeURL := logs.ScopeLogs().At(j).SchemaUrl()
				scopeName := logs.ScopeLogs().At(j).Scope().Name()
				scopeVersion := logs.ScopeLogs().At(j).Scope().Version()
				scopeAttr := attributesToMap(logs.ScopeLogs().At(j).Scope().Attributes())
				for k := 0; k < rs.Len(); k++ {
					r := rs.At(k)
					logID, _ := r.Attributes().Get(sophonzsemconv.LogID)
					logAttr := attributesToMap(r.Attributes())
					_, err = statement.ExecContext(ctx,
						r.Timestamp().AsTime(),
						idutil.TraceIDToHexOrEmptyString(r.TraceID()),
						idutil.SpanIDToHexOrEmptyString(r.SpanID()),
						logID.Str(),
						uint32(r.Flags()),
						r.SeverityText(),
						int32(r.SeverityNumber()),
						r.EventName(),
						serviceNamespace,
						serviceName,
						serviceVersion,
						clientPlatform,
						webVersion,
						r.Body().AsString(),
						resURL,
						resAttr,
						scopeURL,
						scopeName,
						scopeVersion,
						scopeAttr,
						logAttr,
						tenantID,
					)
					if err != nil {
						return fmt.Errorf("ExecContext:%w", err)
					}
				}
			}
		}
		return nil
	})
	duration := time.Since(start)
	e.logger.Debug("insert logs", zap.Int("records", ld.LogRecordCount()),
		zap.String("cost", duration.String()))
	return err
}

func attributesToMap(attributes pcommon.Map) map[string]string {
	m := make(map[string]string, attributes.Len())
	for k, v := range attributes.All() {
		m[k] = v.AsString()
	}
	return m
}

const (
	// createLogsTableSQL is only executed when create_schema is enabled (it
	// defaults to false; cmd/sophonzschemamigrator is the authoritative owner of
	// the sophonz_logs schema). It is kept column-for-column in sync with
	// migrationmanager/migrators/logs/migrations -- same columns, types, codecs,
	// skip indexes and ORDER BY -- so that a table created here accepts the exact
	// INSERT issued by insertLogsSQLTemplate. Only the table name, ON CLUSTER
	// clause, engine and TTL stay parameterised from Config.
	// language=ClickHouse SQL
	createLogsTableSQL = `
CREATE TABLE IF NOT EXISTS %s %s (
     Timestamp DateTime64(9) CODEC(Delta, ZSTD(1)),
     TraceId FixedString(32) CODEC(ZSTD(1)),
     SpanId FixedString(16) CODEC(ZSTD(1)),
     Id String CODEC(ZSTD(1)),
     TraceFlags UInt32 CODEC(ZSTD(1)),
     SeverityText LowCardinality(String) CODEC(ZSTD(1)),
     SeverityNumber Int32 CODEC(ZSTD(1)),
     EventName LowCardinality(String) CODEC(ZSTD(1)),
     ServiceNamespace LowCardinality(String) CODEC(ZSTD(1)),
     ServiceName LowCardinality(String) CODEC(ZSTD(1)),
     ServiceVersion LowCardinality(String) CODEC(ZSTD(1)),
     ClientPlatform LowCardinality(String) CODEC(ZSTD(1)),
     WebVersion LowCardinality(String) CODEC(ZSTD(1)),
     Body String CODEC(ZSTD(1)),
     ResourceSchemaUrl String CODEC(ZSTD(1)),
     ResourceAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
     ScopeSchemaUrl String CODEC(ZSTD(1)),
     ScopeName String CODEC(ZSTD(1)),
     ScopeVersion String CODEC(ZSTD(1)),
     ScopeAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
     LogAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
     TenantID LowCardinality(String) CODEC(ZSTD(1)),
     INDEX idx_service_namespace ServiceNamespace TYPE set(0) GRANULARITY 4,
     INDEX idx_service_name ServiceName TYPE set(0) GRANULARITY 4,
     INDEX idx_service_version ServiceVersion TYPE set(0) GRANULARITY 4,
     INDEX idx_client_platform ClientPlatform TYPE set(0) GRANULARITY 4,
     INDEX idx_web_version WebVersion TYPE set(0) GRANULARITY 4,
     INDEX idx_event_name EventName TYPE set(0) GRANULARITY 4,
     INDEX idx_tenant_id TenantID TYPE set(0) GRANULARITY 4,
     INDEX idx_severity_text SeverityText TYPE bloom_filter GRANULARITY 4,
     INDEX idx_severity_number SeverityNumber TYPE bloom_filter GRANULARITY 4,
     INDEX idx_trace_id TraceId TYPE bloom_filter(0.001) GRANULARITY 4,
     INDEX idx_id Id TYPE bloom_filter GRANULARITY 4,
     INDEX idx_res_attr_key mapKeys(ResourceAttributes) TYPE bloom_filter(0.01) GRANULARITY 4,
     INDEX idx_res_attr_value mapValues(ResourceAttributes) TYPE bloom_filter(0.01) GRANULARITY 4,
     INDEX idx_scope_attr_key mapKeys(ScopeAttributes) TYPE bloom_filter(0.01) GRANULARITY 4,
     INDEX idx_scope_attr_value mapValues(ScopeAttributes) TYPE bloom_filter(0.01) GRANULARITY 4,
     INDEX idx_log_attr_key mapKeys(LogAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_log_attr_value mapValues(LogAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_body Body TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 1
) ENGINE = %s
%s
PARTITION BY toDate(Timestamp)
ORDER BY (ServiceNamespace, ClientPlatform, ServiceVersion, SeverityText, WebVersion, toUnixTimestamp(Timestamp), Id)
SETTINGS index_granularity=8192, ttl_only_drop_parts = 1;
`
	// insertLogsSQLTemplate names 22 columns, matching the sophonz_logs.logs_v2
	// schema owned by cmd/sophonzschemamigrator, and the 22 arguments passed to
	// ExecContext in pushLogsData (in this same order).
	//
	// The VALUES clause below is dead text and its placeholder count is
	// deliberately not maintained. clickhouse-go v2 does not use it: the std
	// driver's PrepareContext calls connect.prepareBatch, which runs the query
	// through extractNormalizedInsertQueryAndColumns (batch.go). That rewrites
	// the statement to "INSERT INTO <table> (<columns>) FORMAT Native" -- the
	// VALUES clause is outside the capture group and is dropped -- then binds
	// positionally against the parsed column list via Block.SortColumns and
	// Block.Append. stdBatch.NumInput() returns -1, so database/sql performs no
	// placeholder/argument count check either. The invariant that matters is
	// len(columns) == len(ExecContext args) == 22, in the same order.
	// language=ClickHouse SQL
	insertLogsSQLTemplate = `INSERT INTO %s (
                        Timestamp,
                        TraceId,
                        SpanId,
                		Id,
                        TraceFlags,
                        SeverityText,
                        SeverityNumber,
                        EventName,
                        ServiceNamespace,
                        ServiceName,
                        ServiceVersion,
                		ClientPlatform,
                        WebVersion,
                        Body,
                        ResourceSchemaUrl,
                        ResourceAttributes,
                        ScopeSchemaUrl,
                        ScopeName,
                        ScopeVersion,
                        ScopeAttributes,
                        LogAttributes,
                        TenantID
                        ) VALUES (
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?,
                                  ?
                                  )`
)

// newClickhouseClient create a clickhouse client.
func newClickhouseClient(cfg *Config) (*sql.DB, error) {
	db, err := cfg.buildDB()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func createDatabase(ctx context.Context, cfg *Config) error {
	// use default database to create new database
	if cfg.Database == defaultDatabase {
		return nil
	}

	// We couldn't set a database in the dsn while creating the database,
	// otherwise, there would be an exception from clickhouse
	targetDatabase := cfg.Database
	cfg.Database = defaultDatabase

	db, err := cfg.buildDB()
	cfg.Database = targetDatabase
	if err != nil {
		return fmt.Errorf("can't connect to clickhouse: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s %s", cfg.Database, cfg.clusterString())
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}
	return nil
}

func createLogsTable(ctx context.Context, cfg *Config, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, renderCreateLogsTableSQL(cfg)); err != nil {
		return fmt.Errorf("exec create logs table sql: %w", err)
	}
	return nil
}

func renderCreateLogsTableSQL(cfg *Config) string {
	ttlExpr := generateTTLExpr(cfg.TTL, "Timestamp")
	return fmt.Sprintf(createLogsTableSQL, cfg.LogsTableName, cfg.clusterString(), cfg.tableEngineString(), ttlExpr)
}

func renderInsertLogsSQL(cfg *Config) string {
	return fmt.Sprintf(insertLogsSQLTemplate, cfg.LogsTableName)
}

func doWithTx(_ context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
