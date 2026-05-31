# SOPHONZ Attribute Processor

Enriches SOPHONZ traces and logs with screen, service, and client metadata, and
normalizes SOPHONZ-specific attributes.

This is the SOPHONZ attribute processor ported from the legacy
`apps/otel-collector/processor/sophonzattributeprocessor` and upgraded to the
current collector API. It applies the SOPHONZ semantic conventions (see
`apps/otel-collector/docs/CONVENTION_MIGRATION.md`) when reading and writing
attribute keys:

| Legacy key | Current key |
|---|---|
| `screen.id` | `app.screen.id` |
| `screen.name` | `app.screen.name` |
| `screen.type` | `app.screen.type` |
| `screen.group.id` | `app.screen.group.id` |
| `screen.group.name` | `app.screen.group.name` |
| `sophonz.service.type` | `client.platform` |
| `span.type` | `app.span.type` |

Unchanged keys: `service.key`, `web.version`, `log.id`, `network.forwarded_for`.

## What it does

- Resolves `app.screen.id` / `app.screen.group.id` / `app.screen.group.name`
  from the screen metadata cache, keyed by service namespace, `app.screen.name`,
  and `app.screen.type`.
- Uppercases `app.span.type` and normalizes `client.platform`.
- Copies client address and `network.forwarded_for` from Envoy-injected request
  metadata onto the telemetry.
- Assigns a KSUID `log.id` to each log record.
- Optionally decrypts the KINFA-encrypted `user.id`.
- Optionally validates the `service.key` and drops unrecognized resources.

The screen and service metadata are loaded and periodically refreshed from a
ClickHouse `sophonz_metadata` database via the shared
`pkg/sophonz/metadata` manager.

## Configuration

| Field | Default | Description |
|---|---|---|
| `enabled` | `false` | Enable the processor; when disabled telemetry passes through untouched |
| `fetch_interval` | (required) | Metadata refresh interval, must be at least `60s` |
| `clickhouse.host` | (required) | ClickHouse host |
| `clickhouse.port` | (required) | ClickHouse port |
| `clickhouse.database` | `default` | Metadata database |
| `clickhouse.username` | (required) | Auth username |
| `clickhouse.password` | | Auth password |
| `service_key_check` | `false` | Validate `service.key` and drop unknown resources |
| `encrypted_user_id_kinfa` | `false` | Decrypt KINFA-encrypted `user.id` (requires `AES_DECRYPT_KEY_KINFA`) |

Example:

```yaml
processors:
  sophonzattribute:
    enabled: true
    fetch_interval: 60s
    service_key_check: true
    clickhouse:
      host: clickhouse
      port: 9000
      username: default
```
