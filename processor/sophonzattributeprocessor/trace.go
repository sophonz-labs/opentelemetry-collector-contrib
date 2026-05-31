// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func (p *SOPHONZAttributeProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	if p == nil || !p.enabled {
		return td, nil
	}

	rss := td.ResourceSpans()

	if p.serviceKeyCheck {
		rss.RemoveIf(func(rs ptrace.ResourceSpans) bool {
			return p.validateServiceKeyAndSetServiceResource(rs.Resource().Attributes())
		})
	}

	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)

		serviceNamespace, _ := rs.Resource().Attributes().Get(string(semconv.ServiceNamespaceKey))
		// set resource attribute - client.platform upper
		p.setServiceTypeUpper(rs.Resource().Attributes())
		ilss := rs.ScopeSpans()
		for j := 0; j < ilss.Len(); j++ {
			ils := ilss.At(j)
			spans := ils.Spans()
			for k := 0; k < spans.Len(); k++ {
				attrs := spans.At(k).Attributes()
				// set resource attribute - client.address, network.forwarded_for from envoy headers
				p.setClientInfo(ctx, attrs)
				// set span attribute - app.screen.id, app.screen.group.name, app.screen.group.id
				p.setScreenAttribute(serviceNamespace, attrs)
				// replace span attribute - user.id for kinfa
				if p.encryptedUserIDKINFA {
					p.decryptUserIDKINFA(attrs)
				}
			}
		}
	}

	return td, nil
}
