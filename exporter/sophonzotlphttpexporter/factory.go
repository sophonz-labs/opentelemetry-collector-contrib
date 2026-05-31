// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate make mdatagen

package sophonzotlphttpexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzotlphttpexporter"

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzotlphttpexporter/internal/metadata"
)

// NewFactory creates a factory for the SOPHONZ OTLP/HTTP exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		metadata.Type,
		createDefaultConfig,
		exporter.WithTraces(createTraces, metadata.TracesStability),
		// SOPHONZ: metrics export is intentionally disabled (preserved from legacy fork).
		// exporter.WithMetrics(createMetrics, metadata.MetricsStability),
		exporter.WithLogs(createLogs, metadata.LogsStability),
	)
}

func createDefaultConfig() component.Config {
	clientConfig := confighttp.NewDefaultClientConfig()
	clientConfig.Timeout = 30 * time.Second
	// Default to gzip compression
	clientConfig.Compression = configcompression.TypeGzip
	// We almost read 0 bytes, so no need to tune ReadBufferSize.
	clientConfig.WriteBufferSize = 512 * 1024

	return &Config{
		RetryConfig:  configretry.NewDefaultBackOffConfig(),
		QueueConfig:  configoptional.Some(exporterhelper.NewDefaultQueueConfig()),
		Encoding:     EncodingProto,
		ClientConfig: clientConfig,
		// SOPHONZ: exporter is opt-in; disabled by default.
		Enabled: false,
	}
}

// composeSignalURL composes the final URL for the signal (traces, metrics, logs) based on the configuration.
// oCfg is the configuration of the exporter.
// signalOverrideURL is the URL specified in the signal specific configuration (empty if not specified).
// signalName is the name of the signal, e.g. "traces", "metrics", "logs".
// signalVersion is the version of the signal, e.g. "v1".
func composeSignalURL(oCfg *Config, signalOverrideURL, signalName, signalVersion string) (*url.URL, error) {
	switch {
	case signalOverrideURL != "":
		u, err := url.Parse(signalOverrideURL)
		if err != nil {
			return nil, fmt.Errorf("%s_endpoint must be a valid URL", signalName)
		}
		return u, nil
	case oCfg.ClientConfig.Endpoint == "":
		return nil, fmt.Errorf("either endpoint or %s_endpoint must be specified", signalName)
	default:
		if strings.HasSuffix(oCfg.ClientConfig.Endpoint, "/") {
			return url.Parse(oCfg.ClientConfig.Endpoint + signalVersion + "/" + signalName)
		}
		return url.Parse(oCfg.ClientConfig.Endpoint + "/" + signalVersion + "/" + signalName)
	}
}

func endpointAttributes(endpointURL *url.URL) []attribute.KeyValue {
	out := []attribute.KeyValue{
		semconv.ServerAddress(endpointURL.Hostname()),
	}
	if endpointURL.Path != "" {
		out = append(out, semconv.URLPath(endpointURL.Path))
	}
	if portNumber, err := strconv.Atoi(endpointURL.Port()); err == nil {
		out = append(out, semconv.ServerPort(portNumber))
	}
	return out
}

func createTraces(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {
	oce, err := newExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	oCfg := cfg.(*Config)

	endpointURL, err := composeSignalURL(oCfg, oCfg.TracesEndpoint, "traces", "v1")
	if err != nil {
		return nil, err
	}
	oce.tracesURL = endpointURL.String()

	return exporterhelper.NewTraces(ctx, set, cfg,
		oce.pushTraces,
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		// SOPHONZ: explicitly disable since we rely on http.Client timeout logic.
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}),
		exporterhelper.WithRetry(oCfg.RetryConfig),
		exporterhelper.WithQueue(oCfg.QueueConfig),
		exporterhelper.WithAttrs(endpointAttributes(endpointURL)...),
	)
}

func createLogs(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Logs, error) {
	oce, err := newExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	oCfg := cfg.(*Config)

	endpointURL, err := composeSignalURL(oCfg, oCfg.LogsEndpoint, "logs", "v1")
	if err != nil {
		return nil, err
	}
	oce.logsURL = endpointURL.String()

	return exporterhelper.NewLogs(ctx, set, cfg,
		oce.pushLogs,
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		// SOPHONZ: explicitly disable since we rely on http.Client timeout logic.
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}),
		exporterhelper.WithRetry(oCfg.RetryConfig),
		exporterhelper.WithQueue(oCfg.QueueConfig),
		exporterhelper.WithAttrs(endpointAttributes(endpointURL)...),
	)
}
