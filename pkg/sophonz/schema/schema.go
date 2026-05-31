// Package schema is the single source of truth for SOPHONZ ClickHouse
// database, table and custom-column names.
//
// It exists to:
//   - break the build-time coupling where the schema migrator imported
//     name constants from the clickhouse traces exporter, and
//   - keep the post-rename custom column names in one place so the exporter
//     (INSERT statements / struct mapping) and the migration DDL agree.
//
// Custom column names reflect the convention migration documented in
// apps/otel-collector/docs/CONVENTION_MIGRATION.md. Migration .sql files are
// edited directly (golang-migrate is file-based), so the constants here are
// the canonical reference the SQL must match.
package schema

// ClickHouse database names.
const (
	DatabaseTraces   = "sophonz_traces"
	DatabaseMetrics  = "sophonz_metrics"
	DatabaseLogs     = "sophonz_logs"
	DatabaseMetadata = "sophonz_metadata"
)

// Traces tables. Local* are the underlying MergeTree tables; the others are
// the Distributed wrappers used for writes/reads.
const (
	TracesLocalIndexTable    = "sophonz_index_v2"
	TracesIndexTable         = "distributed_sophonz_index_v2"
	TracesErrorIndexTable    = "distributed_sophonz_error_index_v2"
	TracesSpansTable         = "distributed_sophonz_spans"
	TracesAttributesTable    = "distributed_span_attributes"
	TracesAttributeKeysTable = "distributed_span_attributes_keys"
	TracesOperationsTable    = "distributed_sophonz_operations"
)

// Logs tables.
const (
	LogsTable = "logs_v2"
)

// Metrics distributed tables.
const (
	MetricsDistributedTimeSeriesV2 = "distributed_time_series_v2"
	MetricsDistributedTimeSeriesV3 = "distributed_time_series_v3"
	MetricsDistributedTimeSeriesV4 = "distributed_time_series_v4"
	MetricsDistributedSamplesV2    = "distributed_samples_v2"
	MetricsDistributedSamplesV4    = "distributed_samples_v4"
	MetricsDistributedExpHist      = "distributed_exp_hist"
	MetricsTimeSeriesV2            = "time_series_v2"
)

// Custom column names on the traces index / error-index tables AFTER the
// convention migration. The trailing comment is the pre-migration name, kept
// for traceability with the query-service migration checklist.
const (
	ColClientPlatform     = "clientPlatform"     // was: serviceType
	ColComponent          = "component"          // was: title
	ColAppSpanType        = "appSpanType"        // was: spanType
	ColAppScreenID        = "appScreenID"        // was: screenID
	ColAppScreenName      = "appScreenName"      // was: screenName
	ColAppScreenType      = "appScreenType"      // was: screenType
	ColAppScreenGroupID   = "appScreenGroupID"   // was: screenGroupID
	ColAppScreenGroupName = "appScreenGroupName" // was: screenGroupName
)

// Custom column names that are unchanged by the convention migration.
const (
	ColSessionID  = "sessionID"
	ColDeviceID   = "deviceID"
	ColUserID     = "userID"
	ColWebVersion = "webVersion"
)
