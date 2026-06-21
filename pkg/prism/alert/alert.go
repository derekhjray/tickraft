// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package alert

import (
	"context"
	"fmt"
	"sync"

	"github.com/tickraft/tickraft/internal/alert"
	"github.com/tickraft/tickraft/pkg/errdefs"
	"github.com/tickraft/tickraft/pkg/event"
	"github.com/tickraft/tickraft/pkg/pool"
	"github.com/tickraft/tickraft/pkg/prism/governance"
	"go.uber.org/zap"
)

// defaultNotificationPoolSize is the worker count used when no pool size is
// configured.
const defaultNotificationPoolSize = 8

// Event is the normalized alert representation shared with governance
// middleware and channels. The canonical definition lives in internal/alert so
// pkg/prism/governance can reference it without a circular dependency on this
// package; it is re-exported here so cross-repository consumers (the extended
// edition) keep using alert.Event / alert.TypeMetric / alert.TypeLog.
type (
	Event         = alert.Event
	Type          = alert.Type
	Violation     = alert.Violation
	MetricContext = alert.MetricContext
	LogContext    = alert.LogContext
	StatusContext = alert.StatusContext
)

const (
	// TypeMetric is emitted by metric threshold violations.
	TypeMetric Type = alert.TypeMetric
	// TypeLog is emitted by log keyword matches.
	TypeLog Type = alert.TypeLog
	// TypeHeartbeat is emitted when a probe stops reporting heartbeats.
	TypeHeartbeat Type = alert.TypeHeartbeat
	// TypeStatus is emitted when an asset transitions to an abnormal state.
	TypeStatus Type = alert.TypeStatus
)

// Violation kind constants identify the category of a single rule violation.
// These mirror the canonical constants in internal/alert so callers across
// repositories reference them as alert.ViolationKind* without importing
// internal packages.
const (
	ViolationKindMetric    = alert.ViolationKindMetric
	ViolationKindLog       = alert.ViolationKindLog
	ViolationKindHeartbeat = alert.ViolationKindHeartbeat
	ViolationKindStatus    = alert.ViolationKindStatus
)

// PrimaryViolation returns the most severe violation carried by the Event,
// or nil when no violations have been recorded. This is the recommended
// accessor for summary scenarios (alert titles, channel messages).
func PrimaryViolation(e Event) *Violation {
	if v, ok := e.PrimaryViolation(); ok {
		return &v
	}
	return nil
}

// MetricName returns the metric name from the primary violation's Metric
// context, or "" when no metric context is present.
func MetricName(e Event) string {
	if v := PrimaryViolation(e); v != nil && v.Metric != nil {
		return v.Metric.Name
	}
	return ""
}

// MetricValue returns the observed metric value from the primary violation's
// Metric context, or 0 when no metric context is present.
func MetricValue(e Event) float64 {
	if v := PrimaryViolation(e); v != nil && v.Metric != nil {
		return v.Metric.Value
	}
	return 0
}

// Threshold returns the configured threshold from the primary violation's
// Metric context, or 0 when no metric context is present.
func Threshold(e Event) float64 {
	if v := PrimaryViolation(e); v != nil && v.Metric != nil {
		return v.Metric.Threshold
	}
	return 0
}

// Metrics returns the full metric value map from the primary violation's
// Metric context, or nil when no metric context is present.
func Metrics(e Event) map[string]float64 {
	if v := PrimaryViolation(e); v != nil && v.Metric != nil {
		return v.Metric.Metrics
	}
	return nil
}

// Severity returns the unified severity from the primary violation, or ""
// when no violation carries a severity.
func Severity(e Event) string {
	if v := PrimaryViolation(e); v != nil {
		return v.Severity
	}
	return ""
}

// Keyword returns the matched log keyword from the primary violation's Log
// context, or "" when no log context is present.
func Keyword(e Event) string {
	if v := PrimaryViolation(e); v != nil && v.Log != nil {
		return v.Log.Keyword
	}
	return ""
}

