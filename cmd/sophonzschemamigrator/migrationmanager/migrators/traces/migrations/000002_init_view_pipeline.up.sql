------------------------------------------------------------------------------------------------------------------------
-- default: stored_dashboard_global_metrics
------------------------------------------------------------------------------------------------------------------------
-- dashboard_global_metrics_5s
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_global_metrics_5s ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace            LowCardinality(String),
    ts                          DateTime,
    requestTotalCount           SimpleAggregateFunction(sum, UInt64),
    requestErrorCount           SimpleAggregateFunction(sum, UInt64),
    userIDUniqState             AggregateFunction(uniq, String),
    userIDUniqHLL12State        AggregateFunction(uniqHLL12, String),
    userIDUniqCombinedState     AggregateFunction(uniqCombined, String),
    sessionIDUniqState          AggregateFunction(uniq, FixedString(32)),
    sessionIDUniqHLL12State     AggregateFunction(uniqHLL12, FixedString(32)),
    sessionIDUniqCombinedState  AggregateFunction(uniqCombined, FixedString(32)),
    loadingTimeAvgState         AggregateFunction(avgOrNull, UInt64),
    responseTimeAvgState        AggregateFunction(avgOrNull, UInt64),
    visitCount                  SimpleAggregateFunction(sum, UInt64),
    errorCount                  SimpleAggregateFunction(sum, UInt64),
    crashCount                  SimpleAggregateFunction(sum, UInt64)
)   ENGINE = AggregatingMergeTree
    PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, ts)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_global_metrics_5s
            TO sophonz_traces.stored_dashboard_global_metrics_5s AS
SELECT
    serviceNamespace,
    toStartOfInterval(timestamp, toIntervalSecond(5))                        AS ts,
    sumSimpleState(toUInt64(appSpanType IN ('xhr', 'fetch')))                   AS requestTotalCount,
    sumSimpleState(toUInt64(
            appSpanType IN ('xhr', 'fetch')
                AND (responseStatusCode = 0 OR responseStatusCode >= 400)))  AS requestErrorCount,
    uniqState(userID)                                                        AS userIDUniqState,
    uniqHLL12State(userID)                                                   AS userIDUniqHLL12State,
    uniqCombinedState(userID)                                                AS userIDUniqCombinedState,
    uniqState(sessionID)                                                     AS sessionIDUniqState,
    uniqHLL12State(sessionID)                                                AS sessionIDUniqHLL12State,
    uniqCombinedState(sessionID)                                             AS sessionIDUniqCombinedState,
    avgOrNullStateIf(durationNano, appSpanType IN ('render', 'post-docs'))      AS loadingTimeAvgState,
    avgOrNullStateIf(durationNano, appSpanType IN ('xhr', 'fetch'))             AS responseTimeAvgState,
    sumSimpleState(toUInt64(appSpanType IN ('render', 'route')))                AS visitCount,
    sumSimpleState(toUInt64(appSpanType = 'error'))                             AS errorCount,
    sumSimpleState(toUInt64(appSpanType = 'crash'))                             AS crashCount
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, ts;

-- dashboard_global_metrics_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_global_metrics_1m ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_global_metrics_5s;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_global_metrics_1m
            TO sophonz_traces.stored_dashboard_global_metrics_1m AS
SELECT
    serviceNamespace,
    toStartOfMinute(ts)                                 AS ts,
    sum(requestTotalCount)                              AS requestTotalCount,
    sum(requestErrorCount)                              AS requestErrorCount,
    uniqMergeState(userIDUniqState)                     AS userIDUniqState,
    uniqHLL12MergeState(userIDUniqHLL12State)           AS userIDUniqHLL12State,
    uniqCombinedMergeState(userIDUniqCombinedState)     AS userIDUniqCombinedState,
    uniqMergeState(sessionIDUniqState)                  AS sessionIDUniqState,
    uniqHLL12MergeState(sessionIDUniqHLL12State)        AS sessionIDUniqHLL12State,
    uniqCombinedMergeState(sessionIDUniqCombinedState)  AS sessionIDUniqCombinedState,
    avgOrNullMergeState(loadingTimeAvgState)            AS loadingTimeAvgState,
    avgOrNullMergeState(responseTimeAvgState)           AS responseTimeAvgState,
    sum(visitCount)                                     AS visitCount,
    sum(errorCount)                                     AS errorCount,
    sum(crashCount)                                     AS crashCount
FROM
    sophonz_traces.stored_dashboard_global_metrics_5s
GROUP BY
    serviceNamespace, ts;


------------------------------------------------------------------------------------------------------------------------
-- default: stored_dashboard_global_metrics_user_count
------------------------------------------------------------------------------------------------------------------------
-- dashboard_global_metrics_user_count_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_global_metrics_user_count_1d ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace         LowCardinality(String),
    ts                       Date,
    userIDUniqState          AggregateFunction(uniq, String),
    userIDUniqHLL12State     AggregateFunction(uniqHLL12, String),
    userIDUniqCombinedState  AggregateFunction(uniqCombined, String)
)   ENGINE = AggregatingMergeTree
    PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, ts)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_global_metrics_user_count_1d
            TO sophonz_traces.stored_dashboard_global_metrics_user_count_1d AS
SELECT
    serviceNamespace,
    toStartOfDay(timestamp)    AS ts,
    uniqState(userID)          AS userIDUniqState,
    uniqHLL12State(userID)     AS userIDUniqHLL12State,
    uniqCombinedState(userID)  AS userIDUniqCombinedState
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, ts;


------------------------------------------------------------------------------------------------------------------------
-- default: stored_dashboard_common_metrics_slowest_request
------------------------------------------------------------------------------------------------------------------------
-- dashboard_common_metrics_slowest_request_5s
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_common_metrics_slowest_request_5s ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace      LowCardinality(String),
    clientPlatform           LowCardinality(String),
    ts                    DateTime,
    requestURL            String,
    requestCount          SimpleAggregateFunction(sum, UInt64),
    responseTimeAvgState  AggregateFunction(avgOrNull, UInt64),
    PROJECTION prjServiceNamespaceTS (SELECT * ORDER BY serviceNamespace, clientPlatform, ts)
)   ENGINE = AggregatingMergeTree
    PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, clientPlatform, requestURL, ts)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1, deduplicate_merge_projection_mode = 'rebuild';

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_common_metrics_slowest_request_5s
            TO sophonz_traces.stored_dashboard_common_metrics_slowest_request_5s AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfInterval(timestamp, toIntervalSecond(5))  AS ts,
    httpURL                                            AS requestURL,
    sumSimpleState(1)                                  AS requestCount,
    avgOrNullState(durationNano)                       AS responseTimeAvgState
FROM
    sophonz_traces.distributed_sophonz_index_v2
WHERE
    appSpanType IN ('xhr', 'fetch')
GROUP BY
    serviceNamespace, clientPlatform, requestURL, ts;

-- dashboard_common_metrics_slowest_request_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_common_metrics_slowest_request_1m ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_common_metrics_slowest_request_5s;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_common_metrics_slowest_request_1m
            TO sophonz_traces.stored_dashboard_common_metrics_slowest_request_1m AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfMinute(ts)                        AS ts,
    requestURL,
    sum(requestCount)                          AS requestCount,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState
