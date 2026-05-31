package metrics

import (
	"context"
	"fmt"
	"strings"

	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/basemigrator"
)

const (
	name                       = "metrics"
	DatabaseNameMetric         = "sophonz_metrics"
	migrationFolder            = "migrationmanager/migrators/metrics/migrations"
	timeSeriesTable            = "time_series_v4"
	distributedTimeSeriesTable = "distributed_time_series_v4"
)

type MetricsMigrator struct {
	*basemigrator.BaseMigrator
}

func (m *MetricsMigrator) Migrate(ctx context.Context) error {
	err := m.BaseMigrator.Migrate(ctx, DatabaseNameMetric, migrationFolder)
	if err != nil {
		return err
	}

	// TODO(srikanthccv): Remove this once we have a better way to handle data and last write
	removeTTL := fmt.Sprintf(`ALTER TABLE %s.%s ON CLUSTER %s REMOVE TTL;`, DatabaseNameMetric, timeSeriesTable, m.Cfg.ClusterName)
	if err = m.DB.Exec(context.Background(), removeTTL); err != nil {
		if !strings.Contains(err.Error(), "Table doesn't have any table TTL expression, cannot remove.") && !strings.Contains(err.Error(), "code: 60") {
			return fmt.Errorf("failed to remove TTL from table %s.%s, err: %s", DatabaseNameMetric, timeSeriesTable, err)
		}
	}

	// TODO(srikanthccv): remove this after a while
	// Why?
	// The table shard key had incorrectly been set to (env, temporality, metric_name, fingerprint, unix_milli) instead of (env, temporality, metric_name, fingerprint)
	var showCreateTable string
	if err = m.DB.QueryRow(context.Background(), fmt.Sprintf("SHOW CREATE TABLE %s.%s", DatabaseNameMetric, distributedTimeSeriesTable)).Scan(&showCreateTable); err != nil {
		if !strings.Contains(err.Error(), "code: 390") {
			return fmt.Errorf("failed to get create table statement for %s.%s, err: %s", DatabaseNameMetric, distributedTimeSeriesTable, err)
		}
	}

	// drop table and it will be recreated with correct shard key
	if strings.Contains(showCreateTable, "cityHash64(env, temporality, metric_name, fingerprint, unix_milli)") {
		// drop table
		if err = m.DB.Exec(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s.%s ON CLUSTER %s", DatabaseNameMetric, distributedTimeSeriesTable, m.Cfg.ClusterName)); err != nil {
			return fmt.Errorf("failed to drop table %s.%s, err: %s", DatabaseNameMetric, distributedTimeSeriesTable, err)
		}
		// create table
		if err = m.DB.Exec(context.Background(), fmt.Sprintf(`CREATE TABLE IF NOT EXISTS sophonz_metrics.distributed_time_series_v4 ON CLUSTER %s AS sophonz_metrics.time_series_v4 ENGINE = Distributed("%s", sophonz_metrics, time_series_v4, cityHash64(env, temporality, metric_name, fingerprint));`, m.Cfg.ClusterName, m.Cfg.ClusterName)); err != nil {
			return fmt.Errorf("failed to create table %s.%s, err: %s", DatabaseNameMetric, distributedTimeSeriesTable, err)
		}
	}
	return nil
}

func (m *MetricsMigrator) Name() string {
	return name
}
