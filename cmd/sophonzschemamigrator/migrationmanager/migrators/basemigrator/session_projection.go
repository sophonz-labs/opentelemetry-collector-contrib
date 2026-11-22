// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package basemigrator

import (
	"context"
	"fmt"
)

// BackfillSessionProjection checks the physical parts on every replica, including
// the currently open partition. A repeat run picks up only unfinished partitions.
func (m *BaseMigrator) BackfillSessionProjection(ctx context.Context, database, table string) error {
	const missingFacts = `SELECT DISTINCT partition_id FROM (
		SELECT hostName(), name, partition_id
		FROM clusterAllReplicas(?, system.parts_columns)
		WHERE database = ? AND table = ? AND active AND rows > 0
		GROUP BY hostName(), name, partition_id
		HAVING countIf(column = 'sessionFactsV1') = 0
	) ORDER BY partition_id`
	const missingProjection = `SELECT DISTINCT partition_id
		FROM clusterAllReplicas(?, system.parts)
		WHERE database = ? AND table = ? AND active AND rows > 0
		  AND NOT has(projections, 'session_by_minute_v1')
		ORDER BY partition_id`

	// A column mutation can build the projection too. Read projection coverage
	// after the column mutations finish so completed partitions stay untouched.
	for _, step := range []struct{ query, target string }{
		{missingFacts, "COLUMN sessionFactsV1"},
		{missingProjection, "PROJECTION session_by_minute_v1"},
	} {
		var partitions []struct {
			ID string `ch:"partition_id"`
		}
		if err := m.DB.Select(ctx, &partitions, step.query, m.Cfg.ClusterName, database, table); err != nil {
			return fmt.Errorf("check %s on %s.%s: %w", step.target, database, table, err)
		}
		alter := fmt.Sprintf("ALTER TABLE %s.%s ON CLUSTER ? MATERIALIZE %s IN PARTITION ID ? SETTINGS mutations_sync=2", database, table, step.target)
		for _, partition := range partitions {
			id := partition.ID
			if err := m.DB.Exec(ctx, alter, m.Cfg.ClusterName, id); err != nil {
				return fmt.Errorf("materialize %s on %s.%s partition %s: %w", step.target, database, table, id, err)
			}
		}
	}
	return nil
}
