// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhousetracesexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousetracesexporter"

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

type Event struct {
	Name         string            `json:"name,omitempty"`
	TimeUnixNano uint64            `json:"timeUnixNano,omitempty"`
	AttributeMap map[string]string `json:"attributeMap,omitempty"`
	IsError      bool              `json:"isError,omitempty"`
}

func (e *Event) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", e.Name)
	enc.AddUint64("timeUnixNano", e.TimeUnixNano)
	enc.AddBool("isError", e.IsError)
	enc.AddString("attributeMap", fmt.Sprintf("%v", e.AttributeMap))
	return nil
}

type TraceModel struct {
	TraceId            string             `json:"traceId,omitempty"`
	SpanId             string             `json:"spanId,omitempty"`
	Name               string             `json:"name,omitempty"`
	DurationNano       uint64             `json:"durationNano,omitempty"`
	StartTimeUnixNano  uint64             `json:"startTimeUnixNano,omitempty"`
	EndTimeUnixNano    uint64             `json:"endTimeUnixNano,omitempty"`
	ServiceNamespace   string             `json:"serviceNamespace,omitempty"`
	ServiceName        string             `json:"serviceName,omitempty"`
	ServiceVersion     string             `json:"serviceVersion,omitempty"`
	ClientPlatform     string             `json:"clientPlatform,omitempty"`
	WebVersion         string             `json:"webVersion,omitempty"`
	SessionID          string             `json:"sessionID,omitempty"`
	DeviceID           string             `json:"deviceID,omitempty"`
	AppSpanType        string             `json:"appSpanType,omitempty"`
	AppScreenID        string             `json:"appScreenID,omitempty"`
	AppScreenName      string             `json:"appScreenName,omitempty"`
	AppScreenGroupID   string             `json:"appScreenGroupID,omitempty"`
	AppScreenGroupName string             `json:"appScreenGroupName,omitempty"`
	UserID             string             `json:"userId,omitempty"`
	Kind               int8               `json:"kind,omitempty"`
	SpanKind           string             `json:"spanKind,omitempty"`
	References         references         `json:"references,omitempty"`
	StatusCode         int16              `json:"statusCode,omitempty"`
	TagMap             map[string]string  `json:"tagMap,omitempty"`
	StringTagMap       map[string]string  `json:"stringTagMap,omitempty"`
	NumberTagMap       map[string]float64 `json:"numberTagMap,omitempty"`
	BoolTagMap         map[string]bool    `json:"boolTagMap,omitempty"`
	Events             []string           `json:"event,omitempty"`
	HasError           bool               `json:"hasError,omitempty"`
	StatusMessage      string             `json:"statusMessage,omitempty"`
	StatusCodeString   string             `json:"statusCodeString,omitempty"`
	Component          string             `json:"component,omitempty"`
}

func (t *TraceModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("traceId", t.TraceId)
	enc.AddString("spanId", t.SpanId)
	enc.AddString("name", t.Name)
	enc.AddUint64("durationNano", t.DurationNano)
	enc.AddUint64("startTimeUnixNano", t.StartTimeUnixNano)
	enc.AddUint64("endTimeUnixNano", t.EndTimeUnixNano)
	enc.AddString("serviceNamespace", t.ServiceNamespace)
	enc.AddString("serviceName", t.ServiceName)
	enc.AddString("serviceVersion", t.ServiceVersion)
	enc.AddString("clientPlatform", t.ClientPlatform)
	enc.AddString("sessionId", t.SessionID)
	enc.AddString("deviceId", t.DeviceID)
	enc.AddString("appSpanType", t.AppSpanType)
	enc.AddString("appScreenId", t.AppScreenID)
	enc.AddString("appScreenName", t.AppScreenName)
	enc.AddString("appScreenGroupId", t.AppScreenGroupID)
	enc.AddString("appScreenGroupName", t.AppScreenGroupName)
	enc.AddInt8("kind", t.Kind)
	enc.AddString("spanKind", t.SpanKind)
	enc.AddInt16("statusCode", t.StatusCode)
	enc.AddBool("hasError", t.HasError)
	enc.AddString("statusMessage", t.StatusMessage)
	enc.AddString("statusCodeString", t.StatusCodeString)
	enc.AddArray("references", &t.References)
	enc.AddString("tagMap", fmt.Sprintf("%v", t.TagMap))
	enc.AddString("event", fmt.Sprintf("%v", t.Events))
	enc.AddString("UserId", fmt.Sprintf("%v", t.UserID))
	enc.AddString("component", t.Component)
	return nil
}

type references []OtelSpanRef

func (s *references) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, e := range *s {
		err := enc.AppendObject(&e)
		if err != nil {
			return err
		}
	}
	return nil
}

