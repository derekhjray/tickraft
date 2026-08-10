// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package quota

import "math"

// Type identifies a resource category subject to quota enforcement.
//
// Type values align with the extended Quota.ResourceType strings so
// that quota records persisted by callers can be matched against
// the policy without translation. The string literals intentionally mirror
// the asset-type categories (task, device, host); unified-policy
// types (prober, scheduled_task, remediation, http_interval, daily_events)
// extend the set without colliding with existing values.
type Type string

// Type constants enumerate every resource category for which the runtime
// exposes a quota ceiling.
const (
	// TypeTask is the scheduled-task asset quota type. It is
	// retained with asset-limit callers that
	// treat generic "task" assets as uncapped (see DefaultQuotaTask).
	TypeTask Type = "task"
	// TypeDevice is the monitored device / monitoring-asset quota type.
	// Under the unified quota policy this maps to
	// DefaultQuotaMonitoringAsset.
	TypeDevice Type = "device"
	// TypeHost is the host asset quota type.
	TypeHost Type = "host"
	// TypeProber is the active prober quota type (e.g. ICMP/TCP/HTTP
	// probes driven by the collector module).
	TypeProber Type = "prober"
	// TypeScheduledTask is the scheduled-task quota type under the
	// unified quota policy. It is distinct from TypeTask, which
	// represents the uncapped asset-limit category.
	TypeScheduledTask Type = "scheduled_task"
	// TypeRemediation is the remediation task quota type.
	TypeRemediation Type = "remediation"
	// TypeHTTPInterval is the minimum allowed HTTP probe interval,
	// expressed in seconds.
	TypeHTTPInterval Type = "http_interval"
	// TypeDailyEvents is the daily event ingestion quota type.
	TypeDailyEvents Type = "daily_events"
)

// Asset-limit constants.
//
// These cover the asset-type categories (task, device, host) and
// are retained because:
//   - pkg/api/handler/asset.go mirrors DefaultQuotaMonitoringAsset via a
//     local constant (maxDeviceQuota = 20) for device-creation enforcement.
//   - The Quotas map below references them for TypeTask / TypeHost.
//
// The unified quota constants (DefaultQuotaMonitoringAsset etc.) cover
// the newer policy categories and supersede DefaultQuotaDevice for
// quota-table lookups.
const (
	// DefaultQuotaTask is the maximum number of tasks in the standalone
	// runtime. The default runtime imposes no task count limit;
	// math.MaxInt32 represents an effectively unlimited ceiling while
	// keeping the constant defined for API symmetry with
	// DefaultQuotaDevice.
	DefaultQuotaTask = math.MaxInt32
	// DefaultQuotaDevice is the maximum number of devices in the
	// standalone runtime. The limit is set to 10 to cover common
	// small-fleet monitoring scenarios without artificial scarcity;
	// larger fleets remain an extension point.
	DefaultQuotaDevice = 10
	// DefaultQuotaHost is the maximum number of hosts in the standalone
	// runtime.
	DefaultQuotaHost = 0
)

// Unified quota constants.
//
// These encode the fixed quotas the runtime enforces per the unified
// policy:
//
//	monitoring assets 20 / probers 20 / scheduled tasks 20 /
//	remediation 5 / HTTP interval 60s / 100k events per day.
//
// Source compilation remains the documented extension point for lifting
// these ceilings.
const (
	// DefaultQuotaMonitoringAsset is the maximum number of monitoring
	// assets (monitored devices/resources) under the unified quota
	// policy. It supersedes DefaultQuotaDevice (10) for quota-table
	// lookups; the device-creation enforcement path in
	// pkg/api/handler/asset.go mirrors this constant via a local
	// maxDeviceQuota = 20. Update both values together when the
	// monitoring-asset ceiling changes.
	DefaultQuotaMonitoringAsset = 20
	// DefaultQuotaProber is the maximum number of active probers in the
	// runtime.
	DefaultQuotaProber = 20
	// DefaultQuotaScheduledTask is the maximum number of scheduled tasks
	// in the runtime under the unified quota policy. It is distinct from
	// DefaultQuotaTask (math.MaxInt32, effectively unlimited), which
	// remains for asset-limit callers that
	// treat generic "task" assets as uncapped.
	DefaultQuotaScheduledTask = 20
	// DefaultQuotaRemediation is the maximum number of remediation tasks
	// in the runtime.
	DefaultQuotaRemediation = 5
	// DefaultQuotaHTTPIntervalSeconds is the minimum allowed HTTP probe
	// interval, in seconds, in the runtime.
	DefaultQuotaHTTPIntervalSeconds = 60
	// DefaultQuotaDailyEvents is the maximum number of events that may be
	// ingested per day in the runtime.
	DefaultQuotaDailyEvents = 100000
)

// Quotas is the single source of truth for quota ceilings keyed by Type.
// The unified policy maps TypeDevice to DefaultQuotaMonitoringAsset
// (superseding DefaultQuotaDevice) so that quota-table lookups enforce
// the newer, higher monitoring-asset ceiling.
var Quotas = map[Type]int{
	TypeTask:          DefaultQuotaTask,
	TypeDevice:        DefaultQuotaMonitoringAsset,
	TypeHost:          DefaultQuotaHost,
	TypeProber:        DefaultQuotaProber,
	TypeScheduledTask: DefaultQuotaScheduledTask,
	TypeRemediation:   DefaultQuotaRemediation,
	TypeHTTPInterval:  DefaultQuotaHTTPIntervalSeconds,
	TypeDailyEvents:   DefaultQuotaDailyEvents,
}

// DefaultQuotaFor returns the quota ceiling for the given type and
// whether the type is recognized. Unknown types return (0, false).
func DefaultQuotaFor(qt Type) (int, bool) {
	v, ok := Quotas[qt]
	return v, ok
}
