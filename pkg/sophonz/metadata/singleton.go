package metadata

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type metadataManagerSingleton struct {
	lock     sync.Mutex
	manager  *MetadataManager
	refCount int
	logger   *zap.Logger
}

var metadataManagerInstance = &metadataManagerSingleton{}

// GetSharedMetadataManager returns a process-wide shared MetadataManager,
// creating and starting it on first use and incrementing a reference count.
func GetSharedMetadataManager(addr, port, database, username, password string, interval time.Duration, logger *zap.Logger) (*MetadataManager, error) {
	metadataManagerInstance.lock.Lock()
	defer metadataManagerInstance.lock.Unlock()

	if metadataManagerInstance.manager == nil {
		smLogger := zap.New(logger.Core()).With(
			zap.String("module", "metadataManager"),
		)
		metadataManagerInstance.logger = smLogger

		pm, err := NewManager(addr, port, database, username, password, interval, metadataManagerInstance.logger)
		if err != nil {
			return nil, err
		}
		metadataManagerInstance.manager = pm

		go pm.Start()
		logger.Info("Shared MetadataManager has started.")
	}

	metadataManagerInstance.refCount++
	logger.Info("MetadataManager reference count increased", zap.Int("refCount", metadataManagerInstance.refCount))
	return metadataManagerInstance.manager, nil
}

// ReleaseMetadataManager decrements the reference count and stops the shared
// manager once no references remain.
func ReleaseMetadataManager() {
	metadataManagerInstance.lock.Lock()
	defer metadataManagerInstance.lock.Unlock()

	if metadataManagerInstance.manager == nil {
		return
	}

	metadataManagerInstance.refCount--
	if metadataManagerInstance.logger != nil {
		metadataManagerInstance.logger.Debug("MetadataManager reference count decreased", zap.Int("refCount", metadataManagerInstance.refCount))
	}

	if metadataManagerInstance.refCount <= 0 {
		metadataManagerInstance.manager.Stop()
		metadataManagerInstance.manager = nil
		if metadataManagerInstance.logger != nil {
			metadataManagerInstance.logger.Info("Shared MetadataManager has stopped.")
		}
	}
}