// Content returns the log content that matched from the primary violation's
// Log context, or "" when no log context is present.
func Content(e Event) string {
	if v := PrimaryViolation(e); v != nil && v.Log != nil {
		return v.Log.Content
	}
	return ""
}

// SourceIP returns the origin address from the primary violation's Source
// field. The Violation.Source field carries the IP for log violations and
// the probe ID for heartbeat violations. Returns "" when no violation
// carries a source.
func SourceIP(e Event) string {
	if v := PrimaryViolation(e); v != nil {
		return v.Source
	}
	return ""
}

// Channel sends an alert notification to an external system.
type Channel interface {
	// Send delivers the alert payload. It must be safe for concurrent use.
	Send(ctx context.Context, alert Event) error
	// Name identifies the channel in logs and metrics.
	Name() string
}

// OnAlertFunc is invoked when an alert event matches the registered rules
// (or when no rules are registered, in which case all alerts match). It is
// called synchronously within dispatch before channel notification, so
// implementations must return quickly to avoid blocking the event bus
// consumer. Persistent side effects (e.g. writing to a store) should be
// offloaded when they may block.
type OnAlertFunc func(ctx context.Context, alert Event)

// PostGuardHook is invoked after an alert passes the governance guard chain
// (every guard returned DecisionPass) and before rule evaluation. The
// callers may use this SPI to notify the Suppressor about active
// source alerts so that future dependent target alerts are suppressed.
//
// The hook receives a pointer to the Event so mutations (e.g. stamping
// enrichment fields) are visible to subsequent rule evaluation and channel
// dispatch. Implementations must be safe for concurrent use and must return
// quickly to avoid blocking the dispatch path; long-running side effects
// should be offloaded. A nil hook is a no-op, preserving the
// behaviour where no post-guard notification is required.
type PostGuardHook func(ctx context.Context, evt *Event)

