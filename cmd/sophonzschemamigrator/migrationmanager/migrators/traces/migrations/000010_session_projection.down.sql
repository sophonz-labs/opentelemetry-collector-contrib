ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER '{{.SOPHONZ_CLUSTER}}' DROP PROJECTION IF EXISTS session_by_minute_v1;
ALTER TABLE sophonz_traces.distributed_sophonz_index_v2 ON CLUSTER '{{.SOPHONZ_CLUSTER}}' DROP COLUMN IF EXISTS sessionMinuteV1;
ALTER TABLE sophonz_traces.distributed_sophonz_index_v2 ON CLUSTER '{{.SOPHONZ_CLUSTER}}' DROP COLUMN IF EXISTS sessionFactsV1;
ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER '{{.SOPHONZ_CLUSTER}}' DROP COLUMN IF EXISTS sessionMinuteV1;
ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER '{{.SOPHONZ_CLUSTER}}' DROP COLUMN IF EXISTS sessionFactsV1;