type Span struct {
	TraceID            string             `json:"traceId,omitempty"`
	SpanID             string             `json:"spanId,omitempty"`
	ParentSpanID       string             `json:"parentSpanId,omitempty"`
	IsRootSpan         bool               `json:"isRootSpan,omitempty"`
	Name               string             `json:"name,omitempty"`
	DurationNano       uint64             `json:"durationNano,omitempty"`
	StartTimeUnixNano  uint64             `json:"startTimeUnixNano,omitempty"`
	EndTimeUnixNano    uint64             `json:"endTimeUnixNano,omitempty"`
	ServiceNamespace   string             `json:"serviceNamespace,omitempty"`
	ServiceName        string             `json:"serviceName,omitempty"`
	ServiceVersion     string             `json:"serviceVersion,omitempty"`
	ClientPlatform     string             `json:"clientPlatform,omitempty"`
	WebVersion         string             `json:"webVersion,omitempty"`
	Kind               int8               `json:"kind,omitempty"`
	SpanKind           string             `json:"spanKind,omitempty"`
	StatusCode         int16              `json:"statusCode,omitempty"`
	HttpURL            string             `json:"httpURL,omitempty"`
	HttpMethod         string             `json:"httpMethod,omitempty"`
	HttpHost           string             `json:"httpHost,omitempty"`
	ResponseStatusCode uint16             `json:"responseStatusCode,omitempty"`
	MsgSystem          string             `json:"msgSystem,omitempty"`
	MsgOperation       string             `json:"msgOperation,omitempty"`
	Events             []string           `json:"event,omitempty"`
	ErrorEvent         ErrorEvent         `json:"errorEvent,omitempty"`
	StringTagMap       map[string]string  `json:"stringTagMap,omitempty"`
	NumberTagMap       map[string]float64 `json:"numberTagMap,omitempty"`
	BoolTagMap         map[string]bool    `json:"boolTagMap,omitempty"`
	ResourceTagsMap    map[string]string  `json:"resourceTagsMap,omitempty"`
	HasError           bool               `json:"hasError,omitempty"`
	StatusMessage      string             `json:"statusMessage,omitempty"`
	StatusCodeString   string             `json:"statusCodeString,omitempty"`
	TraceModel         TraceModel         `json:"traceModel,omitempty"`
	Tenant             *string            `json:"-"`
	SpanAttributes     []SpanAttribute    `json:"spanAttributes,omitempty"`
	SessionID          string             `json:"sessionID,omitempty"`
	DeviceID           string             `json:"deviceID,omitempty"`
	AppSpanType        string             `json:"appSpanType,omitempty"`
	AppScreenID        uint32             `json:"appScreenID,omitempty"`
	AppScreenName      string             `json:"appScreenName,omitempty"`
	AppScreenType      string             `json:"appScreenType,omitempty"`
	AppScreenGroupID   uint32             `json:"appScreenGroupID,omitempty"`
	AppScreenGroupName string             `json:"appScreenGroupName,omitempty"`
	UserID             string             `json:"userID,omitempty"`
	Component          string             `json:"component,omitempty"`
}

type ErrorEvent struct {
	ErrorID             string `json:"errorID,omitempty"`
	GroupID             string `json:"groupID,omitempty"`
	ExceptionType       string `json:"exceptionType,omitempty"`
	ExceptionMessage    string `json:"exceptionMessage,omitempty"`
	ExceptionStacktrace string `json:"exceptionStacktrace,omitempty"`
	ExceptionEscaped    bool   `json:"exceptionEscaped,omitempty"`
	AppSpanType         string `json:"appSpanType,omitempty"`
}

func (e *ErrorEvent) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("errorId", e.ErrorID)
	enc.AddString("groupId", e.GroupID)
	enc.AddString("exceptionType", e.ExceptionType)
	enc.AddString("exceptionMessage", e.ExceptionMessage)
	enc.AddString("exceptionStacktrace", e.ExceptionStacktrace)
	enc.AddBool("exceptionEscaped", e.ExceptionEscaped)
	enc.AddString("appSpanType", e.AppSpanType)
	return nil
}

type SpanAttribute struct {
	Key         string
	TagType     string
	DataType    string
	StringValue string
	NumberValue float64
	IsColumn    bool
}

func (s *Span) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("traceId", s.TraceID)
	enc.AddString("spanId", s.SpanID)
	enc.AddString("parentSpanId", s.ParentSpanID)
	enc.AddBool("isRootSpan", s.IsRootSpan)
	enc.AddString("name", s.Name)
	enc.AddUint64("durationNano", s.DurationNano)
	enc.AddUint64("startTimeUnixNano", s.StartTimeUnixNano)
	enc.AddString("serviceNamespace", s.ServiceNamespace)
	enc.AddString("serviceName", s.ServiceName)
	enc.AddString("serviceVersion", s.ServiceVersion)
	enc.AddString("clientPlatform", s.ClientPlatform)
	enc.AddInt8("kind", s.Kind)
	enc.AddString("spanKind", s.SpanKind)
	enc.AddInt16("statusCode", s.StatusCode)
	enc.AddString("httpUrl", s.HttpURL)
	enc.AddString("httpMethod", s.HttpMethod)
	enc.AddString("httpHost", s.HttpHost)
	enc.AddUint16("responseStatusCode", s.ResponseStatusCode)
	enc.AddString("msgSystem", s.MsgSystem)
	enc.AddString("msgOperation", s.MsgOperation)
	enc.AddBool("hasError", s.HasError)
	enc.AddString("statusMessage", s.StatusMessage)
	enc.AddString("statusCodeString", s.StatusCodeString)
	enc.AddObject("errorEvent", &s.ErrorEvent)
	enc.AddObject("traceModel", &s.TraceModel)
	enc.AddString("event", fmt.Sprintf("%v", s.Events))
	enc.AddString("sessionId", s.SessionID)
	enc.AddString("deviceId", s.DeviceID)
	enc.AddString("appSpanType", s.AppSpanType)
	enc.AddUint32("appScreenId", s.AppScreenID)
	enc.AddString("appScreenName", s.AppScreenName)
	enc.AddString("appScreenType", s.AppScreenType)
	enc.AddUint32("appScreenGroupId", s.AppScreenGroupID)
	enc.AddString("appScreenGroupName", s.AppScreenGroupName)
	enc.AddString("userId", s.UserID)
	enc.AddString("component", s.Component)

	return nil
}

type OtelSpanRef struct {
	TraceId string `json:"traceId,omitempty"`
	SpanId  string `json:"spanId,omitempty"`
	RefType string `json:"refType,omitempty"`
}

func (r *OtelSpanRef) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("traceId", r.TraceId)
	enc.AddString("spanId", r.SpanId)
	enc.AddString("refType", r.RefType)
	return nil
}
