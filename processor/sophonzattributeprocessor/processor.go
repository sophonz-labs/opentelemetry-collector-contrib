// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"context"
	"strconv"
	"sync"

	"github.com/segmentio/ksuid"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	sophonzmetadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/metadata"
)

type SOPHONZAttributeProcessor struct {
	logger          *zap.Logger
	MetadataManager *sophonzmetadata.MetadataManager
	enabled         bool
	shutdownOnce    sync.Once
	ksuid           ksuid.KSUID
	muKSUID         sync.Mutex
	keyMode         ServiceKeyMode
	tenant          *tenantCounters
	// custom
	encryptedUserID bool
}

func newSOPHONZAttributeProcessor(cfg *Config, telemetry component.TelemetrySettings) *SOPHONZAttributeProcessor {
	logger := telemetry.Logger
	if !cfg.Enabled {
		return &SOPHONZAttributeProcessor{
			enabled:         cfg.Enabled,
			encryptedUserID: cfg.EncryptedUserID,
			logger:          logger,
		}
	}

	metadataManager, err := sophonzmetadata.GetSharedMetadataManager(
		cfg.Clickhouse.Host,
		strconv.Itoa(cfg.Clickhouse.Port),
		cfg.Clickhouse.Database,
		cfg.Clickhouse.UserName,
		cfg.Clickhouse.Password,
		cfg.Interval,
		logger,
	)
	if err != nil {
		logger.Warn("Failed to create sophonz metadata manager", zap.Error(err))
	}

	// KeyMode is resolved by Config.Validate; default it here too so a config
	// that never went through validation cannot silently enable the lookup.
	keyMode := cfg.KeyMode
	if keyMode == "" {
		keyMode = ServiceKeyModeOff
	}

	p := &SOPHONZAttributeProcessor{
		enabled:         cfg.Enabled,
		encryptedUserID: cfg.EncryptedUserID,
		logger:          logger,
		ksuid:           ksuid.New(),
		MetadataManager: metadataManager,
		keyMode:         keyMode,
		tenant:          newTenantCounters(telemetry.MeterProvider, logger),
	}

	return p
}

func (p *SOPHONZAttributeProcessor) Shutdown(_ context.Context) error {
	p.logger.Info("Shutting down an sophonzattributeprocessor instance")
	p.shutdownOnce.Do(func() {
		if p.MetadataManager != nil {
			sophonzmetadata.ReleaseMetadataManager()
		}
	})
	return nil
}
