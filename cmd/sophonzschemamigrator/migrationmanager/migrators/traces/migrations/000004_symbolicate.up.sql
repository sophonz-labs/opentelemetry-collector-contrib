CREATE TABLE IF NOT EXISTS sophonz_traces.sophonz_stacktrace ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    timestamp DateTime64(9) CODEC(Delta(8), ZSTD(1)),
    serviceNamespace LowCardinality(String) CODEC(ZSTD(1)),
    clientPlatform LowCardinality(String) CODEC(ZSTD(1)),
    serviceVersion LowCardinality(String) CODEC(ZSTD(1)),
    serviceName LowCardinality(String) CODEC(ZSTD(1)),
    traceID String CODEC(ZSTD(1)),
    spanID String CODEC(ZSTD(1)),
    obfuscationType LowCardinality(String) CODEC(ZSTD(1)),
    stacktrace String CODEC(ZSTD(1)),
    symbolicatedStacktrace String CODEC(ZSTD(1)),
    symbolicationError String CODEC(ZSTD(1)),
    INDEX idx_serviceNamespace serviceNamespace TYPE set(0) GRANULARITY 4,
    INDEX idx_clientPlatform clientPlatform TYPE set(0) GRANULARITY 4,
    INDEX idx_serviceVersion serviceVersion TYPE bloom_filter GRANULARITY 4,
    INDEX idx_serviceName serviceName TYPE bloom_filter GRANULARITY 4,
    INDEX idx_obfuscationType obfuscationType TYPE bloom_filter GRANULARITY 4,
    INDEX idx_trace_id traceID TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_span_id spanID TYPE bloom_filter(0.001) GRANULARITY 1,
)   ENGINE = {{.SOPHONZ_REPLICATED}}MergeTree PARTITION BY toDate(timestamp)
    ORDER BY (serviceNamespace, clientPlatform, serviceName, serviceVersion, spanID)
    TTL toDateTime(timestamp) + toIntervalSecond(2592000)
    SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1;

ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_traceID traceID TYPE bloom_filter(0.001) GRANULARITY 1;
ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_spanID spanID TYPE bloom_filter(0.001) GRANULARITY 1;

ALTER TABLE sophonz_traces.sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_traceID traceID TYPE bloom_filter(0.001) GRANULARITY 1;
ALTER TABLE sophonz_traces.sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_spanID spanID TYPE bloom_filter(0.001) GRANULARITY 1;
