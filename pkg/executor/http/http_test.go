// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package http

import (
	"context"
	"testing"
	"time"

	"github.com/tickraft/tickraft/pkg/executor"
	"github.com/tickraft/tickraft/pkg/types"
)

func TestName(t *testing.T) {
	p := New(10 * time.Second)
	if got := p.Name(); got != "http" {
		t.Errorf("Name: got %q, want %q", got, "http")
	}
}

func TestCapabilities(t *testing.T) {
	p := New(10 * time.Second)
	if got := p.Capabilities(); got != executor.CapProbe {
		t.Errorf("Capabilities: got %v, want %v", got, executor.CapProbe)
	}
}

// TestExecuteMissingAddress verifies that Execute returns an abnormal
// result when the executor config does not carry an address (URL).
func TestExecuteMissingAddress(t *testing.T) {
	p := New(10 * time.Second)
	result, err := p.Execute(context.Background(), executor.ExecutionRequest{
		ExecutorName: "http",
		Config:       `{}`,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.Status != types.AssetStatusAbnormal {
		t.Errorf("Status: got %q, want %q", result.Status, types.AssetStatusAbnormal)
	}
}

// TestExecuteInvalidConfig verifies that Execute returns a parse error
// when the executor config is not valid JSON.
func TestExecuteInvalidConfig(t *testing.T) {
	p := New(10 * time.Second)
	_, err := p.Execute(context.Background(), executor.ExecutionRequest{
		ExecutorName: "http",
		Config:       `{invalid`,
	})
	if err == nil {
		t.Fatal("Execute with invalid config: expected error, got nil")
	}
}
