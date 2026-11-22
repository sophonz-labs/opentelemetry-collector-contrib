// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor

import (
	"testing"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/crypto"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/metadata"
	sc "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/semconv"
)

func TestTenantEnforcementAcrossSignals(t *testing.T) {
	manager := &metadata.MetadataManager{}
	manager.Service.Store(metadata.ServiceKeyMap{
		"test-a":     {CompanyID: "tenant-a"},
		"test-b":     {CompanyID: "tenant-b"},
		"restricted": {CompanyID: "tenant-c", AllowedOrigins: []string{"https://allowed.example"}, OriginEnforced: true},
	})
	p := &SOPHONZAttributeProcessor{enabled: true, logger: zap.NewNop(), MetadataManager: manager, keyMode: ServiceKeyModeEnforce, tenant: &tenantCounters{}}
	ctx := client.NewContext(t.Context(), client.Info{Metadata: client.NewMetadata(map[string][]string{"origin": {"https://app.example"}})})
	traces, logs := ptrace.NewTraces(), plog.NewLogs()
	for _, key := range []string{"test-a", "invalid", "test-b", "restricted", ""} {
		rs, rl := traces.ResourceSpans().AppendEmpty(), logs.ResourceLogs().AppendEmpty()
		rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
		rl.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
		rs.Resource().Attributes().PutStr(sc.ServiceKey, key)
		rs.Resource().Attributes().PutStr(sc.TenantID, "untrusted")
		rs.Resource().CopyTo(rl.Resource())
	}
	td, err := p.processTraces(ctx, traces)
	if err != nil {
		t.Fatal(err)
	}
	ld, err := p.processLogs(ctx, logs)
	if err != nil {
		t.Fatal(err)
	}
	if td.SpanCount() != 2 || ld.LogRecordCount() != 2 {
		t.Fatalf("kept spans %d logs %d", td.SpanCount(), ld.LogRecordCount())
	}
	for i, want := range []string{"tenant-a", "tenant-b"} {
		traceTenant, _ := td.ResourceSpans().At(i).Resource().Attributes().Get(sc.TenantID)
		logTenant, _ := ld.ResourceLogs().At(i).Resource().Attributes().Get(sc.TenantID)
		if traceTenant.Str() != want || logTenant.Str() != want {
			t.Errorf("resource %d: got %q/%q, want %q", i, traceTenant.Str(), logTenant.Str(), want)
		}
	}
	if p.tenant.nDropped.Load() != 4 || p.tenant.nOriginDropped.Load() != 2 {
		t.Errorf("key/origin drops = %d/%d", p.tenant.nDropped.Load(), p.tenant.nOriginDropped.Load())
	}
}

func TestLogEnrichment(t *testing.T) {
	manager := &metadata.MetadataManager{}
	manager.Screen.Store(metadata.ServiceScreenMap{
		{ServiceNamespace: "reader", ScreenName: "Home", ScreenType: "NATIVE"}: {ScreenID: 17, ScreenGroupID: 3, ScreenGroupName: "main"},
	})
	crypto.InitializeKey("0123456789abcdef0123456789abcdef")
	t.Cleanup(func() { crypto.InitializeKey("") })
	ciphertext, err := crypto.Encrypt("reader-user")
	if err != nil {
		t.Fatal(err)
	}
	p := &SOPHONZAttributeProcessor{enabled: true, encryptedUserID: true, logger: zap.NewNop(), MetadataManager: manager, keyMode: ServiceKeyModeOff}
	ctx := client.NewContext(t.Context(), client.Info{Metadata: client.NewMetadata(map[string][]string{"x-client-address": {"192.0.2.1"}, "x-network-forwarded-for": {"192.0.2.1, 192.0.2.2"}})})
	ld := plog.NewLogs()
	rs := ld.ResourceLogs().AppendEmpty()
	rs.Resource().Attributes().PutStr("service.namespace", "reader")
	rs.Resource().Attributes().PutStr(sc.ClientPlatform, "web")
	record := rs.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	attrs := record.Attributes()
	attrs.PutStr(sc.AppScreenType, "native")
	attrs.PutStr(sc.AppScreenName, "Home")
	attrs.PutStr("user.id", ciphertext)
	if _, err := p.processLogs(ctx, ld); err != nil {
		t.Fatal(err)
	}
	if v, _ := rs.Resource().Attributes().Get(sc.ClientPlatform); v.Str() != "WEB" {
		t.Errorf("client platform = %q", v.Str())
	}
	for k, want := range map[string]string{sc.AppScreenType: "NATIVE", sc.AppScreenID: "17", sc.AppScreenGroupID: "3", sc.AppScreenGroupName: "main", "user.id": "reader-user", "client.address": "192.0.2.1", sc.NetworkForwardedFor: "192.0.2.1, 192.0.2.2"} {
		v, ok := attrs.Get(k)
		if !ok || v.AsString() != want {
			t.Errorf("%s = %q, want %q", k, v.AsString(), want)
		}
	}
	if v, ok := attrs.Get(sc.LogID); !ok || v.Str() == "" {
		t.Error("missing log ID")
	}
}
