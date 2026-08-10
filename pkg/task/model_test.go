// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package task

import (
	"testing"
	"time"
)

func TestScheduleTaskToTask_ValidMetadata(t *testing.T) {
	m := &ScheduleTask{
		ID:             1,
		TenantID:       100,
		AssetID:        200,
		ExecutorType:   "webhook",
		ExecutorConfig: `{"url":"https://example.com"}`,
		Timeout:        30,
		Priority:       5,
		DependsOn:      0,
		Metadata:       `{"schedule_type":"cron","cron_expr":"*/5 * * * *"}`,
	}

	tk, err := m.ToTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tk.ID != 1 {
		t.Errorf("ID = %d, want 1", tk.ID)
	}
	if tk.TenantID != 100 {
		t.Errorf("TenantID = %d, want 100", tk.TenantID)
	}
	if tk.AssetID != 200 {
		t.Errorf("AssetID = %d, want 200", tk.AssetID)
	}
	if tk.ExecutorName != "webhook" {
		t.Errorf("ExecutorName = %q, want %q", tk.ExecutorName, "webhook")
	}
	if tk.Config != `{"url":"https://example.com"}` {
		t.Errorf("Config = %q, unexpected", tk.Config)
	}
	if tk.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", tk.Timeout)
	}
	if tk.Priority != 5 {
		t.Errorf("Priority = %d, want 5", tk.Priority)
	}
	if tk.DependsOn != 0 {
		t.Errorf("DependsOn = %d, want 0", tk.DependsOn)
	}
	if tk.Metadata == nil {
		t.Fatal("Metadata is nil, want non-nil map")
	}
	if v, ok := tk.Metadata["schedule_type"]; !ok || v != "cron" {
		t.Errorf("Metadata[schedule_type] = %q, want %q", v, "cron")
	}
	if v, ok := tk.Metadata["cron_expr"]; !ok || v != "*/5 * * * *" {
		t.Errorf("Metadata[cron_expr] = %q, want %q", v, "*/5 * * * *")
	}
}

func TestScheduleTaskToTask_EmptyMetadata(t *testing.T) {
	m := &ScheduleTask{
		ID:           1,
		TenantID:     100,
		ExecutorType: "webhook",
		Timeout:      30,
		Metadata:     "",
	}

	tk, err := m.ToTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tk.Metadata != nil {
		t.Errorf("Metadata = %v, want nil for empty string", tk.Metadata)
	}
}

func TestScheduleTaskToTask_InvalidMetadata(t *testing.T) {
	m := &ScheduleTask{
		ID:           1,
		TenantID:     100,
		ExecutorType: "webhook",
		Timeout:      30,
		Metadata:     `{invalid json}`,
	}

	_, err := m.ToTask()
	if err == nil {
		t.Error("expected error for invalid JSON metadata")
	}
}

func TestScheduleTaskToTask_EmptyJSONObjectMetadata(t *testing.T) {
	m := &ScheduleTask{
		ID:           1,
		TenantID:     100,
		ExecutorType: "webhook",
		Timeout:      30,
		Metadata:     `{}`,
	}

	tk, err := m.ToTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tk.Metadata == nil {
		t.Fatal("Metadata is nil for empty JSON object, want non-nil empty map")
	}
	if len(tk.Metadata) != 0 {
		t.Errorf("Metadata has %d entries, want 0", len(tk.Metadata))
	}
}

func TestScheduleTaskToTask_GroupAndTags(t *testing.T) {
	m := &ScheduleTask{
		ID:           1,
		TenantID:     100,
		ExecutorType: "webhook",
		Timeout:      30,
		Group:        "backup",
		Tags:         "critical,nightly",
	}

	tk, err := m.ToTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tk.Group != "backup" {
		t.Errorf("Group = %q, want %q", tk.Group, "backup")
	}
	if len(tk.Tags) != 2 {
		t.Fatalf("Tags len = %d, want 2", len(tk.Tags))
	}
	if tk.Tags[0] != "critical" || tk.Tags[1] != "nightly" {
		t.Errorf("Tags = %v, want [critical nightly]", tk.Tags)
	}
}

func TestScheduleTaskToTask_EmptyGroupAndTags(t *testing.T) {
	m := &ScheduleTask{
		ID:           1,
		TenantID:     100,
		ExecutorType: "webhook",
		Timeout:      30,
		Group:        "",
		Tags:         "",
	}

	tk, err := m.ToTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tk.Group != "" {
		t.Errorf("Group = %q, want empty", tk.Group)
	}
	// Empty tags string must yield nil (not []string{""}) so the JSON
	// omitempty tag drops the field.
	if tk.Tags != nil {
		t.Errorf("Tags = %v, want nil for empty string", tk.Tags)
	}
}

func TestScheduleTaskToTask_SingleTag(t *testing.T) {
	m := &ScheduleTask{
		ID:           1,
		TenantID:     100,
		ExecutorType: "webhook",
		Timeout:      30,
		Tags:         "urgent",
	}

	tk, err := m.ToTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tk.Tags) != 1 || tk.Tags[0] != "urgent" {
		t.Errorf("Tags = %v, want [urgent]", tk.Tags)
	}
}
