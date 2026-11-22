// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package basemigrator

import (
	"context"
	"fmt"
)

type sessionProjection struct {
	column string
	name   string
}

// BackfillSessionProjection checks physical parts on every replica. New parts
// already have both projections. Retried runs resume at the oldest incomplete
// daily partition and finish both mutations before starting the next partition.
func (m *BaseMigrator) BackfillSessionProjection(ctx context.Context, database, table string) error {
	for _, projection := range []sessionProjection{
		{"sessionFactsV1", "session_by_minute_v1"},
		{"sessionSdkV1", "session_by_sdk_minute_v1"},
	} {
		if err := m.backfillProjection(ctx, database, table, projection); err != nil {
			return err
		}
	}
	return nil
}

func (m *BaseMigrator) backfillProjection(ctx context.Context, database, table string, projection sessionProjection) error {
	// Parts use YYYYMMDD partition IDs. Include the currently open partition.
	const missingColumn = `SELECT partition_id FROM clusterAllReplicas(?, system.parts_columns)
		WHERE database = ? AND table = ? AND active AND rows > 0
		GROUP BY hostName(), name, partition_id HAVING countIf(column = ?) = 0`
	const missingProjection = `SELECT partition_id FROM clusterAllReplicas(?, system.parts)
		WHERE database = ? AND table = ? AND active AND rows > 0 AND NOT has(projections, ?)`
	candidates := "SELECT DISTINCT partition_id FROM (" + missingColumn + " UNION ALL " + missingProjection + ") ORDER BY partition_id"
	cluster := m.Cfg.ClusterName
	var partitions []struct {
		ID string `ch:"partition_id"`
	}
	if err := m.DB.Select(ctx, &partitions, candidates, cluster, database, table, projection.column, cluster, database, table, projection.name); err != nil {
		return fmt.Errorf("find missing %s on %s.%s: %w", projection.name, database, table, err)
	}

	for _, partition := range partitions {
		for _, step := range []struct{ query, objectType, name string }{
			{missingColumn, "COLUMN", projection.column}, {missingProjection, "PROJECTION", projection.name},
		} {
			// Inserts and background merges can finish a candidate while earlier
			// partitions are running. A column mutation can also build a projection.
			check := "SELECT count() FROM (" + step.query + ") WHERE partition_id = ?"
			var missing uint64
			if err := m.DB.QueryRow(ctx, check, cluster, database, table, step.name, partition.ID).Scan(&missing); err != nil {
				return fmt.Errorf("check %s on %s.%s partition %s: %w", step.name, database, table, partition.ID, err)
			}
			if missing == 0 {
				continue
			}
			alter := fmt.Sprintf("ALTER TABLE %s.%s ON CLUSTER ? MATERIALIZE %s %s IN PARTITION ID ? SETTINGS mutations_sync=2", database, table, step.objectType, step.name)
			if err := m.DB.Exec(ctx, alter, cluster, partition.ID); err != nil {
				return fmt.Errorf("materialize %s on %s.%s partition %s: %w", step.name, database, table, partition.ID, err)
			}
		}
	}
	return nil
}
