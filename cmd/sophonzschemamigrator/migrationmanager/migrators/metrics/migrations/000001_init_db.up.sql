-- tables
CREATE TABLE IF NOT EXISTS sophonz_metrics.exp_hist
(
    `env`         LowCardinality(String) DEFAULT 'default',
    `temporality` LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name` LowCardinality(String),
    `fingerprint` UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`  Int64 CODEC(DoubleDelta, ZSTD(1)),
    `count`       UInt64 CODEC(ZSTD(1)),
    `sum`         Float64 CODEC(Gorilla, ZSTD(1)),
    `min`         Float64 CODEC(Gorilla, ZSTD(1)),
    `max`         Float64 CODEC(Gorilla, ZSTD(1)),
    `sketch`      AggregateFunction(quantilesDD(0.01, 0.5, 0.75, 0.9, 0.95, 0.99), UInt64) CODEC(ZSTD(1))
) ENGINE = MergeTree PARTITION BY toDate(unix_milli / 1000) ORDER BY (env, temporality, metric_name, fingerprint, unix_milli) TTL toDateTime(unix_milli / 1000) + toIntervalSecond(2592000) SETTINGS ttl_only_drop_parts = 1, index_granularity = 8192;

CREATE TABLE IF NOT EXISTS sophonz_metrics.samples_v2
(
    `metric_name`  LowCardinality(String),
    `fingerprint`  UInt64 CODEC(DoubleDelta, LZ4),
    `timestamp_ms` Int64 CODEC(DoubleDelta, LZ4),
    `value`        Float64 CODEC(Gorilla, LZ4)
) ENGINE = MergeTree PARTITION BY toDate(timestamp_ms / 1000) ORDER BY (metric_name, fingerprint, timestamp_ms) TTL toDateTime(timestamp_ms / 1000) + toIntervalSecond(2592000) SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1;

CREATE TABLE IF NOT EXISTS sophonz_metrics.samples_v4
(
    `env`         LowCardinality(String) DEFAULT 'default',
    `temporality` LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name` LowCardinality(String),
    `fingerprint` UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`  Int64 CODEC(DoubleDelta, ZSTD(1)),
    `value`       Float64 CODEC(Gorilla, ZSTD(1))
) ENGINE = MergeTree PARTITION BY toDate(unix_milli / 1000) ORDER BY (env, temporality, metric_name, fingerprint, unix_milli) TTL toDateTime(unix_milli / 1000) + toIntervalSecond(2592000) SETTINGS ttl_only_drop_parts = 1, index_granularity = 8192;

CREATE TABLE IF NOT EXISTS sophonz_metrics.samples_v4_agg_30m
(
    `env`         LowCardinality(String) DEFAULT 'default',
    `temporality` LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name` LowCardinality(String),
    `fingerprint` UInt64 CODEC(ZSTD(1)),
    `unix_milli`  Int64 CODEC(DoubleDelta, ZSTD(1)),
    `last`        SimpleAggregateFunction(anyLast, Float64) CODEC(ZSTD(1)),
    `min`         SimpleAggregateFunction( min, Float64) CODEC(ZSTD(1)),
    `max`         SimpleAggregateFunction( max, Float64) CODEC(ZSTD(1)),
    `sum`         SimpleAggregateFunction( sum, Float64) CODEC(ZSTD(1)),
    `count`       SimpleAggregateFunction( sum, UInt64) CODEC(ZSTD(1))
) ENGINE = AggregatingMergeTree PARTITION BY toDate(unix_milli / 1000) ORDER BY (env, temporality, metric_name, fingerprint, unix_milli) TTL toDateTime(unix_milli / 1000) + toIntervalSecond(2592000) SETTINGS ttl_only_drop_parts = 1, index_granularity = 8192;

CREATE TABLE IF NOT EXISTS sophonz_metrics.samples_v4_agg_5m
(
    `env`         LowCardinality(String) DEFAULT 'default',
    `temporality` LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name` LowCardinality(String),
    `fingerprint` UInt64 CODEC(ZSTD(1)),
    `unix_milli`  Int64 CODEC(DoubleDelta, ZSTD(1)),
    `last`        SimpleAggregateFunction(anyLast, Float64) CODEC(ZSTD(1)),
    `min`         SimpleAggregateFunction( min, Float64) CODEC(ZSTD(1)),
    `max`         SimpleAggregateFunction( max, Float64) CODEC(ZSTD(1)),
    `sum`         SimpleAggregateFunction( sum, Float64) CODEC(ZSTD(1)),
    `count`       SimpleAggregateFunction( sum, UInt64) CODEC(ZSTD(1))
) ENGINE = AggregatingMergeTree PARTITION BY toDate(unix_milli / 1000) ORDER BY (env, temporality, metric_name, fingerprint, unix_milli) TTL toDateTime(unix_milli / 1000) + toIntervalSecond(2592000) SETTINGS ttl_only_drop_parts = 1, index_granularity = 8192;

