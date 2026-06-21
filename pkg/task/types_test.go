// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package task

import "testing"

func TestScheduleTypeConstants(t *testing.T) {
	tests := []struct {
		typ ScheduleType
		str string
	}{
		{ScheduleTypeCron, "cron"},
		{ScheduleTypeInterval, "interval"},
		{ScheduleTypeOnce, "once"},
		{ScheduleTypeEvent, "event"},
	}
	for _, tt := range tests {
		if string(tt.typ) != tt.str {
			t.Errorf("ScheduleType %v = %q, want %q", tt.typ, string(tt.typ), tt.str)
		}
	}
}

func TestTaskFields(t *testing.T) {
	tk := Task{
		ID:           1,
		TenantID:     100,
		AssetID:      200,
		ExecutorName: "webhook",
		Config:       `{"url":"http://example.com"}`,
		Timeout:      30,
		Priority:     5,
		DependsOn:    0,
		Metadata:     map[string]string{"key": "value"},
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
	if tk.Config != `{"url":"http://example.com"}` {
		t.Errorf("Config = %q, unexpected", tk.Config)
	}
	if tk.Timeout != 30 {
		t.Errorf("Timeout = %v, want 30", tk.Timeout)
	}
	if tk.Priority != 5 {
		t.Errorf("Priority = %d, want 5", tk.Priority)
	}
	if tk.DependsOn != 0 {
		t.Errorf("DependsOn = %d, want 0", tk.DependsOn)
	}
	if tk.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %q, want %q", tk.Metadata["key"], "value")
	}
}
