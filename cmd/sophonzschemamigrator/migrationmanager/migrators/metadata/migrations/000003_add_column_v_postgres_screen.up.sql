DROP VIEW IF EXISTS sophonz_metadata.v_postgres_screen ON CLUSTER {{.SOPHONZ_CLUSTER}};

CREATE VIEW IF NOT EXISTS sophonz_metadata.v_postgres_screen ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS
SELECT
    screen.id AS id,
    screen.screenName,
    screen.screenType,
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