FROM
    sophonz_traces.stored_dashboard_common_metrics_slowest_request_5s
GROUP BY
    serviceNamespace, clientPlatform, requestURL, ts;

-- dashboard_common_metrics_slowest_request_5m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_common_metrics_slowest_request_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_common_metrics_slowest_request_1m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_common_metrics_slowest_request_5m
            TO sophonz_traces.stored_dashboard_common_metrics_slowest_request_5m AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfFiveMinute(ts)                    AS ts,
    requestURL,
    sum(requestCount)                          AS requestCount,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState
FROM
    sophonz_traces.stored_dashboard_common_metrics_slowest_request_1m
GROUP BY
    serviceNamespace, clientPlatform, requestURL, ts;

-- dashboard_common_metrics_slowest_request_1h
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_common_metrics_slowest_request_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_common_metrics_slowest_request_5m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_common_metrics_slowest_request_1h
            TO sophonz_traces.stored_dashboard_common_metrics_slowest_request_1h AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfHour(ts)                          AS ts,
    requestURL,
    sum(requestCount)                          AS requestCount,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState
FROM
    sophonz_traces.stored_dashboard_common_metrics_slowest_request_5m
GROUP BY
    serviceNamespace, clientPlatform, requestURL, ts;

-- dashboard_common_metrics_slowest_request_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_common_metrics_slowest_request_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_common_metrics_slowest_request_1h;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_common_metrics_slowest_request_1d
            TO sophonz_traces.stored_dashboard_common_metrics_slowest_request_1d AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfDay(ts)                           AS ts,
    requestURL,
    sum(requestCount)                          AS requestCount,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState
FROM
    sophonz_traces.stored_dashboard_common_metrics_slowest_request_1h
GROUP BY
    serviceNamespace, clientPlatform, requestURL, ts;


------------------------------------------------------------------------------------------------------------------------
-- default: stored_dashboard_service_metrics_top_screen_visit
------------------------------------------------------------------------------------------------------------------------
-- dashboard_service_metrics_top_screen_visit_5s
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_5s ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace      LowCardinality(String),
    clientPlatform           LowCardinality(String),
    ts                    DateTime,
    appScreenName            String,
    visitCount            SimpleAggregateFunction(sum, UInt64),
    loadingTimeAvgState   AggregateFunction(avgOrNull, UInt64),
    responseTimeAvgState  AggregateFunction(avgOrNull, UInt64),
    PROJECTION prjServiceNamespaceTS (SELECT * ORDER BY serviceNamespace, clientPlatform, ts)
)   ENGINE = AggregatingMergeTree
    PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, clientPlatform, appScreenName, ts)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1, deduplicate_merge_projection_mode = 'rebuild';

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_top_screen_visit_5s
            TO sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_5s AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfInterval(timestamp, toIntervalSecond(5))                    AS ts,
    appScreenName,
    sumSimpleState(toUInt64(appSpanType IN ('render', 'route')))            AS visitCount,
    avgOrNullIfState(durationNano, appSpanType IN ('render', 'post-docs'))  AS loadingTimeAvgState,
    avgOrNullIfState(durationNano, appSpanType IN ('xhr', 'fetch'))         AS responseTimeAvgState
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, clientPlatform, appScreenName, ts;

-- dashboard_service_metrics_top_screen_visit_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_1m ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_5s;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_top_screen_visit_1m
            TO sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_1m AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfMinute(ts)                        AS ts,
    appScreenName,
    sum(visitCount)                            AS visitCount,
    avgOrNullMergeState(loadingTimeAvgState)   AS loadingTimeAvgState,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState    
FROM
    sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_5s
GROUP BY
    serviceNamespace, clientPlatform, appScreenName, ts;

-- dashboard_service_metrics_top_screen_visit_5m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_1m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_top_screen_visit_5m
            TO sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_5m AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfFiveMinute(ts)                    AS ts,
    appScreenName,
    sum(visitCount)                            AS visitCount,
    avgOrNullMergeState(loadingTimeAvgState)   AS loadingTimeAvgState,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState
FROM
    sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_1m
GROUP BY
    serviceNamespace, clientPlatform, appScreenName, ts;

-- dashboard_service_metrics_top_screen_visit_1h
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_5m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_top_screen_visit_1h
            TO sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_1h AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfHour(ts)                          AS ts,
    appScreenName,
    sum(visitCount)                            AS visitCount,
    avgOrNullMergeState(loadingTimeAvgState)   AS loadingTimeAvgState,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState
FROM
    sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_5m
GROUP BY
    serviceNamespace, clientPlatform, appScreenName, ts;

-- dashboard_service_metrics_top_screen_visit_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_1h;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_top_screen_visit_1d
            TO sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_1d AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfDay(ts)                           AS ts,
    appScreenName,
    sum(visitCount)                            AS visitCount,
    avgOrNullMergeState(loadingTimeAvgState)   AS loadingTimeAvgState,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState
FROM
    sophonz_traces.stored_dashboard_service_metrics_top_screen_visit_1h
GROUP BY
    serviceNamespace, clientPlatform, appScreenName, ts;


------------------------------------------------------------------------------------------------------------------------
-- default: stored_dashboard_common_metrics_top_error
------------------------------------------------------------------------------------------------------------------------
-- dashboard_common_metrics_top_error_5s
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_common_metrics_top_error_5s ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace            LowCardinality(String),
    clientPlatform                 LowCardinality(String),
    serviceVersion              LowCardinality(String),
    ts                          DateTime,
    appSpanType                    LowCardinality(String),
    exceptionType               LowCardinality(String),
    count                       SimpleAggregateFunction(sum, UInt64),
    sessionIDUniqState          AggregateFunction(uniq, FixedString(32)),
    sessionIDUniqHLL12State     AggregateFunction(uniqHLL12, FixedString(32)),
    sessionIDUniqCombinedState  AggregateFunction(uniqCombined, FixedString(32)),
    PROJECTION prjServiceNamespaceTS (SELECT * ORDER BY serviceNamespace, ts)
)   ENGINE = AggregatingMergeTree
    PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, clientPlatform, serviceVersion, appSpanType, exceptionType, ts)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1, deduplicate_merge_projection_mode = 'rebuild';

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_common_metrics_top_error_5s
            TO sophonz_traces.stored_dashboard_common_metrics_top_error_5s AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    toStartOfInterval(timestamp, toIntervalSecond(5))  AS ts,
    appSpanType,
    exceptionType,
    sumSimpleState(1)                                  AS count,
    uniqState(sessionID)                               AS sessionIDUniqState,
    uniqHLL12State(sessionID)                          AS sessionIDUniqHLL12State,
    uniqCombinedState(sessionID)                       AS sessionIDUniqCombinedState
FROM
    sophonz_traces.distributed_sophonz_error_index_v2
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, appSpanType, exceptionType, ts;

