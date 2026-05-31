package migrationmanager

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/basemigrator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/logs"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/metrics"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/traces"
	"go.uber.org/zap"
)

type MigrationManager struct {
	Migrators []migrators.Migrator
	logger    *zap.Logger
}

func New(dsn string, clusterName string, isDurationSortFeatureDisabled, isTimestampSortFeatureDisabled, verboseLoggingEnabled, replicationEnabled bool) (*MigrationManager, error) {
	logger := zap.L().With(zap.String("component", "migrationmanager"))
	cfg := migrators.MigratorConfig{
		DSN:                            dsn,
		ClusterName:                    clusterName,
		IsDurationSortFeatureDisabled:  isDurationSortFeatureDisabled,
		IsTimestampSortFeatureDisabled: isTimestampSortFeatureDisabled,
		VerboseLoggingEnabled:          verboseLoggingEnabled,
		ReplicationEnabled:             replicationEnabled,
	}

	logsMigrator, err := createNewMigrator("logs", cfg)
	if err != nil {
		logger.Error("Failed to create logs migrator", zap.Error(err))
		return nil, err
	}
	metricsMigrator, err := createNewMigrator("metrics", cfg)
	if err != nil {
		logger.Error("Failed to create metrics migrator", zap.Error(err))
		return nil, err
	}
	tracesMigrator, err := createNewMigrator("traces", cfg)
	if err != nil {
		logger.Error("Failed to create traces migrator", zap.Error(err))
		return nil, err
	}
	metadataMigrator, err := createNewMigrator("metadata", cfg)
	if err != nil {
		logger.Error("Failed to create metadata migrator", zap.Error(err))
		return nil, err
	}
	return &MigrationManager{
		Migrators: []migrators.Migrator{
			logsMigrator,
			metricsMigrator,
			tracesMigrator,
			metadataMigrator,
		},
		logger: logger,
	}, nil
}

func createNewMigrator(migratorType string, cfg migrators.MigratorConfig) (migrators.Migrator, error) {
	logger := zap.L().With(zap.String("migrator", migratorType))
	b, err := basemigrator.New(cfg, logger)
	if err != nil {
		logger.Error("Failed to create base migrator", zap.Error(err))
		return nil, err
	}
	switch migratorType {
	case "traces":
		return &traces.TracesMigrator{BaseMigrator: b}, nil
	case "metrics":
		return &metrics.MetricsMigrator{BaseMigrator: b}, nil
	case "logs":
		return &logs.LogsMigrator{BaseMigrator: b}, nil
	case "metadata":
		return &metadata.MetadataMigrator{BaseMigrator: b}, nil
	default:
		return nil, fmt.Errorf("invalid migrator type: %s", migratorType)
	}
}

func (m *MigrationManager) Migrate(ctx context.Context) error {
	m.logger.Info("Running migrations for all migrators")
	for _, migrator := range m.Migrators {
		m.logger.Info(fmt.Sprintf("Running migrations for %s", migrator.Name()), zap.String("migrator", migrator.Name()))
		err := migrator.Migrate(ctx)
		if err != nil {
			m.logger.Error("Failed to run migrations for migrator", zap.String("migrator", migrator.Name()), zap.Error(err))
			return err
		}
	}
	m.logger.Info("Finished running migrations for all migrators")
	return nil
}

func (m *MigrationManager) Close() error {
	for _, migrator := range m.Migrators {
		if err := migrator.Close(); err != nil {
			return err
		}
	}
	return nil
}
