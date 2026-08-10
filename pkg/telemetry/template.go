// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package telemetry

import "time"

// Template is the GORM model for the sys_telemetry_template table.
// It stores reusable telemetry monitoring point templates. Built-in templates
// are seeded at startup and marked IsBuiltin=true; custom templates are
// created by users through the API.
type Template struct {
	// ID is the unique identifier of the template.
	ID int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	// Name is the template name, unique across all templates.
	Name string `gorm:"column:name;type:varchar(128);not null;uniqueIndex" json:"name"`
	// Description is the human-readable description of the template.
	Description string `gorm:"column:description;type:varchar(512)" json:"description"`
	// Category classifies the template (e.g. "network", "web", "database").
	Category string `gorm:"column:category;type:varchar(64);index" json:"category"`
	// ExecutorType is the probe/executor type: icmp, tcp, http, dns, etc.
	ExecutorType string `gorm:"column:executor_type;type:varchar(32);not null" json:"executor_type"`
	// Config is the JSON-encoded monitoring point configuration.
	Config string `gorm:"column:config;type:text;not null" json:"config"`
	// IsBuiltin marks system-seeded templates. Built-in templates are
	// read-only and cannot be deleted.
	IsBuiltin bool `gorm:"column:is_builtin;default:false" json:"is_builtin"`
	// CreatedAt is the template creation timestamp.
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	// UpdatedAt is the last update timestamp.
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName returns the database table name for Template.
func (Template) TableName() string { return "sys_telemetry_template" }