-- dashboard_common_metrics_top_error_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_common_metrics_top_error_1m ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_common_metrics_top_error_5s;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_common_metrics_top_error_1m
            TO sophonz_traces.stored_dashboard_common_metrics_top_error_1m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    toStartOfMinute(ts)                                 AS ts,
    appSpanType,
    exceptionType,
    sum(count)                                          AS count,
    uniqMergeState(sessionIDUniqState)                  AS sessionIDUniqState,
    uniqHLL12MergeState(sessionIDUniqHLL12State)        AS sessionIDUniqHLL12State,
    uniqCombinedMergeState(sessionIDUniqCombinedState)  AS sessionIDUniqCombinedState
FROM
    sophonz_traces.stored_dashboard_common_metrics_top_error_5s
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, appSpanType, exceptionType, ts;

-- dashboard_common_metrics_top_error_5m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_common_metrics_top_error_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_common_metrics_top_error_1m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_common_metrics_top_error_5m
            TO sophonz_traces.stored_dashboard_common_metrics_top_error_5m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    toStartOfMinute(ts)                                 AS ts,
    appSpanType,
    exceptionType,
    sum(count)                                          AS count,
    uniqMergeState(sessionIDUniqState)                  AS sessionIDUniqState,
    uniqHLL12MergeState(sessionIDUniqHLL12State)        AS sessionIDUniqHLL12State,
    uniqCombinedMergeState(sessionIDUniqCombinedState)  AS sessionIDUniqCombinedState
FROM
    sophonz_traces.stored_dashboard_common_metrics_top_error_1m
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, appSpanType, exceptionType, ts;

-- dashboard_common_metrics_top_error_1h
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_common_metrics_top_error_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_common_metrics_top_error_5m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_common_metrics_top_error_1h
            TO sophonz_traces.stored_dashboard_common_metrics_top_error_1h AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    toStartOfMinute(ts)                                 AS ts,
    appSpanType,
    exceptionType,
    sum(count)                                          AS count,
    uniqMergeState(sessionIDUniqState)                  AS sessionIDUniqState,
    uniqHLL12MergeState(sessionIDUniqHLL12State)        AS sessionIDUniqHLL12State,
    uniqCombinedMergeState(sessionIDUniqCombinedState)  AS sessionIDUniqCombinedState
FROM
    sophonz_traces.stored_dashboard_common_metrics_top_error_5m
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, appSpanType, exceptionType, ts;

-- dashboard_common_metrics_top_error_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_common_metrics_top_error_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_common_metrics_top_error_1h;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_common_metrics_top_error_1d
            TO sophonz_traces.stored_dashboard_common_metrics_top_error_1d AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    toStartOfMinute(ts)                                 AS ts,
    appSpanType,
    exceptionType,
    sum(count)                                          AS count,
    uniqMergeState(sessionIDUniqState)                  AS sessionIDUniqState,
    uniqHLL12MergeState(sessionIDUniqHLL12State)        AS sessionIDUniqHLL12State,
    uniqCombinedMergeState(sessionIDUniqCombinedState)  AS sessionIDUniqCombinedState
FROM
    sophonz_traces.stored_dashboard_common_metrics_top_error_1h
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, appSpanType, exceptionType, ts;


------------------------------------------------------------------------------------------------------------------------
-- default: stored_dashboard_service_metrics
------------------------------------------------------------------------------------------------------------------------
-- dashboard_service_metrics_5s
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_5s ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace            LowCardinality(String),
    clientPlatform                 LowCardinality(String),
    serviceName                 LowCardinality(String),
    serviceVersion              LowCardinality(String),
    ts                          DateTime,
    loadingTimeAvgState         AggregateFunction(avgOrNull, UInt64),
    loadingTimeQuantilesState   AggregateFunction(quantiles(0.5, 0.9, 0.95, 0.99), UInt64),
    responseTimeAvgState        AggregateFunction(avgOrNull, UInt64),
    responseTimeQuantilesState  AggregateFunction(quantiles(0.5, 0.9, 0.95, 0.99), UInt64),
    sessionIDUniqState          AggregateFunction(uniq, FixedString(32)),
    sessionIDUniqHLL12State     AggregateFunction(uniqHLL12, FixedString(32)),
    sessionIDUniqCombinedState  AggregateFunction(uniqCombined, FixedString(32)),
    userIDUniqState             AggregateFunction(uniq, String),
    userIDUniqHLL12State        AggregateFunction(uniqHLL12, String),
    userIDUniqCombinedState     AggregateFunction(uniqCombined, String),
    requestTotalCount           SimpleAggregateFunction(sum, UInt64),
    requestErrorCount           SimpleAggregateFunction(sum, UInt64),
    traceCount                  SimpleAggregateFunction(sum, UInt64),
    visitCount                  SimpleAggregateFunction(sum, UInt64),
    crashCount                  SimpleAggregateFunction(sum, UInt64),
    errorCount                  SimpleAggregateFunction(sum, UInt64),
    PROJECTION prjServiceNamespaceTS (SELECT * ORDER BY serviceNamespace, ts)
)   ENGINE = AggregatingMergeTree
    PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, clientPlatform, serviceName, serviceVersion, ts)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1, deduplicate_merge_projection_mode = 'rebuild';

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_5s
            TO sophonz_traces.stored_dashboard_service_metrics_5s AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceName,
    serviceVersion,
    toStartOfInterval(timestamp, toIntervalSecond(5))                                                                 AS ts,
    avgOrNullStateIf(durationNano, appSpanType IN ('render', 'post-docs'))                                               AS loadingTimeAvgState,
    quantilesStateIf(0.5, 0.9, 0.95, 0.99)(durationNano, appSpanType IN ('render', 'post-docs'))                         AS loadingTimeQuantilesState,
    avgOrNullStateIf(durationNano, appSpanType IN ('xhr', 'fetch'))                                                      AS responseTimeAvgState,
    quantilesStateIf(0.5, 0.9, 0.95, 0.99)(durationNano, appSpanType IN ('xhr', 'fetch'))                                AS responseTimeQuantilesState,
    uniqState(sessionID)                                                                                              AS sessionIDUniqState,
    uniqHLL12State(sessionID)                                                                                         AS sessionIDUniqHLL12State,
    uniqCombinedState(sessionID)                                                                                      AS sessionIDUniqCombinedState,
    uniqState(userID)                                                                                                 AS userIDUniqState,
    uniqHLL12State(userID)                                                                                            AS userIDUniqHLL12State,
    uniqCombinedState(userID)                                                                                         AS userIDUniqCombinedState,
    sumSimpleState(toUInt64(appSpanType IN ('xhr', 'fetch')))                                                            AS requestTotalCount,
    sumSimpleState(toUInt64(appSpanType IN ('xhr', 'fetch') AND (responseStatusCode = 0 OR responseStatusCode >= 400)))  AS requestErrorCount,
    sumSimpleState(toUInt64(isRootSpan = true))                                                                       AS traceCount,
    sumSimpleState(toUInt64(appSpanType IN ('render', 'route')))                                                         AS visitCount,
    sumSimpleState(toUInt64(appSpanType = 'crash'))                                                                      AS crashCount,
    sumSimpleState(toUInt64(appSpanType = 'error'))                                                                      AS errorCount
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, clientPlatform, serviceName, serviceVersion, ts;

