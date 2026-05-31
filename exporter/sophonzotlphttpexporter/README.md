# SOPHONZ OTLP HTTP Exporter

Exports telemetry over OTLP/HTTP. This is a SOPHONZ-specific fork of the core
`otlphttpexporter`, rebased onto the current collector API and carrying the
SOPHONZ behavioral changes:

- `enabled` config flag (default `false`): when disabled, the exporter drops
  data without sending it.
- Traces: only spans whose status code is `Error` are exported.
- Logs: only log records that carry an exception stacktrace attribute
  (`exception.stacktrace`) are exported.
- Metrics export is disabled (the metrics push path is preserved in code but
  the signal is not registered with the factory).
- HTTP error responses are mapped to gRPC status codes via the vendored
  `internal/httphelper` package.
- The exporterhelper timeout is set to `0` so timeouts are handled by the
  underlying `http.Client`.

## Configuration

| Field | Default | Description |
|---|---|---|
| `enabled` | `false` | Enables sending. When `false`, traces/logs are dropped |
| `endpoint` | | Base endpoint. Signal URLs are derived as `endpoint/v1/<signal>` |
| `traces_endpoint` | | Full URL for traces (overrides `endpoint`) |
| `logs_endpoint` | | Full URL for logs (overrides `endpoint`) |
| `metrics_endpoint` | | Full URL for metrics (config field preserved; signal disabled) |
| `encoding` | `proto` | `proto` or `json` |
| `compression` | `gzip` | Compression algorithm |
| `sending_queue`, `retry_on_failure` | | Standard exporterhelper settings |

Example:

```yaml
exporters:
  sophonzotlphttp:
    enabled: true
    endpoint: https://collector.example.com
    encoding: proto
```
