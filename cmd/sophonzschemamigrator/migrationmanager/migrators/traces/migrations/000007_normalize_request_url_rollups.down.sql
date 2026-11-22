-- Restore the rollups to grouping by the raw URL, query string included.

ALTER TABLE sophonz_traces.mv_dashboard_common_metrics_slowest_request_5s MODIFY QUERY
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfInterval(timestamp, toIntervalSecond(5)) AS ts,
    httpURL AS requestURL,
    sumSimpleState(1) AS requestCount,
    avgOrNullState(durationNano) AS responseTimeAvgState
FROM sophonz_traces.distributed_sophonz_index_v2
WHERE appSpanType IN ('xhr', 'fetch')
GROUP BY
    serviceNamespace,
    clientPlatform,
    requestURL,
    ts;

ALTER TABLE sophonz_traces.mv_analysis_http_request_response_time_1m MODIFY QUERY
SELECT
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    toStartOfMinute(timestamp) AS ts,
    httpURL AS httpURL,
    httpHost,
    httpMethod,
    sumSimpleState(1) AS requestCount,
    avgOrNullState(durationNano) AS responseTimeAvgState
FROM sophonz_traces.distributed_sophonz_index_v2
WHERE appSpanType IN ('xhr', 'fetch')
GROUP BY
    serviceNamespace,
    clientPlatform,
    serviceVersion,
    webVersion,
    httpURL,
    httpMethod,
    ts,
    httpHost;