-- dashboard_service_metrics_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_1m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_dashboard_service_metrics_5s;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_1m
            TO sophonz_traces.stored_dashboard_service_metrics_1m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceName,
    serviceVersion,
    toStartOfMinute(ts)                                                    AS ts,
    avgOrNullMergeState(loadingTimeAvgState)                               AS loadingTimeAvgState,
    quantilesMergeState(0.5, 0.9, 0.95, 0.99)(loadingTimeQuantilesState)   AS loadingTimeQuantilesState,
    avgOrNullMergeState(responseTimeAvgState)                              AS responseTimeAvgState,
    quantilesMergeState(0.5, 0.9, 0.95, 0.99)(responseTimeQuantilesState)  AS responseTimeQuantilesState,
    uniqMergeState(sessionIDUniqState)                                     AS sessionIDUniqState,
    uniqHLL12MergeState(sessionIDUniqHLL12State)                           AS sessionIDUniqHLL12State,
    uniqCombinedMergeState(sessionIDUniqCombinedState)                     AS sessionIDUniqCombinedState,
    uniqMergeState(userIDUniqState)                                        AS userIDUniqState,
    uniqHLL12MergeState(userIDUniqHLL12State)                              AS userIDUniqHLL12State,
    uniqCombinedMergeState(userIDUniqCombinedState)                        AS userIDUniqCombinedState,
    sum(requestTotalCount)                                                 AS requestTotalCount,
    sum(requestErrorCount)                                                 AS requestErrorCount,
    sum(traceCount)                                                        AS traceCount,
    sum(visitCount)                                                        AS visitCount,
    sum(crashCount)                                                        AS crashCount,
    sum(errorCount)                                                        AS errorCount
FROM
    sophonz_traces.stored_dashboard_service_metrics_5s
GROUP BY
    serviceNamespace, clientPlatform, serviceName, serviceVersion, ts;

-- dashboard_service_metrics_5m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_dashboard_service_metrics_1m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_5m
            TO sophonz_traces.stored_dashboard_service_metrics_5m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceName,
    serviceVersion,
    toStartOfFiveMinutes(ts)                                               AS ts,
    avgOrNullMergeState(loadingTimeAvgState)                               AS loadingTimeAvgState,
    quantilesMergeState(0.5, 0.9, 0.95, 0.99)(loadingTimeQuantilesState)   AS loadingTimeQuantilesState,
    avgOrNullMergeState(responseTimeAvgState)                              AS responseTimeAvgState,
    quantilesMergeState(0.5, 0.9, 0.95, 0.99)(responseTimeQuantilesState)  AS responseTimeQuantilesState,
    uniqMergeState(sessionIDUniqState)                                     AS sessionIDUniqState,
    uniqHLL12MergeState(sessionIDUniqHLL12State)                           AS sessionIDUniqHLL12State,
    uniqCombinedMergeState(sessionIDUniqCombinedState)                     AS sessionIDUniqCombinedState,
    uniqMergeState(userIDUniqState)                                        AS userIDUniqState,
    uniqHLL12MergeState(userIDUniqHLL12State)                              AS userIDUniqHLL12State,
    uniqCombinedMergeState(userIDUniqCombinedState)                        AS userIDUniqCombinedState,
    sum(requestTotalCount)                                                 AS requestTotalCount,
    sum(requestErrorCount)                                                 AS requestErrorCount,
    sum(traceCount)                                                        AS traceCount,
    sum(visitCount)                                                        AS visitCount,
    sum(crashCount)                                                        AS crashCount,
    sum(errorCount)                                                        AS errorCount
FROM
    sophonz_traces.stored_dashboard_service_metrics_1m
GROUP BY
    serviceNamespace, clientPlatform, serviceName, serviceVersion, ts;

-- dashboard_service_metrics_1h
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_dashboard_service_metrics_5m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_1h
            TO sophonz_traces.stored_dashboard_service_metrics_1h AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceName,
    serviceVersion,
    toStartOfHour(ts)                                                      AS ts,
    avgOrNullMergeState(loadingTimeAvgState)                               AS loadingTimeAvgState,
    quantilesMergeState(0.5, 0.9, 0.95, 0.99)(loadingTimeQuantilesState)   AS loadingTimeQuantilesState,
    avgOrNullMergeState(responseTimeAvgState)                              AS responseTimeAvgState,
    quantilesMergeState(0.5, 0.9, 0.95, 0.99)(responseTimeQuantilesState)  AS responseTimeQuantilesState,
    uniqMergeState(sessionIDUniqState)                                     AS sessionIDUniqState,
    uniqHLL12MergeState(sessionIDUniqHLL12State)                           AS sessionIDUniqHLL12State,
    uniqCombinedMergeState(sessionIDUniqCombinedState)                     AS sessionIDUniqCombinedState,
    uniqMergeState(userIDUniqState)                                        AS userIDUniqState,
    uniqHLL12MergeState(userIDUniqHLL12State)                              AS userIDUniqHLL12State,
    uniqCombinedMergeState(userIDUniqCombinedState)                        AS userIDUniqCombinedState,
    sum(requestTotalCount)                                                 AS requestTotalCount,
    sum(requestErrorCount)                                                 AS requestErrorCount,
    sum(traceCount)                                                        AS traceCount,
    sum(visitCount)                                                        AS visitCount,
    sum(crashCount)                                                        AS crashCount,
    sum(errorCount)                                                        AS errorCount
FROM
    sophonz_traces.stored_dashboard_service_metrics_5m
GROUP BY
    serviceNamespace, clientPlatform, serviceName, serviceVersion, ts;

-- dashboard_service_metrics_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_dashboard_service_metrics_1h;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_1d
            TO sophonz_traces.stored_dashboard_service_metrics_1d AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceName,
    serviceVersion,
    toStartOfDay(ts)                                                       AS ts,
    avgOrNullMergeState(loadingTimeAvgState)                               AS loadingTimeAvgState,
    quantilesMergeState(0.5, 0.9, 0.95, 0.99)(loadingTimeQuantilesState)   AS loadingTimeQuantilesState,
    avgOrNullMergeState(responseTimeAvgState)                              AS responseTimeAvgState,
    quantilesMergeState(0.5, 0.9, 0.95, 0.99)(responseTimeQuantilesState)  AS responseTimeQuantilesState,
    uniqMergeState(sessionIDUniqState)                                     AS sessionIDUniqState,
    uniqHLL12MergeState(sessionIDUniqHLL12State)                           AS sessionIDUniqHLL12State,
    uniqCombinedMergeState(sessionIDUniqCombinedState)                     AS sessionIDUniqCombinedState,
    uniqMergeState(userIDUniqState)                                        AS userIDUniqState,
    uniqHLL12MergeState(userIDUniqHLL12State)                              AS userIDUniqHLL12State,
    uniqCombinedMergeState(userIDUniqCombinedState)                        AS userIDUniqCombinedState,
    sum(requestTotalCount)                                                 AS requestTotalCount,
    sum(requestErrorCount)                                                 AS requestErrorCount,
    sum(traceCount)                                                        AS traceCount,
    sum(visitCount)                                                        AS visitCount,
    sum(crashCount)                                                        AS crashCount,
    sum(errorCount)                                                        AS errorCount
FROM
    sophonz_traces.stored_dashboard_service_metrics_1h
