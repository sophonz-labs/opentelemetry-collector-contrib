# SOPHONZ ClickHouse Logs Exporter

Exports OpenTelemetry logs to a SOPHONZ ClickHouse `logs_v2` table.

This is the SOPHONZ-specific logs exporter ported from the legacy
`clickhouselogsexporterv2` and upgraded to the current collector API. It
writes to the SOPHONZ ClickHouse schema (database `sophonz_logs`) and applies
the SOPHONZ semantic conventions (see
`apps/otel-collector/docs/CONVENTION_MIGRATION.md`): the resource attribute
`client.platform` is written to the `ClientPlatform` column.

## Configuration

| Field | Default | Description |
|---|---|---|
| `host` | (required) | ClickHouse host |
| `port` | (required) | ClickHouse port |
| `username` | | Auth username |
| `password` | | Auth password |
| `database` | `sophonz_logs` | Target database |
| `logs_table_name` | `logs_v2` | Target table |
| `ttl` | `0` | Data TTL (e.g. `48h`); `0` disables |
| `cluster_name` | | Appends `ON CLUSTER` when creating tables |
| `create_schema` | `false` | Run DDL to create database/table |
| `compress` | `lz4` | Compression algorithm |
| `async_insert` | `true` | Enable async inserts |
| `connection_params` | | Extra DSN params |
| `sending_queue`, `retry_on_failure`, `timeout` | | Standard exporterhelper settings |

Example:

```yaml
exporters:
  sophonzclickhouselogs:
    host: clickhouse
    port: "9000"
    database: sophonz_logs
    logs_table_name: logs_v2
```
