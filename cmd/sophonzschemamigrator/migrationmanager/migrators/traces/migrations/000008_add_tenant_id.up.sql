-- tenantID carries the organization the telemetry belongs to, resolved from the
-- SDK app key by the sophonz attribute processor.
--
-- The column is appended at the very END of both index tables, with no AFTER
-- clause. The traces exporter writes them with a positional
-- statement.Append(...) against an "INSERT INTO <table>" that has no column
-- list, so the argument order must equal the physical column order. Putting the
-- new column anywhere else would silently shift every value after it by one
-- position instead of failing.

ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS tenantID LowCardinality(String) CODEC(ZSTD(1));
ALTER TABLE sophonz_traces.distributed_sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS tenantID LowCardinality(String) CODEC(ZSTD(1));

ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_tenantID tenantID TYPE set(0) GRANULARITY 4;

ALTER TABLE sophonz_traces.sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS tenantID LowCardinality(String) CODEC(ZSTD(1));
ALTER TABLE sophonz_traces.distributed_sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS tenantID LowCardinality(String) CODEC(ZSTD(1));

ALTER TABLE sophonz_traces.sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_tenantID tenantID TYPE set(0) GRANULARITY 4;
