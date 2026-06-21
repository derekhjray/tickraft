// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package rule

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Record is the GORM persistence model for the sys_prism_rule
// table. It stores compiled-rule source text alongside tenant scoping
// and lifecycle metadata, and is converted to the runtime Rule view
// via toRule.
type Record struct {
	// ID is the auto-incremented primary key.
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// TenantID scopes the rule to a tenant for multi-tenant isolation.
	TenantID int64 `gorm:"column:tenant_id;not null;index"`
	// Name is the human-readable rule name.
	Name string `gorm:"column:name;type:varchar(255);not null"`
	// Description is an optional free-form rule description.
	Description string `gorm:"column:description;type:text"`
	// Scene selects which matching scene the rule participates in.
	Scene string `gorm:"column:scene;type:varchar(32);not null;index"`
	// Expression is the expr-lang source text compiled by the Compiler.
	Expression string `gorm:"column:expression;type:text;not null"`
	// Enabled indicates whether the rule participates in matching.
	Enabled bool `gorm:"column:enabled;not null;default:true"`
	// Priority orders rules within a scene; higher values fire first.
	Priority int `gorm:"column:priority;not null;default:0"`
	// GroupID is the resource group the rule belongs to. nil means the
	// rule is tenant-wide (visible to all members); a non-nil value
	// restricts visibility to members assigned to that group (B1-04).
	GroupID *int64 `gorm:"column:group_id;index"`
	// Metadata is the JSON-encoded extension key-value pairs.
	Metadata string `gorm:"column:metadata;type:text"`
	// CreatedAt is the rule creation timestamp.
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt is the rule last-update timestamp.
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// DeletedAt records the soft-delete timestamp.
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// TableName returns the database table name for Record.
func (Record) TableName() string { return "sys_prism_rule" }

// toRule converts the persistence model into the runtime Rule view.
// Metadata parsing failures are ignored: a malformed metadata payload
// does not block rule loading, the Metadata field is left nil instead.
func (m Record) toRule() Rule {
	return Rule{
		ID:         m.ID,
		TenantID:   m.TenantID,
		Name:       m.Name,
		Scene:      Scene(m.Scene),
		Expression: m.Expression,
		Priority:   m.Priority,
		Enabled:    m.Enabled,
		Metadata:   parseMetadata(m.Metadata),
	}
}

// parseMetadata decodes a JSON-encoded metadata blob into a string
// map. An empty input returns nil so the resulting Rule has a nil
// Metadata field rather than an empty map, keeping expr-lang field
// access semantics consistent.
func parseMetadata(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	var metadata map[string]string
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		return nil
	}
	if len(metadata) == 0 {
		return nil
	}
	return metadata
}