GROUP BY
    serviceNamespace, clientPlatform, serviceName, serviceVersion, ts;


------------------------------------------------------------------------------------------------------------------------
-- default: stored_dashboard_service_metrics_crash_device
------------------------------------------------------------------------------------------------------------------------
-- dashboard_service_metrics_crash_device_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_crash_device_1m ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace  LowCardinality(String),
    clientPlatform       LowCardinality(String),
    ts                DateTime,
    deviceModel       LowCardinality(String),
    crashCount        SimpleAggregateFunction(sum, UInt64)
)   ENGINE = AggregatingMergeTree
    PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, clientPlatform, deviceModel, ts)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_crash_device_1m
            TO sophonz_traces.stored_dashboard_service_metrics_crash_device_1m AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfMinute(timestamp)                  AS ts,
    resourceTagsMap['device.model.identifier']  AS deviceModel,
    sumSimpleState(1)                           AS crashCount
FROM
    sophonz_traces.distributed_sophonz_index_v2
WHERE
    appSpanType = 'crash'
GROUP BY
    serviceNamespace, clientPlatform, deviceModel, ts;

-- dashboard_service_metrics_crash_device_5m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_crash_device_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_dashboard_service_metrics_crash_device_1m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_crash_device_5m
            TO sophonz_traces.stored_dashboard_service_metrics_crash_device_5m AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfFiveMinutes(ts)  AS ts,
    deviceModel,
    sum(crashCount)           AS crashCount
FROM
    sophonz_traces.stored_dashboard_service_metrics_crash_device_1m
GROUP BY
    serviceNamespace, clientPlatform, deviceModel, ts;

-- dashboard_service_metrics_crash_device_1h
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_crash_device_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_dashboard_service_metrics_crash_device_5m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_crash_device_1h
            TO sophonz_traces.stored_dashboard_service_metrics_crash_device_1h AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfHour(ts)  AS ts,
    deviceModel,
    sum(crashCount)    AS crashCount
FROM
    sophonz_traces.stored_dashboard_service_metrics_crash_device_5m
GROUP BY
    serviceNamespace, clientPlatform, deviceModel, ts;

-- dashboard_service_metrics_crash_device_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_crash_device_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_dashboard_service_metrics_crash_device_1h;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_crash_device_1d
            TO sophonz_traces.stored_dashboard_service_metrics_crash_device_1d AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfDay(ts)  AS ts,
    deviceModel,
    sum(crashCount)   AS crashCount
FROM
    sophonz_traces.stored_dashboard_service_metrics_crash_device_1h
GROUP BY
    serviceNamespace, clientPlatform, deviceModel, ts;

------------------------------------------------------------------------------------------------------------------------
-- stored_analysis_session
------------------------------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_session ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace              LowCardinality(String),
    clientPlatform                   LowCardinality(String),
    serviceName                   LowCardinality(String),
    serviceVersion                LowCardinality(String),
    webVersionState               AggregateFunction(groupUniqArray, LowCardinality(String)),
    sessionID                     FixedString(32),
    ts                            DateTime,
    startTimestamp                SimpleAggregateFunction(min, DateTime64(9)),
    endTimestamp                  SimpleAggregateFunction(max, DateTime64(9)),
    userIDState                   AggregateFunction(groupUniqArray, String),
    traceCount                    SimpleAggregateFunction(sum, UInt64),
    hasError                      SimpleAggregateFunction(max, UInt8),
    deviceModel                   LowCardinality(String),
    osNameVersionMajor            LowCardinality(String),
    networkConnectionTypeState    AggregateFunction(groupUniqArray, String),
    browserNameVersionMajorState  AggregateFunction(groupUniqArray, String),
    PROJECTION prjServiceNameTS (SELECT * ORDER BY serviceNamespace, ts)
)   ENGINE = AggregatingMergeTree
    PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, clientPlatform, serviceName, serviceVersion, ts, sessionID)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS deduplicate_merge_projection_mode = 'rebuild';

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_session
            TO sophonz_traces.stored_analysis_session AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceName,
    serviceVersion,
    groupUniqArrayState(webVersion)              AS webVersionState,
    sessionID,
    min(timestamp)                               AS ts,
    minSimpleState(timestamp)                    AS startTimestamp,
    maxSimpleState(endTimestamp)                 AS endTimestamp,
    groupUniqArrayState(userID)                  AS userIDState,
    sumSimpleState(toUInt64(isRootSpan = true))  AS traceCount,
    maxSimpleState(hasError)                     AS hasError,
    any(resourceTagsMap['device.model.identifier'])                                                                        AS deviceModel,
    any(concat(resourceTagsMap['os.name'], ' ', resourceTagsMap['os.version_major']))                                      AS osNameVersionMajor,
    groupUniqArrayState(stringTagMap['network.connection.type'])                                                           AS networkConnectionTypeState,
    groupUniqArrayState(concat(resourceTagsMap['sophonz.browser.name'], ' ', resourceTagsMap['sophonz.browser.version_major']))  AS browserNameVersionMajorState
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, clientPlatform, serviceName, serviceVersion, sessionID;

------------------------------------------------------------------------------------------------------------------------
-- default: stored_analysis_session
------------------------------------------------------------------------------------------------------------------------
-- analysis_session_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_session_1m ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace              LowCardinality(String),
    clientPlatform                   LowCardinality(String),
    serviceVersion                LowCardinality(String),
    webVersionState               AggregateFunction(groupUniqArray, LowCardinality(String)),
    ts                            DateTime,
    sessionID                     FixedString(32),
    startTimestamp                SimpleAggregateFunction(min, DateTime64(9)),
    endTimestamp                  SimpleAggregateFunction(max, DateTime64(9)),
    userIDState                   AggregateFunction(groupUniqArray, String),
    deviceModel                   LowCardinality(String),
    osNameVersionMajor            LowCardinality(String),
    networkConnectionTypeState    AggregateFunction(groupUniqArray, String),
    browserNameVersionMajorState  AggregateFunction(groupUniqArray, String),
    traceCount                    SimpleAggregateFunction(sum, UInt64),
    hasError                      SimpleAggregateFunction(max, UInt8),
    PROJECTION prjServiceNamespaceSessionIDTS (SELECT * ORDER BY serviceNamespace, sessionID, ts)
)   ENGINE = AggregatingMergeTree PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, sessionID, ts)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS deduplicate_merge_projection_mode = 'rebuild';

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_session_1m
            TO sophonz_traces.stored_analysis_session_1m AS
SELECT
    serviceNamespace,
    any(clientPlatform)                                                                                                       AS clientPlatform,
    any(serviceVersion)                                                                                                    AS serviceVersion,
    groupUniqArrayState(webVersion)                                                                                        AS webVersionState,
    toStartOfMinute(timestamp)                                                                                             AS ts,
    sessionID,
    minSimpleState(timestamp)                                                                                              AS startTimestamp,
    maxSimpleState(endTimestamp)                                                                                           AS endTimestamp,
    groupUniqArrayState(userID)                                                                                            AS userIDState,
    any(resourceTagsMap['device.model.identifier'])                                                                        AS deviceModel,
    any(concat(resourceTagsMap['os.name'], ' ', resourceTagsMap['os.version_major']))                                      AS osNameVersionMajor,
    groupUniqArrayState(stringTagMap['network.connection.type'])                                                           AS networkConnectionTypeState,
    groupUniqArrayState(concat(resourceTagsMap['sophonz.browser.name'], ' ', resourceTagsMap['sophonz.browser.version_major']))  AS browserNameVersionMajorState,
    sumSimpleState(toUInt64(isRootSpan = true))                                                                            AS traceCount,
    maxSimpleState(hasError)                                                                                               AS hasError
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, sessionID, ts;

