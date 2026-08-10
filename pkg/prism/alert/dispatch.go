// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package alert

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/tickraft/tickraft/pkg/event"
	"github.com/tickraft/tickraft/pkg/pool"
	"github.com/tickraft/tickraft/pkg/prism/governance"
	"go.uber.org/zap"
)

// DispatchResult is the structured outcome of a synchronous Dispatch call.
type DispatchResult struct {
	// Accepted indicates whether the alert was accepted for processing,
	// i.e. at least one rule matched or no rules are registered
	// (default-allow).
	Accepted bool
	// EventID is the deduplication / tracking identifier assigned to the
	// alert by the engine. It is stable for a single Dispatch call, stamped
	// on the Event before any channel or callback observes it, and
	// suitable for correlating with delivery records.
	EventID string
	// MatchedRules lists the names of rules that matched the alert. Rules
	// that do not implement NamedMatcher are counted toward the accept
	// decision but omitted from this slice.
	MatchedRules []string
	// DispatchedChannels lists the channel names that received a
	// notification dispatch request. Channel sends happen asynchronously
	// via the worker pool; this slice reflects the channels the engine
	// submitted to the pool, not the channels that have completed
	// delivery.
	DispatchedChannels []string
	// Message is a short human-readable acknowledgement message intended
	// for callers.
	Message string
}

// dispatch is the event-bus subscription callback. It delegates to Dispatch
// and discards the result so the existing fire-and-forget semantics are
// preserved.
func (e *Engine) dispatch(ctx context.Context, alert Event) {
	// Dispatch returns a DispatchResult carrying observability metadata (matched
	// rules, dispatched channels). The event-bus callback has no caller to
	// surface this to, and Dispatch already logs every relevant outcome
	// (suppression, channel send failure, onAlert error) via zap, so the result
	// is intentionally not captured here.
	e.Dispatch(ctx, alert)
}

