-- Re-map v_postgres_service onto the current Prisma schema and expose the
-- owning organization, which is the tenant the collector stamps on telemetry.
--
-- The version number is reused rather than bumped because this migration was
-- never applied anywhere: sophonz_metadata.schema_migrations is still at
-- version 1. Its previous body targeted a schema (Service.publicId,
-- Service.serviceNamespaceId, ServiceNamespace) that no longer exists, so
-- rewriting it in place is the only way it can ever run.
--
-- Postgres identifiers here are camelCase and therefore case-sensitive, and the
-- postgresql() table function surfaces them with their exact casing, and
-- ClickHouse identifiers are case-sensitive too, so they must be spelled
-- exactly as Prisma created them.

DROP VIEW IF EXISTS sophonz_metadata.v_postgres_service ON CLUSTER {{.SOPHONZ_CLUSTER}};

CREATE VIEW IF NOT EXISTS sophonz_metadata.v_postgres_service ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS
SELECT
    service.id AS id,
    project.namespace AS serviceNamespace,
    service.name AS serviceName,
    -- "alias" has no direct counterpart in the current schema — packageName is
    -- the closest analogue (the app bundle/package identifier).
    ifNull(service.packageName, '') AS alias,
    -- ifNull keeps the view header non-Nullable: pkg/sophonz/metadata scans
    -- these into plain Go strings, which fails on a Nullable(String) column.
    ifNull(service.platform, '') AS type,
    ifNull(service.serviceKey, '') AS key,
    -- The tenant. Organization is reached through Project, so services of the
    -- same project always resolve to the same tenant.
    project.organizationId AS companyId
FROM
    postgresql(postgres_creds, table='Service') AS service
        INNER JOIN postgresql(postgres_creds, table='Project') AS project
                   ON service.projectId = project.id;
