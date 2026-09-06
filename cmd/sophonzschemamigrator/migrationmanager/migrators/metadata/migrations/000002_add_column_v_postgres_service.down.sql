-- Only a DROP: the 000001 definition of this view cannot be recreated any
-- more. It reads Service.publicId / Service.serviceNamespaceId and the
-- ServiceNamespace table, none of which exist in the current Prisma schema, and
-- ClickHouse resolves a postgresql() view's column types at CREATE time — so
-- restoring it would fail rather than roll back.

DROP VIEW IF EXISTS sophonz_metadata.v_postgres_service ON CLUSTER {{.SOPHONZ_CLUSTER}};
