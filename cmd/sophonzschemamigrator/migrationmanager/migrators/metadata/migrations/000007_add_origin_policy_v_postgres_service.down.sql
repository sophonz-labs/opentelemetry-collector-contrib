-- Restore the 000002 definition: the same view without the origin policy
-- columns. Unlike the 000002 rollback this one can actually recreate what it
-- replaced, because 000002 targets the schema that is still current.

DROP VIEW IF EXISTS sophonz_metadata.v_postgres_service ON CLUSTER {{.SOPHONZ_CLUSTER}};

CREATE VIEW IF NOT EXISTS sophonz_metadata.v_postgres_service ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS
SELECT
    service.id AS id,
    project.namespace AS serviceNamespace,
    service.name AS serviceName,
    ifNull(service.packageName, '') AS alias,
    ifNull(service.platform, '') AS type,
    ifNull(service.serviceKey, '') AS key,
    project.organizationId AS companyId
FROM
    postgresql(postgres_creds, table='Service') AS service
        INNER JOIN postgresql(postgres_creds, table='Project') AS project
                   ON service.projectId = project.id;
