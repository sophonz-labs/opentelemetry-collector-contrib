DROP VIEW IF EXISTS sophonz_metadata.v_postgres_service ON CLUSTER {{.SOPHONZ_CLUSTER}};

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
