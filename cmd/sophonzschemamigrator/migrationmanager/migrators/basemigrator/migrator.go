package basemigrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/sophonzschemamigrator/migrationmanager/migrators"
	"go.uber.org/zap"
)

type BaseMigrator struct {
	Cfg    migrators.MigratorConfig
	Logger *zap.Logger
	DB     driver.Conn
}

func New(cfg migrators.MigratorConfig, logger *zap.Logger) (*BaseMigrator, error) {
	dbConn, err := createClickhouseConnection(cfg.DSN)
	if err != nil {
		logger.Error("Failed to create clickhouse connection", zap.Error(err))
		return nil, err
	}

	return &BaseMigrator{
		Cfg:    cfg,
		Logger: logger,
		DB:     dbConn,
	}, nil
}

func (m *BaseMigrator) Migrate(ctx context.Context, database string, migrationFolder string) error {
	if err := m.createDB(ctx, database); err != nil {
		return err
	}

	if err := m.runSqlMigrations(ctx, migrationFolder, database); err != nil {
		m.Logger.Warn("Failed to run migration", zap.String("database", database), zap.Error(err))
	}
	return nil
}

func (m *BaseMigrator) Close() error {
	return m.DB.Close()
}

func (m *BaseMigrator) createDB(ctx context.Context, database string) error {
	q := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s ON CLUSTER %s;", database, m.Cfg.ClusterName)
	err := m.DB.Exec(ctx, q)
	if err != nil {
		return fmt.Errorf("failed to create database, err: %s", err)
	}
	return nil
}

func (m *BaseMigrator) dropSchemaMigrationsTable(ctx context.Context, database string) error {
	err := m.DB.Exec(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %s.%s ON CLUSTER %s;`, database, "schema_migrations", m.Cfg.ClusterName))
	if err != nil {
		return fmt.Errorf("error dropping schema_migrations table: %v", err)
	}
	return nil
}

func (m *BaseMigrator) runSqlMigrations(ctx context.Context, migrationFolder, database string) error {
	clickhouseUrl, err := m.buildClickhouseMigrateURL(database)
	if err != nil {
		return fmt.Errorf("failed to build clickhouse migrate url, err: %s", err)
	}
	if err = m.clearDirtyState(ctx, database); err != nil {
		return fmt.Errorf("failed to clear dirty state: %v", err)
	}
	migrator, err := migrate.New("file://"+migrationFolder, clickhouseUrl)
	if err != nil {
		return fmt.Errorf("failed to create migrator, err: %s", err)
	}
	migrator.Log = newZapLoggerAdapter(m.Logger, m.Cfg.VerboseLoggingEnabled)
	migrator.EnableTemplating = true

	err = migrator.Up()
	if err != nil && !strings.HasSuffix(err.Error(), "no change") {
		return fmt.Errorf("clickhouse migrate failed to run, error: %s", err)
	}
	return nil
}

func (m *BaseMigrator) buildClickhouseMigrateURL(database string) (string, error) {
	var clickhouseUrl, migrationsTableEngine string
	options, err := clickhouse.ParseDSN(m.Cfg.DSN)
	if err != nil {
		return "", err
	}

	if m.Cfg.ReplicationEnabled {
		migrationsTableEngine = "ReplicatedMergeTree"
	} else {
		migrationsTableEngine = "MergeTree"
	}

	if len(options.Auth.Username) > 0 && len(options.Auth.Password) > 0 {
		clickhouseUrl = fmt.Sprintf("clickhouse://%s:%s@%s/%s?x-multi-statement=true&x-cluster-name=%s&x-migrations-table=schema_migrations&x-migrations-table-engine=%s", options.Auth.Username, options.Auth.Password, options.Addr[0], database, m.Cfg.ClusterName, migrationsTableEngine)
	} else {
		clickhouseUrl = fmt.Sprintf("clickhouse://%s/%s?x-multi-statement=true&x-cluster-name=%s&x-migrations-table=schema_migrations&x-migrations-table-engine=%s", options.Addr[0], database, m.Cfg.ClusterName, migrationsTableEngine)
	}
	return clickhouseUrl, nil
}

func createClickhouseConnection(dsn string) (driver.Conn, error) {
	options, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	db, err := clickhouse.Open(options)
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	return db, nil
}

func (m *BaseMigrator) clearDirtyState(ctx context.Context, database string) error {
	var tableExists uint8
	tableExistsQuery := fmt.Sprintf("SELECT 1 FROM system.tables WHERE database = '%s' AND name = 'schema_migrations'", database)
	err := m.DB.QueryRow(ctx, tableExistsQuery).Scan(&tableExists)
	if err != nil || tableExists == 0 {
		m.Logger.Info("schema_migrations table does not exist, no dirty state to check.", zap.String("database", database))
		return nil
	}

	dirtyCheckQuery := fmt.Sprintf(`SELECT version FROM %s.schema_migrations WHERE dirty = 1 LIMIT 1`, database)
	var version int64
	err = m.DB.QueryRow(ctx, dirtyCheckQuery).Scan(&version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to find dirty state: %w", err)
	}
	m.Logger.Warn("Found dirty migration state, cleaning up...", zap.Int64("version", version))

	cleanupQuery := fmt.Sprintf(`ALTER TABLE %s.schema_migrations DELETE WHERE dirty = 1`, database)
	if err = m.DB.Exec(ctx, cleanupQuery); err != nil {
		return fmt.Errorf("failed to clean dirty state: %w", err)
	}

	return nil
}