-- analysis_session_5m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_session_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_session_1m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_session_5m
            TO sophonz_traces.stored_analysis_session_5m AS
SELECT
    serviceNamespace,
    any(clientPlatform)                                        AS clientPlatform,
    any(serviceVersion)                                     AS serviceVersion,
    groupUniqArrayMergeState(webVersionState)               AS webVersionState,
    toStartOfFiveMinute(ts)                                 AS ts,
    sessionID,
    min(startTimestamp)                                     AS startTimestamp,
    max(endTimestamp)                                       AS endTimestamp,
    groupUniqArrayMergeState(userIDState)                   AS userIDState,
    any(deviceModel)                                        AS deviceModel,
    any(osNameVersionMajor)                                 AS osNameVersionMajor,
    groupUniqArrayMergeState(networkConnectionTypeState)    AS networkConnectionTypeState,
    groupUniqArrayMergeState(browserNameVersionMajorState)  AS browserNameVersionMajorState,
    sum(traceCount)                                         AS traceCount,
    max(hasError)                                           AS hasError
FROM
    sophonz_traces.stored_analysis_session_1m
GROUP BY
    serviceNamespace, sessionID, ts;

-- analysis_session_1h
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_session_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_session_5m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_session_1h
            TO sophonz_traces.stored_analysis_session_1h AS
SELECT
    serviceNamespace,
    any(clientPlatform)                                        AS clientPlatform,
    any(serviceVersion)                                     AS serviceVersion,
    groupUniqArrayMergeState(webVersionState)               AS webVersionState,
    toStartOfHour(ts)                                       AS ts,
    sessionID,
    min(startTimestamp)                                     AS startTimestamp,
    max(endTimestamp)                                       AS endTimestamp,
    groupUniqArrayMergeState(userIDState)                   AS userIDState,
    any(deviceModel)                                        AS deviceModel,
    any(osNameVersionMajor)                                 AS osNameVersionMajor,
    groupUniqArrayMergeState(networkConnectionTypeState)    AS networkConnectionTypeState,
    groupUniqArrayMergeState(browserNameVersionMajorState)  AS browserNameVersionMajorState,
    sum(traceCount)                                         AS traceCount,
    max(hasError)                                           AS hasError
FROM
    sophonz_traces.stored_analysis_session_5m
GROUP BY
    serviceNamespace, sessionID, ts;

-- analysis_session_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_session_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_session_1h;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_session_1d
            TO sophonz_traces.stored_analysis_session_1d AS
SELECT
    serviceNamespace,
    any(clientPlatform)                                        AS clientPlatform,
    any(serviceVersion)                                     AS serviceVersion,
    groupUniqArrayMergeState(webVersionState)               AS webVersionState,
    toStartOfDay(ts)                                        AS ts,
    sessionID,
    min(startTimestamp)                                     AS startTimestamp,
    max(endTimestamp)                                       AS endTimestamp,
    groupUniqArrayMergeState(userIDState)                   AS userIDState,
    any(deviceModel)                                        AS deviceModel,
    any(osNameVersionMajor)                                 AS osNameVersionMajor,
    groupUniqArrayMergeState(networkConnectionTypeState)    AS networkConnectionTypeState,
    groupUniqArrayMergeState(browserNameVersionMajorState)  AS browserNameVersionMajorState,
    sum(traceCount)                                         AS traceCount,
    max(hasError)                                           AS hasError
FROM
    sophonz_traces.stored_analysis_session_1h
GROUP BY
    serviceNamespace, sessionID, ts;


------------------------------------------------------------------------------------------------------------------------
-- default: stored_analysis_screen_loading_time
------------------------------------------------------------------------------------------------------------------------
-- analysis_screen_loading_time_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_screen_loading_time_1m ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace      LowCardinality(String),
    clientPlatform           LowCardinality(String),
    serviceVersion        LowCardinality(String),
    ts                    DateTime,
    appScreenType            LowCardinality(String),
    appScreenName            String,
    appScreenGroupName       LowCardinality(String),
    loadingTimeAvgState   AggregateFunction(avgOrNull, UInt64),
    responseTimeAvgState  AggregateFunction(avgOrNull, UInt64),
    visitCount            SimpleAggregateFunction(sum, UInt64),
    errorCount            SimpleAggregateFunction(sum, UInt64),
    PROJECTION prjServiceNamespaceScreenNameTS (SELECT * ORDER BY serviceNamespace, appScreenName, ts),
    PROJECTION prjServiceNamespaceScreenGroupNameTS (SELECT * ORDER BY serviceNamespace, appScreenGroupName, ts)
)   ENGINE = AggregatingMergeTree PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, clientPlatform, serviceVersion, appScreenName, appScreenGroupName, ts, appScreenType)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1, deduplicate_merge_projection_mode = 'rebuild';

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_screen_loading_time_1m
            TO sophonz_traces.stored_analysis_screen_loading_time_1m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    toStartOfMinute(timestamp)                                           AS ts,
    appScreenType,
    appScreenName,
    appScreenGroupName,
    avgOrNullStateIf(durationNano, appSpanType IN ('render', 'post-docs'))  AS loadingTimeAvgState,
    avgOrNullStateIf(durationNano, appSpanType IN ('xhr', 'fetch'))         AS responseTimeAvgState,
    sumSimpleState(toUInt64(appSpanType IN ('render', 'route')))            AS visitCount,
    sumSimpleState(toUInt64(statusCode = 2))                             AS errorCount
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, appScreenName, appScreenGroupName, ts, appScreenType;

-- analysis_screen_loading_time_5m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_screen_loading_time_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_screen_loading_time_1m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_screen_loading_time_5m
            TO sophonz_traces.stored_analysis_screen_loading_time_5m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    toStartOfFiveMinute(ts)                    AS ts,
    appScreenType,
    appScreenName,
    appScreenGroupName,
    avgOrNullMergeState(loadingTimeAvgState)   AS loadingTimeAvgState,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState,
    sum(visitCount)                            AS visitCount,
    sum(errorCount)                            AS errorCount
FROM
    sophonz_traces.stored_analysis_screen_loading_time_1m
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, appScreenName, appScreenGroupName, ts, appScreenType;

-- analysis_screen_loading_time_1h
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_screen_loading_time_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_screen_loading_time_5m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_screen_loading_time_1h
            TO sophonz_traces.stored_analysis_screen_loading_time_1h AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    toStartOfHour(ts)                          AS ts,
    appScreenType,
    appScreenName,
    appScreenGroupName,
    avgOrNullMergeState(loadingTimeAvgState)   AS loadingTimeAvgState,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState,
    sum(visitCount)                            AS visitCount,
    sum(errorCount)                            AS errorCount
