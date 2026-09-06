ALTER TABLE sophonz_traces.sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF EXISTS idx_tenantID;

ALTER TABLE sophonz_traces.distributed_sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP COLUMN IF EXISTS tenantID;
ALTER TABLE sophonz_traces.sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP COLUMN IF EXISTS tenantID;

ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF EXISTS idx_tenantID;

ALTER TABLE sophonz_traces.distributed_sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP COLUMN IF EXISTS tenantID;
ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP COLUMN IF EXISTS tenantID;
