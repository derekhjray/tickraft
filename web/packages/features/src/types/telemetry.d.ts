// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

/**
 * Prober type enum (aligns with backend ProberRegistry)
 */
export type ProberType =
  | 'icmp'
  | 'tcp'
  | 'http'
  | 'dns'
  | 'udp'
  | 'redis'
  | 'ssh'
  | 'snmp'
  | 'database'
  | 'ssl'

/**
 * Listener type enum (aligns with backend ListenerRegistry)
 */
export type ListenerType = 'webhook' | 'syslog' | 'snmp' | 'mqtt'

/**
 * Monitor mode enum — aligns with backend TelemetryTask.Mode.
 * "active" = probed by ProberService; "passive" = receives via listener.
 */
export type MonitorMode = 'active' | 'passive'

/**
 * Monitor type enum — active points use prober types, passive points use
 * listener types. CE supports: icmp, tcp, http, udp (active), webhook (passive).
 */
export type MonitorType = ProberType | ListenerType

/**
 * Monitor point model — aligns with backend handler.TelemetryTask.
 * Unified model merging prober (active) and listener (passive) into a single
 * table with a Mode field.
 */
export interface MonitorPoint {
  id: number
  name: string
  description?: string
  /** Asset type (host, service, website, device) */
  assetType: string
  /** Monitoring mode: "active" (probed) or "passive" (receives) */
  mode: MonitorMode
  /** Prober executor type or listener type */
  type: MonitorType
  /** Schedule: cron expression or interval string */
  schedule: string
  enabled: boolean
  /** Type-specific configuration (host, port, url, etc.) */
  config?: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

/**
 * Monitor point creation parameters — matches backend TelemetryTask request body
 * for POST /api/v1/telemetry/monitors.
 */
export interface MonitorCreateParams {
  name: string
  description?: string
  assetType: string
  mode: MonitorMode
  type: MonitorType
  schedule: string
  enabled: boolean
  config?: Record<string, unknown>
}

/**
 * Monitor point update parameters — matches backend TelemetryTask request body
 * for PUT /api/v1/telemetry/monitors/:id. The backend replaces all fields,
 * so all required fields must be provided.
 */
export interface MonitorUpdateParams {
  name: string
  description?: string
  assetType: string
  mode: MonitorMode
  type: MonitorType
  schedule: string
  enabled: boolean
  config?: Record<string, unknown>
}

/**
 * Monitor status response — aligns with backend monitorStatus.
 */
export interface MonitorStatus {
  id: number
  name: string
  enabled: boolean
  status: string
}

/**
 * Prober type metadata — aligns with backend ProberType struct returned by
 * GET /api/v1/telemetry/probers.
 */
export interface ProberTypeInfo {
  /** Prober identifier (icmp, tcp, http, udp) */
  type: string
  /** Human-readable display name */
  name: string
  /** Short summary of the prober capability */
  description?: string
}

/**
 * Listener type metadata — aligns with backend ListenerType struct returned by
 * GET /api/v1/telemetry/listeners.
 */
export interface ListenerTypeInfo {
  /** Listener identifier (webhook) */
  type: string
  /** Human-readable display name */
  name: string
  /** Short summary of the listener capability */
  description?: string
}

/**
 * Monitor history entry — aligns with backend monitorHistoryEntry returned by
 * GET /api/v1/telemetry/monitors/:id/history (paginated).
 */
export interface MonitorHistoryEntry {
  /** Timestamp of the data point */
  timestamp: string
  /** Measured value */
  value: unknown
  /** Derived status string */
  status: string
}

/**
 * Monitor log entry — aligns with backend monitorLogEntry returned by
 * GET /api/v1/telemetry/monitors/:id/logs (paginated).
 */
export interface MonitorLog {
  /** Timestamp of the log entry */
  timestamp: string
  /** Log level (info, warning, error, etc.) */
  level: string
  /** Log message */
  message: string
}

/**
 * Metric data point (for asset metric trend)
 */
export interface MetricPoint {
  /** Time label */
  timestamp: string
  /** Average value */
  valueAvg: number
  /** Max value */
  valueMax: number
  /** Min value */
  valueMin: number
}

/**
 * Telemetry template model (pre-configured probe/monitor recipe) — aligns with
 * backend templateResponse. Field names are camelCase; the case conversion
 * layer transparently handles snake_case ↔ camelCase at the API boundary.
 */
export interface TelemetryTemplate {
  id: number
  name: string
  description: string
  category: string
  executorType: string
  config: Record<string, unknown>
  isBuiltin: boolean
  createdAt: string
  updatedAt: string
}

/**
 * Template creation parameters — matches backend templateRequest for
 * POST /api/v1/telemetry/templates.
 */
export interface TemplateCreateParams {
  name: string
  description: string
  category: string
  executorType: string
  config: Record<string, unknown>
}

/**
 * Template update parameters — same shape as creation, used for
 * PUT /api/v1/telemetry/templates/:id.
 */
export interface TemplateUpdateParams {
  name: string
  description: string
  category: string
  executorType: string
  config: Record<string, unknown>
}

/**
 * Apply template parameters — matches backend applyTemplateRequest for
 * POST /api/v1/telemetry/templates/:id/apply. Both fields are optional;
 * when omitted, the template defaults are used.
 */
export interface ApplyTemplateParams {
  /** Override the monitoring point name */
  name?: string
  /** Merge into the template config to customise the resulting monitor */
  overrides?: Record<string, unknown>
}
