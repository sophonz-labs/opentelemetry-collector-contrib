// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/crypto"
)

func (p *SOPHONZAttributeProcessor) decryptUserIDKINFA(attrs pcommon.Map) {
	userID, ok := attrs.Get(string(semconv.UserIDKey))
	if ok && userID.Str() != "" {
		decryptedStr, err := crypto.Decrypt(userID.Str())
		if decryptedStr == "" || err != nil {
			return
		}
		attrs.Remove(string(semconv.UserIDKey))
		attrs.PutStr(string(semconv.UserIDKey), decryptedStr)
	}
}
