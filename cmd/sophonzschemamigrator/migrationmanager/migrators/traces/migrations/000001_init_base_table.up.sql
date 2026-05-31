-- sophonz_index_v2
CREATE TABLE IF NOT EXISTS sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    timestamp DateTime64(9) CODEC(DoubleDelta, ZSTD(1)),
    endTimestamp DateTime64(9) CODEC(DoubleDelta, ZSTD(1)),
    traceID FixedString(32) CODEC(ZSTD(1)),
    spanID FixedString(16) CODEC(ZSTD(1)),
    parentSpanID FixedString(16) CODEC(ZSTD(1)),
    isRootSpan Bool CODEC(ZSTD(1)),
    sessionID FixedString(32) CODEC(ZSTD(1)),
    serviceNamespace LowCardinality(String) CODEC(ZSTD(1)),
    clientPlatform LowCardinality(String) CODEC(ZSTD(1)),
    serviceVersion LowCardinality(String) CODEC(ZSTD(1)),
    serviceName LowCardinality(String) CODEC(ZSTD(1)),
    webVersion LowCardinality(String) CODEC(ZSTD(1)),
    name LowCardinality(String) CODEC(ZSTD(1)),
    component String CODEC(ZSTD(1)),
    appSpanType LowCardinality(String) CODEC(ZSTD(1)),
    appScreenID UInt32 CODEC(ZSTD(1)),
    appScreenName String CODEC(ZSTD(1)),
    appScreenType LowCardinality(String) CODEC(ZSTD(1)),
    appScreenGroupID UInt32 CODEC(ZSTD(1)),
    appScreenGroupName LowCardinality(String) CODEC(ZSTD(1)),
    durationNano UInt64 CODEC(T64, ZSTD(1)),
    userID String CODEC(ZSTD(1)),
    statusCode Int16 CODEC(T64, ZSTD(1)),
    statusCodeString String CODEC(ZSTD(1)),
    statusMessage String CODEC(ZSTD(1)),
    events Array(String) CODEC(ZSTD(2)),
    httpMethod LowCardinality(String) CODEC(ZSTD(1)),
    httpURL String CODEC(ZSTD(1)),
    httpHost LowCardinality(String) CODEC(ZSTD(1)),
    responseStatusCode UInt16 CODEC(ZSTD(1)),
    msgSystem LowCardinality(String) CODEC(ZSTD(1)),
    msgOperation LowCardinality(String) CODEC(ZSTD(1)),
    hasError Bool CODEC(ZSTD(1)),
    stringTagMap Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    numberTagMap Map(LowCardinality(String), Float64) CODEC(ZSTD(1)),
    boolTagMap Map(LowCardinality(String), Bool) CODEC(ZSTD(1)),
    resourceTagsMap Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    kind Int8 CODEC(T64, ZSTD(1)),
    spanKind String CODEC(ZSTD(1)),
    INDEX idx_serviceNamespace serviceNamespace TYPE set(0) GRANULARITY 4,
    INDEX idx_clientPlatform clientPlatform TYPE set(0) GRANULARITY 4,
    INDEX idx_serviceVersion serviceVersion TYPE bloom_filter GRANULARITY 4,
    INDEX idx_serviceName serviceName TYPE bloom_filter GRANULARITY 4,
    INDEX idx_sessionID sessionID TYPE bloom_filter GRANULARITY 4,
    INDEX idx_ngram_sessionID sessionID TYPE ngrambf_v1(3, 10000, 3, 7) GRANULARITY 4,
    INDEX idx_name name TYPE bloom_filter GRANULARITY 4,
    INDEX idx_ngram_name name TYPE ngrambf_v1(3, 10000, 3, 7) GRANULARITY 4,
    INDEX idx_app_span_type appSpanType TYPE set(0) GRANULARITY 4,
    INDEX idx_appScreenID appScreenID TYPE bloom_filter GRANULARITY 4,
    INDEX idx_appScreenName appScreenName TYPE bloom_filter GRANULARITY 4,
    INDEX idx_ngram_appScreenName appScreenName TYPE ngrambf_v1(3, 10000, 3, 7) GRANULARITY 4,
    INDEX idx_appScreenType appScreenType TYPE set(0) GRANULARITY 4,
    INDEX idx_appScreenGroupID appScreenGroupID TYPE bloom_filter GRANULARITY 4,
    INDEX idx_appScreenGroupName appScreenGroupName TYPE bloom_filter GRANULARITY 4,
    INDEX idx_ngram_appScreenGroupName appScreenGroupName TYPE ngrambf_v1(3, 10000, 3, 7) GRANULARITY 4,
    INDEX idx_duration durationNano TYPE minmax GRANULARITY 1,
    INDEX idx_userID userID TYPE bloom_filter GRANULARITY 4,
    INDEX idx_ngram_userID userID TYPE ngrambf_v1(3, 10000, 3, 7) GRANULARITY 4,
    INDEX idx_hasError hasError TYPE set(2) GRANULARITY 1,
    INDEX idx_httpURL httpURL TYPE bloom_filter GRANULARITY 4,
    INDEX idx_ngram_httpURL httpURL TYPE ngrambf_v1(3, 10000, 3, 7) GRANULARITY 4,
    INDEX idx_httpHost httpHost TYPE bloom_filter GRANULARITY 4,
    INDEX idx_ngram_httpHost httpHost TYPE ngrambf_v1(3, 10000, 3, 7) GRANULARITY 4,
    INDEX idx_httpMethod httpMethod TYPE set(0) GRANULARITY 4,
    INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1,
    INDEX idx_responseStatusCode responseStatusCode TYPE set(0) GRANULARITY 1,
    INDEX idx_stringTagMapKeys mapKeys(stringTagMap) TYPE bloom_filter(0.01) GRANULARITY 64,
    INDEX idx_stringTagMapValues mapValues(stringTagMap) TYPE bloom_filter(0.01) GRANULARITY 64,
    INDEX idx_resourceTagsMapKeys mapKeys(resourceTagsMap) TYPE bloom_filter(0.01) GRANULARITY 64,
    INDEX idx_resourceTagsMapValues mapValues(resourceTagsMap) TYPE bloom_filter(0.01) GRANULARITY 64,
    INDEX idx_statusCodeString statusCodeString TYPE set(3) GRANULARITY 4,
    INDEX idx_spanKind spanKind TYPE set(5) GRANULARITY 4,
    INDEX idx_isRootSpan isRootSpan TYPE set(2) GRANULARITY 1,
    PROJECTION timestampSort (SELECT * ORDER BY timestamp)
)   ENGINE = {{.SOPHONZ_REPLICATED}}MergeTree PARTITION BY toDate(timestamp)
    ORDER BY (serviceNamespace, clientPlatform, serviceVersion, serviceName, timestamp, appSpanType, statusCode, appScreenName, appScreenGroupName, userID, sessionID)
    TTL toDateTime(timestamp) + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1;

