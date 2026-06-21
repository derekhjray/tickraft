// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

import { ref, computed, onBeforeUnmount } from 'vue'
import type { Ref, ComputedRef } from 'vue'
import { getToken } from '../utils/request'

/** WebSocket connection status */
type WSStatus = 'connecting' | 'connected' | 'reconnecting' | 'disconnected'

/** Exponential backoff reconnection config: initial 1s, doubles each time, capped at 30s */
const RECONNECT_CONFIG = {
  /** Initial backoff delay (ms) */
  initialDelay: 1000,
  /** Maximum backoff delay (ms) */
  maxDelay: 30000,
  /** Default maximum reconnection attempts */
  maxRetries: 10,
  /** Backoff multiplier */
  multiplier: 2,
} as const

/** Heartbeat ping interval (ms) */
const HEARTBEAT_INTERVAL = 30000
/** Heartbeat pong timeout (ms): if no pong is received within 60s, the connection is considered lost and a reconnect is triggered */
const PONG_TIMEOUT = 60000
/** Default maximum message queue length; oldest messages are dropped when exceeded */
const DEFAULT_MAX_QUEUE = 100

/** useWebSocket options */
export interface UseWebSocketOptions {
  /** WebSocket path (the full URL is built from the current host) */
  url: string
  /** Whether to auto-connect; defaults to true */
  autoConnect?: boolean
  /** Whether to auto-reconnect after disconnection; defaults to true */
  reconnect?: boolean
  /** Maximum reconnection attempts; defaults to 10 */
  maxRetries?: number
  /** Maximum message queue length; defaults to 100; oldest messages are dropped when exceeded */
  maxQueueSize?: number
  /** Business message callback (triggered when a non-heartbeat message is received) */
  onMessage?: (message: Record<string, unknown>) => void
  /** Reconnect failed callback (triggered when the maximum reconnection attempts is reached; equivalent to the 'reconnect-failed' event) */
  onReconnectFailed?: () => void
}

/** useWebSocket return type */
export interface UseWebSocketReturn {
  /** Connection status */
  status: Ref<WSStatus>
  /** Latest business message */
  data: Ref<Record<string, unknown> | null>
  /** Latest error message */
  lastError: Ref<string | null>
  /** Whether connected */
  connected: ComputedRef<boolean>
  /** Whether reconnecting */
  reconnecting: ComputedRef<boolean>
  /** Current reconnection count */
  reconnectCount: Ref<number>
  /** Establish a connection */
  connect: () => void
  /** Disconnect (prevents auto-reconnect) */
  disconnect: () => void
  /** Send a message (cached to the queue while disconnected; auto-flushed after reconnect) */
  send: (message: Record<string, unknown>) => void
  /** Manually trigger a reconnect (exponential backoff) */
  reconnect: (attempt?: number) => void
  /** Clear the pending message queue */
  clearQueue: () => void
}

/**
 * WebSocket connection management composable.
 *
 * - Exponential backoff reconnection: delay = min(1000 * 2^attempt, 30000), up
 *   to 10 attempts (configurable); on failure, the `onReconnectFailed` callback
 *   is triggered (equivalent to the 'reconnect-failed' event).
 * - Heartbeat detection: sends a ping every 30s; if no pong is received within
 *   60s, the connection is considered lost and a reconnect is triggered.
 * - Message queue: messages sent while disconnected are cached in a queue and
 *   auto-flushed after reconnect; queue max length is 100 (configurable), and
 *   the oldest messages are dropped when exceeded.
 * @param options - WebSocket options
 * @returns connection status and control methods
 */
