ALTER TABLE sophonz_logs.logs_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS EventName LowCardinality(String) CODEC(ZSTD(1)) AFTER SeverityNumber;
ALTER TABLE sophonz_logs.distributed_logs_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS EventName LowCardinality(String) CODEC(ZSTD(1)) AFTER SeverityNumber;

ALTER TABLE sophonz_logs.logs_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_event_name EventName TYPE set(0) GRANULARITY 4;
