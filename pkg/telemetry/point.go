// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package telemetry

// Mode defines the monitoring point mode. A monitoring point is either
// actively probed (ModeActive) or passively receiving data (ModePassive).
// The unified MonitorPoint GORM model (defined in model.go) uses this field
// to distinguish the two operational modes within a single table.
type Mode string

const (
	// ModeActive means the point is actively probed by the ProberService.
	// The Type field identifies the prober executor (icmp, tcp, http, udp).
	ModeActive Mode = "active"
	// ModePassive means the point passively receives data via a listener.
	// The Type field identifies the listener type (webhook).
	ModePassive Mode = "passive"
)

// Monitoring point status constants. These are stored in the MonitorPoint
// Status column and reported by the monitor status API endpoint.
const (
	// MonitorStatusActive indicates the monitoring point is running and
	// producing healthy results.
	MonitorStatusActive = "active"
	// MonitorStatusInactive indicates the monitoring point is disabled or
	// has not yet been started.
	MonitorStatusInactive = "inactive"
	// MonitorStatusError indicates the monitoring point encountered an
	// error during its last probe or reception cycle.
	MonitorStatusError = "error"
	// MonitorStatusPending indicates the monitoring point is registered
	// but awaiting its first probe or data reception.
	MonitorStatusPending = "pending"
)

// ValidMode reports whether the given Mode is a recognized value.
func ValidMode(m Mode) bool {
	return m == ModeActive || m == ModePassive
}
