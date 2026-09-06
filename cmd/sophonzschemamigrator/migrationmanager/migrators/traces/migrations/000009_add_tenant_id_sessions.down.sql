-- Restore every view to the body it had before 000009, and only then take the
-- columns away. A view left emitting a column its target no longer has fails the
-- insert, and that failure propagates back into the span batch, so the views go
-- first and the tables second. The bodies below are the ones 000006 and 000002
-- installed, not stubs, so a rollback leaves the pipeline writing exactly what it
-- wrote before.

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
    max(hasError)                                           AS hasError
FROM
    sophonz_traces.stored_analysis_session_1h
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
    max(hasError)                                           AS hasError
FROM
    sophonz_traces.stored_analysis_session_5m
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
    max(hasError)                                           AS hasError
FROM
    sophonz_traces.stored_analysis_session_1m
GROUP BY
    serviceNamespace, sessionID, ts;

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
    maxSimpleState(hasError)                                                                                               AS hasError
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, sessionID, ts;

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
    groupUniqArrayState(stringTagMap['client.address'])                                                                    AS clientAddressState
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, clientPlatform, serviceName, serviceVersion, sessionID;

ALTER TABLE sophonz_traces.stored_analysis_session_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF EXISTS idx_tenantID;
ALTER TABLE sophonz_traces.stored_analysis_session_1d ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP COLUMN IF EXISTS tenantID;

ALTER TABLE sophonz_traces.stored_analysis_session_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF EXISTS idx_tenantID;
ALTER TABLE sophonz_traces.stored_analysis_session_1h ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP COLUMN IF EXISTS tenantID;

ALTER TABLE sophonz_traces.stored_analysis_session_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF EXISTS idx_tenantID;
ALTER TABLE sophonz_traces.stored_analysis_session_5m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP COLUMN IF EXISTS tenantID;

ALTER TABLE sophonz_traces.stored_analysis_session_1m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF EXISTS idx_tenantID;
ALTER TABLE sophonz_traces.stored_analysis_session_1m ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP COLUMN IF EXISTS tenantID;

ALTER TABLE sophonz_traces.stored_analysis_session ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF EXISTS idx_tenantID;
ALTER TABLE sophonz_traces.stored_analysis_session ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP COLUMN IF EXISTS tenantID;
