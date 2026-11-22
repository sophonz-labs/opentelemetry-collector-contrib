-- tenantID carries the organization the telemetry belongs to, resolved from the
-- SDK app key by the sophonz attribute processor. The session views aggregate
-- distributed_sophonz_index_v2, which has carried tenantID since 000008, so the
-- value is read straight off the source span. any() is enough because tenantID
-- is a property of serviceNamespace, which is already in every GROUP BY here,
-- so it is constant across the rows folded into one session. The 5m, 1h and 1d
-- views read it from the stored table one step below them for the same reason.
--
-- Sessions whose source spans predate 000008 have no tenantID to read and
-- aggregate to the empty string, the LowCardinality(String) default. That is the
-- same unscoped marker sophonz_index_v2 and logs_v2 already carry for rows
-- ingested before enforcement, so a reader can apply one rule to all three.
--
-- The column is appended at the very END of every table, with no AFTER clause,
-- and no ORDER BY is touched. MODIFY ORDER BY can only extend a sort key, never
-- prepend to it, so making tenant the leading sort column needs a full table
-- rebuild that stays deferred, exactly as it was for the index tables in 000008.
--
-- These rollup tables have no distributed_ counterpart. Only the two index
-- tables are sharded, and the views read the distributed index, so there is no
-- second copy of the schema to keep in positional agreement.
--
-- Every ADD COLUMN runs before every MODIFY QUERY. A view that emits a column
-- its target table does not have yet fails the insert into the target, and that
-- failure propagates back into the INSERT on the index table and drops the span
-- batch. In this order the worst case inside the migration window is the
-- opposite and much cheaper one: the table already has the column while the view
-- is still on its old body, so those sessions land with an empty tenantID and
-- are corrected the next time the same session key is written. The .down.sql
-- reverses the order for the same reason.

ALTER TABLE sophonz_traces.stored_analysis_session ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS tenantID LowCardinality(String);
ALTER TABLE sophonz_traces.stored_analysis_session ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_tenantID tenantID TYPE set(0) GRANULARITY 4;

ALTER TABLE sophonz_traces.stored_analysis_session_1m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS tenantID LowCardinality(String);
ALTER TABLE sophonz_traces.stored_analysis_session_1m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_tenantID tenantID TYPE set(0) GRANULARITY 4;

ALTER TABLE sophonz_traces.stored_analysis_session_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS tenantID LowCardinality(String);
ALTER TABLE sophonz_traces.stored_analysis_session_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_tenantID tenantID TYPE set(0) GRANULARITY 4;

ALTER TABLE sophonz_traces.stored_analysis_session_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS tenantID LowCardinality(String);
ALTER TABLE sophonz_traces.stored_analysis_session_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_tenantID tenantID TYPE set(0) GRANULARITY 4;

ALTER TABLE sophonz_traces.stored_analysis_session_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS tenantID LowCardinality(String);
ALTER TABLE sophonz_traces.stored_analysis_session_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD INDEX IF NOT EXISTS idx_tenantID tenantID TYPE set(0) GRANULARITY 4;

ALTER TABLE sophonz_traces.mv_analysis_session MODIFY QUERY
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
    groupUniqArrayState(concat(resourceTagsMap['sophonz.browser.name'], ' ', resourceTagsMap['sophonz.browser.version_major']))  AS browserNameVersionMajorState,
    any(deviceID)                                                                                                          AS deviceID,
    groupUniqArrayState(stringTagMap['client.address'])                                                                    AS clientAddressState,
    any(tenantID)                                                                                                          AS tenantID
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, clientPlatform, serviceName, serviceVersion, sessionID;

ALTER TABLE sophonz_traces.mv_analysis_session_1m MODIFY QUERY
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
    maxSimpleState(hasError)                                                                                               AS hasError,
    any(tenantID)                                                                                                          AS tenantID
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, sessionID, ts;

ALTER TABLE sophonz_traces.mv_analysis_session_5m MODIFY QUERY
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
    max(hasError)                                           AS hasError,
    any(tenantID)                                           AS tenantID
FROM
    sophonz_traces.stored_analysis_session_1m
GROUP BY
    serviceNamespace, sessionID, ts;

ALTER TABLE sophonz_traces.mv_analysis_session_1h MODIFY QUERY
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
    max(hasError)                                           AS hasError,
    any(tenantID)                                           AS tenantID
FROM
    sophonz_traces.stored_analysis_session_5m
GROUP BY
    serviceNamespace, sessionID, ts;

ALTER TABLE sophonz_traces.mv_analysis_session_1d MODIFY QUERY
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
    max(hasError)                                           AS hasError,
    any(tenantID)                                           AS tenantID
FROM
    sophonz_traces.stored_analysis_session_1h
GROUP BY
    serviceNamespace, sessionID, ts;
