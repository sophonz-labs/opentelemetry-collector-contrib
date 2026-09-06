// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"context"
	"net/url"
	"strings"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/pcommon"

	sophonzmetadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/metadata"
	sophonzsemconv "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/semconv"
)

// applyOriginPolicy stamps the request's browser Origin on the resource and
// checks it against the app's allowlist. attrs must be the RESOURCE attribute
// map and service must be the one the key resolved to — an unresolved key has
// no policy to apply, which is why this is only ever reached from the resolved
// branch of resolveTenant.
//
// It returns true when the resource must be dropped. That happens only for a
// service that has switched originEnforced on itself. service_key_mode does not
// gate it in either direction: the app key rollout is the operator's decision
// and the origin allowlist is the customer's, and a customer who has turned
// enforcement on has asked for the rejection regardless of how far along the
// operator's own rollout is.
func (p *SOPHONZAttributeProcessor) applyOriginPolicy(ctx context.Context, attrs pcommon.Map, service sophonzmetadata.Service) bool {
	// Derived from the request, so a forged value on the resource is never
	// evidence of anything. Clear first, then stamp only what we actually read.
	attrs.Remove(sophonzsemconv.ClientOrigin)
	attrs.Remove(sophonzsemconv.ClientOriginViolation)

	origin := requestOrigin(ctx)
	if origin == "" {
		// No Origin header: a native or server SDK, or a receiver running
		// without include_metadata. Neither is subject to a browser origin
		// policy, and counting them as violations would make the counters
		// useless for deciding whether enforcement is safe.
		return false
	}

	attrs.PutStr(sophonzsemconv.ClientOrigin, origin)

	// An empty allowlist is the default and means "allow every origin". Not
	// counted either, so `origin.allowed` measures apps that opted in.
	if len(service.AllowedOrigins) == 0 {
		return false
	}

	if originAllowed(origin, service.AllowedOrigins) {
		add(ctx, p.tenant.originAllowed, &p.tenant.nOriginAllowed)
		return false
	}

	attrs.PutStr(sophonzsemconv.ClientOriginViolation, origin)
	if !service.OriginEnforced {
		add(ctx, p.tenant.originViolation, &p.tenant.nOriginViolation)
		return false
	}
	add(ctx, p.tenant.originDropped, &p.tenant.nOriginDropped)
	return true
}

// requestOrigin reads the Origin header off the request metadata, which the
// OTLP receiver only populates with include_metadata: true. With it off this
// returns "" for every request and the whole policy is inert — the same
// silent no-op setClientInfo has.
func requestOrigin(ctx context.Context) string {
	vals := client.FromContext(ctx).Metadata.Get("origin")
	if len(vals) == 0 {
		return ""
	}
	return strings.TrimSpace(vals[0])
}

// originAllowed reports whether origin satisfies allowed, which the caller has
// already established is non-empty.
//
// Matching is on scheme + host + port as the browser sent them. Two entry forms
// are understood:
//
//   - an exact origin, "https://app.acme.com" or "https://app.acme.com:8443",
//     which matches only that scheme, host and port
//   - a wildcard subdomain, "*.acme.com", which matches ANY scheme and ANY
//     port on any host under acme.com, at any depth ("a.acme.com" and
//     "a.b.acme.com" both match). It does NOT match the bare apex
//     "acme.com" — list the apex separately if you want it — and it does not
//     match a host that merely ends in the same letters ("evilacme.com"),
//     because the dot is part of the compared suffix.
//
// Loopback is always allowed, allowlist or not: the first thing an allowlist
// would otherwise break is the customer's own local development.
func originAllowed(origin string, allowed []string) bool {
	// Scheme and host are case-insensitive, and the allowlist entries were
	// lower-cased at cache-build time.
	normalized := strings.TrimSuffix(strings.ToLower(origin), "/")

	host := originHost(normalized)
	if isLoopbackHost(host) {
		return true
	}

	for _, entry := range allowed {
		if suffix, ok := strings.CutPrefix(entry, "*"); ok {
			// entry "*.acme.com" -> suffix ".acme.com". HasSuffix alone
			// excludes the apex, which has no leading dot.
			if host != "" && strings.HasSuffix(host, suffix) {
				return true
			}
			continue
		}
		if entry == normalized {
			return true
		}
	}
	return false
}

// originHost extracts the host of an origin without its port. It returns "" for
// anything that is not a parseable scheme://host origin, which includes the
// literal "null" a browser sends from a sandboxed iframe or a file:// page.
// Those are unattributable and fall through to the allowlist, where they match
// nothing.
func originHost(origin string) string {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" {
		return ""
	}
	return u.Hostname()
}

// isLoopbackHost reports whether the host is the developer's own machine. IPv6
// loopback is included because a dev server bound to ::1 sends
// "http://[::1]:3000" and is no less local than 127.0.0.1.
func isLoopbackHost(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}
