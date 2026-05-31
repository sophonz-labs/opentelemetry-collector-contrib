// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhousetracesexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousetracesexporter"

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"go.opencensus.io/stats/view"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/xconfmap"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv25 "go.opentelemetry.io/otel/semconv/v1.25.0"
	semconv37 "go.opentelemetry.io/otel/semconv/v1.37.0"
	semconv9 "go.opentelemetry.io/otel/semconv/v1.9.0"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/idutil"
	sophonzsemconv "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/semconv"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/usage"
)

const (
	hasIsRemoteMask uint32 = 0x00000100
	isRemoteMask    uint32 = 0x00000200
)

// columnKeyMap is the set of attribute keys (resource and span) that are
// promoted to dedicated ClickHouse columns. Keys are stored as plain strings:
// otel semconv keys are converted with string(...), while the SOPHONZ semconv
// constants are already strings.
var columnKeyMap = map[string]struct{}{
	string(semconv37.ServiceNamespaceKey): {},
	string(semconv37.ServiceNameKey):      {},
	string(semconv37.ServiceVersionKey):   {},
	string(semconv37.URLFullKey):          {},
	string(semconv25.HTTPURLKey):          {},
	string(semconv9.HTTPHostKey):          {},
	string(semconv25.HTTPMethodKey):       {},
	string(semconv37.UserIDKey):           {},
	string(semconv37.SessionIDKey):        {},
	string(semconv37.DeviceIDKey):         {},
	sophonzsemconv.ClientPlatform:         {},
	sophonzsemconv.WebVersion:             {},
	sophonzsemconv.AppSpanType:            {},
	sophonzsemconv.AppScreenID:            {},
	sophonzsemconv.AppScreenName:          {},
	sophonzsemconv.AppScreenType:          {},
	sophonzsemconv.AppScreenGroupID:       {},
	sophonzsemconv.AppScreenGroupName:     {},
}

// newExporter creates a new exporter.
func newExporter(cfg component.Config, logger *zap.Logger) (*storage, error) {
	if err := xconfmap.Validate(cfg); err != nil {
		return nil, err
	}

	config := cfg.(*Config)

	id := uuid.New()

	var datasource string
	if config.UserName == "" || config.Password == "" {
		datasource = fmt.Sprintf("tcp://%s:%s/%s", config.Host, config.Port, config.Database)
	} else {
		datasource = fmt.Sprintf("tcp://%s:%s@%s:%s/%s", config.UserName, config.Password, config.Host, config.Port, config.Database)
	}

	numConsumers := 0
	if config.QueueSettings.HasValue() {
		numConsumers = config.QueueSettings.Get().NumConsumers
	}

	f := ClickHouseNewFactory(id, config.Migrations, datasource, config.DockerMultiNodeCluster, numConsumers)

	err := f.Initialize(logger)
	if err != nil {
		return nil, err
	}
	spanWriter, err := f.CreateSpanWriter()
	if err != nil {
		return nil, err
	}

	collector := usage.NewUsageCollector(
		id,
		f.db,
		usage.Options{ReportingInterval: usage.DefaultCollectionInterval},
		DefaultTraceDatabase,
		UsageExporter,
	)
	if err != nil {
		log.Fatalf("Error creating usage collector for traces: %v", err)
	}
	collector.Start()

	if err := view.Register(SpansCountView, SpansCountBytesView); err != nil {
		return nil, err
	}

	storage := storage{
		id:             id,
		Writer:         spanWriter,
		usageCollector: collector,
		config: storageConfig{
			lowCardinalExceptionGrouping: config.LowCardinalExceptionGrouping,
		},
		wg:        new(sync.WaitGroup),
		closeChan: make(chan struct{}),
	}

	return &storage, nil
}

type storage struct {
	id             uuid.UUID
	Writer         Writer
	usageCollector *usage.UsageCollector
	config         storageConfig
	wg             *sync.WaitGroup
	closeChan      chan struct{}
}

type storageConfig struct {
	lowCardinalExceptionGrouping bool
}

