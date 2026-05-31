// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"go.opentelemetry.io/collector/pdata/pcommon"

	sophonzsemconv "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/semconv"
)

func (p *SOPHONZAttributeProcessor) setKSUID(attrs pcommon.Map) {
	p.muKSUID.Lock()
	defer p.muKSUID.Unlock()

	attrs.PutStr(sophonzsemconv.LogID, p.ksuid.String())
	p.ksuid = p.ksuid.Next()
}
