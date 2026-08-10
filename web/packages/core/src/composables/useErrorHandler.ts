// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

/**
 * useErrorHandler - standardized 3-layer error feedback for components.
 *
 * Establishes a consistent error-feedback model across all frontends:
 *
 * - **Form-level** (`reportFieldError` / `fieldErrors`): inline error messages
 *   rendered beneath the offending form field (red text). Use for validation
 *   failures bound to a specific field. Cleared by `clearFieldError` or
 *   `clear`.
 * - **Operation-level** (`notify`): short-lived transient feedback via
 *   `ElMessage` for user-initiated actions (save/delete/submit). Auto-dismisses.
 * - **Page-level** (`alert`): persistent notification via `ElNotification`
 *   for critical/blocking errors the user must acknowledge (e.g. data load
 *   failure that leaves the page unusable).
 *
 * The composable also tracks the latest error (`error`/`message`/`hasError`)
 * so callers can render an inline page-level error banner if desired.
 *
 * Element Plus is lazy-imported so the composable has no module-load side
 * effects and works in any Element Plus host project.
 */
import { computed, ref, type ComputedRef, type Ref } from 'vue'

/** Severity levels for transient and persistent notifications. */
export type ErrorSeverity = 'error' | 'warning' | 'info'

/** Layer at which an error is surfaced. */
export type ErrorLayer = 'form' | 'operation' | 'page'

export interface UseErrorHandlerOptions {
  /**
   * Custom transient notifier (defaults to lazy ElMessage). Override to plug
   * in a different toast library or i18n-aware notifier.
   */
  notifier?: (message: string, severity: ErrorSeverity) => void
  /**
   * Custom persistent notifier (defaults to lazy ElNotification). Override for
   * a different banner/region mechanism.
   */
  alerter?: (title: string, message: string, severity: ErrorSeverity) => void
  /** Custom message extractor (defaults to {@link extractErrorMessage}). */
  extractor?: (err: unknown, fallback?: string) => string
  /** Whether operation-level errors auto-notify via ElMessage (default `true`). */
  autoNotify?: boolean
  /** Whether to log every handled error to `console.error` (default `true`). */
  logToConsole?: boolean
}

export interface UseErrorHandlerReturn {
  /** The last captured error (`null` when cleared or never set). */
  error: Ref<unknown>
  /** Human-readable message extracted from the last error. */
  message: ComputedRef<string>
  /** Whether an error is currently active. */
  hasError: ComputedRef<boolean>
  /** Form-field error map (field name -> message). Reactive. */
  fieldErrors: Ref<Record<string, string>>
  /** Whether any form-field error is present. */
  hasFieldError: ComputedRef<boolean>

  /**
   * Record an error and surface it at the chosen layer.
   *
   * @param err - the thrown value (Error / ApiError / unknown)
   * @param fallbackMessage - i18n message shown when the error has no message
   * @param layer - feedback layer (default `operation`)
   */
  handleError: (err: unknown, fallbackMessage?: string, layer?: ErrorLayer) => void
  /**
   * Set an inline error for a specific form field (form-level feedback).
   *
   * @param field - form field key (must match the field's `prop`/`error` slot)
   * @param message - inline error message (pass empty string to clear)
   */
  reportFieldError: (field: string, message: string) => void
  /** Clear the inline error for a single form field. */
  clearFieldError: (field: string) => void
  /** Show an operation-level transient toast (ElMessage). */
  notify: (message: string, severity?: ErrorSeverity) => void
  /** Show a page-level persistent notification (ElNotification). */
  alert: (title: string, message: string, severity?: ErrorSeverity) => void
  /**
   * Wrap an async function so any thrown error is auto-handled at the
   * operation layer. Returns `undefined` on failure.
   */
  guard: <TArgs extends unknown[], TResult>(
    fn: (...args: TArgs) => Promise<TResult>,
    fallbackMessage?: string,
  ) => (...args: TArgs) => Promise<TResult | undefined>
  /** Clear all error state (last error + form-field errors). */
  clear: () => void
}

/**
 * Extract a human-readable message from a thrown value.
 *
 * Recognizes:
 * - `null`/`undefined` -> fallback
 * - Plain strings
 * - `Error` instances (uses `.message`)
 * - Objects with a `.message` string property (ApiError, etc.)
 * - Everything else falls back to `String(value)` or the provided fallback
 */