func makeJaegerProtoReferences(
	links ptrace.SpanLinkSlice,
	parentSpanID pcommon.SpanID,
	traceID pcommon.TraceID,
) []OtelSpanRef {
	parentSpanIDSet := len([8]byte(parentSpanID)) != 0
	if !parentSpanIDSet && links.Len() == 0 {
		return nil
	}

	refsCount := links.Len()
	if parentSpanIDSet {
		refsCount++
	}

	refs := make([]OtelSpanRef, 0, refsCount)

	// Put parent span ID at the first place because usually backends look for it
	// as the first CHILD_OF item in the model.SpanRef slice.
	if parentSpanIDSet {
		refs = append(refs, OtelSpanRef{
			TraceId: idutil.TraceIDToHexOrEmptyString(traceID),
			SpanId:  idutil.SpanIDToHexOrEmptyString(parentSpanID),
			RefType: "CHILD_OF",
		})
	}

	for i := 0; i < links.Len(); i++ {
		link := links.At(i)

		refs = append(refs, OtelSpanRef{
			TraceId: idutil.TraceIDToHexOrEmptyString(link.TraceID()),
			SpanId:  idutil.SpanIDToHexOrEmptyString(link.SpanID()),

			// Since Jaeger RefType is not captured in internal data,
			// use SpanRefType_FOLLOWS_FROM by default.
			// SpanRefType_CHILD_OF supposed to be set only from parentSpanID.
			RefType: "FOLLOWS_FROM",
		})
	}

	return refs
}

func populateOtherDimensions(attributes pcommon.Map, span *Span) {
	attributes.Range(func(k string, v pcommon.Value) bool {
		if k == "http.status_code" || k == "http.response.status_code" {
			// Handle both string/int http status codes.
			statusString, err := strconv.Atoi(v.Str())
			statusInt := v.Int()
			if err == nil && statusString != 0 {
				statusInt = int64(statusString)
			}
			span.ResponseStatusCode = uint16(statusInt)
		} else if k == "http.host" || k == "server.address" ||
			k == "client.address" || k == "http.request.header.host" && span.Kind == 3 {
			span.HttpHost = v.Str()
		} else if (k == "http.method" || k == "http.request.method") && span.Kind == 3 {
			span.HttpMethod = v.Str()
		} else if (k == "url.full") && span.Kind == 3 {
			valueUrl, err := url.Parse(v.Str())
			if err != nil {
				return true
			}
			span.HttpURL = valueUrl.String()
		} else if (k == "http.method" || k == "http.request.method") && span.Kind != 3 {
			span.HttpMethod = v.Str()
		} else if k == "messaging.system" {
			span.MsgSystem = v.Str()
		} else if k == "messaging.operation" {
			span.MsgOperation = v.Str()
		} else if k == "rpc.grpc.status_code" {
			// Handle both string/int status code in GRPC spans.
			statusString, err := strconv.Atoi(v.Str())
			statusInt := v.Int()
			if err == nil && statusString != 0 {
				statusInt = int64(statusString)
			}
			span.ResponseStatusCode = uint16(statusInt)
		}
		return true
	})
}

func populateEvents(events ptrace.SpanEventSlice, span *Span, lowCardinalExceptionGrouping bool) {
	for i := 0; i < events.Len(); i++ {
		event := Event{}
		event.Name = events.At(i).Name()
		event.TimeUnixNano = uint64(events.At(i).Timestamp())
		event.AttributeMap = map[string]string{}
		event.IsError = false
		events.At(i).Attributes().Range(func(k string, v pcommon.Value) bool {
			event.AttributeMap[k] = v.AsString()
			return true
		})
		stringEvent, _ := json.Marshal(event)
		span.Events = append(span.Events, string(stringEvent))
	}
}

