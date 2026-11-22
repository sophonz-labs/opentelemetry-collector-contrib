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
