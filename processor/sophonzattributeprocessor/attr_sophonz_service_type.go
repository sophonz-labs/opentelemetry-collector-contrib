// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"go.opentelemetry.io/collector/pdata/pcommon"

	sophonzmetadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/metadata"
	sophonzsemconv "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/semconv"
)

func (p *SOPHONZAttributeProcessor) setServiceTypeUpper(attrs pcommon.Map) sophonzmetadata.ServiceType {
	serviceType, exist := attrs.Get(sophonzsemconv.ClientPlatform)
	if exist {
		st := sophonzmetadata.ToServiceType(serviceType.Str())
		attrs.Remove(sophonzsemconv.ClientPlatform)
		attrs.PutStr(sophonzsemconv.ClientPlatform, string(st))
		return st
	}
	return ""
}
