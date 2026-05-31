DROP TABLE IF EXISTS sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}};
DROP TABLE IF EXISTS sophonz_traces.sophonz_spans ON CLUSTER {{.SOPHONZ_CLUSTER}};
DROP TABLE IF EXISTS sophonz_traces.span_attributes ON CLUSTER {{.SOPHONZ_CLUSTER}};
DROP TABLE IF EXISTS sophonz_traces.span_attributes_keys ON CLUSTER {{.SOPHONZ_CLUSTER}};
DROP TABLE IF EXISTS sophonz_traces.sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}};

DROP TABLE IF EXISTS sophonz_traces.distributed_sophonz_index_v2;
DROP TABLE IF EXISTS sophonz_traces.distributed_sophonz_spans;
DROP TABLE IF EXISTS sophonz_traces.distributed_span_attributes;
DROP TABLE IF EXISTS sophonz_traces.distributed_span_attributes_keys;
DROP TABLE IF EXISTS sophonz_traces.distributed_sophonz_error_index_v2;