// Dispatch synchronously evaluates the alert against the registered rules
// and, if it matches (or no rules are registered), submits notification
// jobs to the worker pool for each registered channel. It returns a
// DispatchResult carrying the matched rule names and the dispatched channel
// names. Channel sends happen asynchronously via the worker pool; this
// method returns as soon as the jobs are submitted (or the alert is
// suppressed).
//
// Before rule evaluation, the engine invokes the registered governance.Middleware
// chain in order. The first middleware that returns governance.DecisionSuppress or
// governance.DecisionAggregate short-circuits the chain: the alert is still recorded
// (the OnAlert callback is invoked) but notification channels are skipped.
// In an single-process deployment the chain is empty, so
// Dispatch proceeds directly to rule evaluation.
//
// The engine assigns a unique EventID to the alert at the start of Dispatch
// and stamps it on the Event before any channel or callback observes
// it, so the same identifier flows through rules, the OnAlert callback,
// every channel Send, and the returned DispatchResult.
//
// All registered rules are evaluated (no short-circuit) so that every
// matching rule name is collected for the response. This differs from the
// historical fire-and-forget dispatch path which broke on the first match;
// the dispatch decision (any-match) is unchanged. Each rule.Match call is
// wrapped with panic recovery so a buggy custom Matcher cannot crash the
// engine; a panicking rule is logged and treated as not matching.
func (e *Engine) Dispatch(ctx context.Context, alert Event) DispatchResult {
	// Defensive check: ensure the event has at least one violation.
	// When Violations is empty, initialize a default violation based on
	// the event type so that downstream consumers (channels, governance)
	// do not receive an empty violation list.
	if len(alert.Violations) == 0 {
		alert.Violations = []Violation{{
			Kind: string(alert.Type),
		}}
	}

	eventID := newEventID()
	alert.EventID = eventID

	rules := e.rulesSnapshot()
	channels := e.channelsSnapshot()
	guards := e.guardsSnapshot()

	// Governance guard chain: invoked before rule evaluation. The
	// first non-Pass decision short-circuits the chain. Suppressed and
	// aggregated alerts are still recorded (OnAlert) but skip channel
	// dispatch. default deployments pass an empty chain, so this loop
	// is a no-op and Dispatch falls through to rule evaluation.
	for _, g := range guards {
		decision := safeProcess(ctx, g, &alert, e.logger)
		switch decision {
		case governance.DecisionSuppress:
			e.recordAlert(ctx, alert, eventID)
			e.logger.Debug("alert suppressed by governance guard",
				zap.String("event_id", eventID),
				zap.String("guard", guardName(g)),
				zap.String("type", string(alert.Type)),
				zap.Int64("asset_id", alert.AssetID),
			)
			return DispatchResult{
				Accepted: false,
				EventID:  eventID,
				Message:  "alert suppressed by governance guard",
			}
		case governance.DecisionAggregate:
			e.recordAlert(ctx, alert, eventID)
			e.logger.Debug("alert aggregated by governance guard",
				zap.String("event_id", eventID),
				zap.String("guard", guardName(g)),
				zap.String("type", string(alert.Type)),
				zap.Int64("asset_id", alert.AssetID),
			)
			return DispatchResult{
				Accepted: false,
				EventID:  eventID,
				Message:  "alert aggregated by governance guard",
			}
		}
	}

	// All governance guards passed (DecisionPass). Invoke the post-guard
	// hook so the callers can notify the Suppressor about active
	// source alerts before rule evaluation. The hook is nil in default
	// deployments, so this is a no-op there.
	if e.postGuardHook != nil {
		e.postGuardHook(ctx, &alert)
	}

	matchedRules := make([]string, 0)
	matched := len(rules) == 0
	// Collect structured violations from any matched rule that implements
	// ViolationMatcher. When a compound rule (e.g. "cpu > 90 && mem > 85")
	// matches multiple conditions, each condition contributes one Violation.
	// When violations are collected, they replace the single violation
	// populated by the payload converter (metricPayloadToAlert /
	// logPayloadToAlert) so downstream consumers (channels, governance
	// fingerprint, record persistence) see the full set of matched
	// conditions. When no ViolationMatcher rules match or none return
	// violations, the payload-populated Event.Violations are preserved.
	var collectedViolations []Violation
	for _, r := range rules {
		if safeMatch(ctx, r, alert, e.logger) {
			matched = true
			if nr, ok := r.(NamedMatcher); ok {
				matchedRules = append(matchedRules, nr.Name())
			}
			if vm, ok := r.(ViolationMatcher); ok {
				collectedViolations = append(collectedViolations,
					safeMatchWithViolations(ctx, vm, alert, e.logger)...)
			}
		}
	}
	if len(collectedViolations) > 0 {
		alert.Violations = collectedViolations
	}
	if !matched {
		e.logger.Debug("alert suppressed by rules",
			zap.String("event_id", eventID),
			zap.String("type", string(alert.Type)),
			zap.Int64("asset_id", alert.AssetID),
		)
		return DispatchResult{
			Accepted: false,
			EventID:  eventID,
			Message:  "alert suppressed by rules",
		}
	}

	// Invoke the OnAlert callback (if registered) so that callers can
	// persist the alert record without the prism alert package depending on a
	// store. Errors from the callback are logged but do not suppress
	// channel notification.
	e.recordAlert(ctx, alert, eventID)

	dispatchedChannels := make([]string, 0, len(channels))
	if len(channels) == 0 {
		// No channels registered: log the alert so it is still
		// observable in deployments without a configured
		// notification sink.
		primary, _ := alert.PrimaryViolation()
		metricName := ""
		if primary.Metric != nil {
			metricName = primary.Metric.Name
		}
		e.logger.Info("alert received (no channels registered)",
			zap.String("event_id", eventID),
			zap.String("type", string(alert.Type)),
			zap.Int64("asset_id", alert.AssetID),
			zap.Int64("tenant_id", alert.TenantID),
			zap.String("metric_name", metricName),
			zap.String("level", primary.Severity),
		)
		return DispatchResult{
			Accepted:     true,
			EventID:      eventID,
			MatchedRules: matchedRules,
			Message:      "alert accepted; no channels registered",
		}
	}

	for _, ch := range channels {
		c := ch
		dispatchedChannels = append(dispatchedChannels, c.Name())
		e.wg.Add(1)
		job := pool.Lambda(func(jobCtx context.Context) error {
			defer e.wg.Done()
			if err := c.Send(jobCtx, alert); err != nil {
				e.logger.Warn("channel send failed",
					zap.String("event_id", eventID),
					zap.String("channel", c.Name()),
					zap.Error(err),
				)
			}
			return nil
		})
		if err := e.notifyPool.Submit(ctx, job); err != nil {
			e.wg.Done()
			e.logger.Warn("notification pool submit failed, alert dropped",
				zap.String("event_id", eventID),
				zap.String("channel", c.Name()),
				zap.Error(err),
			)
		}
	}

	return DispatchResult{
		Accepted:           true,
		EventID:            eventID,
		MatchedRules:       matchedRules,
		DispatchedChannels: dispatchedChannels,
		Message:            "alert accepted",
	}
}

