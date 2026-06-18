-- Root spans are exported with parentSpanID = "" (idutil.SpanIDToHexOrEmptyString), but the
-- column is FixedString(16), which pads empty values to 16 NUL bytes. Downstream consumers
-- (e.g. the Grafana trace gantt) detect a root span by parentSpanID = '' and otherwise fail to
-- build the trace tree ("Cannot read properties of undefined (reading 'meta')").
-- Converting the column to String drops the trailing NUL padding (for existing rows too, since
-- ClickHouse trims trailing zeros on FixedString -> String), so root spans read back as true ''.
ALTER TABLE sophonz_traces.sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    MODIFY COLUMN parentSpanID String;

ALTER TABLE sophonz_traces.distributed_sophonz_index_v2 ON CLUSTER {{.SOPHONZ_CLUSTER}}
    MODIFY COLUMN parentSpanID String;
