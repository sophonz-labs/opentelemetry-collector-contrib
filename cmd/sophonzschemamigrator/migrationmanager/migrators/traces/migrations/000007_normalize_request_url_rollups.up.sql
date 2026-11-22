-- Group the request rollups by the URL without its query string.
--
-- `httpURL` comes from `url.full` on client spans, which is the URL actually
-- requested. Query strings make that unbounded: a Next.js app measured here
-- produced 799 distinct URLs out of 815 from the `_rsc` prefetch token alone,
-- against 58 once the query string was removed. Both of these views key a
-- per-interval aggregate on it, so an unbounded key means one output row per
-- input span — an AggregatingMergeTree that aggregates nothing, with the key
-- sitting in the sort order and in a projection.
--
-- The browser SDK drops known cache busters at source (@sophonz/redaction), but
-- that list can only ever cover the names someone thought of. Cutting the query
-- string here bounds the key by construction, whatever the SDK sends.
--
-- Only the two base views need it. The 1m/5m/1h/1d rollups read from the
-- previous stored table, so they inherit an already-normalized URL.
--
-- Individual spans keep their full URL — this affects the rollups only.

ALTER TABLE sophonz_traces.mv_dashboard_common_metrics_slowest_request_5s MODIFY QUERY
SELECT
    serviceNamespace,
    clientPlatform,
    toStartOfInterval(timestamp, toIntervalSecond(5)) AS ts,
    cutQueryStringAndFragment(httpURL) AS requestURL,
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
    cutQueryStringAndFragment(httpURL) AS httpURL,
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
