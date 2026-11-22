// Package semconv defines the SOPHONZ-specific semantic conventions
// (OTel attribute keys) used by the sophonz collector components
// (sophonzattributeprocessor, clickhouse exporters).
//
// This is the single source of truth for attribute keys. Renaming a value
// here propagates automatically to every component that reads attributes via
// these constants. ClickHouse column names are NOT defined here — see
// pkg/sophonz/schema for those (they must be kept in sync manually because
// the column identifiers are separate string literals in DDL/INSERT code).
//
// Convention migration (legacy -> current), see
// apps/otel-collector/docs/CONVENTION_MIGRATION.md:
//
//	screen.id            -> app.screen.id
//	screen.name          -> app.screen.name
//	screen.type          -> app.screen.type
//	screen.group.id      -> app.screen.group.id
//	screen.group.name    -> app.screen.group.name
//	sophonz.service.type -> client.platform
//	span.type            -> app.span.type
//
// Keys keep plain string type so they can be passed directly to
// pcommon.Map.Get / PutStr without conversion.
package semconv

// Resource attributes.
const (
	// ServiceKey is unchanged.
	ServiceKey = "service.key"

	// TenantID is the organization that owns the telemetry. It is derived by
	// the collector from ServiceKey against the service metadata cache and is
	// never read from what the SDK sends, so it can be trusted for tenant
	// isolation; anything a client puts under this key is overwritten.
	TenantID = "sophonz.tenant.id"

	// ClientPlatform replaces the legacy "sophonz.service.type".
	// Identifies the client platform (web, iOS, Android, ...).
	ClientPlatform = "client.platform"

	// WebVersion is unchanged.
	WebVersion = "web.version"

	// NetworkForwardedFor is unchanged.
	NetworkForwardedFor = "network.forwarded_for"

	// ClientOrigin is the browser Origin header of the request that carried the
	// telemetry, copied verbatim from the request metadata on every resource
	// whose key resolved. It exists so the product can answer "which origins
	// report under this app" without a schema change: as a resource attribute
	// it lands in resourceTagsMap on traces and ResourceAttributes on logs.
	//
	// Like TenantID it is derived from the request, never from what the SDK
	// sent, and anything a client puts under this key is removed.
	ClientOrigin = "sophonz.client.origin"

	// ClientOriginViolation carries the offending origin on a resource whose
	// Origin failed the app's allowlist. It duplicates ClientOrigin on purpose:
	// a violation has to be selectable on its own, without joining against the
	// app's current allowlist, which changes.
	ClientOriginViolation = "sophonz.client.origin.violation"
)

// Span attributes.
const (
	// AppScreenID replaces the legacy "screen.id".
	AppScreenID = "app.screen.id"
	// AppScreenName replaces the legacy "screen.name".
	AppScreenName = "app.screen.name"
	// AppScreenType replaces the legacy "screen.type".
	AppScreenType = "app.screen.type"
	// AppScreenGroupID replaces the legacy "screen.group.id".
	AppScreenGroupID = "app.screen.group.id"
	// AppScreenGroupName replaces the legacy "screen.group.name".
	AppScreenGroupName = "app.screen.group.name"

	// AppSpanType replaces the legacy "span.type".
	// Values are determined per platform, e.g.:
	//   web, iOS, Android : xhr, render, event, error
	//   iOS, Android      : crash
	//   Android           : anr
	//   web               : webvitals, websocket, route, log, visibility,
	//                        longtask, connectivity, post-docs
	AppSpanType = "app.span.type"
)

// Log attributes.
const (
	// LogID is unchanged.
	LogID = "log.id"
)
