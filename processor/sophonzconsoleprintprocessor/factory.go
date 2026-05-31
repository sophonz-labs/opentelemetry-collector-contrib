// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate make mdatagen

package sophonzconsoleprintprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzconsoleprintprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzconsoleprintprocessor/internal/metadata"
)

var processorCapabilities = consumer.Capabilities{MutatesData: false}

// NewFactory creates a new processor factory for the SOPHONZ console print processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, metadata.TracesStability),
		processor.WithLogs(createLogsProcessor, metadata.LogsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createTracesProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	config := cfg.(*Config)
	proc := newSOPHONZConsolePrintProcessor(config, set.Logger)
	return processorhelper.NewTraces(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processTraces,
		processorhelper.WithShutdown(proc.Shutdown),
		processorhelper.WithCapabilities(processorCapabilities),
	)
}

func createLogsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	config := cfg.(*Config)
	proc := newSOPHONZConsolePrintProcessor(config, set.Logger)
	return processorhelper.NewLogs(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processLogs,
		processorhelper.WithShutdown(proc.Shutdown),
		processorhelper.WithCapabilities(processorCapabilities),
	)
}
