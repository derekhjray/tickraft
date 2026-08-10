// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package alert

import (
	"context"

	"go.uber.org/zap"
)

// Matcher evaluates whether an alert event should be dispatched to channels.
// A matcher returns true to forward the alert, false to suppress it.
type Matcher interface {
	// Match returns true if the alert event should be dispatched.
	Match(ctx context.Context, alert Event) bool
}

// NamedMatcher is an optional interface that rules may implement to expose their
// name for observability and logging. Rules that do not implement NamedMatcher
// are counted as matched but omitted from the returned name list.
type NamedMatcher interface {
	Matcher
	// Name returns the human-readable rule identifier.
	Name() string
}

// ViolationMatcher is an optional interface that rules may implement to
// return structured Violations for a matched alert event. When a rule
// implements ViolationMatcher, Dispatch calls MatchWithViolations after
// Match returns true and collects the returned violations into
// Event.Violations, replacing the single violation populated by the
// payload converter. This enables compound rules (e.g.
// "cpu > 90 && mem > 85") to contribute one Violation per matched
// condition.
//
// Rules that do not implement ViolationMatcher are unaffected: Dispatch
// continues to use the payload-populated Event.Violations as-is.
type ViolationMatcher interface {
	// MatchWithViolations evaluates the rule and returns all violations
	// for the matched comparison sub-conditions. Returns nil or an empty
	// slice when the rule does not match or produces no structured
	// violations; in that case Dispatch preserves the existing
	// Event.Violations.
	MatchWithViolations(ctx context.Context, alert Event) []Violation
}

// MatcherFunc adapts a function into a Matcher.
type MatcherFunc func(ctx context.Context, alert Event) bool

// Match implements Matcher.
func (f MatcherFunc) Match(ctx context.Context, alert Event) bool {
	return f(ctx, alert)
}

// AddRule registers an alert rule. Rules are evaluated in registration
// order; if any rule matches, the alert is dispatched. When no rules are
// registered, all alerts are dispatched (default-allow). It must be
// called before Start.
func (e *Engine) AddRule(rule Matcher) {
	if rule == nil {
		return
	}
	e.mu.Lock()
	e.rules = append(e.rules, rule)
	e.mu.Unlock()
}

// Rules returns the registered alert rules. The returned slice is a copy
// and safe to read concurrently with AddRule.
func (e *Engine) Rules() []Matcher {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]Matcher, len(e.rules))
	copy(out, e.rules)
	return out
}

// UpdateRule replaces the rule identified by oldName with newMatcher. If
// no rule with oldName is registered, newMatcher is appended so that an
// updated rule not previously tracked by the engine (for example, one that
// was disabled at load time) is still registered. If newMatcher is nil the
// rule is removed, making this call equivalent to RemoveRule. Rules that do
// not implement NamedMatcher are skipped during the lookup. It is safe to
// call concurrently with AddRule, RemoveRule and Dispatch.
func (e *Engine) UpdateRule(oldName string, newMatcher Matcher) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, r := range e.rules {
		nr, ok := r.(NamedMatcher)
		if !ok {
			continue
		}
		if nr.Name() == oldName {
			if newMatcher == nil {
				e.rules = append(e.rules[:i], e.rules[i+1:]...)
			} else {
				e.rules[i] = newMatcher
			}
			return
		}
	}
	// Rule not found: append newMatcher so an updated rule that was not
	// previously registered is tracked.
	if newMatcher != nil {
		e.rules = append(e.rules, newMatcher)
	}
}

// RemoveRule removes the rule identified by name from the engine. If no
// rule with the given name is registered, the call is a no-op. Rules that
// do not implement NamedMatcher are skipped during the lookup. It is safe
// to call concurrently with AddRule, UpdateRule and Dispatch.
func (e *Engine) RemoveRule(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, r := range e.rules {
		nr, ok := r.(NamedMatcher)
		if !ok {
			continue
		}
		if nr.Name() == name {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			return
		}
	}
}

// safeMatch invokes rule.Match with panic recovery so that a buggy custom
// Matcher implementation cannot crash the prism alert engine. A panic is recovered,
// logged at error level, and treated as not matching so the alert evaluation
// continues with the remaining rules.
func safeMatch(ctx context.Context, rule Matcher, alert Event, logger *zap.Logger) (matched bool) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("rule Match panicked",
				zap.String("event_id", alert.EventID),
				zap.String("type", string(alert.Type)),
				zap.Int64("asset_id", alert.AssetID),
				zap.Any("panic", r),
			)
		}
	}()
	return rule.Match(ctx, alert)
}

// safeMatchWithViolations invokes vm.MatchWithViolations with panic recovery
// so that a buggy custom ViolationMatcher implementation cannot crash the
// prism alert engine. A panic is recovered, logged at error level, and
// treated as returning no violations so the alert evaluation continues with
// the remaining rules. The returned slice may be nil.
func safeMatchWithViolations(ctx context.Context, vm ViolationMatcher, alert Event, logger *zap.Logger) (violations []Violation) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("rule MatchWithViolations panicked",
				zap.String("event_id", alert.EventID),
				zap.String("type", string(alert.Type)),
				zap.Int64("asset_id", alert.AssetID),
				zap.Any("panic", r),
			)
			violations = nil
		}
	}()
	return vm.MatchWithViolations(ctx, alert)
}
