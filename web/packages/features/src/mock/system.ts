// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

import type { MockMethod } from './types'

/** System basic configuration (aligned with prototype systemSettings) */
const mockSettings = {
  site_name: 'Tickraft',
  logo_url: '',
  default_lang: 'zh-Hans',
  default_theme: 'light',
  log_level: 'info',
  /** Data retention days (7-365) */
  retention_days: 30,
  /** Login session timeout (minutes) */
  session_timeout: 720,
}

/** Runtime info (aligned with prototype systemInfo, includes Go version) */
const mockRuntimeInfo = {
  version: 'v0.1.0',
  build_tags: 'ce',
  go_version: 'go1.22.3',
  start_time: '2026-06-29 10:00:00',
  uptime: '2d 4h 32m',
}

export default [
  {
    url: '/api/v1/system/config',
    method: 'get',
    response: () => ({
      code: 0,
      message: 'success',
      data: { ...mockSettings },
    }),
  },
  {
    url: '/api/v1/system/config',
    method: 'put',
    response: ({ body }: { body: Record<string, unknown> }) => {
      Object.assign(mockSettings, body)
      return {
        code: 0,
        message: 'success',
        data: { ...mockSettings },
      }
    },
  },
  {
    url: '/api/v1/system/info',
    method: 'get',
    response: () => ({
      code: 0,
      message: 'success',
      data: { ...mockRuntimeInfo },
    }),
  },
  {
    url: '/api/v1/health',
    method: 'get',
    response: () => ({
      code: 0,
      message: 'success',
      data: {
        status: 'ok',
        version: mockRuntimeInfo.version,
        uptime: 186_720,
      },
    }),
  },
] as MockMethod[]
