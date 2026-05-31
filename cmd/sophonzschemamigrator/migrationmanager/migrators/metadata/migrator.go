package metadata

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/basemigrator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/logs"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/metrics"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators/traces"
	"go.uber.org/zap"
)

const (
	name                      = "metadata"
	database                  = "sophonz_metadata"
	migrationFolder           = "migrationmanager/migrators/metadata/migrations"
	TTLConfigurationTableName = "schema_ttl_config"
)

type MetadataMigrator struct {
	*basemigrator.BaseMigrator
}

type DatabaseTable struct {
	databaseName string
	tableName    string
}

func (m *MetadataMigrator) Migrate(ctx context.Context) error {
	if err := m.BaseMigrator.Migrate(ctx, database, migrationFolder); err != nil {
		return err
	}
	if err := m.setTTLSchemaANDInsertTables(); err != nil {
		m.Logger.Error("Failed to set ttl schema and insert tables", zap.Error(err))
	}
	return nil
}

func (m *MetadataMigrator) Name() string {
	return name
}

func (m *MetadataMigrator) setTTLSchemaANDInsertTables() error {
	ctx := context.Background()

	clusterName := os.Getenv("SOPHONZ_CLUSTER")
	if clusterName == "" {
		return fmt.Errorf("env var SOPHONZ_CLUSTER is not set")
	}

	databases := []string{traces.DatabaseNameTrace, logs.DatabaseNameLog, metrics.DatabaseNameMetric}
	var dbList []string
	for _, db := range databases {
		dbList = append(dbList, fmt.Sprintf("'%s'", db))
	}
	dbClause := strings.Join(dbList, ",")

	getTableQuery := fmt.Sprintf(`
SELECT
    database, table
FROM
    system.tables
WHERE
    database IN (%s) AND engine ILIKE '%%mergetree'
`, dbClause)
	tables := make(map[DatabaseTable]struct{})
	rows, err := m.DB.Query(ctx, getTableQuery)
	if err != nil {
		return fmt.Errorf("failed to select tables from system table: %w", err)
	}
	for rows.Next() {
		var result DatabaseTable
		if err = rows.Scan(&result.databaseName, &result.tableName); err != nil {
			return fmt.Errorf("failed to scan table lisdt: %w", err)
		}
		tables[result] = struct{}{}
	}
	if len(tables) == 0 {
		return nil
	}

	// get saved tale name list from ttl config table
	getSavedTableQuery := fmt.Sprintf(`
SELECT
	database_name, table_name
FROM
    %s.%s
`, database, TTLConfigurationTableName)

	rows, err = m.DB.Query(ctx, getSavedTableQuery)
	if err != nil {
		return fmt.Errorf("failed to select tables from %s.%s: %w", database, TTLConfigurationTableName, err)
	}
	var savedTables []DatabaseTable
	for rows.Next() {
		var result DatabaseTable
		if err = rows.Scan(&result.databaseName, &result.tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		savedTables = append(savedTables, result)
	}

	// remove already saved table name in total table name
	for _, savedTable := range savedTables {
		delete(tables, savedTable)
	}
	if len(tables) == 0 {
		return nil
	}

	// insert table name list in ttl config table
	var values []string
	for table, _ := range tables {
		values = append(values, fmt.Sprintf("('%s', '%s')", table.databaseName, table.tableName))
	}

	insertRowQuery := fmt.Sprintf(`
INSERT INTO %s.%s (database_name, table_name) VALUES %s;
`, database, TTLConfigurationTableName, strings.Join(values, ","))

	if err = m.DB.Exec(ctx, insertRowQuery); err != nil {
		return fmt.Errorf("failed to insert tables in %s.%s: %w", database, TTLConfigurationTableName, err)
	}

	// setting ttl for tables with "_5s" suffix in tableName
	suffix5stableUpdateQuery := "ALTER TABLE %s.%s UPDATE column_name = 'ts', ttl = %d, unit = 'HOUR' WHERE database_name = '%s' AND table_name = '%s';"
	for table, _ := range tables {
		if strings.HasSuffix(table.tableName, "_5s") {
			query := fmt.Sprintf(suffix5stableUpdateQuery, database, TTLConfigurationTableName, 4, table.databaseName, table.tableName)
			if err = m.DB.Exec(ctx, query); err != nil {
				return fmt.Errorf("failed to insert tables in %s.%s: %w", database, TTLConfigurationTableName, err)
			}
		}
	}

	m.Logger.Info("Successfully set up schema_ttl_config table:", zap.String("database", database))
	return nil
}
