// Package usage meters exporter throughput (count/size per tenant) and writes
// encrypted usage rows to ClickHouse. It is shared by the sophonz clickhouse
// exporters.
package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/metricexport"
)

// Options provides options for the usage collector.
type Options struct {
	// ReportingInterval is a time interval between two successive metrics export.
	ReportingInterval time.Duration
}

// Usage is a single tenant's usage sample.
type Usage struct {
	TimeStamp time.Time
	Count     int64
	Size      int64
}

// UsageCollector reads OpenCensus metrics on an interval and persists usage.
type UsageCollector struct {
	exporterID           uuid.UUID
	reader               *metricexport.Reader
	ir                   *metricexport.IntervalReader
	initReaderOnce       sync.Once
	o                    Options
	db                   clickhouse.Conn
	dbName               string
	tableName            string
	distributedTableName string
	usageParser          func(metrics []*metricdata.Metric, exporterID uuid.UUID) (map[string]Usage, error)
	prevCount            int64
	prevSize             int64
	ttl                  int
}

// CollectorID identifies this collector process for usage attribution.
var CollectorID uuid.UUID

func init() {
	CollectorID = uuid.New()
}

const cluster = "cluster"

// NewUsageCollector constructs a UsageCollector for the given exporter.
func NewUsageCollector(exporterId uuid.UUID, db clickhouse.Conn, options Options, dbName string, usageParser func(metrics []*metricdata.Metric, id uuid.UUID) (map[string]Usage, error)) *UsageCollector {
	return &UsageCollector{
		exporterID:           exporterId,
		reader:               metricexport.NewReader(),
		o:                    options,
		db:                   db,
		dbName:               dbName,
		tableName:            UsageTableName,
		distributedTableName: "distributed_" + UsageTableName,
		usageParser:          usageParser,
		prevCount:            0,
		prevSize:             0,
		ttl:                  3,
	}
}

// Start begins the interval reader.
func (e *UsageCollector) Start() error {
	e.initReaderOnce.Do(func() {
		e.ir, _ = metricexport.NewIntervalReader(&metricexport.Reader{}, e)
	})
	e.ir.ReportingInterval = e.o.ReportingInterval
	return e.ir.Start()
}

// Stop halts the interval reader.
func (c *UsageCollector) Stop() error {
	c.ir.Stop()
	return nil
}

// ExportMetrics is the OpenCensus exporter callback that persists usage.
func (e *UsageCollector) ExportMetrics(ctx context.Context, metrics []*metricdata.Metric) error {
	usages, err := e.usageParser(metrics, e.exporterID)
	if err != nil {
		return err
	}
	now := time.Now()
	for tenant, usage := range usages {
		usage.TimeStamp = now
		usageBytes, err := json.Marshal(usage)
		if err != nil {
			return err
		}
		encryptedData, err := Encrypt([]byte(e.exporterID.String())[:32], usageBytes)
		if err != nil {
			return err
		}

		// insert everything as a new row
		err = e.db.Exec(ctx, fmt.Sprintf("insert into %s.%s values ($1, $2, $3, $4, $5)", e.dbName, e.distributedTableName), tenant, CollectorID.String(), e.exporterID.String(), now, string(encryptedData))
		if err != nil {
			return err
		}
	}
	return nil
}
