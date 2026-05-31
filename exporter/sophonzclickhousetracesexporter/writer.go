// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhousetracesexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousetracesexporter"

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/collector/pipeline"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/usage"
)

type Encoding string

const (
	// EncodingJSON is used for spans encoded as JSON.
	EncodingJSON Encoding = "json"
	// EncodingProto is used for spans encoded as Protobuf.
	EncodingProto Encoding = "protobuf"
)

// SpanWriter for writing spans to ClickHouse.
type SpanWriter struct {
	logger            *zap.Logger
	db                clickhouse.Conn
	traceDatabase     string
	indexTable        string
	errorTable        string
	spansTable        string
	attributeTable    string
	attributeKeyTable string
	encoding          Encoding
	exporterId        uuid.UUID
}

type WriterOptions struct {
	logger            *zap.Logger
	db                clickhouse.Conn
	traceDatabase     string
	spansTable        string
	indexTable        string
	errorTable        string
	attributeTable    string
	attributeKeyTable string
	encoding          Encoding
	exporterId        uuid.UUID
}

var skipKeys = map[string]struct{}{
	"traceID":      {},
	"spanID":       {},
	"parentSpanID": {},
	"sessionID":    {},
	"durationNano": {},
	"user.id":      {},
	"session.id":   {},
	"device.id":    {},
}

// NewSpanWriter returns a SpanWriter for the database.
func NewSpanWriter(options WriterOptions) *SpanWriter {
	if err := view.Register(SpansCountView, SpansCountBytesView); err != nil {
		return nil
	}
	writer := &SpanWriter{
		logger:            options.logger,
		db:                options.db,
		traceDatabase:     options.traceDatabase,
		indexTable:        options.indexTable,
		errorTable:        options.errorTable,
		spansTable:        options.spansTable,
		attributeTable:    options.attributeTable,
		attributeKeyTable: options.attributeKeyTable,
		encoding:          options.encoding,
		exporterId:        options.exporterId,
	}

	return writer
}

func (w *SpanWriter) writeIndexBatch(ctx context.Context, batchSpans []*Span) error {
	var statement driver.Batch
	var err error

	defer func() {
		if statement != nil {
			_ = statement.Abort()
		}
	}()
	statement, err = w.db.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s.%s", w.traceDatabase, w.indexTable), driver.WithReleaseConnection())
	if err != nil {
		w.logger.Error("Could not prepare batch for index table: ", zap.Error(err))
		return err
	}

	for _, span := range batchSpans {
		err = statement.Append(
			time.Unix(0, int64(span.StartTimeUnixNano)),
			time.Unix(0, int64(span.EndTimeUnixNano)),
			span.TraceID,
			span.SpanID,
			span.ParentSpanID,
			span.IsRootSpan,
			span.SessionID,
			span.DeviceID,
			span.ServiceNamespace,
			span.ClientPlatform,
			span.ServiceVersion,
			span.ServiceName,
			span.WebVersion,
			span.Name,
			span.Component,
			span.AppSpanType,
			span.AppScreenID,
			span.AppScreenName,
			span.AppScreenType,
			span.AppScreenGroupID,
			span.AppScreenGroupName,
			span.DurationNano,
			span.UserID,
			span.StatusCode,
			span.StatusCodeString,
			span.StatusMessage,
			span.Events,
			span.HttpMethod,
			span.HttpURL,
			span.HttpHost,
			span.ResponseStatusCode,
			span.MsgSystem,
			span.MsgOperation,
			span.HasError,
			span.StringTagMap,
			span.NumberTagMap,
			span.BoolTagMap,
			span.ResourceTagsMap,
			span.Kind,
			span.SpanKind,
		)
		if err != nil {
			w.logger.Error("Could not append span to batch: ", zap.Object("span", span), zap.Error(err))
			return err
		}
	}

	start := time.Now()

	err = statement.Send()

	ctx, _ = tag.New(ctx,
		tag.Upsert(exporterKey, pipeline.SignalTraces.String()),
		tag.Upsert(tableKey, w.indexTable),
	)
	stats.Record(ctx, writeLatencyMillis.M(int64(time.Since(start).Milliseconds())))
	return err
}

