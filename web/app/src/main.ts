// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

import { createApp, type Component } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import ElementPlus from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'
import 'virtual:uno.css'

// Global styles must load before component rendering to ensure CSS variables
// (Design Tokens) are available. Loaded via @tickraft/core/styles subpath:
// tokens → themes → element → common
import '@tickraft/core/styles'

import { App, createRouter, createI18n, vFeature, BASE_MENUS_KEY } from '@tickraft/core'
import { baseRoutes, baseMessages, baseMenus } from '@tickraft/features'

// Open-source standalone entry: builds router and i18n from kernel base data.
// Extension app (tickraft-x/web/src/main.ts) merges extension data before
// calling the same kernel factories.
const router = createRouter(baseRoutes)
const i18n = createI18n(baseMessages)

const app = createApp(App)

// Pinia state management
const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)
app.use(pinia)

// i18n
app.use(i18n)

// Router
app.use(router)

// Element Plus
app.use(ElementPlus)

// Globally register Element Plus icon components.
// Menu icons resolve via <component :is="iconName"> using these registrations.
// Login.vue etc. also reference icons by string name (prefix-icon="User").
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component as Component)
}

// Custom directive
app.directive('feature', vFeature)

// Provide base menus to DefaultLayout via injection key.
// This breaks the circular dependency: core DefaultLayout needs baseMenus
// from features, but features depends on core. App root provides it here.
app.provide(BASE_MENUS_KEY, baseMenus)

// Dev error handler for diagnosing render issues
app.config.errorHandler = (err, _instance, info) => {
  console.error('[Web Error]', err, info)
}

app.mount('#app')