CREATE TABLE IF NOT EXISTS sophonz_traces.distributed_sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}} AS sophonz_traces.sophonz_index_v2
    ENGINE = Distributed('{{.SOPHONZ_CLUSTER}}', 'sophonz_traces', 'sophonz_index_v2', cityHash64(traceID));

-- sophonz_spans
CREATE TABLE IF NOT EXISTS sophonz_traces.sophonz_spans ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    timestamp DateTime64(9) CODEC(DoubleDelta, LZ4),
    traceID   FixedString(32) CODEC(ZSTD(1)),
    model     String CODEC(ZSTD(9))
)   ENGINE = {{.SOPHONZ_REPLICATED}}MergeTree PARTITION BY toDate(timestamp)
    ORDER BY traceID
    TTL toDateTime(timestamp) + toIntervalSecond(2592000)
    SETTINGS index_granularity = 1024, ttl_only_drop_parts = 1;

CREATE TABLE IF NOT EXISTS sophonz_traces.distributed_sophonz_spans ON CLUSTER {{.SOPHONZ_CLUSTER}} AS sophonz_traces.sophonz_spans
    ENGINE = Distributed('{{.SOPHONZ_CLUSTER}}', 'sophonz_traces', 'sophonz_spans', cityHash64(traceID));

-- span_attributes
CREATE TABLE IF NOT EXISTS sophonz_traces.span_attributes ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    timestamp DateTime CODEC(DoubleDelta, ZSTD(1)),
    serviceNamespace LowCardinality(String) CODEC(ZSTD(1)),
    tagKey LowCardinality(String) CODEC(ZSTD(1)),
    tagType Enum8('tag' = 1, 'resource' = 2) CODEC(ZSTD(1)),
    dataType Enum8('string' = 1, 'bool' = 2, 'float64' = 3) CODEC(ZSTD(1)),
    stringTagValue String CODEC(ZSTD(1)),
    float64TagValue Nullable(Float64) CODEC(ZSTD(1)),
    isColumn Bool CODEC(ZSTD(1))
)   ENGINE = {{.SOPHONZ_REPLICATED}}ReplacingMergeTree PARTITION BY toDate(timestamp)
    ORDER BY (serviceNamespace, timestamp, tagKey, tagType, dataType, stringTagValue, float64TagValue, isColumn)
    TTL timestamp + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1, allow_nullable_key = 1;

CREATE TABLE IF NOT EXISTS sophonz_traces.distributed_span_attributes ON CLUSTER {{.SOPHONZ_CLUSTER}} AS sophonz_traces.span_attributes
    ENGINE = Distributed('{{.SOPHONZ_CLUSTER}}', 'sophonz_traces', 'span_attributes', cityHash64(rand()));

-- span_attributes_keys
CREATE TABLE IF NOT EXISTS sophonz_traces.span_attributes_keys ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace LowCardinality(String) CODEC(ZSTD(1)),
    tagKey LowCardinality(String) CODEC(ZSTD(1)),
    tagType Enum8('tag' = 1, 'resource' = 2) CODEC(ZSTD(1)),
    dataType Enum8('string' = 1, 'bool' = 2, 'float64' = 3) CODEC(ZSTD(1)),
    isColumn Bool CODEC(ZSTD(1))
)   ENGINE = {{.SOPHONZ_REPLICATED}}ReplacingMergeTree
    ORDER BY (serviceNamespace, tagKey, tagType, dataType, isColumn);

