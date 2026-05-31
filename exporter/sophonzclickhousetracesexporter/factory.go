// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate make mdatagen

package sophonzclickhousetracesexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousetracesexporter"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousetracesexporter/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/schema"
)

const (
	primaryNamespace = "clickhouse"
	archiveNamespace = "clickhouse-archive"
)

// NewFactory creates a factory for the SOPHONZ ClickHouse traces exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		metadata.Type,
		createDefaultConfig,
		exporter.WithTraces(createTracesExporter, metadata.TracesStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Database:                     schema.DatabaseTraces,
		UserName:                     "sophonz",
		Password:                     "",
		Host:                         "localhost",
		Port:                         "9000",
		Migrations:                   "",
		DockerMultiNodeCluster:       false,
		LowCardinalExceptionGrouping: false,
		TimeoutSettings:              exporterhelper.NewDefaultTimeoutConfig(),
		BackOffConfig:                configretry.NewDefaultBackOffConfig(),
		QueueSettings:                configoptional.Some(exporterhelper.NewDefaultQueueConfig()),
	}
}

func createTracesExporter(
	ctx context.Context,
	params exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {
	c := cfg.(*Config)
	oce, err := newExporter(cfg, params.Logger)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewTraces(
		ctx,
		params,
		cfg,
		oce.pushTraceData,
		exporterhelper.WithShutdown(oce.Shutdown),
		exporterhelper.WithTimeout(c.TimeoutSettings),
		exporterhelper.WithQueue(c.QueueSettings),
		exporterhelper.WithRetry(c.BackOffConfig))
}