// Engine is the alert evaluation and notification dispatch engine. It
// subscribes to telemetry alert events on the event bus, evaluates
// registered rules against each event, and dispatches matching alerts to
// registered channels through a bounded worker pool.
//
// A governance.Guard chain is invoked before rule evaluation. In a
// single-process default deployment the chain is empty, so Dispatch
// proceeds directly to rule evaluation. The callers may inject the
// full governance chain (silence → aggregator → suppressor → storm) at
// startup.
type Engine struct {
	bus    event.Bus
	logger *zap.Logger

	mu       sync.RWMutex
	rules    []Matcher
	channels []Channel
	guards   []governance.Guard

	onAlert OnAlertFunc

	// postGuardHook is invoked after the governance guard chain passes and
	// before rule evaluation. nil in default deployment (no-op).
	postGuardHook PostGuardHook

	notifyPool pool.Pool
	poolOwned  bool

	startMu sync.Mutex
	started bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// Option configures an Engine.
type Option interface {
	apply(*options)
}

type options struct {
	bus                  event.Bus
	logger               *zap.Logger
	notificationPoolSize int
	notifyPool           pool.Pool
	onAlert              OnAlertFunc
	postGuardHook        PostGuardHook
	guards               []governance.Guard
}

type funcOption func(*options)

func (f funcOption) apply(o *options) { f(o) }

// WithEventBus sets the event bus used to subscribe to alert events.
func WithEventBus(bus event.Bus) Option {
	return funcOption(func(o *options) { o.bus = bus })
}

// WithLogger sets the structured logger.
func WithLogger(logger *zap.Logger) Option {
	return funcOption(func(o *options) { o.logger = logger })
}

// WithNotificationPoolSize sets the goroutine pool size for sending
// notifications. A non-positive value defaults to 8. Ignored when
// WithPool is used to inject an externally-owned pool.
func WithNotificationPoolSize(n int) Option {
	return funcOption(func(o *options) { o.notificationPoolSize = n })
}

// WithPool injects an externally-owned worker pool for notification
// dispatch. When set, the engine does not create or shut down its own
// pool; the caller is responsible for the pool lifecycle.
func WithPool(p pool.Pool) Option {
	return funcOption(func(o *options) { o.notifyPool = p })
}

// WithOnAlert registers a callback invoked when an alert event matches the
// registered rules (or when no rules are registered). The callback is
// called synchronously within dispatch before channel notification, so it
// must return quickly. It is typically used to persist alert records to a
// store without introducing a dependency from the prism alert package to the
// store package.
func WithOnAlert(fn OnAlertFunc) Option {
	return funcOption(func(o *options) { o.onAlert = fn })
}

// WithPostGuardHook registers a hook invoked after the governance guard chain
// passes (every guard returned DecisionPass) and before rule evaluation. The
// callers may use this to notify the Suppressor about active source
// alerts. Passing nil clears any previously registered hook. In an
// single-process deployment no hook is registered, so Dispatch
// proceeds directly to rule evaluation.
func WithPostGuardHook(h PostGuardHook) Option {
	return funcOption(func(o *options) { o.postGuardHook = h })
}

// WithGuards sets the governance guard chain invoked before rule
// evaluation in Dispatch. The chain is called in order; the first non-Pass
// decision short-circuits the chain. In an single-process
// deployment the chain is empty, so Dispatch proceeds directly to rule
// evaluation.
func WithGuards(guards ...governance.Guard) Option {
	return funcOption(func(o *options) {
		for _, g := range guards {
			if g != nil {
				o.guards = append(o.guards, g)
			}
		}
	})
}

// New creates a new alert Engine with the given options.
//
// When no notification pool is injected via WithPool, the engine creates
// and owns a bounded IO pool sized by WithNotificationPoolSize (default 8).
// The owned pool is shut down on Stop.
//
// Returns an error if the internally-created notification pool cannot be
// initialized. This path is unreachable in practice because the worker
// count is sanitized to a positive value, but the error is returned
// rather than panicking to honor the "no panic in business logic" rule.
func New(opts ...Option) (*Engine, error) {
	o := &options{
		logger:               zap.NewNop(),
		notificationPoolSize: defaultNotificationPoolSize,
	}
	for _, opt := range opts {
		opt.apply(o)
	}

	e := &Engine{
		bus:           o.bus,
		logger:        o.logger,
		onAlert:       o.onAlert,
		postGuardHook: o.postGuardHook,
		guards:        o.guards,
	}

	if o.notifyPool != nil {
		e.notifyPool = o.notifyPool
		e.poolOwned = false
	} else {
		size := o.notificationPoolSize
		if size <= 0 {
			size = defaultNotificationPoolSize
		}
		p, err := pool.New(
			pool.WithWorkers(size),
			pool.WithRejectionPolicy(pool.RejectionCallerRuns),
		)
		if err != nil {
			// Unreachable in practice: size is sanitized to >= 1
			// above. Returned as an error (not panic) to honor
			// the "no panic in business logic" rule.
			return nil, fmt.Errorf("alert: create notification pool: %w", err)
		}
		e.notifyPool = p
		e.poolOwned = true
	}

	return e, nil
}

// AddChannel registers a notification channel. Channels are notified
// concurrently for each dispatched alert. It must be called before Start.
func (e *Engine) AddChannel(ch Channel) {
	if ch == nil {
		return
	}
	e.mu.Lock()
	e.channels = append(e.channels, ch)
	e.mu.Unlock()
}

// Channels returns the registered notification channels. The returned
// slice is a copy and safe to read concurrently with AddChannel.
func (e *Engine) Channels() []Channel {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]Channel, len(e.channels))
	copy(out, e.channels)
	return out
}

// AddGuard registers a governance guard invoked before rule
// evaluation in Dispatch. Guards are called in registration order. It
// must be called before Start.
func (e *Engine) AddGuard(g governance.Guard) {
	if g == nil {
		return
	}
	e.mu.Lock()
	e.guards = append(e.guards, g)
	e.mu.Unlock()
}