export function useWebSocket(options: UseWebSocketOptions): UseWebSocketReturn {
  const {
    autoConnect = true,
    reconnect: autoReconnect = true,
    maxRetries = RECONNECT_CONFIG.maxRetries,
    maxQueueSize = DEFAULT_MAX_QUEUE,
    onMessage,
    onReconnectFailed,
  } = options

  const status: Ref<WSStatus> = ref('disconnected')
  const data: Ref<Record<string, unknown> | null> = ref(null)
  const lastError: Ref<string | null> = ref(null)
  const reconnectAttempts = ref(0)

  /** Whether connected */
  const connected: ComputedRef<boolean> = computed(() => status.value === 'connected')
  /** Whether reconnecting */
  const reconnecting: ComputedRef<boolean> = computed(() => status.value === 'reconnecting')
  /** Current reconnection count (semantic alias) */
  const reconnectCount = reconnectAttempts

  /** Pending message queue (cached while the connection is down) */
  const messageQueue: Array<Record<string, unknown>> = []

  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null
  let heartbeatTimeoutTimer: ReturnType<typeof setTimeout> | null = null
  /** Timestamp of the last received pong */
  let lastPongTime = 0
  /** Whether reconnect-failed has already been emitted (avoids duplicate triggers) */
  let reconnectFailedEmitted = false

  /**
   * Get the full WebSocket URL
   * @returns full URL including the token
   */
  function getWsUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const token = getToken()
    return `${protocol}//${host}${options.url}?token=${token}`
  }

  /**
   * Compute the exponential backoff delay
   * @param attempt - current reconnection attempt
   * @returns delay in milliseconds
   */
  function getReconnectDelay(attempt: number): number {
    return Math.min(
      RECONNECT_CONFIG.initialDelay * Math.pow(RECONNECT_CONFIG.multiplier, attempt),
      RECONNECT_CONFIG.maxDelay,
    )
  }

  /**
   * Start heartbeat detection
   *
   * Sends a ping every 30s and checks whether more than 60s has elapsed since
   * the last pong; if so, the connection is considered lost and a forced
   * reconnect is triggered.
   */
  function startHeartbeat(): void {
    stopHeartbeat()
    lastPongTime = Date.now()
    heartbeatTimer = setInterval(() => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'ping' }))
        // Check whether the pong timeout is exceeded (no pong for 60s)
        if (Date.now() - lastPongTime > PONG_TIMEOUT) {
          forceReconnect()
          return
        }
      }
    }, HEARTBEAT_INTERVAL)
  }

  /**
   * Stop heartbeat detection
   */
  function stopHeartbeat(): void {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
    if (heartbeatTimeoutTimer) {
      clearTimeout(heartbeatTimeoutTimer)
      heartbeatTimeoutTimer = null
    }
  }

  /**
   * Enqueue a pending message
   *
   * Drops the oldest messages when the max length is exceeded to avoid
   * unbounded memory growth.
   * @param message - message body
   */
  function enqueueMessage(message: Record<string, unknown>): void {
    while (messageQueue.length >= maxQueueSize) {
      messageQueue.shift()
    }
    messageQueue.push(message)
  }

  /**
   * Flush the pending message queue (called after a successful reconnect)
   */
  function flushMessageQueue(): void {
    while (messageQueue.length > 0) {
      const message = messageQueue.shift()
      if (message !== undefined && ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(message))
      }
    }
  }

  /**
   * Clear the pending message queue
   */
  function clearQueue(): void {
    messageQueue.length = 0
  }

  /**
   * Handle connection open
   */
  function handleOpen(): void {
    status.value = 'connected'
    reconnectAttempts.value = 0
    lastError.value = null
    reconnectFailedEmitted = false
    startHeartbeat()
    flushMessageQueue()
  }

  /**
   * Handle a received message
   *
   * pong messages refresh the heartbeat timer; business messages are written
   * to data and trigger the onMessage callback.
   * @param event - message event
   */
  function handleMessage(event: MessageEvent): void {
    try {
      const parsed = JSON.parse(event.data as string) as Record<string, unknown>
      if (parsed.type === 'pong') {
        lastPongTime = Date.now()
        if (heartbeatTimeoutTimer) {
          clearTimeout(heartbeatTimeoutTimer)
          heartbeatTimeoutTimer = null
        }
        return
      }
      data.value = parsed
      onMessage?.(parsed)
    } catch {
      // Non-JSON messages are ignored
    }
  }

  /**
   * Schedule an auto-reconnect
   *
   * When the maximum reconnection attempts is reached, the 'reconnect-failed'
   * event is triggered via onReconnectFailed (only triggered once).
   */
  function scheduleReconnect(): void {
    if (!autoReconnect) return
    if (reconnectAttempts.value >= maxRetries) {
      status.value = 'disconnected'
      if (!reconnectFailedEmitted) {
        reconnectFailedEmitted = true
        onReconnectFailed?.()
      }
      return
    }
    status.value = 'reconnecting'
    const delay = getReconnectDelay(reconnectAttempts.value)
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
    }
    reconnectTimer = setTimeout(() => {
      reconnectAttempts.value += 1
      connect()
    }, delay)
  }

  /**
   * Handle connection close
   */
  function handleClose(): void {
    status.value = 'disconnected'
    stopHeartbeat()
    scheduleReconnect()
  }

  /**
   * Handle connection error
   */
  function handleError(): void {
    lastError.value = 'WebSocket connection error'
    status.value = 'disconnected'
    stopHeartbeat()
    // onError is usually followed by onClose, which schedules the reconnect
  }

  /**
   * Force a reconnect (triggered by heartbeat timeout)
   */
  function forceReconnect(): void {
    if (ws) {
      ws.onclose = null
      ws.close()
      ws = null
    }
    stopHeartbeat()
    status.value = 'disconnected'
    scheduleReconnect()
  }

  /**
   * Establish a connection
   */
  function connect(): void {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    stopHeartbeat()
    if (ws) {
      ws.onclose = null
      ws.close()
      ws = null
    }
    status.value = 'connecting'
    try {
      ws = new WebSocket(getWsUrl())
      ws.onopen = handleOpen
      ws.onmessage = handleMessage
      ws.onclose = handleClose
      ws.onerror = handleError
    } catch (err) {
      lastError.value = err instanceof Error ? err.message : 'Failed to create WebSocket'
      status.value = 'disconnected'
    }
  }

  /**
   * Disconnect (prevents auto-reconnect)
   */
  function disconnect(): void {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    stopHeartbeat()
    reconnectAttempts.value = maxRetries
    if (ws) {
      ws.onclose = null
      ws.close()
      ws = null
    }
    status.value = 'disconnected'
  }

  /**
   * Manually trigger a reconnect (exponential backoff)
   * @param attempt - starting reconnection attempt; defaults to the current count
   */
  function reconnect(attempt: number = reconnectAttempts.value): void {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    stopHeartbeat()
    if (ws) {
      ws.onclose = null
      ws.close()
      ws = null
    }
    reconnectAttempts.value = Math.min(attempt, maxRetries)
    reconnectFailedEmitted = false
    const delay = getReconnectDelay(reconnectAttempts.value)
    status.value = 'reconnecting'
    reconnectTimer = setTimeout(() => {
      reconnectAttempts.value += 1
      connect()
    }, delay)
  }

  /**
   * Send a message (cached to the queue while disconnected; auto-flushed after reconnect)
   * @param message - message body
   */
  function send(message: Record<string, unknown>): void {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(message))
    } else {
      enqueueMessage(message)
    }
  }

  // Auto-connect
  if (autoConnect) {
    connect()
  }

  // Disconnect on unmount to avoid memory leaks
  onBeforeUnmount(() => {
    disconnect()
  })

  return {
    status,
    data,
    lastError,
    connected,
    reconnecting,
    reconnectCount,
    connect,
    disconnect,
    send,
    reconnect,
    clearQueue,
  }
}
