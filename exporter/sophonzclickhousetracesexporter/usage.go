// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhousetracesexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousetracesexporter"

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
	SOPHONZSentSpansKey      = "sophonz_sent_spans"
	SOPHONZSentSpansBytesKey = "sophonz_sent_spans_bytes"
	SOPHONZSpansCount        = "sophonz_spans_count"
	SOPHONZSpansBytes        = "sophonz_spans_bytes"
)

var (
	// Measures for usage
	ExporterSOPHONZSentSpans = stats.Int64(
		SOPHONZSentSpansKey,
		"Number of SOPHONZ log records successfully sent to destination.",
		stats.UnitDimensionless)
	ExporterSOPHONZSentSpansBytes = stats.Int64(
		SOPHONZSentSpansBytesKey,
		"Total size of SOPHONZ log records successfully sent to destination.",
		stats.UnitDimensionless)

	// Views for usage
	SpansCountView = &view.View{
		Name:        SOPHONZSpansCount,
		Measure:     ExporterSOPHONZSentSpans,
		Description: "The number of spans exported to SOPHONZ",
		Aggregation: view.Sum(),
		TagKeys:     []tag.Key{usage.TagTenantKey, usage.TagExporterIdKey},
	}
	SpansCountBytesView = &view.View{
		Name:        SOPHONZSpansBytes,
		Measure:     ExporterSOPHONZSentSpansBytes,
		Description: "The size of spans exported to SOPHONZ",
		Aggregation: view.Sum(),
		TagKeys:     []tag.Key{usage.TagTenantKey, usage.TagExporterIdKey},
	}
)

func UsageExporter(metrics []*metricdata.Metric, id uuid.UUID) (map[string]usage.Usage, error) {
	data := map[string]usage.Usage{}
	for _, metric := range metrics {
		if !strings.Contains(metric.Descriptor.Name, SOPHONZSpansCount) && !strings.Contains(metric.Descriptor.Name, SOPHONZSpansBytes) {
			continue
		}
		exporterIndex := usage.GetIndexOfLabel(metric.Descriptor.LabelKeys, usage.ExporterIDKey)
		tenantIndex := usage.GetIndexOfLabel(metric.Descriptor.LabelKeys, usage.TenantKey)
		if exporterIndex == -1 || tenantIndex == -1 {
			return nil, fmt.Errorf("usage: failed to get index of labels")
		}
		if strings.Contains(metric.Descriptor.Name, SOPHONZSpansCount) {
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
		} else if strings.Contains(metric.Descriptor.Name, SOPHONZSpansBytes) {
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
