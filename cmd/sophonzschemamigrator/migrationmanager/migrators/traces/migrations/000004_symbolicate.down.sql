DROP TABLE IF EXISTS sophonz_traces.sophonz_stacktrace ON CLUSTER {{.SOPHONZ_CLUSTER}};

ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF NOT EXISTS idx_traceID traceID TYPE bloom_filter(0.001) GRANULARITY 1;
ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF NOT EXISTS idx_spanID spanID TYPE bloom_filter(0.001) GRANULARITY 1;

ALTER TABLE sophonz_traces.sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF NOT EXISTS idx_traceID traceID TYPE bloom_filter(0.001) GRANULARITY 1;
ALTER TABLE sophonz_traces.sophonz_error_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    DROP INDEX IF NOT EXISTS idx_spanID spanID TYPE bloom_filter(0.001) GRANULARITY 1;
