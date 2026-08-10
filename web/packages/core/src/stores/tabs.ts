// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { TabItem } from '../types/global'

/**
 * Multi-tab management store.
 */
export const useTabsStore = defineStore('tabs', () => {
  /** Tab list */
  const tabs = ref<TabItem[]>([])

  /** Path of the currently active tab */
  const activeTab = ref('')

  /**
   * Add a tab.
   */
  function addTab(tab: TabItem): void {
    const exists = tabs.value.some((item) => item.path === tab.path)
    if (!exists) {
      tabs.value.push(tab)
    }
    activeTab.value = tab.path
  }

  /**
   * Close a tab.
   */
  function closeTab(path: string): string | null {
    const index = tabs.value.findIndex((tab) => tab.path === path)
    if (index === -1) return null

    tabs.value.splice(index, 1)

    // If the closed tab was the active one, switch to a neighboring tab
    if (activeTab.value === path) {
      if (tabs.value.length === 0) return '/dashboard'
      const nextIndex = Math.min(index, tabs.value.length - 1)
      return tabs.value[nextIndex].path
    }

    return null
  }

  /**
   * Close other tabs.
   */
  function closeOtherTabs(path: string): void {
    tabs.value = tabs.value.filter((tab) => tab.path === path || !tab.closable)
    activeTab.value = path
  }

  /**
   * Close all closable tabs.
   */
  function closeAllTabs(): string {
    tabs.value = tabs.value.filter((tab) => !tab.closable)
    return tabs.value.length > 0 ? tabs.value[0].path : '/dashboard'
  }

  return {
    tabs,
    activeTab,
    addTab,
    closeTab,
    closeOtherTabs,
    closeAllTabs,
  }
})