func (w *SpanWriter) writeTagBatch(ctx context.Context, batchSpans []*Span) error {
	var tagKeyStatement driver.Batch
	var tagStatement driver.Batch
	var err error

	defer func() {
		if tagKeyStatement != nil {
			_ = tagKeyStatement.Abort()
		}
		if tagStatement != nil {
			_ = tagStatement.Abort()
		}
	}()
	tagStatement, err = w.db.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s.%s", w.traceDatabase, w.attributeTable), driver.WithReleaseConnection())
	if err != nil {
		w.logger.Error("Could not prepare batch for span attributes table due to error: ", zap.Error(err))
		return err
	}
	tagKeyStatement, err = w.db.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s.%s", w.traceDatabase, w.attributeKeyTable), driver.WithReleaseConnection())
	if err != nil {
		w.logger.Error("Could not prepare batch for span attributes key table due to error: ", zap.Error(err))
		return err
	}
	// create map of span attributes of key, tagType, dataType and isColumn to avoid duplicates in batch
	mapOfSpanAttributeKeys := make(map[string]struct{})

	// create map of span attributes of key, tagType, dataType, isColumn and value to avoid duplicates in batch
	mapOfSpanAttributeValues := make(map[string]struct{})

	for _, span := range batchSpans {
		for _, spanAttribute := range span.SpanAttributes {

			// form a map key of span attribute key, tagType, dataType, isColumn and value
			mapOfSpanAttributeValueKey := spanAttribute.Key +
				spanAttribute.TagType + spanAttribute.DataType +
				strconv.FormatBool(spanAttribute.IsColumn) +
				spanAttribute.StringValue +
				strconv.FormatFloat(spanAttribute.NumberValue, 'f', -1, 64) +
				span.ServiceNamespace

			// check if mapOfSpanAttributeValueKey already exists in map
			_, ok := mapOfSpanAttributeValues[mapOfSpanAttributeValueKey]
			if ok {
				continue
			}
			// add mapOfSpanAttributeValueKey to map
			mapOfSpanAttributeValues[mapOfSpanAttributeValueKey] = struct{}{}

			// form a map key of span attribute key, tagType, dataType and isColumn
			mapOfSpanAttributeKey := spanAttribute.Key +
				spanAttribute.TagType +
				spanAttribute.DataType +
				strconv.FormatBool(spanAttribute.IsColumn) +
				span.ServiceNamespace

			// check if mapOfSpanAttributeKey already exists in map
			_, ok = mapOfSpanAttributeKeys[mapOfSpanAttributeKey]
			if !ok {
				err = tagKeyStatement.Append(
					span.ServiceNamespace,
					spanAttribute.Key,
					spanAttribute.TagType,
					spanAttribute.DataType,
					spanAttribute.IsColumn,
				)
				if err != nil {
					w.logger.Error("Could not append span to tagKey Statement to batch due to error: ", zap.Error(err), zap.Object("span", span))
					return err
				}
			}
			// add mapOfSpanAttributeKey to map
			mapOfSpanAttributeKeys[mapOfSpanAttributeKey] = struct{}{}

			if _, skip := skipKeys[spanAttribute.Key]; skip {
				continue
			}

			if spanAttribute.DataType == "string" {
				err = tagStatement.Append(
					time.Unix(0, int64(span.StartTimeUnixNano)),
					span.ServiceNamespace,
					spanAttribute.Key,
					spanAttribute.TagType,
					spanAttribute.DataType,
					spanAttribute.StringValue,
					nil,
					spanAttribute.IsColumn,
				)
			} else if spanAttribute.DataType == "float64" {
				err = tagStatement.Append(
					time.Unix(0, int64(span.StartTimeUnixNano)),
					span.ServiceNamespace,
					spanAttribute.Key,
					spanAttribute.TagType,
					spanAttribute.DataType,
					nil,
					spanAttribute.NumberValue,
					spanAttribute.IsColumn,
				)
			} else if spanAttribute.DataType == "bool" {
				err = tagStatement.Append(
					time.Unix(0, int64(span.StartTimeUnixNano)),
					span.ServiceNamespace,
					spanAttribute.Key,
					spanAttribute.TagType,
					spanAttribute.DataType,
					nil,
					nil,
					spanAttribute.IsColumn,
				)
			}
			if err != nil {
				w.logger.Error("Could not append span to tag Statement batch due to error: ", zap.Error(err), zap.Object("span", span))
				return err
			}
		}
	}

	tagStart := time.Now()
	err = tagStatement.Send()
	stats.RecordWithTags(ctx,
		[]tag.Mutator{
			tag.Upsert(exporterKey, pipeline.SignalTraces.String()),
			tag.Upsert(tableKey, w.attributeTable),
		},
		writeLatencyMillis.M(int64(time.Since(tagStart).Milliseconds())),
	)
	if err != nil {
		w.logger.Error("Could not write to span attributes table due to error: ", zap.Error(err))
		return err
	}

	tagKeyStart := time.Now()
	err = tagKeyStatement.Send()
	stats.RecordWithTags(ctx,
		[]tag.Mutator{
			tag.Upsert(exporterKey, pipeline.SignalTraces.String()),
			tag.Upsert(tableKey, w.attributeKeyTable),
		},
		writeLatencyMillis.M(int64(time.Since(tagKeyStart).Milliseconds())),
	)
	if err != nil {
		w.logger.Error("Could not write to span attributes key table due to error: ", zap.Error(err))
		return err
	}

	return err
}

