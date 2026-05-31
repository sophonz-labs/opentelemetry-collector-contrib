CREATE TABLE IF NOT EXISTS sophonz_logs.logs_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    Timestamp DateTime64(9) CODEC(Delta, ZSTD(1)),
    TraceId FixedString(32) CODEC(ZSTD(1)),
    SpanId FixedString(16) CODEC(ZSTD(1)),
    Id String CODEC(ZSTD(1)),
    TraceFlags UInt32 CODEC(ZSTD(1)),
    SeverityText LowCardinality(String) CODEC(ZSTD(1)),
    SeverityNumber Int32 CODEC(ZSTD(1)),
    ServiceNamespace LowCardinality(String) CODEC(ZSTD(1)),
    ServiceName LowCardinality(String) CODEC(ZSTD(1)),
    ServiceVersion LowCardinality(String) CODEC(ZSTD(1)),
    ClientPlatform LowCardinality(String) CODEC(ZSTD(1)),
    WebVersion LowCardinality(String) CODEC(ZSTD(1)),
    Body String CODEC(ZSTD(1)),
    ResourceSchemaUrl String CODEC(ZSTD(1)),
    ResourceAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    ScopeSchemaUrl String CODEC(ZSTD(1)),
    ScopeName String CODEC(ZSTD(1)),
    ScopeVersion String CODEC(ZSTD(1)),
    ScopeAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    LogAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    INDEX idx_service_namespace ServiceNamespace TYPE set(0) GRANULARITY 4,
    INDEX idx_service_name ServiceName TYPE set(0) GRANULARITY 4,
    INDEX idx_service_version ServiceVersion TYPE set(0) GRANULARITY 4,
    INDEX idx_client_platform ClientPlatform TYPE set(0) GRANULARITY 4,
    INDEX idx_web_version WebVersion TYPE set(0) GRANULARITY 4,
    INDEX idx_severity_text SeverityText TYPE bloom_filter GRANULARITY 4,
    INDEX idx_severity_number SeverityNumber TYPE bloom_filter GRANULARITY 4,
    INDEX idx_trace_id TraceId TYPE bloom_filter(0.001) GRANULARITY 4,
    INDEX idx_id Id TYPE bloom_filter GRANULARITY 4,
    INDEX idx_res_attr_key mapKeys(ResourceAttributes) TYPE bloom_filter(0.01) GRANULARITY 4,
    INDEX idx_res_attr_value mapValues(ResourceAttributes) TYPE bloom_filter(0.01) GRANULARITY 4,
    INDEX idx_scope_attr_key mapKeys(ScopeAttributes) TYPE bloom_filter(0.01) GRANULARITY 4,
    INDEX idx_scope_attr_value mapValues(ScopeAttributes) TYPE bloom_filter(0.01) GRANULARITY 4,
    INDEX idx_log_attr_key mapKeys(LogAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_log_attr_value mapValues(LogAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_body Body TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 1
) ENGINE = {{.SOPHONZ_REPLICATED}}MergeTree
    PARTITION BY toDate(Timestamp)
    ORDER BY (ServiceNamespace, ClientPlatform, ServiceVersion, SeverityText, WebVersion, toUnixTimestamp(Timestamp), Id)
    TTL toDateTime(Timestamp) + toIntervalSecond(2592000)
    SETTINGS index_granularity=8192, ttl_only_drop_parts = 1;


CREATE TABLE IF NOT EXISTS sophonz_logs.distributed_logs_v2 AS sophonz_logs.logs_v2
    ENGINE = Distributed('{{.SOPHONZ_CLUSTER}}', 'sophonz_logs', 'logs_v2', cityHash64(Id));
