// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	sophonzsemconv "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/semconv"
)

// setClientInfo extracts client address information from the incoming request
// context metadata (populated by the OTLP receiver when include_metadata=true)
// and sets them as resource attributes.
//
// Expected headers (set by Envoy sidecar Lua filter):
//   - x-client-address: leftmost IP from X-Forwarded-For (original client)
//   - x-network-forwarded-for: full X-Forwarded-For value
func (p *SOPHONZAttributeProcessor) setClientInfo(ctx context.Context, attrs pcommon.Map) {
	cl := client.FromContext(ctx)
	md := cl.Metadata

	if vals := md.Get("x-client-address"); len(vals) > 0 && vals[0] != "" {
		attrs.PutStr(string(semconv.ClientAddressKey), vals[0])
	}

	if vals := md.Get("x-network-forwarded-for"); len(vals) > 0 && vals[0] != "" {
		attrs.PutStr(sophonzsemconv.NetworkForwardedFor, vals[0])
	}
}
