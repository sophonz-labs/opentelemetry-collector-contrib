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

	// ClientPlatform replaces the legacy "sophonz.service.type".
	// Identifies the client platform (web, iOS, Android, ...).
	ClientPlatform = "client.platform"

	// WebVersion is unchanged.
	WebVersion = "web.version"

	// NetworkForwardedFor is unchanged.
	NetworkForwardedFor = "network.forwarded_for"
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
