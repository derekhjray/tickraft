// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

/**
 * useLoading - unified loading-state primitive with flicker debounce.
 *
 * Standardizes how loading state is tracked and rendered across the app:
 * - `mode` declares which loading affordance a view uses (`skeleton` for
 *   content-placeholder skeletons, `spinner` for overlay spinners). Consumers
 *   read `mode` to decide which template branch to render.
 * - `debounce` delays flipping `isLoading` to `true` so very fast responses
 *   don't flash a spinner/skeleton. `done()` cancels any pending timer and
 *   hides the loading state immediately.
 *
 * This is the low-level manual-control primitive. For wrapping an async
 * function with automatic loading/error tracking, see `@tickraft/ui`'s
 * `useLoading` async-wrapper variant.
 */
import { onBeforeUnmount, readonly, ref, type Ref } from 'vue'

/** Loading affordance the consuming view renders. */
export type LoadingMode = 'skeleton' | 'spinner'

export interface UseLoadingOptions {
  /** Loading affordance declared for the view (default `spinner`) */
  mode?: LoadingMode
  /**
   * Delay in ms before `isLoading` becomes `true` after `start()` is called.
   * Fast operations that finish within this window never show a loading state,
   * avoiding flicker. Set to `0` to show immediately (default `0`).
   */
  debounce?: number
}

export interface UseLoadingReturn {
  /** Whether loading is currently visible (readonly reactive flag) */
  isLoading: Readonly<Ref<boolean>>
  /** Declared loading affordance for the view (readonly) */
  mode: Readonly<Ref<LoadingMode>>
  /** Mark the operation as started (shows loading after the debounce window) */
  start: () => void
  /** Mark the operation as finished (hides loading immediately) */
  done: () => void
  /** Toggle loading state directly */
  toggle: (value?: boolean) => void
}

/**
 * Create a reactive loading state with optional flicker debounce.
 *
 * The debounce timer is cleared on unmount so no stale timer can flip
 * `isLoading` after the component is gone.
 *
 * @param options - loading configuration
 */
export function useLoading(options: UseLoadingOptions = {}): UseLoadingReturn {
  const { mode = 'spinner', debounce = 0 } = options

  const isLoading = ref(false)
  const modeRef = ref<LoadingMode>(mode)
  let debounceTimer: ReturnType<typeof setTimeout> | null = null

  function clearTimer(): void {
    if (debounceTimer !== null) {
      clearTimeout(debounceTimer)
      debounceTimer = null
    }
  }

  function start(): void {
    clearTimer()
    if (debounce > 0) {
      debounceTimer = setTimeout(() => {
        isLoading.value = true
      }, debounce)
    } else {
      isLoading.value = true
    }
  }

  function done(): void {
    clearTimer()
    isLoading.value = false
  }

  function toggle(value?: boolean): void {
    if (typeof value === 'boolean') {
      if (value) {
        start()
      } else {
        done()
      }
    } else {
      // Flip the current state.
      if (isLoading.value) {
        done()
      } else {
        start()
      }
    }
  }

  onBeforeUnmount(() => {
    clearTimer()
  })

  return {
    isLoading: readonly(isLoading),
    mode: readonly(modeRef),
    start,
    done,
    toggle,
  }
}
