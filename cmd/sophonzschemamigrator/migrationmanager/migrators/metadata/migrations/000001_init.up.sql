CREATE TABLE IF NOT EXISTS sophonz_metadata.schema_ttl_config ON CLUSTER {{.SOPHONZ_CLUSTER}} (
    database_name LowCardinality(String),
    table_name LowCardinality(String),
    column_name LowCardinality(String),
    ttl UInt32,
    unit enum('NOT SET', 'SECOND', 'MINUTE', 'HOUR', 'DAY', 'WEEK', 'MONTH', 'QUARTER', 'YEAR')
) ENGINE = MergeTree ORDER BY (database_name, table_name);

-- v_postgres_alarm is deliberately not created either: nothing in this
-- deployment reads it and the application has no alarm model, so an empty
-- stand-in would only pretend the concept exists.
--
-- Note for anyone editing these files: the migration runner splits on ';', so a
-- chunk containing only comments — a trailing note after the last statement, or
-- a semicolon inside a comment — is handed to ClickHouse as an empty query and
-- fails the whole migration.
--
-- This deployment has no screen registry: the application models projects and
-- apps, not screens, so there is nothing for this view to read. It is created
-- with the right shape and no rows, because the collector's metadata manager
-- selects from it every refresh — an empty result is the truth, an error is
-- just noise in the log.
--
-- Column types match pkg/sophonz/metadata.Screen so the driver can scan them.
DROP VIEW IF EXISTS sophonz_metadata.v_postgres_screen ON CLUSTER {{.SOPHONZ_CLUSTER}};

CREATE VIEW IF NOT EXISTS sophonz_metadata.v_postgres_screen ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS
SELECT
    CAST(0, 'Int32') AS id,
    CAST('', 'String') AS screenName,
    CAST('', 'String') AS screenType,
    CAST(0, 'Int32') AS groupID,
    CAST('', 'String') AS groupName,
    CAST('', 'String') AS serviceNamespace
WHERE 0;
