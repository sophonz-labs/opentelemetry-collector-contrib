# SOPHONZ ClickHouse Traces Exporter

Exports OpenTelemetry traces to the SOPHONZ ClickHouse trace schema (database
`sophonz_traces`): the index, spans, error-index and span-attribute tables.

This is the SOPHONZ-specific traces exporter ported from the legacy
`clickhousetracesexporter` and upgraded to the current collector API. It applies
the SOPHONZ semantic conventions (see
`apps/otel-collector/docs/CONVENTION_MIGRATION.md`):

| Attribute (read) | ClickHouse column |
|---|---|
| `client.platform` (was `sophonz.service.type`) | `clientPlatform` (was `serviceType`) |
| `app.span.type` (was `span.type`) | `appSpanType` (was `spanType`) |
| `app.screen.id` (was `screen.id`) | `appScreenID` (was `screenID`) |
| `app.screen.name` (was `screen.name`) | `appScreenName` (was `screenName`) |
| `app.screen.type` (was `screen.type`) | `appScreenType` (was `screenType`) |
| `app.screen.group.id` (was `screen.group.id`) | `appScreenGroupID` (was `screenGroupID`) |
| `app.screen.group.name` (was `screen.group.name`) | `appScreenGroupName` (was `screenGroupName`) |
| (span title) | `component` (was `title`) |

## Configuration

| Field | Default | Description |
|---|---|---|
| `host` | (required) | ClickHouse host |
| `port` | (required) | ClickHouse port |
| `username` | `sophonz` | Auth username |
| `password` | | Auth password |
| `database` | `sophonz_traces` | Target database |
| `migrations` | | Migrations path |
| `docker_multi_node_cluster` | `false` | Enable docker multi-node cluster |
| `low_cardinal_exception_grouping` | `false` | Group exceptions by serviceName + exceptionType |
| `sending_queue`, `retry_on_failure`, `timeout` | | Standard exporterhelper settings |

Example:

```yaml
exporters:
  sophonzclickhousetraces:
    host: clickhouse
    port: "9000"
    database: sophonz_traces
```
