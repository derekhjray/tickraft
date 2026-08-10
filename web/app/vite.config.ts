// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import legacy from '@vitejs/plugin-legacy'
import UnoCSS from 'unocss/vite'
import Icons from 'unplugin-icons/vite'
import IconsResolver from 'unplugin-icons/resolver'
import Components from 'unplugin-vue-components/vite'
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite'
import { mockServerPlugin } from './vite-plugins/mock-server'
import path from 'path'

export default defineConfig({
  plugins: [
    vue(),
    UnoCSS(),
    Icons({ compiler: 'vue3', autoInstall: true }),
    Components({
      resolvers: [IconsResolver({ enabledCollections: ['ep'] })],
    }),
    VueI18nPlugin({
      // 同时纳入内核（common 命名空间）与开源业务（auth/scheduler/collector/alert/system）的语言包
      // Note: paths intentionally use runtime JSON import fallback; pre-compilation is
      // disabled because some locale messages contain intentional HTML (e.g. <code>).
      include: [
        path.resolve(__dirname, '../../packages/core/src/i18n/locales/**'),
        path.resolve(__dirname, '../../packages/features/src/i18n/locales/**'),
      ],
    }),
    mockServerPlugin({
      // mock 定义已迁移至 @tickraft/features 包内
      // Vite 以 web/app 为 cwd，packages 在 web/ 下，故用 ../packages（一个 .. ）
      // Mock disabled by default — frontend now connects to the real backend
      // via the vite proxy below. Set enable: true to use mock data for
      // standalone frontend development when the backend is unavailable.
      mockPath: '../packages/features/src/mock',
      enable: true,
      logger: true,
    }),
    legacy({
      targets: ['defaults and not dead', '> 0.5%', 'Firefox ESR', 'last 2 versions'],
      renderLegacyChunks: true,
      additionalLegacyPolyfills: ['core-js/stable'],
    }),
  ],
  resolve: {
    // @tickraft/core 与 @tickraft/features 通过 pnpm workspace 符号链接自动解析，无需额外 alias
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        // Use the Sass modern JS API to remove the Dart Sass legacy-js-api deprecation warning.
        api: 'modern',
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:6153',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:6153',
        ws: true,
        changeOrigin: true,
      },
    },
  },
  build: {
    // ECharts is a lazy chunk (loaded on first chart render) and Element Plus
    // is a large vendor chunk; the legacy plugin further inflates Element Plus
    // for older browsers. Raise the limit so these expected third-party sizes
    // do not raise noise while app code stays under scrutiny.
    chunkSizeWarningLimit: 1700,
    rollupOptions: {
      output: {
        manualChunks(id) {
          // Group Element Plus into a vendor chunk. ECharts is left to Rollup's
          // automatic chunking: it is reached only via a dynamic import
          // (`import('./echarts')` in `useChart.ts`), so Rollup emits it as a
          // single lazy, tree-shaken chunk without an explicit entry here.
          if (id.includes('/node_modules/element-plus/')) {
            return 'element-plus'
          }
        },
      },
    },
  },
})
