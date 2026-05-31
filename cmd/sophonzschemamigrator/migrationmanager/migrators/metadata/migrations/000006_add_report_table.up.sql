CREATE TABLE IF NOT EXISTS sophonz_metadata.report_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    createdAt DateTime64 CODEC(DoubleDelta, ZSTD(1)),
    creationType LowCardinality(String) CODEC(ZSTD(1)),
    requestedUserID LowCardinality(String) CODEC(ZSTD(1)),
    code LowCardinality(String) CODEC(ZSTD(1)),
    serviceNamespace LowCardinality(String) CODEC(ZSTD(1)),
    reportType LowCardinality(String) CODEC(ZSTD(1)),
    startDate DateTime CODEC(DoubleDelta, ZSTD(1)),
    endDate DateTime CODEC(DoubleDelta, ZSTD(1)),
    payload String CODEC(ZSTD(1)),
    supplementaryPayload String CODEC(ZSTD(1))
)   ENGINE = {{.SOPHONZ_REPLICATED}}MergeTree PARTITION BY toDate(createdAt)
    ORDER BY (serviceNamespace, createdAt)
    SETTINGS ttl_only_drop_parts = 1;

CREATE TABLE IF NOT EXISTS sophonz_metadata.batch_log ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    timestamp DateTime64 CODEC(DoubleDelta, ZSTD(1)),
    startTime DateTime64 CODEC(DoubleDelta, ZSTD(1)),
    endTime DateTime64 CODEC(DoubleDelta, ZSTD(1)),
    jobName LowCardinality(String) CODEC(ZSTD(1)),
    batchTime String CODEC(ZSTD(1)),
    status LowCardinality(String) CODEC(ZSTD(1)),
    error String CODEC(ZSTD(1))
)   ENGINE = {{.SOPHONZ_REPLICATED}}MergeTree PARTITION BY toDate(timestamp)
    ORDER BY (timestamp)
    SETTINGS ttl_only_drop_parts = 1;
