// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"context"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"

	sophonzmetadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/metadata"
	sophonzsemconv "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/semconv"
)

// tenantSummaryInterval bounds how often the aggregated counter snapshot is
// logged. The counters are per-resource, so anything finer would put a log line
// in the hot path.
const tenantSummaryInterval = time.Minute

// tenantCounters mirrors the OTel counters as plain atomics so a summary can be
// logged on collectors whose internal telemetry is not scraped. Both are fed
// from the same call sites; the OTel side is the one to build alerts on.
type tenantCounters struct {
	resolved   metric.Int64Counter
	missingKey metric.Int64Counter
	unknownKey metric.Int64Counter
	dropped    metric.Int64Counter
	degraded   metric.Int64Counter

	nResolved   atomic.Int64
	nMissingKey atomic.Int64
	nUnknownKey atomic.Int64
	nDropped    atomic.Int64
	nDegraded   atomic.Int64

	// lastSummaryUnixNano rate-limits both the summary and the fail-open
	// warning; it is compared-and-swapped so concurrent consumers emit once.
	lastSummaryUnixNano atomic.Int64
}

// newTenantCounters builds the shadow-mode counters. A provider that refuses to
// create an instrument leaves it nil; add() tolerates that rather than failing
// the pipeline over telemetry.
func newTenantCounters(mp metric.MeterProvider, logger *zap.Logger) *tenantCounters {
	c := &tenantCounters{}
	if mp == nil {
		return c
	}
	meter := mp.Meter("github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor")

	newCounter := func(name, desc string) metric.Int64Counter {
		counter, err := meter.Int64Counter(name, metric.WithDescription(desc), metric.WithUnit("{resource}"))
		if err != nil {
			logger.Warn("failed to create tenant counter", zap.String("name", name), zap.Error(err))
			return nil
		}
		return counter
	}

	c.resolved = newCounter("sophonzattribute.tenant.resolved", "Resources whose service.key resolved to a tenant.")
	c.missingKey = newCounter("sophonzattribute.tenant.missing_key", "Resources that carried no service.key.")
	c.unknownKey = newCounter("sophonzattribute.tenant.unknown_key", "Resources whose service.key is absent from the metadata cache.")
	c.dropped = newCounter("sophonzattribute.tenant.dropped", "Resources dropped because no tenant could be resolved.")
	c.degraded = newCounter("sophonzattribute.tenant.degraded", "Resources passed through untouched because the metadata cache was unavailable.")
	return c
}

func add(ctx context.Context, counter metric.Int64Counter, mirror *atomic.Int64) {
	mirror.Add(1)
	if counter != nil {
		counter.Add(ctx, 1)
	}
}

// due reports whether enough time has passed to emit another rate-limited log
// line, claiming the slot for the caller.
func (c *tenantCounters) due(now time.Time) bool {
	last := c.lastSummaryUnixNano.Load()
	next := now.UnixNano()
	if next-last < int64(tenantSummaryInterval) {
		return false
	}
	return c.lastSummaryUnixNano.CompareAndSwap(last, next)
}

func (c *tenantCounters) logSummary(logger *zap.Logger, mode ServiceKeyMode) {
	logger.Info("tenant resolution summary",
		zap.String("mode", string(mode)),
		zap.Int64("resolved", c.nResolved.Load()),
		zap.Int64("missing_key", c.nMissingKey.Load()),
		zap.Int64("unknown_key", c.nUnknownKey.Load()),
		zap.Int64("dropped", c.nDropped.Load()),
		zap.Int64("degraded", c.nDegraded.Load()),
	)
}

// resolveTenant looks the raw service.key up in the service metadata cache and,
// on a hit, stamps the owning organization on the resource as
// sophonz.tenant.id. attrs must be the RESOURCE attribute map.
//
// It returns true when the caller should drop the resource, which only ever
// happens in enforce mode.
//
// The key is an opaque token minted by the app (Service.serviceKey, `sk_...`);
// it is matched verbatim against the cache rather than decrypted, which is what
// replaced the old AES-CTR payload with its hardcoded key.
func (p *SOPHONZAttributeProcessor) resolveTenant(ctx context.Context, attrs pcommon.Map) bool {
	// Fail open on an infrastructure problem: with no cache there is no way to
	// tell a forged key from a valid one, and dropping everything would turn a
	// metadata outage into total data loss.
	if p.MetadataManager == nil {
		p.reportDegraded(ctx, "metadata manager is not initialized")
		return false
	}
	skm, ok := p.MetadataManager.Service.Load().(sophonzmetadata.ServiceKeyMap)
	if !ok || len(skm) == 0 {
		p.reportDegraded(ctx, "service key map is empty")
		return false
	}

	value, exist := attrs.Get(sophonzsemconv.ServiceKey)
	if !exist || value.Str() == "" {
		add(ctx, p.tenant.missingKey, &p.tenant.nMissingKey)
		return p.dropUnresolved(ctx)
	}

	service, found := skm[value.Str()]
	if !found {
		add(ctx, p.tenant.unknownKey, &p.tenant.nUnknownKey)
		return p.dropUnresolved(ctx)
	}

	// Overwrite unconditionally: whatever the client sent under this key is not
	// evidence of anything.
	attrs.PutStr(sophonzsemconv.TenantID, service.CompanyID)
	add(ctx, p.tenant.resolved, &p.tenant.nResolved)
	p.maybeLogSummary()
	return false
}

func (p *SOPHONZAttributeProcessor) dropUnresolved(ctx context.Context) bool {
	p.maybeLogSummary()
	if p.keyMode != ServiceKeyModeEnforce {
		return false
	}
	add(ctx, p.tenant.dropped, &p.tenant.nDropped)
	return true
}

func (p *SOPHONZAttributeProcessor) reportDegraded(ctx context.Context, reason string) {
	add(ctx, p.tenant.degraded, &p.tenant.nDegraded)
	if p.tenant.due(time.Now()) {
		p.logger.Error("tenant resolution degraded, passing telemetry through untouched",
			zap.String("reason", reason),
			zap.String("mode", string(p.keyMode)),
			zap.Int64("degraded_resources", p.tenant.nDegraded.Load()),
		)
	}
}

func (p *SOPHONZAttributeProcessor) maybeLogSummary() {
	if p.tenant.due(time.Now()) {
		p.tenant.logSummary(p.logger, p.keyMode)
	}
}
