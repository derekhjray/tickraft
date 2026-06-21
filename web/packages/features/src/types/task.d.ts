// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

/**
 * Executor type enum (aligned with backend handler.Task.Executor)
 *
 * Built-in executors: http / tcp / icmp / local / webhook
 * Extension executors: ssh / mysql / redis (extended by the extension)
 */
export type ExecutorType =
  | 'http'
  | 'tcp'
  | 'icmp'
  | 'local'
  | 'webhook'
  | 'ssh'
  | 'mysql'
  | 'redis'
  | 'telnet'
  | 'snmp'
  | 'notify'

/**
 * Schedule type for form UI state only.
 * The backend uses a single `schedule` string; this type helps the form
 * distinguish between cron expression, fixed interval, and event-driven
 * (empty schedule) for UI rendering. It is NOT sent to the backend.
 */
export type ScheduleType = 'cron' | 'interval' | 'event'

/**
 * Retry policy enum (aligned with backend handler.Task.RetryPolicy)
 */
export type RetryPolicy = 'fixed' | 'exponential'

/**
 * Task model (aligned with backend handler.Task)
 */
export interface TaskModel {
  id: number
  name: string
  description?: string
  executor: string
  schedule: string
  enabled: boolean
  config?: Record<string, unknown>
  group?: string
  tags?: string[]
  runId?: string
  retryPolicy?: string
  concurrency?: number
  createdAt: string
  updatedAt: string
}

/**
 * Execution log model (aligned with backend handler.Execution)
 */
export interface LogModel {
  id: number
  taskId: number
  status: string // pending, running, success, failed
  output: string
  error?: string
  startedAt: string
  finishedAt?: string
  /** Display-only fields enriched by backend list joins */
  taskName?: string
  executorType?: string
  duration?: number
  statusCode?: number
  retryCount?: number
}

/**
 * Execution stats (aligned with backend handler.ExecutionStats)
 */
export interface ExecutionStats {
  totalExecutions: number
  successCount: number
  failureCount: number
  successRate: number
  averageDurationMs: number
}

/**
 * Task creation parameters (aligned with backend handler.Task request body)
 */
export interface TaskCreateParams {
  name: string
  description?: string
  executor: string
  schedule: string
  enabled: boolean
  config?: Record<string, unknown>
  group?: string
  tags?: string[]
  retryPolicy?: string
  concurrency?: number
}

/**
 * Task update parameters (same fields as create; id is in the URL path)
 */
export type TaskUpdateParams = TaskCreateParams

/**
 * Task form data (shared by Create/Edit)
 *
 * Contains backend-compatible fields plus form-only UI fields used to
 * compose the `schedule` string. The form-only fields (scheduleType,
 * cronExpr, interval) are stripped by the parent component before
 * sending the API request.
 */
export interface TaskFormData {
  // Backend-compatible fields
  name: string
  description: string
  executor: ExecutorType
  schedule: string
  config: Record<string, unknown>
  group: string
  tags: string[]
  enabled: boolean
  retryPolicy: string
  concurrency: number

  // Form-only UI fields (not sent to backend)
  scheduleType: ScheduleType
  cronExpr: string
  interval: number
}
