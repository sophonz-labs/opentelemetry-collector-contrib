// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzconsoleprintprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzconsoleprintprocessor"

import (
	"context"

	"go.uber.org/zap"
)

type sophonzConsolePrintProcessor struct {
	traceEnabled bool
	logEnabled   bool
	logger       *zap.Logger
}

func newSOPHONZConsolePrintProcessor(cfg *Config, logger *zap.Logger) *sophonzConsolePrintProcessor {
	return &sophonzConsolePrintProcessor{
		traceEnabled: cfg.TraceEnabled,
		logEnabled:   cfg.LogEnabled,
		logger:       logger,
	}
}

func (p *sophonzConsolePrintProcessor) Shutdown(_ context.Context) error {
	p.logger.Info("Shutting down sophonzconsoleprintprocessor")
	return nil
}