FROM
    sophonz_traces.stored_analysis_screen_loading_time_5m
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, appScreenName, appScreenGroupName, ts, appScreenType;

-- analysis_screen_loading_time_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_screen_loading_time_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_screen_loading_time_1h;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_screen_loading_time_1d
            TO sophonz_traces.stored_analysis_screen_loading_time_1d AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    toStartOfDay(ts)                           AS ts,
    appScreenType,
    appScreenName,
    appScreenGroupName,
    avgOrNullMergeState(loadingTimeAvgState)   AS loadingTimeAvgState,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState,
    sum(visitCount)                            AS visitCount,
    sum(errorCount)                            AS errorCount
FROM
    sophonz_traces.stored_analysis_screen_loading_time_1h
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, appScreenName, appScreenGroupName, ts, appScreenType;


------------------------------------------------------------------------------------------------------------------------
-- default: stored_analysis_http_request_response_time
------------------------------------------------------------------------------------------------------------------------
-- analysis_http_request_response_time_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_http_request_response_time_1m ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace      LowCardinality(String),
    clientPlatform           LowCardinality(String),
    serviceVersion        LowCardinality(String),
    webVersion            LowCardinality(String),
    ts                    DateTime,
    httpURL               String,
    httpHost              LowCardinality(String),
    httpMethod            LowCardinality(String),
    requestCount          SimpleAggregateFunction(sum, UInt64),
    responseTimeAvgState  AggregateFunction(avgOrNull, UInt64),
    PROJECTION prjServiceNamespaceTShttpURL (SELECT * ORDER BY serviceNamespace, ts, httpURL)
)   ENGINE = AggregatingMergeTree PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, clientPlatform, serviceVersion, webVersion, httpURL, httpMethod, ts, httpHost)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1, deduplicate_merge_projection_mode = 'rebuild';

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_http_request_response_time_1m
            TO sophonz_traces.stored_analysis_http_request_response_time_1m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfMinute(timestamp)    AS ts,
    httpURL,
    httpHost,
    httpMethod,
    sumSimpleState(1)             AS requestCount,
    avgOrNullState(durationNano)  AS responseTimeAvgState
FROM
    sophonz_traces.distributed_sophonz_index_v2
WHERE
    appSpanType IN ('xhr', 'fetch')
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, httpURL, httpMethod, ts, httpHost;

-- analysis_http_request_response_time_5m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_http_request_response_time_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_http_request_response_time_1m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_http_request_response_time_5m
            TO sophonz_traces.stored_analysis_http_request_response_time_5m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfHour(ts)                          AS ts,
    httpURL,
    httpHost,
    httpMethod,
    sum(requestCount)                          AS requestCount,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState
FROM
    sophonz_traces.stored_analysis_http_request_response_time_1m
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, httpURL, httpMethod, ts, httpHost;

-- analysis_http_request_response_time_1h
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_http_request_response_time_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_http_request_response_time_5m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_http_request_response_time_1h
            TO sophonz_traces.stored_analysis_http_request_response_time_1h AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfHour(ts)                          AS ts,
    httpURL,
    httpHost,
    httpMethod,
    sum(requestCount)                          AS requestCount,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState
FROM
    sophonz_traces.stored_analysis_http_request_response_time_5m
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, httpURL, httpMethod, ts, httpHost;

-- analysis_http_request_response_time_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_http_request_response_time_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_http_request_response_time_1h;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_http_request_response_time_1d
            TO sophonz_traces.stored_analysis_http_request_response_time_1d AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfDay(ts)                           AS ts,
    httpURL,
    httpHost,
    httpMethod,
    sum(requestCount)                          AS requestCount,
    avgOrNullMergeState(responseTimeAvgState)  AS responseTimeAvgState
FROM
    sophonz_traces.stored_analysis_http_request_response_time_1h
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, httpURL, httpMethod, ts, httpHost;


------------------------------------------------------------------------------------------------------------------------
-- default: stored_analysis_user_resource_summary
------------------------------------------------------------------------------------------------------------------------
-- analysis_user_resource_summary_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_user_resource_summary_1m ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace         LowCardinality(String),
    clientPlatform              LowCardinality(String),
    serviceVersion           LowCardinality(String),
    webVersion               LowCardinality(String),
    ts                       Datetime,
    userIDUniqState          AggregateFunction(uniq, String),
    userIDUniqHLL12State     AggregateFunction(uniqHLL12, String),
    userIDUniqCombinedState  AggregateFunction(uniqCombined, String),
    deviceModel              LowCardinality(String),
    osName                   LowCardinality(String),
    osNameVersionMajor       LowCardinality(String),
    networkConnectionType    LowCardinality(String)
)   ENGINE = AggregatingMergeTree PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, clientPlatform, serviceVersion, webVersion, networkConnectionType, osName, osNameVersionMajor, deviceModel, ts)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_user_resource_summary_1m
            TO sophonz_traces.stored_analysis_user_resource_summary_1m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfMinute(timestamp)                                                    AS ts,
    uniqState(userID)                                                             AS userIDUniqState,
    uniqHLL12State(userID)                                                        AS userIDUniqHLL12State,
    uniqCombinedState(userID)                                                     AS userIDUniqCombinedState,
    resourceTagsMap['device.model.identifier']                                    AS deviceModel,
    resourceTagsMap['os.name']                                                    AS osName,
    concat(resourceTagsMap['os.name'], ' ', resourceTagsMap['os.version_major'])  AS osNameVersionMajor,
    stringTagMap['network.connection.type']                                       AS networkConnectionType
FROM
    sophonz_traces.distributed_sophonz_index_v2
WHERE
    userID != ''
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, ts, networkConnectionType, osName, osNameVersionMajor, deviceModel;

-- analysis_user_resource_summary_5m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_user_resource_summary_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_user_resource_summary_1m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_user_resource_summary_5m
            TO sophonz_traces.stored_analysis_user_resource_summary_5m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfFiveMinute(ts)                          AS ts,
    uniqMergeState(userIDUniqState)                  AS userIDUniqState,
    uniqHLL12MergeState(userIDUniqHLL12State)        AS userIDUniqHLL12State,
    uniqCombinedMergeState(userIDUniqCombinedState)  AS userIDUniqCombinedState,
    deviceModel,
    osName,
    osNameVersionMajor,
    networkConnectionType
FROM
    sophonz_traces.stored_analysis_user_resource_summary_1m
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, ts, networkConnectionType, osName, osNameVersionMajor, deviceModel;

-- analysis_user_resource_summary_1h
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_user_resource_summary_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_user_resource_summary_5m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_user_resource_summary_1h
            TO sophonz_traces.stored_analysis_user_resource_summary_1h AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfHour(ts)                                AS ts,
    uniqMergeState(userIDUniqState)                  AS userIDUniqState,
    uniqHLL12MergeState(userIDUniqHLL12State)        AS userIDUniqHLL12State,
    uniqCombinedMergeState(userIDUniqCombinedState)  AS userIDUniqCombinedState,
    deviceModel,
    osName,
    osNameVersionMajor,
    networkConnectionType
FROM
    sophonz_traces.stored_analysis_user_resource_summary_5m
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, ts, networkConnectionType, osName, osNameVersionMajor, deviceModel;

