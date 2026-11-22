-- TenantID carries the organization the telemetry belongs to, resolved from the
-- SDK app key by the sophonz attribute processor.
--
-- Appended at the very END of the table (no AFTER clause) so it stays the last
-- entry of the column list the logs exporter names in insertLogsSQLTemplate —
-- that keeps the template's column order identical to the table's and lets the
-- new column be a pure append to both the SQL and its ExecContext arguments.

ALTER TABLE sophonz_logs.logs_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS TenantID LowCardinality(String) CODEC(ZSTD(1));
ALTER TABLE sophonz_logs.distributed_logs_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS TenantID LowCardinality(String) CODEC(ZSTD(1));

ALTER TABLE sophonz_logs.logs_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_tenant_id TenantID TYPE set(0) GRANULARITY 4;