export function extractErrorMessage(err: unknown, fallback?: string): string {
  if (err === null || err === undefined) {
    return fallback ?? 'Unknown error'
  }
  if (typeof err === 'string') return err
  if (err instanceof Error) return err.message || (fallback ?? 'Unknown error')
  if (typeof err === 'object' && 'message' in err) {
    const msg = (err as { message: unknown }).message
    if (typeof msg === 'string' && msg.length > 0) return msg
  }
  return fallback ?? String(err)
}

/** Default operation-level notifier using ElMessage (lazy-loaded). */
async function defaultNotifier(message: string, severity: ErrorSeverity): Promise<void> {
  const { ElMessage } = await import('element-plus')
  if (severity === 'warning') {
    ElMessage.warning(message)
  } else if (severity === 'info') {
    ElMessage.info(message)
  } else {
    ElMessage.error(message)
  }
}

/** Default page-level alerter using ElNotification (lazy-loaded). */
async function defaultAlerter(
  title: string,
  message: string,
  severity: ErrorSeverity,
): Promise<void> {
  const { ElNotification } = await import('element-plus')
  const type: 'error' | 'warning' | 'info' =
    severity === 'warning' ? 'warning' : severity === 'info' ? 'info' : 'error'
  ElNotification({
    title,
    message,
    type,
    duration: 0,
  })
}

/**
 * Create a standardized 3-layer error handler for component-level feedback.
 *
 * @param options - configuration for notification, logging, and extraction
 */
export function useErrorHandler(
  options: UseErrorHandlerOptions = {},
): UseErrorHandlerReturn {
  const {
    notifier,
    alerter,
    extractor,
    autoNotify = true,
    logToConsole = true,
  } = options
  const resolveMessage = extractor ?? extractErrorMessage

  const error = ref<unknown>(null)
  const fieldErrors = ref<Record<string, string>>({})

  const message = computed(() => resolveMessage(error.value))
  const hasError = computed(() => error.value !== null)
  const hasFieldError = computed(() => Object.keys(fieldErrors.value).length > 0)

  function notify(msg: string, severity: ErrorSeverity = 'error'): void {
    if (notifier) {
      notifier(msg, severity)
    } else {
      void defaultNotifier(msg, severity)
    }
  }

  function alert(title: string, msg: string, severity: ErrorSeverity = 'error'): void {
    if (alerter) {
      alerter(title, msg, severity)
    } else {
      void defaultAlerter(title, msg, severity)
    }
  }

  function reportFieldError(field: string, msg: string): void {
    if (msg === '') {
      clearFieldError(field)
      return
    }
    fieldErrors.value = { ...fieldErrors.value, [field]: msg }
  }

  function clearFieldError(field: string): void {
    if (!(field in fieldErrors.value)) return
    const next = { ...fieldErrors.value }
    delete next[field]
    fieldErrors.value = next
  }

  function handleError(
    err: unknown,
    fallbackMessage?: string,
    layer: ErrorLayer = 'operation',
  ): void {
    error.value = err
    const msg = resolveMessage(err, fallbackMessage)

    if (logToConsole) {
      console.error('[useErrorHandler]', err)
    }

    if (layer === 'form') {
      // Form-level: caller must supply a field via reportFieldError; here we
      // only store the error state and skip transient notification.
      return
    }

    if (layer === 'page') {
      alert(fallbackMessage ?? 'Error', msg)
      return
    }

    // operation layer
    if (autoNotify) {
      notify(msg)
    }
  }

  function guard<TArgs extends unknown[], TResult>(
    fn: (...args: TArgs) => Promise<TResult>,
    fallbackMessage?: string,
  ): (...args: TArgs) => Promise<TResult | undefined> {
    return async (...args: TArgs): Promise<TResult | undefined> => {
      try {
        return await fn(...args)
      } catch (err) {
        handleError(err, fallbackMessage)
        return undefined
      }
    }
  }

  function clear(): void {
    error.value = null
    fieldErrors.value = {}
  }

  return {
    error,
    message,
    hasError,
    fieldErrors,
    hasFieldError,
    handleError,
    reportFieldError,
    clearFieldError,
    notify,
    alert,
    guard,
    clear,
  }
}
