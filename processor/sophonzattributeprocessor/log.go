// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/pdata/plog"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func (p *SOPHONZAttributeProcessor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	if p == nil || !p.enabled {
		return ld, nil
	}

	rls := ld.ResourceLogs()

	if p.serviceKeyCheck {
		rls.RemoveIf(func(rs plog.ResourceLogs) bool {
			return p.validateServiceKeyAndSetServiceResource(rs.Resource().Attributes())
		})
	}

	for i := 0; i < rls.Len(); i++ {
		rs := rls.At(i)

		serviceNamespace, _ := rs.Resource().Attributes().Get(string(semconv.ServiceNamespaceKey))
		// set resource attribute - client.platform upper
		p.setServiceTypeUpper(rs.Resource().Attributes())
		ilss := rs.ScopeLogs()
		for j := 0; j < ilss.Len(); j++ {
			ils := ilss.At(j)
			logs := ils.LogRecords()
			for k := 0; k < logs.Len(); k++ {
				lr := logs.At(k)
				attrs := lr.Attributes()
				// set resource attribute - client.address, network.forwarded_for from envoy headers
				p.setClientInfo(ctx, attrs)
				// set attribute(app.screen.id, app.screen.group.name, app.screen.group.id)
				p.setScreenAttribute(serviceNamespace, attrs)
				// set attribute(log.id)
				p.setKSUID(attrs)
				// replace attribute(user.id) for kinfa
				if p.encryptedUserIDKINFA {
					p.decryptUserIDKINFA(attrs)
				}
			}
		}
	}

	return ld, nil
}
