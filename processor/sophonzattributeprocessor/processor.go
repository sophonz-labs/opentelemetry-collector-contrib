// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"context"
	"strconv"
	"sync"

	"github.com/segmentio/ksuid"
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
	serviceKeyCheck bool
	// custom
	encryptedUserIDKINFA bool
}

func newSOPHONZAttributeProcessor(cfg *Config, logger *zap.Logger) *SOPHONZAttributeProcessor {
	if !cfg.Enabled {
		return &SOPHONZAttributeProcessor{
			enabled:              cfg.Enabled,
			encryptedUserIDKINFA: cfg.EncryptedUserIDKINFA,
			logger:               logger,
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

	p := &SOPHONZAttributeProcessor{
		enabled:              cfg.Enabled,
		encryptedUserIDKINFA: cfg.EncryptedUserIDKINFA,
		logger:               logger,
		ksuid:                ksuid.New(),
		MetadataManager:      metadataManager,
		serviceKeyCheck:      cfg.ServiceKeyCheck,
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
