ALTER TABLE sophonz_logs.logs_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF EXISTS idx_event_name;

ALTER TABLE sophonz_logs.distributed_logs_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP COLUMN IF EXISTS EventName;
ALTER TABLE sophonz_logs.logs_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP COLUMN IF EXISTS EventName;
