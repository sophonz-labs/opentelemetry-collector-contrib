ALTER TABLE sophonz_traces.distributed_sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    MODIFY COLUMN parentSpanID FixedString(16);

ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    MODIFY COLUMN parentSpanID FixedString(16);
