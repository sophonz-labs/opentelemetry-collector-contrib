CREATE TABLE IF NOT EXISTS sophonz_metadata.schema_ttl_config ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    database_name LowCardinality(String),
    table_name LowCardinality(String),
    column_name LowCardinality(String),
    ttl UInt32,
    unit enum('NOT SET', 'SECOND', 'MINUTE', 'HOUR', 'DAY', 'WEEK', 'MONTH', 'QUARTER', 'YEAR')
) ENGINE = MergeTree ORDER BY (database_name, table_name);

CREATE VIEW IF NOT EXISTS sophonz_metadata.v_postgres_screen ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS
SELECT
    screen.id AS id,
    screen.screenName,
    screen.screenGroupId AS groupID,
    screenGroup.groupName,
    serviceNamespace.name AS serviceNamespace,
    service.name AS serviceName,
    serviceVersion.version AS serviceVersion,
    service.type AS serviceType
FROM
    postgresql(postgres_creds, table='Screen') AS screen
        INNER JOIN postgresql(postgres_creds, table='ScreenGroup') AS screenGroup
                   ON screen.screenGroupId = screenGroup.id
        INNER JOIN postgresql(postgres_creds, table='Service') AS service
                   ON screen.serviceId = service.id
        INNER JOIN postgresql(postgres_creds, table='ServiceVersion') AS serviceVersion
                   ON screen.serviceVersionId = serviceVersion.id
        INNER JOIN postgresql(postgres_creds, table='ServiceNamespace') AS serviceNamespace
                   ON service.serviceNamespaceId = serviceNamespace.id;

CREATE VIEW IF NOT EXISTS sophonz_metadata.v_postgres_service ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS
SELECT
    service.publicId AS id,
    serviceNamespace.name AS serviceNamespace,
    service.name AS serviceName
FROM
    postgresql(postgres_creds, table='Service') AS service
        INNER JOIN postgresql(postgres_creds, table='ServiceNamespace') AS serviceNamespace
                   ON service.serviceNamespaceId = serviceNamespace.id;

CREATE VIEW IF NOT EXISTS sophonz_metadata.v_postgres_alarm ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS
SELECT
    *
FROM
    postgresql(postgres_creds, table='Alarm') AS alarm