// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/route"
)

// TestTelemetryReportHandlerAdapter forwards a Hertz POST request to a
// net/http.Handler and copies the response back. It verifies that the
// adapter correctly bridges the two interfaces: the wrapped handler
// receives the body, headers, and method from the Hertz request, and the
// Hertz client receives the status code and body from the handler.
func TestTelemetryReportHandlerAdapter(t *testing.T) {
	// Build a net/http.HandlerFunc that echoes back the request body and
	// sets a custom header so we can verify the round-trip.
	netHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("X-Test-Echo", "ok")
		w.WriteHeader(http.StatusAccepted)
		// Echo the request body back as the response body.
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		_, _ = w.Write(buf)
	})

	adapter := NewTelemetryReportHandlerAdapter(netHandler, nil)

	h := route.NewEngine(config.NewOptions(nil))
	h.POST("/api/v1/telemetry", func(ctx context.Context, arc *app.RequestContext) {
		adapter.Report(ctx, arc)
	})

	body := map[string]any{"kind": "heartbeat", "asset_id": 1}
	bodyBytes, _ := json.Marshal(body)

	w := ut.PerformRequest(h, "POST", "/api/v1/telemetry",
		&ut.Body{Body: bytes.NewReader(bodyBytes), Len: len(bodyBytes)},
		ut.Header{Key: "X-Tickraft-Asset-Key", Value: "test-key"})

	if w.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", w.Code, http.StatusAccepted)
	}
	if got := w.Header().Get("X-Test-Echo"); got != "ok" {
		t.Errorf("X-Test-Echo header = %q, want %q", got, "ok")
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if result["kind"] != "heartbeat" {
		t.Errorf("response kind = %v, want heartbeat", result["kind"])
	}
}

// TestTelemetryReportHandlerAdapter_NetHTTPError verifies that error
// status codes and bodies from the net/http handler are correctly
// propagated to the Hertz response.
func TestTelemetryReportHandlerAdapter_NetHTTPError(t *testing.T) {
	netHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	adapter := NewTelemetryReportHandlerAdapter(netHandler, nil)

	h := route.NewEngine(config.NewOptions(nil))
	h.POST("/api/v1/telemetry", func(ctx context.Context, arc *app.RequestContext) {
		adapter.Report(ctx, arc)
	})

	w := ut.PerformRequest(h, "POST", "/api/v1/telemetry",
		&ut.Body{Body: bytes.NewReader([]byte(`{}`)), Len: 2})

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestResponseRecorder verifies the responseRecorder captures status,
// headers, and body correctly.
func TestResponseRecorder(t *testing.T) {
	rec := newResponseRecorder()
	if rec.statusCode != http.StatusOK {
		t.Errorf("initial status = %d, want %d", rec.statusCode, http.StatusOK)
	}
	rec.WriteHeader(http.StatusCreated)
	if rec.statusCode != http.StatusCreated {
		t.Errorf("status after WriteHeader = %d, want %d", rec.statusCode, http.StatusCreated)
	}
	rec.Header().Set("X-Custom", "value")
	if rec.Header().Get("X-Custom") != "value" {
		t.Errorf("header X-Custom = %q, want %q", rec.Header().Get("X-Custom"), "value")
	}
	_, _ = rec.Write([]byte("hello"))
	if string(rec.body) != "hello" {
		t.Errorf("body = %q, want %q", string(rec.body), "hello")
	}
}
