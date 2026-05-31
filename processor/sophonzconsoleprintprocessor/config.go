// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzconsoleprintprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzconsoleprintprocessor"

import (
	"go.opentelemetry.io/collector/component"
)

// Config holds configuration for the SOPHONZ console print processor.
type Config struct {
	TraceEnabled bool `mapstructure:"trace_enabled"`
	LogEnabled   bool `mapstructure:"log_enabled"`
}

var _ component.Config = (*Config)(nil)

// Validate checks that the configuration is valid.
func (cfg *Config) Validate() error {
	return nil
}