func populateErrorEvents(attr pcommon.Map, span *Span, lowCardinalExceptionGrouping bool) {
	errorID := strings.Replace(uuid.New().String(), "-", "", -1)

	var exceptionType, exceptionMessage, exceptionStacktrace, appSpanType string
	var exceptionEscaped bool
	if v, ok := attr.Get(string(semconv37.ExceptionTypeKey)); ok {
		exceptionType = v.AsString()
	}
	if v, ok := attr.Get(string(semconv37.ExceptionMessageKey)); ok {
		exceptionMessage = v.AsString()
	}
	if v, ok := attr.Get(string(semconv37.ExceptionStacktraceKey)); ok {
		exceptionStacktrace = v.AsString()
	}
	if v, ok := attr.Get(string(semconv25.ExceptionEscapedKey)); ok {
		exceptionEscaped = v.Bool()
	}
	if v, ok := attr.Get(sophonzsemconv.AppSpanType); ok {
		appSpanType = v.AsString()
	}

	var hash [16]byte
	if lowCardinalExceptionGrouping {
		hash = md5.Sum([]byte(span.ServiceNamespace + span.ClientPlatform + exceptionType))
	} else {
		if appSpanType == "error" {
			hash = md5.Sum([]byte(span.ServiceNamespace + span.ClientPlatform + span.ServiceVersion + span.WebVersion + exceptionType + exceptionMessage))
		} else {
			hash = md5.Sum([]byte(span.ServiceNamespace + span.ClientPlatform + exceptionType + exceptionMessage))
		}
	}
	groupID := fmt.Sprintf("%x", hash)

	errorEvent := ErrorEvent{
		ErrorID:             errorID,
		GroupID:             groupID,
		ExceptionType:       exceptionType,
		ExceptionMessage:    exceptionMessage,
		ExceptionStacktrace: exceptionStacktrace,
		ExceptionEscaped:    exceptionEscaped,
		AppSpanType:         appSpanType,
	}

	span.ErrorEvent = errorEvent
}

func populateTraceModel(span *Span) {
	span.TraceModel.Component = span.Component
	span.TraceModel.Events = span.Events
	span.TraceModel.HasError = span.HasError
}

func getColumnMapValue[T string | int | bool | int64](data map[string]interface{}, key string) T {
	val, ok := data[key]
	if !ok || val == nil {
		var zero T
		return zero
	}

	v, valid := val.(T)
	if !valid {
		var zero T
		return zero
	}
	return v
}

