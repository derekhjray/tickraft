// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

import type { MockMethod } from './types'

/** Asset type identifier (aligned with telemetry asset types) */
export type AssetType = 'task' | 'device' | 'host' | 'port' | 'website' | 'service'

/** Asset distribution data item */
export interface AssetDistributionItem {
  type: AssetType
  value: number
}

/** Alert severity level */
export type AlertSeverity = 'critical' | 'warning' | 'info'

/** Alert trend data item (per day, with counts by severity) */
export interface AlertTrendItem {
  date: string
  count: number
  critical: number
  warning: number
  info: number
}

/** Asset status distribution item */
export interface AssetStatusItem {
  status: 'normal' | 'warning' | 'error' | 'offline' | 'maintenance' | 'unknown'
  count: number
}

/** Dashboard overview statistics */
export interface DashboardOverview {
  totalAssets: number
  onlineRate: number
  alertCount: number
  taskCount: number
}

/** Dashboard aggregate stats (aligned with prototype dashboardStats) */
export interface DashboardStats {
  totalAssets: number
  onlineAssets: number
  firingAlerts: number
  todayExecutions: number
  successRate: number
  activeTasks: number
}

/** System health info item */
export interface SystemHealthItem {
  labelKey: string
  value: string
  state: 'ok' | 'warn' | ''
}

/** System information (aligned with prototype systemInfo) */
export interface SystemInfo {
  version: string
  uptime: string
  goroutines: number
  dbSize: string
}

/** Recent alert record for dashboard table */
export interface RecentAlert {
  id: number
  severity: AlertSeverity
  assetName: string
  message: string
  status: 'firing' | 'resolved'
  firedAt: string
}

/** Task execution trend data item (per day) */
export interface TaskExecTrendItem {
  date: string
  success: number
  failed: number
}

/** Alert threshold (for trend chart markLine) */
export const ALERT_THRESHOLD = 10

/** Asset distribution by type mock (6 asset types) */
export const assetDistribution: AssetDistributionItem[] = [
  { type: 'host', value: 64 },
  { type: 'service', value: 32 },
  { type: 'website', value: 16 },
  { type: 'device', value: 10 },
  { type: 'port', value: 8 },
  { type: 'task', value: 6 },
]

/** Asset status distribution mock (aligned with prototype) */
export const assetStatusDistribution: AssetStatusItem[] = [
  { status: 'normal', count: 12 },
  { status: 'warning', count: 3 },
  { status: 'error', count: 2 },
  { status: 'offline', count: 2 },
  { status: 'maintenance', count: 1 },
]

/** Last 7 days alert trend mock (with severity breakdown) */
export const alertTrend: AlertTrendItem[] = [
  { date: '07-19', count: 13, critical: 2, warning: 8, info: 3 },
  { date: '07-20', count: 9, critical: 1, warning: 6, info: 2 },
  { date: '07-21', count: 18, critical: 3, warning: 11, info: 4 },
  { date: '07-22', count: 6, critical: 0, warning: 5, info: 1 },
  { date: '07-23', count: 11, critical: 2, warning: 7, info: 2 },
  { date: '07-24', count: 13, critical: 1, warning: 9, info: 3 },
  { date: '07-25', count: 10, critical: 2, warning: 6, info: 2 },
]

/** Task execution trend mock (last 7 days, success/failed) */
export const taskExecTrend: TaskExecTrendItem[] = [
  { date: '07-19', success: 168, failed: 12 },
  { date: '07-20', success: 172, failed: 8 },
  { date: '07-21', success: 165, failed: 15 },
  { date: '07-22', success: 178, failed: 6 },
  { date: '07-23', success: 170, failed: 10 },
  { date: '07-24', success: 175, failed: 9 },
  { date: '07-25', success: 182, failed: 7 },
]

/** Dashboard overview statistics mock */
export const dashboardOverview: DashboardOverview = {
  totalAssets: 128,
  onlineRate: 96.8,
  alertCount: 5,
  taskCount: 47,
}

/** Dashboard aggregate stats mock (aligned with prototype) */
export const dashboardStats: DashboardStats = {
  totalAssets: 20,
  onlineAssets: 15,
  firingAlerts: 5,
  todayExecutions: 1284,
  successRate: 96.8,
  activeTasks: 21,
}

/** System information mock (aligned with prototype) */
export const systemInfo: SystemInfo = {
  version: 'v1.2.0',
  uptime: '32d 14h 26m',
  goroutines: 128,
  dbSize: '1.84 GB',
}

/** System health items for dashboard health card */
export const systemHealthItems: SystemHealthItem[] = [
  { labelKey: 'common.dashboard.health.version', value: systemInfo.version, state: '' },
  { labelKey: 'common.dashboard.health.uptime', value: systemInfo.uptime, state: '' },
  { labelKey: 'common.dashboard.health.activeTasks', value: String(dashboardStats.activeTasks), state: 'ok' },
  {
    labelKey: 'common.dashboard.health.successRate',
    value: `${dashboardStats.successRate.toFixed(1)}%`,
    state: dashboardStats.successRate >= 95 ? 'ok' : 'warn',
  },
  { labelKey: 'common.dashboard.health.goroutines', value: String(systemInfo.goroutines), state: '' },
  { labelKey: 'common.dashboard.health.dbSize', value: systemInfo.dbSize, state: '' },
]

/** Recent alert records for dashboard table mock */
export const recentAlerts: RecentAlert[] = [
  {
    id: 30000,
    severity: 'critical',
    assetName: 'prod-web-01',
    message: 'CPU 使用率超过 90%',
    status: 'firing',
    firedAt: '2026-07-25 13:58:00',
  },
  {
    id: 30001,
    severity: 'critical',
    assetName: 'prod-db-01',
    message: '磁盘空间不足 10%',
    status: 'firing',
    firedAt: '2026-07-25 13:45:00',
  },
  {
    id: 30002,
    severity: 'warning',
    assetName: 'prod-db-02',
    message: '内存使用率超过 85%',
    status: 'firing',
    firedAt: '2026-07-25 13:30:00',
  },
  {
    id: 30003,
    severity: 'warning',
    assetName: 'cert-tickraft-io',
    message: 'SSL 证书将于 7 天后过期',
    status: 'firing',
    firedAt: '2026-07-25 13:00:00',
  },
  {
    id: 30004,
    severity: 'info',
    assetName: 'prod-cache-01',
    message: 'Redis 内存碎片率过高',
    status: 'resolved',
    firedAt: '2026-07-25 12:15:00',
  },
  {
    id: 30005,
    severity: 'warning',
    assetName: 'prod-api-03',
    message: 'HTTP 5xx 错误率超过 5%',
    status: 'resolved',
    firedAt: '2026-07-25 11:30:00',
  },
  {
    id: 30006,
    severity: 'info',
    assetName: 'prod-web-02',
    message: '网络丢包率超过 1%',
    status: 'resolved',
    firedAt: '2026-07-25 10:45:00',
  },
  {
    id: 30007,
    severity: 'critical',
    assetName: 'prod-db-02',
    message: '数据库连接池已满',
    status: 'resolved',
    firedAt: '2026-07-25 09:20:00',
  },
]

export default [
  {
    url: '/api/v1/dashboard/overview',
    method: 'get',
    response: () => ({
      code: 0,
      message: 'success',
      data: {
        overview: dashboardOverview,
        stats: dashboardStats,
        assetDistribution,
        assetStatusDistribution,
        alertTrend,
        taskExecTrend,
        systemInfo,
        systemHealthItems,
        recentAlerts,
        alertThreshold: ALERT_THRESHOLD,
      },
    }),
  },
] as MockMethod[]