func (w *SpanWriter) writeErrorBatch(ctx context.Context, batchSpans []*Span) error {
	var statement driver.Batch
	var err error

	defer func() {
		if statement != nil {
			_ = statement.Abort()
		}
	}()
	statement, err = w.db.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s.%s", w.traceDatabase, w.errorTable), driver.WithReleaseConnection())
	if err != nil {
		w.logger.Error("Could not prepare batch for error table: ", zap.Error(err))
		return err
	}

	for _, span := range batchSpans {
		if !span.HasError {
			continue
		}
		err = statement.Append(
			time.Unix(0, int64(span.StartTimeUnixNano)),
			span.ErrorEvent.ErrorID,
			span.ErrorEvent.GroupID,
			span.TraceID,
			span.SpanID,
			span.ServiceNamespace,
			span.ClientPlatform,
			span.ServiceVersion,
			span.ServiceName,
			span.WebVersion,
			span.SessionID,
			span.DeviceID,
			span.ErrorEvent.AppSpanType,
			span.ErrorEvent.ExceptionType,
			span.ErrorEvent.ExceptionMessage,
			span.ErrorEvent.ExceptionStacktrace,
			span.ErrorEvent.ExceptionEscaped,
			span.ResourceTagsMap,
		)
		if err != nil {
			w.logger.Error("Could not append span to batch: ", zap.Object("span", span), zap.Error(err))
			return err
		}
	}

	start := time.Now()

	err = statement.Send()

	ctx, _ = tag.New(ctx,
		tag.Upsert(exporterKey, pipeline.SignalTraces.String()),
		tag.Upsert(tableKey, w.errorTable),
	)
	stats.Record(ctx, writeLatencyMillis.M(int64(time.Since(start).Milliseconds())))
	return err
}

func (w *SpanWriter) writeModelBatch(ctx context.Context, batchSpans []*Span) error {
	var statement driver.Batch
	var err error

	defer func() {
		if statement != nil {
			_ = statement.Abort()
		}
	}()

	statement, err = w.db.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s.%s", w.traceDatabase, w.spansTable), driver.WithReleaseConnection())
	if err != nil {
		w.logger.Error("Could not prepare batch for model table: ", zap.Error(err))
		return err
	}

	metrics := map[string]usage.Metric{}
	for _, span := range batchSpans {
		var serialized []byte
		usageMap := span.TraceModel
		usageMap.TagMap = map[string]string{}
		serialized, err = json.Marshal(span.TraceModel)
		if err != nil {
			return err
		}
		serializedUsage, err := json.Marshal(usageMap)
		if err != nil {
			return err
		}

		err = statement.Append(time.Unix(0, int64(span.StartTimeUnixNano)), span.TraceID, string(serialized))
		if err != nil {
			w.logger.Error("Could not append span to batch: ", zap.Object("span", span), zap.Error(err))
			return err
		}

		usage.AddMetric(metrics, *span.Tenant, 1, int64(len(serializedUsage)))
	}
	start := time.Now()

	err = statement.Send()
	ctx, _ = tag.New(ctx,
		tag.Upsert(exporterKey, pipeline.SignalTraces.String()),
		tag.Upsert(tableKey, w.spansTable),
	)
	stats.Record(ctx, writeLatencyMillis.M(int64(time.Since(start).Milliseconds())))
	if err != nil {
		return err
	}
	for k, v := range metrics {
		stats.RecordWithTags(ctx, []tag.Mutator{tag.Upsert(usage.TagTenantKey, k), tag.Upsert(usage.TagExporterIdKey, w.exporterId.String())}, ExporterSOPHONZSentSpans.M(int64(v.Count)), ExporterSOPHONZSentSpansBytes.M(int64(v.Size)))
	}

	return nil
}

// WriteBatchOfSpans writes the encoded batch of spans.
func (w *SpanWriter) WriteBatchOfSpans(ctx context.Context, batch []*Span) error {
	if w.spansTable != "" {
		if err := w.writeModelBatch(ctx, batch); err != nil {
			w.logger.Error("Could not write a batch of spans to model table: ", zap.Error(err))
			return err
		}
	}
	if w.indexTable != "" {
		if err := w.writeIndexBatch(ctx, batch); err != nil {
			w.logger.Error("Could not write a batch of spans to index table: ", zap.Error(err))
			return err
		}
	}
	if w.errorTable != "" {
		if err := w.writeErrorBatch(ctx, batch); err != nil {
			w.logger.Error("Could not write a batch of spans to error table: ", zap.Error(err))
			return err
		}
	}
	if w.attributeTable != "" && w.attributeKeyTable != "" {
		if err := w.writeTagBatch(ctx, batch); err != nil {
			w.logger.Error("Could not write a batch of spans to tag/tagKey tables: ", zap.Error(err))
			return err
		}
	}
	return nil
}

// Close closes the writer.
func (w *SpanWriter) Close() error {
	if w.db != nil {
		return w.db.Close()
	}
	return nil
}
