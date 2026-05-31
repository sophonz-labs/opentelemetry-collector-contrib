// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzconsoleprintprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzconsoleprintprocessor"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func (p *sophonzConsolePrintProcessor) processTraces(_ context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	if p.traceEnabled {
		m := ptrace.JSONMarshaler{}
		b, err := m.MarshalTraces(td)
		if err != nil {
			p.logger.Warn("Failed to marshal trace", zap.Error(err))
			return td, nil
		}
		var pretty bytes.Buffer
		if err = json.Indent(&pretty, b, "", "  "); err != nil {
			p.logger.Warn("Failed to indent trace", zap.Error(err))
			return td, nil
		}
		fmt.Println("================================================================================")
		fmt.Println(pretty.String())
	}
	return td, nil
}
