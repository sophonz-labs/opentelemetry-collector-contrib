// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhousetracesexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousetracesexporter"

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Config defines configuration for the SOPHONZ ClickHouse traces exporter.
type Config struct {
	Options    `mapstructure:",squash"`
	UserName   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	Host       string `mapstructure:"host"`
	Port       string `mapstructure:"port"`
	Database   string `mapstructure:"database"`
	Migrations string `mapstructure:"migrations"`
	// DockerMultiNodeCluster is a flag to enable the docker multi node cluster. Default is false.
	DockerMultiNodeCluster bool `mapstructure:"docker_multi_node_cluster"`
	// LowCardinalExceptionGrouping is a flag to enable exception grouping by serviceName + exceptionType. Default is false.
	LowCardinalExceptionGrouping bool `mapstructure:"low_cardinal_exception_grouping"`

	TimeoutSettings exporterhelper.TimeoutConfig `mapstructure:",squash"`
	// QueueSettings is the sending queue configuration.
	// As of collector v0.150+ this is wrapped in configoptional.Optional.
	QueueSettings             configoptional.Optional[exporterhelper.QueueBatchConfig] `mapstructure:"sending_queue"`
	configretry.BackOffConfig `mapstructure:"retry_on_failure"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid.
//
// As of collector v0.150+, the sending queue is wrapped in
// configoptional.Optional: an absent value means the queue is disabled (the
// legacy `enabled` flag is gone). The QueueBatchConfig type validates its own
// QueueSize/NumConsumers (> 0) when present, so we only guard against
// negative values here to preserve the legacy error surface.
func (cfg *Config) Validate() error {
	if !cfg.QueueSettings.HasValue() {
		return nil
	}
	qs := cfg.QueueSettings.Get()
	if qs.QueueSize < 0 {
		return fmt.Errorf("remote write queue size can't be negative")
	}

	if qs.QueueSize == 0 {
		return fmt.Errorf("a 0 size queue will drop all the data")
	}

	if qs.NumConsumers < 0 {
		return fmt.Errorf("remote write consumer number can't be negative")
	}
	return nil
}
