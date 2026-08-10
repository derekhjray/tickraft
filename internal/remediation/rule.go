// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package remediation

import "time"

// Rule is the GORM model for the sys_remediation_rule table. It persists a
// remediation rule definition: which source event to match, an optional
// asset filter and match condition, and the executor action to dispatch
// when the rule fires.
type Rule struct {
	// ID is the primary key of the rule.
	ID int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	// Name is the human-readable rule name.
	Name string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	// Description is an optional free-form description of the rule.
	Description string `gorm:"column:description;type:varchar(1024)" json:"description,omitempty"`
	// Enabled indicates whether the rule is active. Only enabled rules are
	// loaded by the Manager on startup.
	Enabled bool `gorm:"column:enabled;not null;default:true" json:"enabled"`
	// EventType is the source event type to match (for example
	// event.TypeAssetFaultDetected).
	EventType string `gorm:"column:event_type;type:varchar(64);not null" json:"event_type"`
	// AssetKey is an optional asset filter. When empty the rule matches all
	// assets; otherwise the rule matches only events whose asset key equals
	// this value.
	AssetKey string `gorm:"column:asset_key;type:varchar(255)" json:"asset_key,omitempty"`
	// Condition is an optional JSON-encoded match condition evaluated
	// against the event payload. An empty string means no extra condition.
	Condition string `gorm:"column:condition;type:varchar(2048)" json:"condition,omitempty"`
	// ActionType is the executor type to invoke (for example "local" or
	// "http"). It is resolved through the executor.Registry at dispatch
	// time.
	ActionType string `gorm:"column:action_type;type:varchar(32);not null" json:"action_type"`
	// ActionPayload is the JSON-encoded action payload consumed by the
	// executor (command, args, target, etc.).
	ActionPayload string `gorm:"column:action_payload;type:varchar(4096)" json:"action_payload,omitempty"`
	// CooldownSeconds is the minimum number of seconds between two
	// triggers of the same rule for the same asset. A value of 0 disables
	// cooldown enforcement for the rule.
	CooldownSeconds int `gorm:"column:cooldown_seconds;not null;default:0" json:"cooldown_seconds"`
	// CreatedAt records the rule creation time, populated by the database.
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	// UpdatedAt records the last rule update time, populated by the
	// database.
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName returns the database table name for Rule.
func (Rule) TableName() string { return "sys_remediation_rule" }

// Record is the GORM model for the sys_remediation_record table. It
// persists the lifecycle of a single remediation dispatch: which rule
// fired, which asset and source event triggered it, the resulting
// executor task, and the final status.
type Record struct {
	// ID is the primary key of the record.
	ID int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	// RuleID references the remediation rule that produced this record.
	RuleID int64 `gorm:"column:rule_id;not null;index" json:"rule_id"`
	// RuleName is a denormalized snapshot of the rule name at trigger
	// time, retained so historical records remain readable even if the
	// rule is later renamed or deleted.
	RuleName string `gorm:"column:rule_name;type:varchar(255);not null" json:"rule_name"`
	// AssetKey is the asset key of the triggering event, when applicable.
	// It is indexed to support per-asset cooldown lookups.
	AssetKey string `gorm:"column:asset_key;type:varchar(255);index" json:"asset_key,omitempty"`
	// SourceEventID is the identifier of the event that triggered this
	// remediation dispatch.
	SourceEventID string `gorm:"column:source_event_id;type:varchar(128)" json:"source_event_id,omitempty"`
	// Status is the dispatch lifecycle state. See the Status* constants for
	// the recognized values.
	Status string `gorm:"column:status;type:varchar(16);not null;default:'triggered'" json:"status"`
	// TaskID is the executor task identifier assigned to this dispatch.
	TaskID string `gorm:"column:task_id;type:varchar(128)" json:"task_id,omitempty"`
	// Error captures the failure message when Status is "failed".
	Error string `gorm:"column:error;type:varchar(2048)" json:"error,omitempty"`
	// StartedAt is the time the executor started working on the dispatch.
	StartedAt time.Time `gorm:"column:started_at" json:"started_at,omitempty"`
	// FinishedAt is the time the executor finished (success or failure).
	FinishedAt time.Time `gorm:"column:finished_at" json:"finished_at,omitempty"`
	// CreatedAt records when this record was inserted, populated by the
	// database.
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName returns the database table name for Record.
func (Record) TableName() string { return "sys_remediation_record" }

// Record status values describe the lifecycle of a remediation dispatch.
const (
	// StatusTriggered indicates a rule match was detected and a dispatch
	// is about to start.
	StatusTriggered = "triggered"
	// StatusStarted indicates the executor has accepted and started the
	// dispatch.
	StatusStarted = "started"
	// StatusCompleted indicates the dispatch finished successfully.
	StatusCompleted = "completed"
	// StatusSkipped indicates the dispatch was skipped due to cooldown,
	// condition mismatch, or circuit-breaker suppression.
	StatusSkipped = "skipped"
	// StatusFailed indicates the dispatch failed and the executor returned
	// an error. The Record's Error field is populated.
	StatusFailed = "failed"
)
