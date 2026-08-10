// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package remediation

import (
	"context"
	"time"
)

// Rule represents a self-healing (remediation) rule definition
// managed through the CRUD API at /api/v1/prism/remediation/rules. A rule
// binds a trigger event type (metric, log, status_change) to an executor
// action (e.g., local script) with optional condition filtering, cooldown,
// and circuit-breaker safety mechanisms. The open-source edition supports
// the "local" executor type; additional types are injected via the Operator
// SPI.
type Rule struct {
	ID                      int64      `json:"id"`
	Name                    string     `json:"name"`
	Description             string     `json:"description,omitempty"`
	AssetID                 int64      `json:"asset_id"`
	TriggerEventType        string     `json:"trigger_event_type"`
	ConditionExpr           string     `json:"condition_expr,omitempty"`
	ExecutorType            string     `json:"executor_type"`
	ExecutorConfig          string     `json:"executor_config,omitempty"`
	Cooldown                int        `json:"cooldown"`
	CircuitBreakerThreshold int        `json:"circuit_breaker_threshold"`
	Enabled                 bool       `json:"enabled"`
	Status                  string     `json:"status,omitempty"`
	LastRunAt               *time.Time `json:"last_run_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// Service defines the operations for managing remediation
// rules. The concrete implementation is injected via the
// WithRemediationRuleService RouteOption; when omitted, the handler package
// falls back to an in-memory implementation.
type Service interface {
	// ListRules returns a page of remediation rules and the total count.
	ListRules(ctx context.Context, page, size int) ([]Rule, int64, error)
	// GetRule returns a single remediation rule by ID.
	GetRule(ctx context.Context, id int64) (*Rule, error)
	// CreateRule creates a new remediation rule from the given request.
	CreateRule(ctx context.Context, req *Rule) (*Rule, error)
	// UpdateRule updates an existing remediation rule identified by ID.
	UpdateRule(ctx context.Context, id int64, req *Rule) (*Rule, error)
	// DeleteRule deletes a remediation rule by ID.
	DeleteRule(ctx context.Context, id int64) error
}
