DROP VIEW IF EXISTS sophonz_metadata.v_postgres_alarm ON CLUSTER {{.SOPHONZ_CLUSTER}};

CREATE VIEW IF NOT EXISTS sophonz_metadata.v_postgres_alarm ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS
SELECT
    alarm.id AS id,
    alarm.createdAt AS createdAt,
    alarm.updatedAt AS updatedAt,
    alarm.name AS name,
    alarm.description AS description,
    alarm.enabled AS enabled,
    alarm.state AS state,
    alarm.serviceNamespaceId AS serviceNamespaceID,
    serviceNamespace.name AS serviceNamespace,
    alarm.appSelectionType AS appSelectionType,
    alarm.selectedApps AS selectedApps,
    alarm.severity AS severity,
    alarm.evaluationPeriod AS evaluationPeriod,
    alarm.screenSelectionType AS screenSelectionType,
    alarm.selectedScreenNames AS selectedScreenNames,
    alarm.selectedScreenGroupIds AS selectedScreenGroupIDs,
    alarm.metric AS metric,
    alarm.metricUnit AS metricUnit,
    alarm.threshold AS threshold,
    alarm.thresholdUnit AS thresholdUnit,
    alarm.thresholdOperator AS thresholdOperator,
    alarm.evaluationInterval AS evaluationInterval,
    alarm.receivers AS receivers,
    alarm.webhookEnabled AS webhookEnabled,
    alarm.webhookUrl AS webhookURL,
    alarm.webhookBody AS webhookBody,
    alarm.webhookContent AS webhookContent,
    alarm.maxNotificationCountPerDay AS maxNotificationCountPerDay,
    alarm.maxNotificationCountPerDayResetTime AS maxNotificationCountPerDayResetTime
FROM postgresql(postgres_creds, `table` = 'Alarm') AS alarm
         INNER JOIN postgresql(postgres_creds, table='ServiceNamespace') AS serviceNamespace
                    ON alarm.serviceNamespaceId = serviceNamespace.id
         LEFT OUTER JOIN postgresql(postgres_creds, table='Service') AS service
                         ON alarm.serviceId = service.id;


CREATE TABLE IF NOT EXISTS sophonz_metadata.alarm_log ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    timestamp DateTime64 CODEC(DoubleDelta, ZSTD(1)),
    alarmID Int32 CODEC(ZSTD(1)),
    alarmName LowCardinality(String) CODEC(ZSTD(1)),
    serviceNamespace LowCardinality(String) CODEC(ZSTD(1)),
    severity LowCardinality(String) CODEC(ZSTD(1)),
    channel LowCardinality(String) CODEC(ZSTD(1)),
    requestResult LowCardinality(String) CODEC(ZSTD(1)),
    requestMessage String CODEC(ZSTD(1)),
    responseResult LowCardinality(String) CODEC(ZSTD(1)),
    responseMessage String CODEC(ZSTD(1))
)   ENGINE = {{.SOPHONZ_REPLICATED}}MergeTree PARTITION BY toDate(timestamp)
    ORDER BY (serviceNamespace, timestamp)
    TTL toDateTime(timestamp) + toIntervalSecond(2592000)
    SETTINGS ttl_only_drop_parts = 1;


CREATE TABLE IF NOT EXISTS sophonz_metadata.t_postgres_alarm_change_log
(
    alarmId     Int32,
    state       String,
    createdAt   DateTime,
    isTriggered UInt8,
    stateReason String
)   ENGINE PostgreSQL(postgres_creds, `table` = 'AlarmChangeLog');
