// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhousemetricsexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousemetricsexporter"

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/usage"
)

const (
	SOPHONZSentMetricPointsKey      = "sophonz_sent_metric_points"
	SOPHONZSentMetricPointsBytesKey = "sophonz_sent_metric_points_bytes"
	SOPHONZMetricPointsCount        = "sophonz_metric_points_count"
	SOPHONZMetricPointsBytes        = "sophonz_metric_points_bytes"
)

var (
	// Measures for usage
	ExporterSOPHONZSentMetricPoints = stats.Int64(
		SOPHONZSentMetricPointsKey,
		"Number of SOPHONZ metric points successfully sent to destination.",
		stats.UnitDimensionless)
	ExporterSOPHONZSentMetricPointsBytes = stats.Int64(
		SOPHONZSentMetricPointsBytesKey,
		"Total size of SOPHONZ metric points successfully sent to destination.",
		stats.UnitDimensionless)

	// Views for usage
	MetricPointsCountView = &view.View{
		Name:        SOPHONZMetricPointsCount,
		Measure:     ExporterSOPHONZSentMetricPoints,
		Description: "The number of metric points exported to SOPHONZ",
		Aggregation: view.Sum(),
		TagKeys:     []tag.Key{usage.TagTenantKey, usage.TagExporterIdKey},
	}
	MetricPointsBytesView = &view.View{
		Name:        SOPHONZMetricPointsBytes,
		Measure:     ExporterSOPHONZSentMetricPointsBytes,
		Description: "The size of metric points exported to SOPHONZ",
		Aggregation: view.Sum(),
		TagKeys:     []tag.Key{usage.TagTenantKey, usage.TagExporterIdKey},
	}
)

func UsageExporter(metrics []*metricdata.Metric, id uuid.UUID) (map[string]usage.Usage, error) {
	data := map[string]usage.Usage{}
	for _, metric := range metrics {
		if !strings.Contains(metric.Descriptor.Name, SOPHONZMetricPointsCount) && !strings.Contains(metric.Descriptor.Name, SOPHONZMetricPointsBytes) {
			continue
		}
		exporterIndex := usage.GetIndexOfLabel(metric.Descriptor.LabelKeys, usage.ExporterIDKey)
		tenantIndex := usage.GetIndexOfLabel(metric.Descriptor.LabelKeys, usage.TenantKey)
		if exporterIndex == -1 || tenantIndex == -1 {
			return nil, fmt.Errorf("usage: failed to get index of labels")
		}
		if strings.Contains(metric.Descriptor.Name, SOPHONZMetricPointsCount) {
			for _, v := range metric.TimeSeries {
				if v.LabelValues[exporterIndex].Value != id.String() {
					continue
				}
				tenant := v.LabelValues[tenantIndex].Value
				if d, ok := data[tenant]; ok {
					d.Count = v.Points[0].Value.(int64)
					data[tenant] = d
				} else {
					data[tenant] = usage.Usage{
						Count: v.Points[0].Value.(int64),
					}
				}
			}
		} else if strings.Contains(metric.Descriptor.Name, SOPHONZMetricPointsBytes) {
			for _, v := range metric.TimeSeries {
				if v.LabelValues[exporterIndex].Value != id.String() {
					continue
				}
				tenant := v.LabelValues[tenantIndex].Value
				if d, ok := data[tenant]; ok {
					d.Size = v.Points[0].Value.(int64)
					data[tenant] = d
				} else {
					data[tenant] = usage.Usage{
						Size: v.Points[0].Value.(int64),
					}
				}
			}
		}
	}
	return data, nil
}