// safeOnAlert invokes the OnAlert callback with panic recovery so that a
// buggy callback cannot crash the prism alert engine. The returned error is
// non-nil only when the callback returned an error; panics are recovered,
// logged, and reported as an error so the caller can log them.
func safeOnAlert(ctx context.Context, fn OnAlertFunc, alert Event, logger *zap.Logger) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("onAlert callback panicked",
				zap.String("type", string(alert.Type)),
				zap.Int64("asset_id", alert.AssetID),
				zap.Any("panic", r),
			)
			err = fmt.Errorf("alert: onAlert callback panicked: %v", r)
		}
	}()
	fn(ctx, alert)
	return nil
}

// recordAlert invokes the OnAlert callback (if registered) so that callers
// can persist the alert record without the prism alert package depending on a
// store. Errors from the callback are logged but do not suppress channel
// notification. It is called both for governance-suppressed alerts (which are
// still recorded) and for alerts that passed rule evaluation.
func (e *Engine) recordAlert(ctx context.Context, alert Event, eventID string) {
	if e.onAlert == nil {
		return
	}
	if err := safeOnAlert(ctx, e.onAlert, alert, e.logger); err != nil {
		e.logger.Warn("onAlert callback returned error",
			zap.String("event_id", eventID),
			zap.String("type", string(alert.Type)),
			zap.Int64("asset_id", alert.AssetID),
			zap.Error(err),
		)
	}
}

// safeProcess invokes a governance.Guard.Process with panic recovery so
// that a buggy guard cannot crash the engine. A panic is recovered,
// logged, and treated as governance.DecisionPass so the alert is not silently swallowed
// by a faulty guard.
func safeProcess(ctx context.Context, g governance.Guard, evt *Event, logger *zap.Logger) (decision governance.Decision) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("governance guard panicked",
				zap.String("type", string(evt.Type)),
				zap.Int64("asset_id", evt.AssetID),
				zap.Any("panic", r),
			)
			decision = governance.DecisionPass
		}
	}()
	return g.Process(ctx, evt)
}

// guardName returns a human-readable name for a governance guard
// for logging. It uses a type-naming fallback when the guard does not
// expose a Name method.
func guardName(g governance.Guard) string {
	if namer, ok := g.(interface{ Name() string }); ok {
		return namer.Name()
	}
	return fmt.Sprintf("%T", g)
}

