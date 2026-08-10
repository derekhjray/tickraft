// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

<script setup lang="ts">
import { reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { DataTable, SearchForm, StatusTag, useTable, formatDate } from '@tickraft/core'
import PrismPageHeader from '../../components/PrismPageHeader.vue'

const { t } = useI18n()

/** Remediation record type */
interface RemediationRecord {
  id: number
  ruleName: string
  alertName: string
  status: string
  actionType: string
  startedAt: string
  finishedAt: string | null
  error: string
}

/** Search form */
const searchModel = reactive<Record<string, unknown>>({
  status: '',
  actionType: '',
})

const searchFields = computed(() => [
  {
    prop: 'status',
    label: t('prism.remediation.list.status'),
    type: 'select' as const,
    placeholder: t('prism.remediation.list.allStatus'),
    span: 8,
    options: [
      { label: t('prism.remediation.list.statusSuccess'), value: 'success' },
      { label: t('prism.remediation.list.statusFailed'), value: 'failed' },
      { label: t('prism.remediation.list.statusRunning'), value: 'running' },
    ],
  },
  {
    prop: 'actionType',
    label: t('prism.remediation.list.actionType'),
    type: 'input' as const,
    placeholder: t('prism.remediation.list.searchActionType'),
    span: 8,
  },
])

/** Table columns */
const columns = computed(() => [
  { prop: 'ruleName', label: t('prism.remediation.list.ruleName'), minWidth: 160 },
  { prop: 'alertName', label: t('prism.remediation.list.alertName'), minWidth: 160 },
  { prop: 'actionType', label: t('prism.remediation.list.actionType'), width: 120, align: 'center' as const },
  { prop: 'status', label: t('prism.remediation.list.status'), width: 120, slot: 'status' },
  { prop: 'startedAt', label: t('prism.remediation.list.startedAt'), width: 180, slot: 'startedAt' },
])

const {
  data,
  loading,
  total,
  page,
  pageSize,
  immediateSearch,
  resetSearch,
  changePage,
  changePageSize,
} = useTable<RemediationRecord>({
  defaultPageSize: 15,
  // The backend does not yet expose a remediation records list endpoint.
  // Return an empty result set so the page renders gracefully instead of
  // hitting a 404. When the backend adds GET /prism/remediation/records,
  // replace this stub with a real request<T> call.
  fetchFn: async () => {
    return { items: [], total: 0 }
  },
})

/** Click search: trigger query with current search model */
function handleSearch(values: Record<string, unknown>): void {
  immediateSearch({
    status: (values.status as string) || '',
    actionType: (values.actionType as string) || '',
  })
}

/** Reset search conditions */
function handleReset(): void {
  resetSearch()
}

/** Pagination change handler */
function handlePageChange(payload: { current: number; pageSize: number }): void {
  if (payload.pageSize !== pageSize.value) {
    changePageSize(payload.pageSize)
  } else {
    changePage(payload.current)
  }
}

onMounted(() => {
  immediateSearch()
})
</script>

<template>
  <div class="tk-prism-remediation-list tk-page-container">
    <PrismPageHeader
      :title="t('prism.remediation.list.title')"
      :subtitle="t('prism.remediation.list.subtitle')"
      :count="total"
      :count-label="t('prism.remediation.list.countLabel')"
    />

    <SearchForm
      v-model="searchModel"
      :fields="searchFields"
      :loading="loading"
      :show-collapse="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <DataTable
      table-id="remediation-records"
      :data="data"
      :columns="columns"
      :loading="loading"
      :total="total"
      :current="page"
      :page-size="pageSize"
      :page-sizes="[10, 15, 20, 50]"
      row-key="id"
      @page-change="handlePageChange"
    >
      <template #status="{ row }">
        <StatusTag
          category="log"
          :status="(row as RemediationRecord).status"
          show-icon
        >
          {{ t(`prism.remediation.list.status${(row as RemediationRecord).status.charAt(0).toUpperCase() + (row as RemediationRecord).status.slice(1)}`) }}
        </StatusTag>
      </template>
      <template #startedAt="{ row }">
        {{ formatDate((row as RemediationRecord).startedAt) }}
      </template>
    </DataTable>
  </div>
</template>

<style scoped lang="scss">
/* PrismPageHeader provides the page header structure */
</style>
