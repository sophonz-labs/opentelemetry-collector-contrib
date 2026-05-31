# SOPHONZ Console Print Processor

| Status                   |                       |
| ------------------------ | --------------------- |
| Stability                | traces: stable        |
|                          | logs: beta            |
| Distributions            | contrib               |
| Warnings                 | None                  |
| Issues                   | Open issues Closed issues |

The SOPHONZ Console Print Processor pretty-prints telemetry data (traces and logs) to stdout as indented JSON for debugging purposes.

## Configuration

```yaml
processors:
  sophonzconsoleprint:
    trace_enabled: true
    log_enabled: true
```

| Field           | Type   | Default | Description                                  |
| --------------- | ------ | ------- | -------------------------------------------- |
| `trace_enabled` | bool   | `false` | Enable pretty-printing of trace data to stdout |
| `log_enabled`   | bool   | `false` | Enable pretty-printing of log data to stdout  |
