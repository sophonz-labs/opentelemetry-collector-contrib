// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzconsoleprintprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzconsoleprintprocessor"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

func (p *sophonzConsolePrintProcessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	if p.logEnabled {
		m := plog.JSONMarshaler{}
		b, err := m.MarshalLogs(ld)
		if err != nil {
			p.logger.Warn("Failed to marshal log", zap.Error(err))
			return ld, nil
		}
		var pretty bytes.Buffer
		if err = json.Indent(&pretty, b, "", "  "); err != nil {
			p.logger.Warn("Failed to indent log", zap.Error(err))
			return ld, nil
		}
		fmt.Println("================================================================================")
		fmt.Println(pretty.String())
	}
	return ld, nil
}
