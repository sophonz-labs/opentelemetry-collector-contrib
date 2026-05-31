# SOPHONZ ClickHouse Metrics Exporter

Exports OpenTelemetry metrics to the SOPHONZ ClickHouse metrics schema. It
converts OTLP metrics to Prometheus remote-write `TimeSeries` and writes them to
the time-series and samples tables (`time_series_v2`/`v4`, `samples_v2`/`v4`,
`exp_hist`).

This is the SOPHONZ-specific metrics exporter ported from the legacy
`clickhousemetricsexporter` and upgraded to the current collector API. It only
uses standard Prometheus label attributes (`__name__`, `tenant`, `env`,
`deployment.environment`), so no SOPHONZ attribute-key renames apply.

## Configuration

| Field | Default | Description |
|---|---|---|
| `endpoint` | (required) | ClickHouse DSN, supplied via the `confighttp` endpoint field |
| `namespace` | | Prefix attached to each exported metric name |
| `external_labels` | | Labels added to every exported series |
| `watcher_interval` | `30s` | Shard-count watcher poll interval |
| `write_ts_to_v4` | `true` | Also write to the v4 time-series/samples tables |
| `remote_write_queue` | | Legacy fan-out queue settings (`enabled`, `queue_size`, `num_consumers`) |
| `sending_queue`, `retry_on_failure`, `timeout` | | Standard exporterhelper settings |

Example:

```yaml
exporters:
  sophonzclickhousemetrics:
    endpoint: tcp://clickhouse:9000/sophonz_metrics
    write_ts_to_v4: true
```
