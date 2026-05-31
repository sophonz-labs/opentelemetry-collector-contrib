ALTER TABLE sophonz_traces.sophonz_index_v2 ADD COLUMN IF NOT EXISTS deviceID FixedString(32) CODEC(ZSTD(1)) AFTER sessionID;
ALTER TABLE sophonz_traces.distributed_sophonz_index_v2 ADD COLUMN IF NOT EXISTS deviceID FixedString(32) CODEC(ZSTD(1)) AFTER sessionID;

ALTER TABLE sophonz_traces.sophonz_error_index_v2 ADD COLUMN IF NOT EXISTS deviceID FixedString(32) CODEC(ZSTD(1)) AFTER sessionID;
ALTER TABLE sophonz_traces.distributed_sophonz_error_index_v2 ADD COLUMN IF NOT EXISTS deviceID FixedString(32) CODEC(ZSTD(1)) AFTER sessionID;

ALTER TABLE sophonz_traces.stored_analysis_session ADD COLUMN IF NOT EXISTS deviceID FixedString(32),
    MODIFY ORDER BY (serviceNamespace, clientPlatform, serviceName, serviceVersion, ts, sessionID, deviceID);

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
    any(concat(resourceTagsMap['os.name'], ' ', resourceTagsMap['os.version_major']))                                      AS osNameVersion,
    groupUniqArrayState(stringTagMap['network.connection.type'])                                                           AS networkConnectionTypeState,
    groupUniqArrayState(concat(resourceTagsMap['sophonz.browser.name'], ' ', resourceTagsMap['sophonz.browser.version_major']))  AS browserNameVersionMajorState,
    any(deviceID)                                AS deviceID
FROM sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, clientPlatform, serviceName, serviceVersion, sessionID;