-- analysis_user_resource_summary_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_user_resource_summary_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
    AS sophonz_traces.stored_analysis_user_resource_summary_1h;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_user_resource_summary_1d
            TO sophonz_traces.stored_analysis_user_resource_summary_1d AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfFiveMinute(ts)                          AS ts,
    uniqMergeState(userIDUniqState)                  AS userIDUniqState,
    uniqHLL12MergeState(userIDUniqHLL12State)        AS userIDUniqHLL12State,
    uniqCombinedMergeState(userIDUniqCombinedState)  AS userIDUniqCombinedState,
    deviceModel,
    osName,
    osNameVersionMajor,
    networkConnectionType
FROM
    sophonz_traces.stored_analysis_user_resource_summary_1h
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, networkConnectionType, osName, osNameVersionMajor, deviceModel, ts;


------------------------------------------------------------------------------------------------------------------------
-- default: stored_analysis_user_event
------------------------------------------------------------------------------------------------------------------------
-- analysis_user_event_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_user_event_1m ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace       LowCardinality(String),
    clientPlatform            LowCardinality(String),
    serviceVersion         LowCardinality(String),
    webVersion             LowCardinality(String),
    ts                     DateTime,
    appScreenName             String,
    eventName              String,
    eventCount             SimpleAggregateFunction(sum, UInt64),
    eventDurationAvgState  AggregateFunction(avgOrNull, UInt64)
)   ENGINE = AggregatingMergeTree PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, clientPlatform, serviceVersion, webVersion, appScreenName, eventName, ts)
    TTL ts + toIntervalSecond(2592000);

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_user_event_1m
            TO sophonz_traces.stored_analysis_user_event_1m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfMinute(timestamp)    AS ts,
    appScreenName,
    name                          AS eventName,
    sumSimpleState(1)             AS eventCount,
    avgOrNullState(durationNano)  AS eventDurationAvgState
FROM
    sophonz_traces.distributed_sophonz_index_v2
WHERE
    appSpanType = 'user_event' AND isRootSpan = true
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, appScreenName, eventName, ts;

-- analysis_user_event_5m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_user_event_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_analysis_user_event_1m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_user_event_5m
            TO sophonz_traces.stored_analysis_user_event_5m AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfFiveMinute(ts)                     AS ts,
    appScreenName,
    eventName,
    sum(eventCount)                             AS eventCount,
    avgOrNullMergeState(eventDurationAvgState)  AS eventDurationAvgState
FROM
    sophonz_traces.stored_analysis_user_event_1m
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, appScreenName, eventName, ts;

-- analysis_user_event_1h
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_user_event_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_analysis_user_event_5m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_user_event_1h
            TO sophonz_traces.stored_analysis_user_event_1h AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfHour(ts)                           AS ts,
    appScreenName,
    eventName,
    sum(eventCount)                             AS eventCount,
    avgOrNullMergeState(eventDurationAvgState)  AS eventDurationAvgState
FROM
    sophonz_traces.stored_analysis_user_event_5m
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, appScreenName, eventName, ts;

-- analysis_user_event_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_analysis_user_event_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_analysis_user_event_1h;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_analysis_user_event_1d
            TO sophonz_traces.stored_analysis_user_event_1d AS
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfDay(ts)                            AS ts,
    appScreenName,
    eventName,
    sum(eventCount)                             AS eventCount,
    avgOrNullMergeState(eventDurationAvgState)  AS eventDurationAvgState
FROM
    sophonz_traces.stored_analysis_user_event_1h
GROUP BY
    serviceNamespace, clientPlatform, serviceVersion, webVersion, appScreenName, eventName, ts;


------------------------------------------------------------------------------------------------------------------------
-- default: stored_dashboard_service_metrics_duration_histogram
------------------------------------------------------------------------------------------------------------------------
-- dashboard_service_metrics_duration_histogram_1m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_duration_histogram_1m ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    serviceNamespace  LowCardinality(String),
    clientPlatform       LowCardinality(String),
    ts                DateTime,
    metricType        LowCardinality(String),
    appSpanType          LowCardinality(String),
    duration100msBucket   UInt8,
    count             SimpleAggregateFunction(sum, UInt64)
)   ENGINE = AggregatingMergeTree
    PARTITION BY toDate(ts)
    ORDER BY (serviceNamespace, clientPlatform, ts, metricType, appSpanType, duration100msBucket)
    TTL ts + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_duration_histogram_1m
            TO sophonz_traces.stored_dashboard_service_metrics_duration_histogram_1m AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfMinute(timestamp) AS ts,
    multiIf(
            appSpanType IN ('xhr', 'fetch'), 'request',
            appSpanType IN ('render', 'post-docs'), 'loading',
            'other'
    ) AS metricType,
    appSpanType,
    CASE
        WHEN durationNano < 5000000000 THEN CAST(durationNano / 100000000 AS UInt8)
        ELSE 50
        END AS duration100msBucket,
    sumSimpleState(1) AS count
FROM sophonz_traces.distributed_sophonz_index_v2
WHERE
    appSpanType IN ('xhr', 'fetch', 'render', 'post-docs')
GROUP BY
    serviceNamespace, clientPlatform, ts, metricType, appSpanType, duration100msBucket;

-- dashboard_service_metrics_duration_histogram_5m
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_duration_histogram_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_service_metrics_duration_histogram_1m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_duration_histogram_5m
            TO sophonz_traces.stored_dashboard_service_metrics_duration_histogram_5m AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfFiveMinute(ts)  AS ts,
    metricType,
    appSpanType,
    duration100msBucket,
    sum(count)               AS count
FROM
    sophonz_traces.stored_dashboard_service_metrics_duration_histogram_1m
GROUP BY
    serviceNamespace, clientPlatform, ts, metricType, appSpanType, duration100msBucket;

-- dashboard_service_metrics_duration_histogram_1h
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_duration_histogram_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_service_metrics_duration_histogram_5m;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_duration_histogram_1h
            TO sophonz_traces.stored_dashboard_service_metrics_duration_histogram_1h AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfHour(ts)    AS ts,
    metricType,
    appSpanType,
    duration100msBucket,
    sum(count)           AS count
FROM
    sophonz_traces.stored_dashboard_service_metrics_duration_histogram_5m
GROUP BY
    serviceNamespace, clientPlatform, ts, metricType, appSpanType, duration100msBucket;

-- dashboard_service_metrics_duration_histogram_1d
CREATE TABLE IF NOT EXISTS sophonz_traces.stored_dashboard_service_metrics_duration_histogram_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS sophonz_traces.stored_dashboard_service_metrics_duration_histogram_1h;

CREATE MATERIALIZED VIEW IF NOT EXISTS sophonz_traces.mv_dashboard_service_metrics_duration_histogram_1d
            TO sophonz_traces.stored_dashboard_service_metrics_duration_histogram_1d AS
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfDay(ts)     AS ts,
    metricType,
    appSpanType,
    duration100msBucket,
    sum(count)           AS count
FROM
    sophonz_traces.stored_dashboard_service_metrics_duration_histogram_1h
GROUP BY
    serviceNamespace, clientPlatform, ts, metricType, appSpanType, duration100msBucket;
