// Package metadata maintains an in-memory, periodically-refreshed cache of
// screen/service metadata read from ClickHouse (the sophonz_metadata
// database). It is consumed by the sophonz attribute processor to enrich
// telemetry with screen and service context.
package metadata

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/schema"
)

const (
	screenView  = "v_postgres_screen"
	serviceView = "v_postgres_service"
)

// MetadataManager periodically syncs screen and service metadata from
// ClickHouse and exposes it via lock-free atomic reads.
type MetadataManager struct {
	DB           driver.Conn
	Screen       atomic.Value
	Service      atomic.Value
	logger       *zap.Logger
	syncInterval time.Duration
	stopChan     chan struct{}
}

// NewManager opens a ClickHouse connection, performs an initial fetch and
// returns a manager ready to be Start()ed.
func NewManager(addr, port, database, username, password string, interval time.Duration, logger *zap.Logger) (*MetadataManager, error) {
	manager := &MetadataManager{
		logger:       logger,
		syncInterval: interval,
		stopChan:     make(chan struct{}),
	}

	conn, err := clickhouseConn(addr, port, database, username, password)
	if err != nil {
		return nil, err
	}
	manager.DB = conn

	initCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = manager.getScreens(initCtx); err != nil {
		manager.logger.Warn("fail to fetch screen data", zap.Error(err))
	}
	if err = manager.getServices(initCtx); err != nil {
		manager.logger.Warn("fail to fetch service data", zap.Error(err))
	}

	return manager, nil
}

// Start runs the sync loop until Stop is called.
func (pm *MetadataManager) Start() {
	ticker := time.NewTicker(pm.syncInterval)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-pm.stopChan:
			cancel()
			return
		case <-ticker.C:
			func() {
				opCtx, opCancel := context.WithTimeout(ctx, 30*time.Second)
				defer opCancel()

				if err := pm.getScreens(opCtx); err != nil {
					pm.logger.Warn("fail to fetch screen data", zap.Error(err))
					return
				}
				if err := pm.getServices(opCtx); err != nil {
					pm.logger.Warn("fail to fetch service data", zap.Error(err))
					return
				}
			}()
		}
	}
}

// Stop terminates the sync loop.
func (pm *MetadataManager) Stop() {
	pm.logger.Info("metadata manager stop")
	close(pm.stopChan)
}

func clickhouseConn(addr, port, database, userName, password string) (driver.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", addr, port)},
		Auth: clickhouse.Auth{Database: database, Username: userName, Password: password},
		DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		},
		Debug: false,
		Debugf: func(format string, v ...any) {
			fmt.Printf(format+"\n", v...)
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:          time.Second * 30,
		MaxOpenConns:         5,
		MaxIdleConns:         5,
		ConnMaxLifetime:      time.Duration(10) * time.Minute,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
	})
	if err != nil {
		return nil, err
	}
	return conn, conn.Ping(context.Background())
}

func (pm *MetadataManager) getScreens(ctx context.Context) error {
	var screens []Screen
	query := fmt.Sprintf(`
SELECT
    id,
    screenName,
    screenType,
    groupID,
    groupName,
    serviceNamespace
FROM
    %s.%s
    `, schema.DatabaseMetadata, screenView)
	if err := pm.DB.Select(ctx, &screens, query); err != nil {
		return err
	}
	if len(screens) > 0 {
		pm.Screen.Store(ConvertToScreenMap(screens))
	}

	return nil
}

func (pm *MetadataManager) getServices(ctx context.Context) error {
	var services []Service
	query := fmt.Sprintf("SELECT id, serviceNamespace, serviceName, type, key FROM %s.%s", schema.DatabaseMetadata, serviceView)
	if err := pm.DB.Select(ctx, &services, query); err != nil {
		return err
	}

	if len(services) > 0 {
		pm.Service.Store(ConvertToServiceMap(services))
	}

	return nil
}