func newStructuredSpan(otelSpan ptrace.Span, resource pcommon.Resource, config storageConfig) *Span {
	durationNano := uint64(otelSpan.EndTimestamp() - otelSpan.StartTimestamp())

	attributes := otelSpan.Attributes()
	resourceAttributes := resource.Attributes()
	tagMap := map[string]string{}
	columnMap := map[string]interface{}{}

	var spanAttributes []SpanAttribute
	stringTagMap := map[string]string{}
	numberTagMap := map[string]float64{}
	boolTagMap := map[string]bool{}
	attributes.Range(func(k string, v pcommon.Value) bool {
		tagMap[k] = v.AsString()

		if _, ok := columnKeyMap[k]; ok {
			columnMap[k] = v.AsRaw()
		}

		spanAttribute := SpanAttribute{
			Key:      k,
			TagType:  "tag",
			IsColumn: false,
		}
		if v.Type() == pcommon.ValueTypeDouble {
			numberTagMap[k] = v.Double()
			spanAttribute.NumberValue = v.Double()
			spanAttribute.DataType = "float64"
		} else if v.Type() == pcommon.ValueTypeInt {
			numberTagMap[k] = float64(v.Int())
			spanAttribute.NumberValue = float64(v.Int())
			spanAttribute.DataType = "float64"
		} else if v.Type() == pcommon.ValueTypeBool {
			boolTagMap[k] = v.Bool()
			spanAttribute.DataType = "bool"
		} else {
			stringTagMap[k] = v.AsString()
			spanAttribute.StringValue = v.AsString()
			spanAttribute.DataType = "string"
		}
		spanAttributes = append(spanAttributes, spanAttribute)

		return true
	})

	resourceAttrs := map[string]string{}
	resourceAttributes.Range(func(k string, v pcommon.Value) bool {
		if k == sophonzsemconv.ServiceKey {
			return true
		}
		tagMap[k] = v.AsString()
		resourceAttrs[k] = v.AsString()
		if _, ok := columnKeyMap[k]; ok {
			columnMap[k] = v.AsRaw()
		}

		return true
	})

	var span = &Span{
		TraceID:           idutil.TraceIDToHexOrEmptyString(otelSpan.TraceID()),
		SpanID:            idutil.SpanIDToHexOrEmptyString(otelSpan.SpanID()),
		ParentSpanID:      idutil.SpanIDToHexOrEmptyString(otelSpan.ParentSpanID()),
		IsRootSpan:        idutil.SpanIDToHexOrEmptyString(otelSpan.ParentSpanID()) == "",
		Name:              otelSpan.Name(),
		StartTimeUnixNano: uint64(otelSpan.StartTimestamp()),
		EndTimeUnixNano:   uint64(otelSpan.EndTimestamp()),
		DurationNano:      durationNano,
		ServiceNamespace:  getColumnMapValue[string](columnMap, string(semconv37.ServiceNamespaceKey)),
		ServiceName:       getColumnMapValue[string](columnMap, string(semconv37.ServiceNameKey)),
		ServiceVersion:    getColumnMapValue[string](columnMap, string(semconv37.ServiceVersionKey)),
		ClientPlatform:    getColumnMapValue[string](columnMap, sophonzsemconv.ClientPlatform),
		WebVersion:        getColumnMapValue[string](columnMap, sophonzsemconv.WebVersion),
		Kind:              int8(otelSpan.Kind()),
		SpanKind:          otelSpan.Kind().String(),
		StatusCode:        int16(otelSpan.Status().Code()),
		StringTagMap:      stringTagMap,
		NumberTagMap:      numberTagMap,
		BoolTagMap:        boolTagMap,
		ResourceTagsMap:   resourceAttrs,
		HasError:          false,
		StatusMessage:     otelSpan.Status().Message(),
		StatusCodeString:  otelSpan.Status().Code().String(),
		TraceModel: TraceModel{
			TraceId:           idutil.TraceIDToHexOrEmptyString(otelSpan.TraceID()),
			SpanId:            idutil.SpanIDToHexOrEmptyString(otelSpan.SpanID()),
			Name:              otelSpan.Name(),
			DurationNano:      durationNano,
			StartTimeUnixNano: uint64(otelSpan.StartTimestamp()),
			EndTimeUnixNano:   uint64(otelSpan.EndTimestamp()),
			ServiceNamespace:  getColumnMapValue[string](columnMap, string(semconv37.ServiceNamespaceKey)),
			ServiceName:       getColumnMapValue[string](columnMap, string(semconv37.ServiceNameKey)),
			ServiceVersion:    getColumnMapValue[string](columnMap, string(semconv37.ServiceVersionKey)),
			ClientPlatform:    getColumnMapValue[string](columnMap, sophonzsemconv.ClientPlatform),
			WebVersion:        getColumnMapValue[string](columnMap, sophonzsemconv.WebVersion),
			UserID:            getColumnMapValue[string](columnMap, string(semconv37.UserIDKey)),
			Kind:              int8(otelSpan.Kind()),
			SpanKind:          otelSpan.Kind().String(),
			References:        makeJaegerProtoReferences(otelSpan.Links(), otelSpan.ParentSpanID(), otelSpan.TraceID()),
			TagMap:            tagMap,
			StringTagMap:      stringTagMap,
			NumberTagMap:      numberTagMap,
			BoolTagMap:        boolTagMap,
			HasError:          false,
			StatusMessage:     otelSpan.Status().Message(),
			StatusCodeString:  otelSpan.Status().Code().String(),
		},
		Tenant:             usage.GetTenantNameFromResource(resource),
		SessionID:          getColumnMapValue[string](columnMap, string(semconv37.SessionIDKey)),
		DeviceID:           getColumnMapValue[string](columnMap, string(semconv37.DeviceIDKey)),
		UserID:             getColumnMapValue[string](columnMap, string(semconv37.UserIDKey)),
		AppSpanType:        strings.ToLower(getColumnMapValue[string](columnMap, sophonzsemconv.AppSpanType)),
		AppScreenID:        uint32(getColumnMapValue[int64](columnMap, sophonzsemconv.AppScreenID)),
		AppScreenName:      getColumnMapValue[string](columnMap, sophonzsemconv.AppScreenName),
		AppScreenType:      getColumnMapValue[string](columnMap, sophonzsemconv.AppScreenType),
		AppScreenGroupID:   uint32(getColumnMapValue[int64](columnMap, sophonzsemconv.AppScreenGroupID)),
		AppScreenGroupName: getColumnMapValue[string](columnMap, sophonzsemconv.AppScreenGroupName),
	}
	if otelSpan.Status().Code() == ptrace.StatusCodeError {
		span.HasError = true
		populateErrorEvents(otelSpan.Attributes(), span, config.lowCardinalExceptionGrouping)
	}
	populateOtherDimensions(attributes, span)
	setSpanComponent(span)
	populateEvents(otelSpan.Events(), span, config.lowCardinalExceptionGrouping)
	populateTraceModel(span)
	spanAttributes = append(spanAttributes, extractSpanAttributesFromSpanIndex(span)...)
	span.SpanAttributes = spanAttributes
	return span
}

