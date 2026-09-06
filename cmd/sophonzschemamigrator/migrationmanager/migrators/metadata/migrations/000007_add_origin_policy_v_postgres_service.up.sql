-- Expose the per-app browser origin policy on v_postgres_service so the
-- collector can decide whether a browser Origin is allowed to report under a
-- given app key. Two columns are added to Postgres' "Service" by the app:
--   "allowedOrigins" text[]    -- empty means allow every origin
--   "originEnforced" boolean   -- false means count violations, never reject
--
-- A NEW version rather than an edit of 000002. 000002 could be rewritten in
-- place only because sophonz_metadata.schema_migrations was still at version 1
-- and it had therefore never run. The chain is at 6 and clean now, so
-- golang-migrate would never replay 000002 and an in-place edit would be a
-- silent no-op on every deployed cluster. The view is recreated in full because
-- ClickHouse has no ALTER VIEW that adds a column to the underlying SELECT.
--
-- The allowlist is flattened to a comma-separated String here instead of being
-- carried through as an Array. Postgres' array_to_string() cannot be used for
-- this: the postgresql() table function in ClickHouse 25.8 takes a remote table
-- NAME only, so nothing in this file runs on the Postgres side. arrayMap+ifNull
-- makes the concatenation independent of whether type inference produced
-- Array(String) or Array(Nullable(String)), and an empty array concatenates to
-- '' rather than NULL, which is what the "empty allowlist means allow
-- everything" default depends on.
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
    project.organizationId AS companyId,
    -- Comma-separated, never NULL. The collector splits it once per cache
    -- refresh, not per request.
    arrayStringConcat(arrayMap(origin -> ifNull(origin, ''), service.allowedOrigins), ',') AS allowedOrigins,
    -- toUInt8 is a no-op on the boolean mapping and pins the column type so the
    -- Go driver can scan it into a bool whatever inference decided.
    toUInt8(service.originEnforced) AS originEnforced
FROM
    postgresql(postgres_creds, table='Service') AS service
        INNER JOIN postgresql(postgres_creds, table='Project') AS project
                   ON service.projectId = project.id;
