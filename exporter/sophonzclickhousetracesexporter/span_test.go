// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhousetracesexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestPromotedDimensions(t *testing.T) {
	resource := pcommon.NewResource()
	for k, v := range map[string]string{"service.key": "test-key", "sophonz.tenant.id": "tenant-a", "service.version": "1.2", "web.version": "3.4"} {
		resource.Attributes().PutStr(k, v)
	}
	input := ptrace.NewSpan()
	input.Attributes().PutStr("messaging.system", "kafka")
	got := newStructuredSpan(input, resource, storageConfig{})
	assert.Equal(t, "tenant-a", got.TenantID)
	assert.Equal(t, "1.2", got.ServiceVersion)
	assert.Equal(t, "3.4", got.WebVersion)
	assert.Equal(t, "kafka", got.MsgSystem)
	assert.NotContains(t, got.ResourceTagsMap, "service.key")
}
