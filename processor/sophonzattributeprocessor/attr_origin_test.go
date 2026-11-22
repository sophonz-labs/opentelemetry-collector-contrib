// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor

import "testing"

// The wildcard and loopback rules are the two places where being wrong is
// expensive: too loose and the allowlist is decorative, too tight and a
// customer drops their own traffic the moment they enforce it.
func TestOriginAllowed(t *testing.T) {
	allowlist := []string{"https://app.acme.com", "https://portal.acme.com:8443", "*.widgets.io"}

	tests := []struct {
		origin string
		want   bool
	}{
		{"https://app.acme.com", true},
		{"HTTPS://APP.ACME.COM", true},         // scheme and host are case-insensitive
		{"http://app.acme.com", false},         // exact entries pin the scheme
		{"https://app.acme.com:8443", false},   // and the port
		{"https://portal.acme.com:8443", true}, //
		{"https://portal.acme.com", false},     //
		{"https://acme.com", false},            // apex is not covered by an exact subdomain entry
		{"https://evil.com", false},            //
		{"https://a.widgets.io", true},         // wildcard: any subdomain
		{"http://a.b.widgets.io:3000", true},   // any depth, any scheme, any port
		{"https://widgets.io", false},          // but not the bare apex
		{"https://notwidgets.io", false},       // and not a lookalike suffix
		{"http://localhost:3000", true},        // loopback is always allowed
		{"https://localhost", true},            //
		{"http://127.0.0.1:5173", true},        //
		{"http://[::1]:5173", true},            //
		{"null", false},                        // opaque origin is unattributable
		{"https://app.acme.com/", true},        // tolerate a stray trailing slash
	}

	for _, tt := range tests {
		if got := originAllowed(tt.origin, allowlist); got != tt.want {
			t.Errorf("originAllowed(%q) = %v, want %v", tt.origin, got, tt.want)
		}
	}
}
