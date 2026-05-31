// Package idutil converts pdata trace/span identifiers to hex strings.
package idutil

import (
	"encoding/hex"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

// TraceIDToHexOrEmptyString returns the hex encoding of traceID, or "" when
// the trace ID is empty.
func TraceIDToHexOrEmptyString(traceID pcommon.TraceID) string {
	if !traceID.IsEmpty() {
		return hex.EncodeToString(traceID[:])
	}
	return ""
}

// SpanIDToHexOrEmptyString returns the hex encoding of spanID, or "" when the
// span ID is empty.
func SpanIDToHexOrEmptyString(spanID pcommon.SpanID) string {
	if !spanID.IsEmpty() {
		return hex.EncodeToString(spanID[:])
	}
	return ""
}
