package logs

import (
	"context"

	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/basemigrator"
)

const (
	name            = "logs"
	DatabaseNameLog = "sophonz_logs"
	migrationFolder = "migrationmanager/migrators/logs/migrations"
)

type LogsMigrator struct {
	*basemigrator.BaseMigrator
}

func (m *LogsMigrator) Migrate(ctx context.Context) error {
	return m.BaseMigrator.Migrate(ctx, DatabaseNameLog, migrationFolder)
}

func (m *LogsMigrator) Name() string {
	return name
}
