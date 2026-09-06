# SOPHONZ Attribute Processor

Enriches SOPHONZ traces and logs with screen, service, and client metadata, and
normalizes SOPHONZ-specific attributes.

This is the SOPHONZ attribute processor ported from the legacy
`apps/otel-collector/processor/sophonzattributeprocessor` and upgraded to the
current collector API. It applies the SOPHONZ semantic conventions (see
`apps/otel-collector/docs/CONVENTION_MIGRATION.md`) when reading and writing
attribute keys:

| Legacy key             | Current key             |
| ---------------------- | ----------------------- |
| `screen.id`            | `app.screen.id`         |
| `screen.name`          | `app.screen.name`       |
| `screen.type`          | `app.screen.type`       |
| `screen.group.id`      | `app.screen.group.id`   |
| `screen.group.name`    | `app.screen.group.name` |
| `sophonz.service.type` | `client.platform`       |
| `span.type`            | `app.span.type`         |

Unchanged keys: `service.key`, `web.version`, `log.id`, `network.forwarded_for`.

## What it does

- Resolves `app.screen.id` / `app.screen.group.id` / `app.screen.group.name`
  from the screen metadata cache, keyed by service namespace, `app.screen.name`,
  and `app.screen.type`.
- Uppercases `app.span.type` and normalizes `client.platform`.
- Copies client address and `network.forwarded_for` from Envoy-injected request
  metadata onto the telemetry.
- Assigns a KSUID `log.id` to each log record.
- Optionally decrypts the encrypted `user.id`.
- Resolves the tenant from the SDK app key (`service.key`) against the service
  metadata cache and stamps it on the resource as `sophonz.tenant.id`. The key
  is an opaque token looked up verbatim; the resolved value is the owning
  organization, so it can be trusted for tenant isolation. Anything the client
  sends under `sophonz.tenant.id` is overwritten.

The screen and service metadata are loaded and periodically refreshed from a
ClickHouse `sophonz_metadata` database via the shared
`pkg/sophonz/metadata` manager.

## Configuration

| Field                     | Default    | Description                                                            |
| ------------------------- | ---------- | ---------------------------------------------------------------------- |
| `enabled`                 | `false`    | Enable the processor; when disabled telemetry passes through untouched |
| `fetch_interval`          | (required) | Metadata refresh interval, must be at least `60s`                      |
| `clickhouse.host`         | (required) | ClickHouse host                                                        |
| `clickhouse.port`         | (required) | ClickHouse port                                                        |
| `clickhouse.database`     | `default`  | Metadata database                                                      |
| `clickhouse.username`     | (required) | Auth username                                                          |
| `clickhouse.password`     |            | Auth password                                                          |
| `service_key_mode`        | `off`      | Tenant resolution mode: `off`, `shadow` or `enforce` (see below)       |
| `service_key_check`       | `false`    | Deprecated alias: `true` means `service_key_mode: enforce`             |
| `encrypted_user_id` | `false`    | Decrypt encrypted `user.id` (requires `AES_DECRYPT_KEY`)   |

### Tenant resolution modes

| Mode      | Lookup | Stamps `sophonz.tenant.id` | Drops resources                |
| --------- | ------ | -------------------------- | ------------------------------ |
| `off`     | no     | no                         | no                             |
| `shadow`  | yes    | yes, on a hit              | never                          |
| `enforce` | yes    | yes, on a hit              | when the key is missing/unknown |

Roll out `off` -> `shadow` -> `enforce`, watching the counters below in shadow
until `missing_key` and `unknown_key` are at zero for the traffic you care
about. `service_key_mode` takes precedence over the deprecated
`service_key_check`; if only the boolean is set, `true` maps to `enforce`.

Resolution fails open: if the metadata manager is unavailable or its service
cache is empty, nothing is dropped in any mode, and the processor logs the
degradation at most once a minute. A metadata outage must not become data loss.

`shadow` and `enforce` emit these counters (and, at most once a minute, an
aggregated `tenant resolution summary` log line):

| Counter                                | Meaning                                       |
| -------------------------------------- | --------------------------------------------- |
| `sophonzattribute.tenant.resolved`     | Key resolved, tenant stamped                  |
| `sophonzattribute.tenant.missing_key`  | Resource carried no `service.key`             |
| `sophonzattribute.tenant.unknown_key`  | Key is not in the metadata cache              |
| `sophonzattribute.tenant.dropped`      | Resource dropped (`enforce` only)             |
| `sophonzattribute.tenant.degraded`     | Passed through because the cache was unusable |

Example:

```yaml
processors:
  sophonzattribute:
    enabled: true
    fetch_interval: 60s
    service_key_mode: shadow
    clickhouse:
      host: clickhouse
      port: 9000
      username: default
```