// newEventID generates a unique 32-char hex identifier for a dispatched alert
// event. It uses crypto/rand so identifiers are unpredictable enough for
// deduplication keys; the prism alert engine does not rely on them for security. On read
// failure (extremely unlikely) it falls back to a timestamp-based value so the
// engine never blocks on entropy availability.
func newEventID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("evt-%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// metricPayloadToAlert converts a typed metric alert event into the
// normalized Event consumed by channels.
func metricPayloadToAlert(ev event.Event[event.MetricExceededPayload]) Event {
	p := ev.Payload
	ts := ev.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	assetID, _ := strconv.ParseInt(p.AssetID, 10, 64)
	tenantID, _ := strconv.ParseInt(p.TenantID, 10, 64)
	severity := p.Severity
	if severity == "" {
		severity = "warning"
	}
	return Event{
		Type:      TypeMetric,
		AssetID:   assetID,
		TenantID:  tenantID,
		Timestamp: ts,
		Violations: []Violation{{
			Kind:     ViolationKindMetric,
			Severity: severity,
			Metric: &MetricContext{
				Name:      p.MetricName,
				Value:     p.MetricValue,
				Threshold: p.Threshold,
				Metrics:   p.Resources,
			},
		}},
	}
}

// logPayloadToAlert converts a typed log alert event into the normalized
// Event consumed by channels.
func logPayloadToAlert(ev event.Event[event.LogMatchedPayload]) Event {
	p := ev.Payload
	ts := ev.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	assetID, _ := strconv.ParseInt(p.AssetID, 10, 64)
	tenantID, _ := strconv.ParseInt(p.TenantID, 10, 64)
	return Event{
		Type:      TypeLog,
		AssetID:   assetID,
		TenantID:  tenantID,
		Timestamp: ts,
		Violations: []Violation{{
			Kind:     ViolationKindLog,
			Severity: mapLogLevel(p.Level),
			Source:   p.SourceIP,
			Log: &LogContext{
				Keyword: p.Keyword,
				Content: p.Content,
			},
		}},
	}
}

// statusPayloadToAlert converts a typed asset status-change event into the
// normalized Event consumed by channels. A transition whose Source is
// "timeout" (heartbeat loss reported by telemetry.MarkOffline) is mapped to
// TypeHeartbeat with ViolationKindHeartbeat; every other transition is mapped
// to TypeStatus with ViolationKindStatus. Transitions to a non-abnormal state
// (healthy/unknown) are skipped by returning ok=false so the engine does not
// emit alert noise for recoveries.
func statusPayloadToAlert(ev event.Event[event.StatusChangePayload]) (Event, bool) {
	p := ev.Payload
	if !isAbnormalStatus(p.CurrStatus) {
		return Event{}, false
	}
	ts := ev.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	assetID, _ := strconv.ParseInt(p.AssetID, 10, 64)
	tenantID, _ := strconv.ParseInt(p.TenantID, 10, 64)

	alertType := TypeStatus
	kind := ViolationKindStatus
	severity := "error"
	message := fmt.Sprintf("asset %s transitioned %s -> %s", p.AssetID, p.PrevStatus, p.CurrStatus)
	if p.Source == "timeout" {
		alertType = TypeHeartbeat
		kind = ViolationKindHeartbeat
		severity = "critical"
		message = fmt.Sprintf("asset %s heartbeat lost, marked offline", p.AssetID)
	}
	if p.Reason != "" {
		message = fmt.Sprintf("%s (%s)", message, p.Reason)
	}

	return Event{
		Type:      alertType,
		AssetID:   assetID,
		TenantID:  tenantID,
		Timestamp: ts,
		Violations: []Violation{{
			Kind:     kind,
			Severity: severity,
			Source:   p.AssetKey,
			Message:  message,
			Status: &StatusContext{
				PrevStatus: p.PrevStatus,
				CurrStatus: p.CurrStatus,
			},
		}},
	}, true
}

// isAbnormalStatus reports whether status represents an alert-worthy
// (non-healthy, non-unknown) asset state. The engine uses it to skip
// recovery transitions so alerts are emitted only for degradations.
func isAbnormalStatus(status string) bool {
	switch status {
	case "offline", "critical", "warning":
		return true
	default:
		return false
	}
}

// mapLogLevel normalizes a raw log level (debug/info/warn/error/fatal) into
// the unified severity scale: critical > error > warning > info > debug.
func mapLogLevel(level string) string {
	switch level {
	case "fatal", "critical":
		return "critical"
	case "error":
		return "error"
	case "warn", "warning":
		return "warning"
	case "info", "notice":
		return "info"
	case "debug":
		return "debug"
	default:
		return level
	}
}
