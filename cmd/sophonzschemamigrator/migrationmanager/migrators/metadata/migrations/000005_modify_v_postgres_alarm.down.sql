DROP VIEW IF EXISTS sophonz_metadata.v_postgres_alarm ON CLUSTER {{.SOPHONZ_CLUSTER}};

CREATE VIEW IF NOT EXISTS sophonz_metadata.v_postgres_alarm ON CLUSTER {{.SOPHONZ_CLUSTER}}
AS
SELECT
    *
FROM
    postgresql(postgres_creds, table='Alarm') AS alarm;

DROP VIEW IF EXISTS sophonz_metadata.t_postgres_alarm_log ON CLUSTER {{.SOPHONZ_CLUSTER}};
DROP VIEW IF EXISTS sophonz_metadata.t_postgres_alarm_change_log ON CLUSTER {{.SOPHONZ_CLUSTER}};
