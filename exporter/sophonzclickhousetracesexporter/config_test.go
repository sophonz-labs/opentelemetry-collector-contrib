// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhousetracesexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap/xconfmap"
)

// TestCreateDefaultConfig verifies the default config is well-formed and valid.
func TestCreateDefaultConfig(t *testing.T) {
	cfg := createDefaultConfig()
	require.NotNil(t, cfg)

	c, ok := cfg.(*Config)
	require.True(t, ok)
	assert.Equal(t, "sophonz_traces", c.Database)
	assert.Equal(t, "sophonz", c.UserName)
	assert.True(t, c.QueueSettings.HasValue())

	require.NoError(t, xconfmap.Validate(cfg))
}
