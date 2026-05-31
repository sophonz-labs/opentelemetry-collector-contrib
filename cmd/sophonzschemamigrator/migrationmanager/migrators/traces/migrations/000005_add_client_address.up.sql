ALTER TABLE sophonz_traces.stored_analysis_session ON CLUSTER {{.SOPHONZ_CLUSTER}}
    ADD COLUMN IF NOT EXISTS clientAddressState AggregateFunction(groupUniqArray, String);

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
    groupUniqArrayState(stringTagMap['client.address'])                                                                   AS clientAddressState
FROM
    sophonz_traces.distributed_sophonz_index_v2
GROUP BY
    serviceNamespace, clientPlatform, serviceName, serviceVersion, sessionID;
