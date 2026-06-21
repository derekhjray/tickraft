// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

import js from '@eslint/js'
import tseslint from 'typescript-eslint'
import vuePlugin from 'eslint-plugin-vue'

export default tseslint.config(
  // Global ignores: build artifacts, dependencies, legacy/storyboard snapshots
  {
    ignores: [
      '**/dist/**',
      '**/node_modules/**',
      '**/sb.bak/**',
      '**/.storyboard/**',
      '**/*.d.ts',
    ],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...vuePlugin.configs['flat/recommended'],
  // Configure TypeScript parser for <script lang="ts"> blocks in Vue SFCs.
  // vue-eslint-parser delegates script parsing to the parser specified here,
  // enabling TS syntax (interface, type annotations, generics, etc.).
  {
    files: ['**/*.vue'],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
      },
    },
  },
  {
    files: ['**/*.{ts,vue}'],
    rules: {
      'vue/multi-word-component-names': 'off',
      'vue/require-v-for-key': 'error',
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
    },
  },
)
