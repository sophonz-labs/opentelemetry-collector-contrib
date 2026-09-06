-- No alarm model in this deployment (see 000001). The view is dropped rather
-- than recreated over Postgres tables that do not exist.
DROP VIEW IF EXISTS sophonz_metadata.v_postgres_alarm ON CLUSTER {{.SOPHONZ_CLUSTER}};
