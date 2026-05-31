// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate make mdatagen

package sophonzclickhousemetricsexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousemetricsexporter"

import (
	"context"
	"errors"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousemetricsexporter/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry"
)

var (
	writeLatencyMillis = stats.Int64("exporter_db_write_latency", "Time taken (in millis) for exporter to write batch", "ms")
	exporterKey        = tag.MustNewKey("exporter")
	tableKey           = tag.MustNewKey("table")
)

// NewFactory creates a new SOPHONZ ClickHouse metrics exporter.
func NewFactory() exporter.Factory {

	writeLatencyDistribution := view.Distribution(100, 250, 500, 750, 1000, 2000, 4000, 8000, 16000, 32000, 64000, 128000, 256000, 512000)

	writeLatencyView := &view.View{
		Name:        "exporter_db_write_latency",
		Measure:     writeLatencyMillis,
		Description: writeLatencyMillis.Description(),
		TagKeys:     []tag.Key{exporterKey, tableKey},
		Aggregation: writeLatencyDistribution,
	}

	view.Register(writeLatencyView)
	return exporter.NewFactory(
		metadata.Type,
		createDefaultConfig,
		exporter.WithMetrics(createMetricsExporter, metadata.MetricsStability))
}

func createMetricsExporter(ctx context.Context, set exporter.Settings,
	cfg component.Config) (exporter.Metrics, error) {

	prwCfg, ok := cfg.(*Config)
	if !ok {
		return nil, errors.New("invalid configuration")
	}

	prwe, err := NewPrwExporter(prwCfg, set)
	if err != nil {
		return nil, err
	}

	// Prometheus remote write samples need to be in chronological order for
	// each timeseries. If we shard the incoming metrics without considering
	// this limitation, we experience "out of order samples" errors, so the
	// queue is always single-consumer.
	// See https://github.com/open-telemetry/opentelemetry-collector/issues/2949.
	queueCfg := exporterhelper.NewDefaultQueueConfig()
	queueCfg.NumConsumers = 1
	queueCfg.QueueSize = int64(prwCfg.RemoteWriteQueue.QueueSize)
	queueOpt := configoptional.Some(queueCfg)
	if !prwCfg.RemoteWriteQueue.Enabled {
		queueOpt = configoptional.None[exporterhelper.QueueBatchConfig]()
	}

	exp, err := exporterhelper.NewMetrics(
		ctx,
		set,
		cfg,
		prwe.PushMetrics,
		exporterhelper.WithTimeout(prwCfg.TimeoutConfig),
		exporterhelper.WithQueue(queueOpt),
		exporterhelper.WithRetry(prwCfg.BackOffConfig),
		exporterhelper.WithStart(prwe.Start),
		exporterhelper.WithShutdown(prwe.Shutdown),
	)

	if err != nil {
		return nil, err
	}

	return resourcetotelemetry.WrapMetricsExporter(prwCfg.ResourceToTelemetrySettings, exp), nil
}

func createDefaultConfig() component.Config {
	return &Config{
		Namespace:      "",
		ExternalLabels: map[string]string{},
		TimeoutConfig:  exporterhelper.NewDefaultTimeoutConfig(),
		BackOffConfig:  configretry.NewDefaultBackOffConfig(),
		HTTPClientSettings: confighttp.ClientConfig{
			Endpoint: "http://some.url:9411/api/prom/push",
			// We almost read 0 bytes, so no need to tune ReadBufferSize.
			ReadBufferSize:  0,
			WriteBufferSize: 512 * 1024,
			Timeout:         exporterhelper.NewDefaultTimeoutConfig().Timeout,
		},
		// TODO(jbd): Adjust the default queue size.
		RemoteWriteQueue: RemoteWriteQueue{
			Enabled:      true,
			QueueSize:    10000,
			NumConsumers: 5,
		},
		WatcherInterval: 30 * time.Second,
		WriteTSToV4:     true,
	}
}
