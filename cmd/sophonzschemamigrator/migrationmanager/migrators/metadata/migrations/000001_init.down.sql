DROP TABLE IF EXISTS sophonz_metadata.schema_ttl_config ON CLUSTER {{.SOPHONZ_CLUSTER}};
DROP VIEW IF EXISTS sophonz_metadata.v_postgres_screen ON CLUSTER {{.SOPHONZ_CLUSTER}};
DROP VIEW IF EXISTS sophonz_metadata.v_postgres_service ON CLUSTER {{.SOPHONZ_CLUSTER}};