CREATE TABLE IF NOT EXISTS sophonz_metrics.schema_migrations
(
    `version`  Int64,
    `dirty`    UInt8,
    `sequence` UInt64
) ENGINE = MergeTree ORDER BY sequence SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS sophonz_metrics.time_series_v2
(
    `metric_name`  LowCardinality(String),
    `fingerprint`  UInt64 CODEC(DoubleDelta, LZ4),
    `timestamp_ms` Int64 CODEC(DoubleDelta, LZ4),
    `labels`       String CODEC(ZSTD(5)),
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified' CODEC(ZSTD(5)),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1)),
    INDEX          temporality_index temporality TYPE SET (3) GRANULARITY 1
) ENGINE = ReplacingMergeTree PARTITION BY toDate(timestamp_ms / 1000) ORDER BY (metric_name, fingerprint) SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1;

CREATE TABLE IF NOT EXISTS sophonz_metrics.time_series_v3
(
    `env`          LowCardinality(String) DEFAULT 'default',
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name`  LowCardinality(String),
    `fingerprint`  UInt64 CODEC(Delta(8), ZSTD(1)),
    `timestamp_ms` Int64 CODEC(Delta(8), ZSTD(1)),
    `labels`       String CODEC(ZSTD(5)),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1))
) ENGINE = ReplacingMergeTree PARTITION BY toDate(timestamp_ms / 1000) ORDER BY (env, temporality, metric_name, fingerprint) SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS sophonz_metrics.time_series_v4
(
    `env`          LowCardinality(String) DEFAULT 'default',
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name`  LowCardinality(String),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1)),
    `fingerprint`  UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`   Int64 CODEC(Delta(8), ZSTD(1)),
    `labels`       String CODEC(ZSTD(5)),
    INDEX          idx_labels labels TYPE ngrambf_v1(4, 1024, 3, 0) GRANULARITY 1
) ENGINE = ReplacingMergeTree PARTITION BY toDate(unix_milli / 1000) ORDER BY (env, temporality, metric_name, fingerprint, unix_milli) TTL toDateTime(unix_milli / 1000) + toIntervalSecond(2592000) SETTINGS ttl_only_drop_parts = 1, index_granularity = 8192;

CREATE TABLE IF NOT EXISTS sophonz_metrics.time_series_v4_1day
(
    `env`          LowCardinality(String) DEFAULT 'default',
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name`  LowCardinality(String),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1)),
    `fingerprint`  UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`   Int64 CODEC(Delta(8), ZSTD(1)),
    `labels`       String CODEC(ZSTD(5)),
    INDEX          idx_labels labels TYPE ngrambf_v1(4, 1024, 3, 0) GRANULARITY 1
) ENGINE = ReplacingMergeTree PARTITION BY toDate(unix_milli / 1000) ORDER BY (env, temporality, metric_name, fingerprint, unix_milli) TTL toDateTime(unix_milli / 1000) + toIntervalSecond(2592000) SETTINGS ttl_only_drop_parts = 1, index_granularity = 8192;

CREATE TABLE IF NOT EXISTS sophonz_metrics.time_series_v4_1week
(
    `env`          LowCardinality(String) DEFAULT 'default',
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name`  LowCardinality(String),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1)),
    `fingerprint`  UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`   Int64 CODEC(Delta(8), ZSTD(1)),
    `labels`       String CODEC(ZSTD(5)),
    INDEX          idx_labels labels TYPE ngrambf_v1(4, 1024, 3, 0) GRANULARITY 1
) ENGINE = ReplacingMergeTree PARTITION BY toDate(unix_milli / 1000) ORDER BY (env, temporality, metric_name, fingerprint, unix_milli) TTL toDateTime(unix_milli / 1000) + toIntervalSecond(2592000) SETTINGS ttl_only_drop_parts = 1, index_granularity = 8192;

CREATE TABLE IF NOT EXISTS sophonz_metrics.time_series_v4_6hrs
(
    `env`          LowCardinality(String) DEFAULT 'default',
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name`  LowCardinality(String),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1)),
    `fingerprint`  UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`   Int64 CODEC(Delta(8), ZSTD(1)),
    `labels`       String CODEC(ZSTD(5)),
    INDEX          idx_labels labels TYPE ngrambf_v1(4, 1024, 3, 0) GRANULARITY 1
) ENGINE = ReplacingMergeTree PARTITION BY toDate(unix_milli / 1000) ORDER BY (env, temporality, metric_name, fingerprint, unix_milli) TTL toDateTime(unix_milli / 1000) + toIntervalSecond(2592000) SETTINGS ttl_only_drop_parts = 1, index_granularity = 8192;

CREATE TABLE IF NOT EXISTS sophonz_metrics.usage
(
    `tenant`       String,
    `collector_id` String,
    `exporter_id`  String,
    `timestamp`    DateTime,
    `data`         String
) ENGINE = MergeTree ORDER BY (tenant, collector_id, exporter_id, timestamp) TTL timestamp + toIntervalDay(3) SETTINGS index_granularity = 8192;


-- distributed tables
CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_exp_hist
(
    `env`         LowCardinality(String) DEFAULT 'default',
    `temporality` LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name` LowCardinality(String),
    `fingerprint` UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`  Int64 CODEC(DoubleDelta, ZSTD(1)),
    `count`       UInt64 CODEC(ZSTD(1)),
    `sum`         Float64 CODEC(Gorilla, ZSTD(1)),
    `min`         Float64 CODEC(Gorilla, ZSTD(1)),
    `max`         Float64 CODEC(Gorilla, ZSTD(1)),
    `sketch`      AggregateFunction(quantilesDD(0.01, 0.5, 0.75, 0.9, 0.95, 0.99), UInt64) CODEC(ZSTD(1))
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'exp_hist', cityHash64(env, temporality, metric_name, fingerprint));

CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_samples_v2
(
    `metric_name`  LowCardinality(String),
    `fingerprint`  UInt64 CODEC(DoubleDelta, LZ4),
    `timestamp_ms` Int64 CODEC(DoubleDelta, LZ4),
    `value`        Float64 CODEC(Gorilla, LZ4)
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'samples_v2', cityHash64(metric_name, fingerprint));

CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_samples_v4
(
    `env`         LowCardinality(String) DEFAULT 'default',
    `temporality` LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name` LowCardinality(String),
    `fingerprint` UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`  Int64 CODEC(DoubleDelta, ZSTD(1)),
    `value`       Float64 CODEC(Gorilla, ZSTD(1))
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'samples_v4', cityHash64(env, temporality, metric_name, fingerprint));

CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_samples_v4_agg_30m
(
    `env`         LowCardinality(String) DEFAULT 'default',
    `temporality` LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name` LowCardinality(String),
    `fingerprint` UInt64 CODEC(ZSTD(1)),
    `unix_milli`  Int64 CODEC(DoubleDelta, ZSTD(1)),
    `last`        SimpleAggregateFunction(anyLast, Float64) CODEC(ZSTD(1)),
    `min`         SimpleAggregateFunction( min, Float64) CODEC(ZSTD(1)),
    `max`         SimpleAggregateFunction( max, Float64) CODEC(ZSTD(1)),
    `sum`         SimpleAggregateFunction( sum, Float64) CODEC(ZSTD(1)),
    `count`       SimpleAggregateFunction( sum, UInt64) CODEC(ZSTD(1))
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'samples_v4_agg_30m', cityHash64(env, temporality, metric_name, fingerprint));

CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_samples_v4_agg_5m
(
    `env`         LowCardinality(String) DEFAULT 'default',
    `temporality` LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name` LowCardinality(String),
    `fingerprint` UInt64 CODEC(ZSTD(1)),
    `unix_milli`  Int64 CODEC(DoubleDelta, ZSTD(1)),
    `last`        SimpleAggregateFunction(anyLast, Float64) CODEC(ZSTD(1)),
    `min`         SimpleAggregateFunction( min, Float64) CODEC(ZSTD(1)),
    `max`         SimpleAggregateFunction( max, Float64) CODEC(ZSTD(1)),
    `sum`         SimpleAggregateFunction( sum, Float64) CODEC(ZSTD(1)),
    `count`       SimpleAggregateFunction( sum, UInt64) CODEC(ZSTD(1))
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'samples_v4_agg_5m', cityHash64(env, temporality, metric_name, fingerprint));

CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_time_series_v2
(
    `metric_name`  LowCardinality(String),
    `fingerprint`  UInt64 CODEC(DoubleDelta, LZ4),
    `timestamp_ms` Int64 CODEC(DoubleDelta, LZ4),
    `labels`       String CODEC(ZSTD(5)),
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified' CODEC(ZSTD(5)),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1))
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'time_series_v2', cityHash64(metric_name, fingerprint));

CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_time_series_v3
(
    `env`          LowCardinality(String) DEFAULT 'default',
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name`  LowCardinality(String),
    `fingerprint`  UInt64 CODEC(Delta(8), ZSTD(1)),
    `timestamp_ms` Int64 CODEC(Delta(8), ZSTD(1)),
    `labels`       String CODEC(ZSTD(5)),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1))
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'time_series_v3', cityHash64(env, temporality, metric_name, fingerprint));

CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_time_series_v4
(
    `env`          LowCardinality(String) DEFAULT 'default',
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name`  LowCardinality(String),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1)),
    `fingerprint`  UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`   Int64 CODEC(Delta(8), ZSTD(1)),
    `labels`       String CODEC(ZSTD(5))
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'time_series_v4', cityHash64(env, temporality, metric_name, fingerprint));

CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_time_series_v4_1day
(
    `env`          LowCardinality(String) DEFAULT 'default',
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name`  LowCardinality(String),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1)),
    `fingerprint`  UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`   Int64 CODEC(Delta(8), ZSTD(1)),
    `labels`       String CODEC(ZSTD(5))
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'time_series_v4_1day', cityHash64(env, temporality, metric_name, fingerprint));

CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_time_series_v4_1week
(
    `env`          LowCardinality(String) DEFAULT 'default',
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name`  LowCardinality(String),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1)),
    `fingerprint`  UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`   Int64 CODEC(Delta(8), ZSTD(1)),
    `labels`       String CODEC(ZSTD(5))
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'time_series_v4_1week', cityHash64(env, temporality, metric_name, fingerprint));

CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_time_series_v4_6hrs
(
    `env`          LowCardinality(String) DEFAULT 'default',
    `temporality`  LowCardinality(String) DEFAULT 'Unspecified',
    `metric_name`  LowCardinality(String),
    `description`  LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `unit`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `type`         LowCardinality(String) DEFAULT '' CODEC(ZSTD(1)),
    `is_monotonic` Bool DEFAULT false CODEC(ZSTD(1)),
    `fingerprint`  UInt64 CODEC(Delta(8), ZSTD(1)),
    `unix_milli`   Int64 CODEC(Delta(8), ZSTD(1)),
    `labels`       String CODEC(ZSTD(5))
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'time_series_v4_6hrs', cityHash64(env, temporality, metric_name, fingerprint));

CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_usage
(
    `tenant`       String,
    `collector_id` String,
    `exporter_id`  String,
    `timestamp`    DateTime,
    `data`         String
) ENGINE = Distributed('cluster', 'sophonz_metrics', 'usage', cityHash64(rand()));


-- materialized views
CREATE
MATERIALIZED VIEW sophonz_metrics.samples_v4_agg_30m_mv TO sophonz_metrics.samples_v4_agg_30m (`env` LowCardinality(String), `temporality` LowCardinality(String), `metric_name` LowCardinality(String), `fingerprint` UInt64, `unix_milli` Int64, `last` SimpleAggregateFunction(anyLast, Float64), `min` SimpleAggregateFunction(min, Float64), `max` SimpleAggregateFunction(max, Float64), `sum` Float64, `count` UInt64) AS
SELECT env,
       temporality,
       metric_name,
       fingerprint,
       intDiv(unix_milli, 1800000) * 1800000 AS unix_milli,
       anyLast(last) AS last, min(min) AS min, max(max) AS max, sum(sum) AS sum, sum(count) AS count
FROM sophonz_metrics.samples_v4_agg_5m
GROUP BY env, temporality, metric_name, fingerprint, unix_milli;

CREATE
MATERIALIZED VIEW sophonz_metrics.samples_v4_agg_5m_mv TO sophonz_metrics.samples_v4_agg_5m (`env` LowCardinality(String), `temporality` LowCardinality(String), `metric_name` LowCardinality(String), `fingerprint` UInt64, `unix_milli` Int64, `last` Float64, `min` Float64, `max` Float64, `sum` Float64, `count` UInt64) AS
SELECT env,
       temporality,
       metric_name,
       fingerprint,
       intDiv(unix_milli, 300000) * 300000 AS unix_milli,
       anyLast(value) AS last, min(value) AS min, max(value) AS max, sum(value) AS sum, count(*) AS count
FROM sophonz_metrics.samples_v4
GROUP BY env, temporality, metric_name, fingerprint, unix_milli;

CREATE
MATERIALIZED VIEW sophonz_metrics.time_series_v4_1day_mv TO sophonz_metrics.time_series_v4_1day (`env` LowCardinality(String), `temporality` LowCardinality(String), `metric_name` LowCardinality(String), `description` LowCardinality(String), `unit` LowCardinality(String), `type` LowCardinality(String), `is_monotonic` Bool, `fingerprint` UInt64, `unix_milli` Float64, `labels` String) AS
SELECT env,
       temporality,
       metric_name,
       description,
       unit,
       type,
       is_monotonic,
       fingerprint,
       floor(unix_milli / 86400000) * 86400000 AS unix_milli,
       labels
FROM sophonz_metrics.time_series_v4;

CREATE
MATERIALIZED VIEW sophonz_metrics.time_series_v4_1week_mv TO sophonz_metrics.time_series_v4_1week (`env` LowCardinality(String), `temporality` LowCardinality(String), `metric_name` LowCardinality(String), `description` LowCardinality(String), `unit` LowCardinality(String), `type` LowCardinality(String), `is_monotonic` Bool, `fingerprint` UInt64, `unix_milli` Float64, `labels` String) AS
SELECT env,
       temporality,
       metric_name,
       description,
       unit,
       type,
       is_monotonic,
       fingerprint,
       floor(unix_milli / 604800000) * 604800000 AS unix_milli,
       labels
FROM sophonz_metrics.time_series_v4_1day;

CREATE
MATERIALIZED VIEW sophonz_metrics.time_series_v4_6hrs_mv TO sophonz_metrics.time_series_v4_6hrs (`env` LowCardinality(String), `temporality` LowCardinality(String), `metric_name` LowCardinality(String), `description` LowCardinality(String), `unit` LowCardinality(String), `type` LowCardinality(String), `is_monotonic` Bool, `fingerprint` UInt64, `unix_milli` Float64, `labels` String) AS
SELECT env,
       temporality,
       metric_name,
       description,
       unit,
       type,
       is_monotonic,
       fingerprint,
       floor(unix_milli / 21600000) * 21600000 AS unix_milli,
       labels
FROM sophonz_metrics.time_series_v4;