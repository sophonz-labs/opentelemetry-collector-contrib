-- Down migrations here used to restore the previous view definition, which
-- read Postgres tables this deployment does not have — rolling back would have
-- failed rather than reverted. Dropping the view is the only honest inverse of
-- creating it.
DROP VIEW IF EXISTS sophonz_metadata.v_postgres_alarm ON CLUSTER {{.SOPHONZ_CLUSTER}};
