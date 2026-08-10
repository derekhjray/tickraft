// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package rule

import "errors"

var (
	// ErrRuleNotFound is returned when a rule cannot be located by its ID.
	ErrRuleNotFound = errors.New("rule: not found")
	// ErrRuleCompileFailed is returned when a rule expression fails to compile.
	ErrRuleCompileFailed = errors.New("rule: compile failed")
	// ErrRuleInvalidScene is returned when an unknown rule scene is encountered.
	ErrRuleInvalidScene = errors.New("rule: invalid scene")
	// ErrRuleDuplicate is returned when a rule with the same tenant and name already exists.
	ErrRuleDuplicate = errors.New("rule: duplicate")
	// ErrRuleTooManyComparisons is returned when a rule expression
	// exceeds the configured MaxComparisons limit. Limiting comparisons
	// keeps the per-rule evaluation cost bounded so a single compound
	// rule cannot dominate the engine's evaluation budget.
	ErrRuleTooManyComparisons = errors.New("rule: too many comparisons")
)
