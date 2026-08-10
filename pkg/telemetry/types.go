// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package telemetry

import (
	"time"

	"github.com/tickraft/tickraft/pkg/types"
)

// Telemetry is the standardized data structure that flows from
// external collectors to Processor. All collection channels produce this.
type Telemetry struct {
	// AssetID is the ID of the asset the telemetry was collected from.
	AssetID int64
	// TenantID is the tenant that owns the asset.
	TenantID int64
	// AssetType categorizes the asset.
	AssetType types.AssetType
	// SourceType identifies the data source (e.g., "webhook").
	SourceType string
	// RemoteAddr is the source address of the telemetry.
	RemoteAddr string
	// CollectedAt is when the data was collected.
	CollectedAt time.Time
	// RawData contains the original unprocessed data.
	RawData []byte
	// Metrics holds extracted numerical metrics (optional).
	Metrics map[string]float64
	// LogContent holds log content (optional).
	LogContent string
	// LogLevel holds the severity level of the log content (optional, defaults to "INFO").
	LogLevel string
	// Status is the pre-judged status (optional, set by collectors).
	Status types.AssetStatus
}

// ProcessResult indicates the outcome of processing a telemetry.
type ProcessResult struct {
	// PrevStatus is the status before the transition.
	PrevStatus types.AssetStatus
	// CurrStatus is the status after the transition.
	CurrStatus types.AssetStatus
	// Reason describes why the status changed.
	Reason string
	// Alerts holds any alerts triggered during processing.
	Alerts []AlertContext
}

// AlertContext describes an alert triggered during processing.
type AlertContext struct {
	// Level is the alert severity (info/warning/critical).
	Level string
	// Title is the alert title.
	Title string
	// Message is the alert detail.
	Message string
	// SourceIP is the origin address of the telemetry that triggered the alert,
	// used to populate LogMatchedPayload.SourceIP. Empty when unavailable.
	SourceIP string
	// Metadata holds additional alert information.
	// For metric alerts, conventionally populated keys are:
	//   - "metric":    the metric name (e.g. "rtt_ms")
	//   - "value":     the observed metric value (stringified float64)
	//   - "threshold": the threshold value (stringified float64)
	//   - "operator":  the comparison operator (e.g. ">"); defaults to ">" when absent
	Metadata map[string]string
}
