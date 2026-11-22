ALTER TABLE sophonz_logs.logs_v2 ON CLUSTER '{{.SOPHONZ_CLUSTER}}' DROP PROJECTION IF EXISTS session_by_sdk_minute_v1;
ALTER TABLE sophonz_logs.distributed_logs_v2 ON CLUSTER '{{.SOPHONZ_CLUSTER}}' DROP COLUMN IF EXISTS sessionSdkV1;
ALTER TABLE sophonz_logs.logs_v2 ON CLUSTER '{{.SOPHONZ_CLUSTER}}' DROP COLUMN IF EXISTS sessionSdkV1;
