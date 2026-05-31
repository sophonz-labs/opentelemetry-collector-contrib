package traces

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/basemigrator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/schema"
)

const (
	name              = "traces"
	DatabaseNameTrace = "sophonz_traces"
	migrationFolder   = "migrationmanager/migrators/traces/migrations"
)

type TracesMigrator struct {
	*basemigrator.BaseMigrator
}

func (m *TracesMigrator) Migrate(ctx context.Context) error {
	err := m.BaseMigrator.Migrate(ctx, DatabaseNameTrace, migrationFolder)
	if err != nil {
		return err
	}

	return m.initFeatures()
}

func (m *TracesMigrator) Name() string {
	return name
}

func (m *TracesMigrator) initFeatures() error {
	if m.Cfg.IsTimestampSortFeatureDisabled {
		err := disableTimestampSortFeature(m.DB, m.Cfg.ClusterName)
		if err != nil {
			return err
		}
	} else {
		err := enableTimestampSortFeature(m.DB, m.Cfg.ClusterName)
		if err != nil {
			return err
		}
	}
	return nil
}

func enableTimestampSortFeature(db clickhouse.Conn, cluster string) error {
	err := db.Exec(context.Background(), fmt.Sprintf(`ALTER TABLE %s.%s ON CLUSTER %s
	ADD PROJECTION IF NOT EXISTS timestampSort 
	( SELECT * ORDER BY timestamp )`, schema.DatabaseTraces, schema.TracesLocalIndexTable, cluster))
	if err != nil {
		return err
	}
	return nil
}

func disableTimestampSortFeature(db clickhouse.Conn, cluster string) error {
	err := db.Exec(context.Background(), fmt.Sprintf(`ALTER TABLE %s.%s ON CLUSTER %s
	DROP PROJECTION IF EXISTS timestampSort`, schema.DatabaseTraces, schema.TracesLocalIndexTable, cluster))
	if err != nil {
		return err
	}
	return nil
}