// Start subscribes to alert events on the event bus. It returns an error
// if the engine is already started or no event bus is configured.
func (e *Engine) Start(ctx context.Context) error {
	e.startMu.Lock()
	defer e.startMu.Unlock()
	if e.started {
		return nil
	}
	if e.bus == nil {
		return errdefs.ErrBusNotConfigured
	}

	runCtx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	e.started = true

	// Subscribe to telemetry alert events. Each handler normalizes the
	// typed payload into an Event and dispatches it through the
	// rule engine and notification channels.
	if _, err := event.Subscribe[event.MetricExceededPayload](e.bus, event.TypeTelemetryMetricExceeded, func(_ context.Context, ev event.Event[event.MetricExceededPayload]) error {
		e.dispatch(runCtx, metricPayloadToAlert(ev))
		return nil
	}); err != nil {
		return fmt.Errorf("alert: subscribe to metric exceeded events: %w", err)
	}
	if _, err := event.Subscribe[event.LogMatchedPayload](e.bus, event.TypeTelemetryLogMatched, func(_ context.Context, ev event.Event[event.LogMatchedPayload]) error {
		e.dispatch(runCtx, logPayloadToAlert(ev))
		return nil
	}); err != nil {
		return fmt.Errorf("alert: subscribe to log matched events: %w", err)
	}
	// Subscribe to asset status-change events. A transition to an abnormal
	// state (offline/critical/warning) is normalized into a heartbeat or
	// status alert Event; transitions to a healthy/unknown state (recoveries)
	// are skipped by statusPayloadToAlert so the engine does not emit alert
	// noise for recoveries. A heartbeat-loss transition (Source == "timeout",
	// published by telemetry.MarkOffline) is mapped to TypeHeartbeat; every
	// other abnormal transition is mapped to TypeStatus.
	if _, err := event.Subscribe[event.StatusChangePayload](e.bus, event.TypeAssetStatusChanged, func(_ context.Context, ev event.Event[event.StatusChangePayload]) error {
		evt, ok := statusPayloadToAlert(ev)
		if !ok {
			return nil
		}
		e.dispatch(runCtx, evt)
		return nil
	}); err != nil {
		return fmt.Errorf("alert: subscribe to status change events: %w", err)
	}

	e.logger.Info("prism alert engine started",
		zap.Int("rules", len(e.rulesSnapshot())),
		zap.Int("channels", len(e.channelsSnapshot())),
	)
	return nil
}

// Stop cancels the run context, waits for in-flight notifications to
// finish or the context to expire, and shuts down the owned notification
// pool. It is idempotent.
func (e *Engine) Stop(ctx context.Context) error {
	e.startMu.Lock()
	if !e.started {
		e.startMu.Unlock()
		return nil
	}
	e.started = false
	if e.cancel != nil {
		e.cancel()
	}
	e.startMu.Unlock()

	done := make(chan struct{})
	// goroutine lifecycle: bounded — drains in-flight notification goroutines
	// spawned by Dispatch; exits after e.wg.Wait() returns. In-flight channel
	// Send jobs observe the cancelled runCtx (cancelled above) and exit
	// promptly, so the wait is bounded in practice.
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	if e.poolOwned && e.notifyPool != nil {
		if err := e.notifyPool.Shutdown(ctx); err != nil {
			e.logger.Warn("notification pool shutdown returned error",
				zap.Error(err),
			)
		}
	}

	e.logger.Info("prism alert engine stopped")
	return nil
}

// rulesSnapshot returns the current rules without holding the read lock
// across the call site.
func (e *Engine) rulesSnapshot() []Matcher {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return append([]Matcher(nil), e.rules...)
}

// channelsSnapshot returns the current channels without holding the read
// lock across the call site.
func (e *Engine) channelsSnapshot() []Channel {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return append([]Channel(nil), e.channels...)
}

// guardsSnapshot returns the current governance guards without
// holding the read lock across the call site.
func (e *Engine) guardsSnapshot() []governance.Guard {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return append([]governance.Guard(nil), e.guards...)
}