// pushTraceData implements OTEL exporterhelper.traceDataPusher.
func (s *storage) pushTraceData(ctx context.Context, td ptrace.Traces) error {
	s.wg.Add(1)
	defer s.wg.Done()

	select {
	case <-s.closeChan:
		return errors.New("shutdown has been called")
	default:
		rss := td.ResourceSpans()
		var batchOfSpans []*Span
		for i := 0; i < rss.Len(); i++ {
			rs := rss.At(i)
			ilss := rs.ScopeSpans()
			for j := 0; j < ilss.Len(); j++ {
				ils := ilss.At(j)

				spans := ils.Spans()

				for k := 0; k < spans.Len(); k++ {
					span := spans.At(k)
					structuredSpan := newStructuredSpan(span, rs.Resource(), s.config)
					batchOfSpans = append(batchOfSpans, structuredSpan)
				}
			}
		}
		err := s.Writer.WriteBatchOfSpans(ctx, batchOfSpans)
		if err != nil {
			zap.S().Error("Error in writing spans to clickhouse: ", err)
			return err
		}
		return nil
	}
}

// Shutdown will shutdown the exporter.
func (s *storage) Shutdown(_ context.Context) error {
	close(s.closeChan)
	s.wg.Wait()

	if s.usageCollector != nil {
		s.usageCollector.Stop()
	}

	if closer, ok := s.Writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func extractSpanAttributesFromSpanIndex(span *Span) []SpanAttribute {
	var spanAttributes []SpanAttribute
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "traceID",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.TraceID,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "spanID",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.SpanID,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "parentSpanID",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.ParentSpanID,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "name",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.Name,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "serviceNamespace",
		TagType:     "resource",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.ServiceNamespace,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "serviceName",
		TagType:     "resource",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.ServiceName,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "serviceVersion",
		TagType:     "resource",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.ServiceVersion,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "clientPlatform",
		TagType:     "resource",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.ClientPlatform,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "kind",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "float64",
		NumberValue: float64(span.Kind),
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "spanKind",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.SpanKind,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "durationNano",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "float64",
		NumberValue: float64(span.DurationNano),
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "statusCode",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "float64",
		NumberValue: float64(span.StatusCode),
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:      "hasError",
		TagType:  "tag",
		IsColumn: true,
		DataType: "bool",
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "statusMessage",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.StatusMessage,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "statusCodeString",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.StatusCodeString,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "httpMethod",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.HttpMethod,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "httpURL",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.HttpURL,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "httpHost",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.HttpHost,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "msgSystem",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.MsgSystem,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "msgOperation",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.MsgOperation,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "responseStatusCode",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "float64",
		NumberValue: float64(span.ResponseStatusCode),
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "sessionID",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.SessionID,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "deviceID",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.DeviceID,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "appSpanType",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.AppSpanType,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "appScreenID",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "float64",
		NumberValue: float64(span.AppScreenID),
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "appScreenName",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.AppScreenName,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "appScreenType",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.AppScreenType,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "appScreenGroupID",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "float64",
		NumberValue: float64(span.AppScreenGroupID),
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "appScreenGroupName",
		TagType:     "tag",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.AppScreenGroupName,
	})
	spanAttributes = append(spanAttributes, SpanAttribute{
		Key:         "webVersion",
		TagType:     "resource",
		IsColumn:    true,
		DataType:    "string",
		StringValue: span.WebVersion,
	})
	return spanAttributes
}

func setSpanComponent(span *Span) {
	switch strings.ToLower(span.AppSpanType) {
	case "xhr", "fetch":
		span.Component = fmt.Sprintf("[%d] %s %s", span.ResponseStatusCode, span.HttpMethod, span.HttpURL)
	case "event":
		element := getElementFromStringTagMap(span.StringTagMap)
		if element != "" {
			span.Component = fmt.Sprintf("%s [%s]", element, span.Name)
		} else {
			span.Component = fmt.Sprintf("[%s]", span.Name)
		}
	default:
		span.Component = span.Name
	}
}

func getElementFromStringTagMap(tagMap map[string]string) string {
	if element, ok := tagMap["target_element_text"]; ok {
		return element
	}
	if element, ok := tagMap["target_element_id"]; ok {
		return element
	}
	if element, ok := tagMap["target_element"]; ok {
		return element
	}
	return ""
}