CREATE TABLE IF NOT EXISTS sophonz_traces.distributed_span_attributes_keys ON CLUSTER {{.SOPHONZ_CLUSTER}} AS sophonz_traces.span_attributes_keys
    ENGINE = Distributed('{{.SOPHONZ_CLUSTER}}', 'sophonz_traces', 'span_attributes_keys', cityHash64(rand()));

-- sophonz_error_index_v2
CREATE TABLE IF NOT EXISTS sophonz_traces.sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    timestamp DateTime64(9) CODEC(DoubleDelta, LZ4),
    errorID FixedString(32) CODEC(ZSTD(1)),
    groupID FixedString(32) CODEC(ZSTD(1)),
    traceID FixedString(32) CODEC(ZSTD(1)),
    spanID FixedString(16) CODEC(ZSTD(1)),
    serviceNamespace LowCardinality(String) CODEC(ZSTD(1)),
    clientPlatform LowCardinality(String) CODEC(ZSTD(1)),
    serviceVersion LowCardinality(String) CODEC(ZSTD(1)),
    serviceName LowCardinality(String) CODEC(ZSTD(1)),
    webVersion LowCardinality(String) CODEC(ZSTD(1)),
    sessionID FixedString(32) CODEC(ZSTD(1)),
    appSpanType LowCardinality(String) CODEC(ZSTD(1)),
    exceptionType LowCardinality(String) CODEC(ZSTD(1)),
    exceptionMessage String CODEC(ZSTD(1)),
    exceptionStacktrace String CODEC(ZSTD(1)),
    exceptionEscaped Bool CODEC(T64, ZSTD(1)),
    resourceTagsMap Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    INDEX idx_serviceNamespace serviceNamespace TYPE set(0) GRANULARITY 4,
    INDEX idx_clientPlatform clientPlatform TYPE set(0) GRANULARITY 4,
    INDEX idx_serviceVersion serviceVersion TYPE set(0) GRANULARITY 4,
    INDEX idx_serviceName serviceName TYPE set(0) GRANULARITY 4,
    INDEX idx_webVersion webVersion TYPE set(0) GRANULARITY 4,
    INDEX idx_groupID groupID TYPE bloom_filter GRANULARITY 4,
    INDEX idx_errorID errorID TYPE bloom_filter GRANULARITY 4,
    INDEX idx_appSpanType appSpanType TYPE set(0) GRANULARITY 4,
    INDEX idx_exceptionType exceptionType TYPE bloom_filter GRANULARITY 4,
    INDEX idx_ngram_exceptionType exceptionType TYPE ngrambf_v1(3, 10000, 3, 7) GRANULARITY 1,
    INDEX idx_exceptionMessage exceptionMessage TYPE bloom_filter GRANULARITY 4,
    INDEX idx_ngram_exceptionMessage exceptionMessage Type ngrambf_v1(3, 10000, 3, 7) GRANULARITY 1,
    INDEX idx_resourceTagsMapKeys mapKeys(resourceTagsMap) TYPE bloom_filter(0.01) GRANULARITY 64,
    INDEX idx_resourceTagsMapValues mapValues(resourceTagsMap) TYPE bloom_filter(0.01) GRANULARITY 64,
    PROJECTION prj_uniq (
        SELECT exceptionType, exceptionMessage, clientPlatform, serviceVersion, groupUniqArray(webVersion), count(spanID), uniq(sessionID), max(timestamp)
        GROUP BY serviceNamespace, clientPlatform, serviceVersion, exceptionType, exceptionMessage
    ),
    PROJECTION prj_uniq_hll12 (
        SELECT exceptionType, exceptionMessage, clientPlatform, serviceVersion, groupUniqArray(webVersion), count(spanID), uniqHLL12(sessionID), max(timestamp)
        GROUP BY serviceNamespace, clientPlatform, serviceVersion, exceptionType, exceptionMessage
    ),
    PROJECTION prj_uniq_combined (
        SELECT exceptionType, exceptionMessage, clientPlatform, serviceVersion, groupUniqArray(webVersion), count(spanID), uniqCombined(sessionID), max(timestamp)
        GROUP BY serviceNamespace, clientPlatform, serviceVersion, exceptionType, exceptionMessage
    )
)   ENGINE = {{.SOPHONZ_REPLICATED}}MergeTree PARTITION BY toDate(timestamp)
    ORDER BY (serviceNamespace, clientPlatform, serviceVersion, webVersion, exceptionType, timestamp, groupID)
    TTL toDateTime(timestamp) + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1;

CREATE TABLE IF NOT EXISTS sophonz_traces.distributed_sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}} AS sophonz_traces.sophonz_error_index_v2
    ENGINE = Distributed('{{.SOPHONZ_CLUSTER}}', 'sophonz_traces', 'sophonz_error_index_v2', cityHash64(groupID));
