// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

/**
 * WebSocket event types (aligned with backend pkg/event/event.go)
 */
export type WSEventType =
  | 'asset.status_change'
  | 'asset.metric_alert'
  | 'asset.log_alert'
  | 'task.execution'
  | 'task.scheduled'
  | 'task.retry'

/**
 * WebSocket event payload base structure
 */
export interface WSEventPayload {
  /** Event type */
  type: WSEventType
  /** Event timestamp */
  timestamp: string
  /** Tenant ID */
  tenantId: number
}

/**
 * Asset status change event payload
 */
export interface StatusChangePayload extends WSEventPayload {
  type: 'asset.status_change'
  assetId: number
  assetType: string
  prevStatus: string
  currStatus: string
  reason: string
}

/**
 * Metric alert event payload
 */
export interface MetricAlertPayload extends WSEventPayload {
  type: 'asset.metric_alert'
  assetId: number
  metricKey: string
  metricValue: number
  threshold: number
  alertLevel: 'info' | 'warning' | 'critical'
}

/**
 * Log alert event payload
 */
export interface LogAlertPayload extends WSEventPayload {
  type: 'asset.log_alert'
  assetId: number
  logPattern: string
  logContent: string
  alertLevel: 'info' | 'warning' | 'critical'
}

/**
 * Task execution event payload
 */
export interface TaskExecutionPayload extends WSEventPayload {
  type: 'task.execution'
  taskId: number
  taskName: string
  executorType: string
  status: 'running' | 'success' | 'failed' | 'timeout'
  output?: string
  errorMsg?: string
  duration: number
}

/**
 * Task scheduled event payload
 */
export interface TaskScheduledPayload extends WSEventPayload {
  type: 'task.scheduled'
  taskId: number
  taskName: string
  scheduleType: string
  nextRunAt: string
}

/**
 * Task retry event payload
 */
export interface TaskRetryPayload extends WSEventPayload {
  type: 'task.retry'
  taskId: number
  taskName: string
  retryCount: number
  maxRetries: number
  nextRetryAt: string
}

/**
 * Union type of all event payloads
 */
export type WSEvent =
  | StatusChangePayload
  | MetricAlertPayload
  | LogAlertPayload
  | TaskExecutionPayload
  | TaskScheduledPayload
  | TaskRetryPayload
