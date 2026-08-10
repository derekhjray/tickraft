// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Edit, Delete, Lightning } from '@element-plus/icons-vue'
import { StatusTag, PageEmpty, ConfirmDialog, DataTable } from '@tickraft/core'
import { formatDate, formatDuration } from '@tickraft/core'
import TrendTab from './TrendTab.vue'
import type { TaskModel, LogModel } from '../../../../types/task'
import { getTask, getLogs, getExecutionStats, triggerTask, deleteTask } from '../../../../api/task'
import type { ExecutionStats } from '../../../../types/task'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const taskId = Number(route.params.id)
const loading = ref(false)
const notFound = ref(false)
const task = ref<TaskModel | null>(null)
const logs = ref<LogModel[]>([])
const execStats = ref<ExecutionStats | null>(null)

const deleteVisible = ref(false)
const deleteLoading = ref(false)
const triggerLoading = ref(false)
const activeTab = ref('basic')

const EXECUTOR_LABELS = computed<Record<string, string>>(() => ({
  http: 'HTTP', tcp: 'TCP', icmp: 'ICMP',
  local: t('task.task.create.executorLocal'), webhook: 'Webhook',
  ssh: 'SSH', mysql: 'MYSQL', redis: 'REDIS',
}))
const PRO_EXECUTORS: Record<string, boolean> = { ssh: true, mysql: true, redis: true }

const totalExec = computed(() => execStats.value?.totalExecutions ?? 0)
const successCount = computed(() => execStats.value?.successCount ?? 0)
const failedCount = computed(() => execStats.value?.failureCount ?? 0)
const successRate = computed(() => {
  const rate = execStats.value?.successRate
  if (rate == null) return '0.0'
  return rate.toFixed(1)
})
const avgDuration = computed(() => execStats.value?.averageDurationMs ?? 0)

/** Execution logs table columns */
const logColumns = computed(() => [
  { prop: 'id', label: t('task.log.list.logId'), width: 80, slot: 'id' },
  { prop: 'status', label: t('task.log.list.status'), width: 100, slot: 'status' },
  { prop: 'startedAt', label: t('task.log.list.startedAt'), minWidth: 160, slot: 'startedAt' },
])

const configJson = computed(() => {
  if (!task.value?.config) return '{}'
  return JSON.stringify(task.value.config, null, 2)
})

const tagsText = computed(() => {
  if (!task.value?.tags || task.value.tags.length === 0) return '—'
  return task.value.tags.join(', ')
})

async function fetchData(): Promise<void> {
  loading.value = true
  try {
    const [taskData, logsData, stats] = await Promise.all([
      getTask(taskId),
      getLogs(taskId, { page: 1, pageSize: 10 }),
      getExecutionStats().catch(() => null),
    ])
    task.value = taskData
    logs.value = logsData.items
    execStats.value = stats
  } catch {
    notFound.value = true
  } finally {
    loading.value = false
  }
}

function handleEdit(): void { router.push(`/task/edit/${taskId}`) }

async function handleTrigger(): Promise<void> {
  try {
    await ElMessageBox.confirm(
      t('task.task.list.triggerConfirm', { name: task.value?.name ?? '' }),
      t('task.task.list.trigger'),
      { type: 'warning' },
    )
    triggerLoading.value = true
    await triggerTask(taskId)
    ElMessage.success(t('task.task.list.triggerSuccess', { name: task.value?.name ?? '' }))
    await fetchData()
  } catch {
    // User cancellation or errors are handled by the interceptor
  } finally {
    triggerLoading.value = false
  }
}

function handleDelete(): void { deleteVisible.value = true }

async function confirmDelete(): Promise<void> {
  deleteLoading.value = true
  try {
    await deleteTask(taskId)
    ElMessage.success(t('task.task.list.deleteSuccess'))
    deleteVisible.value = false
    router.push('/task/list')
  } catch {
    // Errors are handled centrally by the interceptor
  } finally {
    deleteLoading.value = false
  }
}

function handleLogDetail(row: LogModel): void { router.push(`/task/log/detail/${taskId}/${row.id}`) }
function handleBack(): void { router.push('/task/list') }
function handleViewAllLogs(): void { router.push('/task/log/list') }

onMounted(() => { void fetchData() })
</script>

<template>
  <div v-loading="loading" class="tk-task-detail tk-page-container">
    <!-- Not found -->
    <template v-if="notFound">
      <PageEmpty :description="t('task.task.detail.notFound', { id: taskId })">
        <el-button type="primary" @click="handleBack">{{ t('task.task.detail.back') }}</el-button>
      </PageEmpty>
    </template>

    <template v-else-if="task">
      <!-- Header: back + eyebrow + title + badges + actions -->
      <div class="tk-detail-header">
        <div class="tk-detail-header__left">
          <button class="tk-detail-header__back" :title="t('task.task.detail.back')" @click="handleBack">
            <el-icon :size="16"><ArrowLeft /></el-icon>
          </button>
          <div class="tk-detail-header__title-block">
            <div class="tk-detail-header__eyebrow">{{ t('task.task.detail.eyebrow', { id: taskId }) }}</div>
            <div class="tk-detail-header__title-row">
              <h1 class="tk-detail-header__title">{{ task.name }}</h1>
              <span class="tk-executor-badge" :class="`tk-executor-badge--${task.executor}`">
                <span class="tk-executor-badge__dot" />
                {{ EXECUTOR_LABELS[task.executor] || task.executor }}
                <span v-if="PRO_EXECUTORS[task.executor]" class="tk-executor-badge__edition">PRO</span>
              </span>
              <span class="tk-status-tag" :class="task.enabled ? 'tk-status-tag--success' : 'tk-status-tag--unknown'">
                <span class="tk-status-tag__dot" />
                {{ task.enabled ? t('common.app.enabled') : t('common.app.disabled') }}
              </span>
              <span class="tk-detail-header__key">TASK-{{ task.id }}</span>
            </div>
          </div>
        </div>
        <div class="tk-detail-header__actions">
          <el-button @click="handleEdit"><el-icon><Edit /></el-icon>{{ t('task.task.detail.edit') }}</el-button>
          <el-button type="primary" :loading="triggerLoading" @click="handleTrigger">
            <el-icon><Lightning /></el-icon>{{ t('task.task.detail.trigger') }}
          </el-button>
          <el-button type="danger" @click="handleDelete"><el-icon><Delete /></el-icon>{{ t('task.task.detail.delete') }}</el-button>
        </div>
      </div>

      <!-- Stat strip: 4 tiles with accent borders -->
      <div class="tk-stat-strip">
        <div class="tk-stat-tile">
          <div class="tk-stat-tile__label">{{ t('task.task.detail.totalExec') }}</div>
          <div class="tk-stat-tile__value">{{ totalExec }}</div>
          <div class="tk-stat-tile__sub">{{ t('task.task.detail.totalExecSub') }}</div>
        </div>
        <div class="tk-stat-tile tk-stat-tile--ok">
          <div class="tk-stat-tile__label">{{ t('task.task.detail.successRate') }}</div>
          <div class="tk-stat-tile__value">{{ successRate }}%</div>
          <div class="tk-stat-tile__sub tk-stat-tile__sub--ok">{{ t('task.task.detail.successRateSub', { count: successCount }) }}</div>
        </div>
        <div class="tk-stat-tile tk-stat-tile--warn">
          <div class="tk-stat-tile__label">{{ t('task.task.detail.failed') }}</div>
          <div class="tk-stat-tile__value">{{ failedCount }}</div>
          <div class="tk-stat-tile__sub tk-stat-tile__sub--fail">{{ t('task.task.detail.failedSub') }}</div>
        </div>
        <div class="tk-stat-tile tk-stat-tile--info">
          <div class="tk-stat-tile__label">{{ t('task.task.detail.avgDuration') }}</div>
          <div class="tk-stat-tile__value">{{ formatDuration(avgDuration) }}</div>
          <div class="tk-stat-tile__sub">{{ t('task.task.detail.avgDurationSub', { count: Math.min(totalExec, 20) }) }}</div>
        </div>
      </div>

      <!-- Tabs card -->
      <div class="tk-detail-card">
        <el-tabs v-model="activeTab" class="tk-detail-tabs">
          <!-- Basic Info -->
          <el-tab-pane :label="t('task.task.detail.basicInfo')" name="basic">
            <div class="tk-descriptions">
              <div class="tk-desc-item"><span class="tk-desc-item__label">{{ t('task.task.list.name') }}</span><span class="tk-desc-item__value">{{ task.name }}</span></div>
              <div class="tk-desc-item"><span class="tk-desc-item__label">ID</span><span class="tk-desc-item__value tk-desc-item__value--mono">{{ task.id }}</span></div>
              <div class="tk-desc-item"><span class="tk-desc-item__label">{{ t('task.task.detail.executorType') }}</span><span class="tk-desc-item__value">{{ EXECUTOR_LABELS[task.executor] || task.executor }}</span></div>
              <div class="tk-desc-item"><span class="tk-desc-item__label">{{ t('task.task.detail.scheduleExpr') }}</span><span class="tk-desc-item__value tk-desc-item__value--mono">{{ task.schedule || '—' }}</span></div>
              <div class="tk-desc-item"><span class="tk-desc-item__label">{{ t('task.task.create.group') }}</span><span class="tk-desc-item__value">{{ task.group || '—' }}</span></div>
              <div class="tk-desc-item"><span class="tk-desc-item__label">{{ t('task.task.list.tags') }}</span><span class="tk-desc-item__value">{{ tagsText }}</span></div>
              <div class="tk-desc-item"><span class="tk-desc-item__label">{{ t('task.task.create.retryPolicy') }}</span><span class="tk-desc-item__value tk-desc-item__value--mono">{{ task.retryPolicy || '—' }}</span></div>
              <div class="tk-desc-item"><span class="tk-desc-item__label">{{ t('task.task.create.concurrency') }}</span><span class="tk-desc-item__value tk-desc-item__value--mono">{{ task.concurrency ?? 0 }}</span></div>
              <div class="tk-desc-item"><span class="tk-desc-item__label">{{ t('task.task.detail.enabled') }}</span><span class="tk-desc-item__value">{{ task.enabled ? t('common.app.enabled') : t('common.app.disabled') }}</span></div>
              <div class="tk-desc-item"><span class="tk-desc-item__label">{{ t('task.task.detail.description') }}</span><span class="tk-desc-item__value">{{ task.description || '—' }}</span></div>
              <div class="tk-desc-item"><span class="tk-desc-item__label">{{ t('task.task.detail.createdAt') }}</span><span class="tk-desc-item__value tk-desc-item__value--mono">{{ formatDate(task.createdAt) }}</span></div>
              <div class="tk-desc-item"><span class="tk-desc-item__label">{{ t('task.task.detail.updatedAt') }}</span><span class="tk-desc-item__value tk-desc-item__value--mono">{{ formatDate(task.updatedAt) }}</span></div>
            </div>
          </el-tab-pane>

          <!-- Execution Logs -->
          <el-tab-pane name="logs">
            <template #label>
              {{ t('task.task.detail.execLogs') }}
              <span class="tk-tab-count">{{ totalExec }}</span>
            </template>
            <DataTable
              v-if="logs.length > 0"
              :data="logs"
              :columns="logColumns"
              :pagination="false"
              density="compact"
              @row-click="handleLogDetail"
            >
              <template #id="{ row }"><span class="tk-mono-id">#{{ (row as LogModel).id }}</span></template>
              <template #status="{ row }"><StatusTag category="log" :status="(row as LogModel).status" size="sm" /></template>
              <template #startedAt="{ row }"><span class="tk-mono-text">{{ formatDate((row as LogModel).startedAt) }}</span></template>
            </DataTable>
            <div v-else class="tk-empty-inline">{{ t('task.task.detail.noLogs') }}</div>
            <div v-if="logs.length > 0" class="tk-logs-footer">
              <el-button link type="primary" @click="handleViewAllLogs">{{ t('task.task.detail.viewAllLogs') }} →</el-button>
            </div>
          </el-tab-pane>

          <!-- Trend -->
          <el-tab-pane :label="t('task.task.detail.trend')" name="trend">
            <TrendTab :logs="logs" />
          </el-tab-pane>

          <!-- Config -->
          <el-tab-pane :label="t('task.task.detail.configJson')" name="config">
            <div class="tk-code-section">
              <div class="tk-code-section__label">{{ t('task.task.detail.configArgs') }}</div>
              <pre class="tk-json-view">{{ configJson }}</pre>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </template>
  </div>

  <ConfirmDialog v-model="deleteVisible" :title="t('task.task.detail.delete')"
    :content="t('task.task.list.deleteConfirm', { name: task?.name ?? '' })"
    :loading="deleteLoading" type="danger" @confirm="confirmDelete" />
</template>

<style scoped lang="scss">
.tk-detail-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: var(--tk-spacing-8); margin-bottom: var(--tk-spacing-8); flex-wrap: wrap;
  &__left { display: flex; align-items: center; gap: var(--tk-spacing-5); min-width: 0; }
  &__back {
    display: inline-flex; align-items: center; justify-content: center;
    width: 36px; height: 36px; border-radius: var(--tk-radius-md);
    color: var(--tk-text-secondary); background: var(--tk-bg-surface);
    border: 1px solid var(--tk-border-color-base); cursor: pointer;
    transition: all var(--tk-duration-fast) var(--tk-ease-out); flex-shrink: 0;
    &:hover { color: var(--tk-primary-color); border-color: var(--tk-primary-color-border); background: var(--tk-primary-color-bg); }
  }
  &__title-block { min-width: 0; }
  &__eyebrow {
    font-family: var(--tk-font-mono); font-size: var(--tk-font-size-xs);
    color: var(--tk-text-secondary); letter-spacing: var(--tk-letter-widest);
    text-transform: uppercase; margin-bottom: 4px;
  }
  &__title-row { display: flex; align-items: center; gap: var(--tk-spacing-4); flex-wrap: wrap; }
  &__title {
    font-family: var(--tk-font-display); font-size: var(--tk-font-size-2xl);
    font-weight: var(--tk-font-weight-bold); color: var(--tk-text-primary);
    letter-spacing: var(--tk-letter-tight); line-height: 1.1; margin: 0;
  }
  &__key {
    font-family: var(--tk-font-mono); font-size: var(--tk-font-size-xs);
    color: var(--tk-text-secondary); background: var(--tk-bg-fill);
    padding: 2px var(--tk-spacing-4); border-radius: var(--tk-radius-sm);
    border: 1px solid var(--tk-border-color-light);
  }
  &__actions { display: flex; align-items: center; gap: var(--tk-spacing-3); }
}

.tk-stat-strip { display: grid; grid-template-columns: repeat(4, 1fr); gap: var(--tk-spacing-4); margin-bottom: var(--tk-spacing-8); }
.tk-stat-tile {
  position: relative; padding: var(--tk-spacing-6) var(--tk-spacing-8);
  background: var(--tk-bg-surface); border: 1px solid var(--tk-border-color-base);
  border-radius: var(--tk-radius-lg); display: flex; flex-direction: column; gap: 4px; overflow: hidden;
  &::before { content: ""; position: absolute; top: 0; left: 0; width: 3px; height: 100%; background-color: var(--tk-primary-color); }
  &--ok::before { background-color: var(--tk-success-color); }
  &--warn::before { background-color: var(--tk-warning-color); }
  &--info::before { background-color: var(--tk-info-color); }
  &__label { font-family: var(--tk-font-mono); font-size: var(--tk-font-size-xs); color: var(--tk-text-secondary); letter-spacing: var(--tk-letter-wider); text-transform: uppercase; }
  &__value { font-family: var(--tk-font-display); font-size: var(--tk-font-size-2xl); font-weight: var(--tk-font-weight-extrabold); color: var(--tk-text-primary); line-height: 1; font-variant-numeric: tabular-nums; letter-spacing: var(--tk-letter-tight); }
  &__sub { font-family: var(--tk-font-mono); font-size: var(--tk-font-size-xs); color: var(--tk-text-secondary); }
  &__sub--ok { color: var(--tk-success-color); }
  &__sub--fail { color: var(--tk-danger-color); }
}

.tk-executor-badge {
  display: inline-flex; align-items: center; gap: var(--tk-spacing-2);
  padding: 2px var(--tk-spacing-6); font-family: var(--tk-font-mono);
  font-size: var(--tk-font-size-xs); font-weight: var(--tk-font-weight-semibold);
  letter-spacing: var(--tk-letter-wide); text-transform: uppercase;
  border-radius: var(--tk-radius-sm); border: 1px solid transparent; white-space: nowrap;
  &__dot { width: 5px; height: 5px; border-radius: var(--tk-radius-circle); background-color: currentColor; flex-shrink: 0; }
  &__edition { margin-left: var(--tk-spacing-2); padding: 0 4px; font-size: 9px; font-weight: var(--tk-font-weight-bold); color: #fff; background: linear-gradient(135deg, #0ea5e9, #0284c7); border-radius: var(--tk-radius-xs); letter-spacing: var(--tk-letter-wider); }
  &--http { color: #2563eb; background-color: rgba(37, 99, 235, 0.10); border-color: rgba(37, 99, 235, 0.25); }
  &--tcp { color: #0891b2; background-color: rgba(8, 145, 178, 0.10); border-color: rgba(8, 145, 178, 0.25); }
  &--icmp { color: #7c3aed; background-color: rgba(124, 58, 237, 0.10); border-color: rgba(124, 58, 237, 0.25); }
  &--local { color: #475569; background-color: rgba(71, 85, 105, 0.10); border-color: rgba(71, 85, 105, 0.25); }
  &--ssh { color: #b45309; background-color: rgba(180, 83, 9, 0.10); border-color: rgba(180, 83, 9, 0.25); }
  &--mysql { color: #15803d; background-color: rgba(21, 128, 61, 0.10); border-color: rgba(21, 128, 61, 0.25); }
  &--redis { color: #dc2626; background-color: rgba(220, 38, 38, 0.10); border-color: rgba(220, 38, 38, 0.25); }
  &--webhook { color: #be185d; background-color: rgba(190, 24, 93, 0.10); border-color: rgba(190, 24, 93, 0.25); }
}

.tk-status-tag {
  display: inline-flex; align-items: center; gap: 6px; padding: 2px var(--tk-spacing-4);
  border-radius: var(--tk-radius-sm); font-size: var(--tk-font-size-xs); font-weight: var(--tk-font-weight-medium); border: 1px solid transparent;
  &__dot { width: 7px; height: 7px; border-radius: var(--tk-radius-circle); flex-shrink: 0; }
  &--success { background: var(--tk-success-color-bg); color: var(--tk-success-color-text); border-color: var(--tk-success-color-border); .tk-status-tag__dot { background-color: var(--tk-success-color); } }
  &--unknown { background: var(--tk-bg-fill); color: var(--tk-text-secondary); border-color: var(--tk-border-color-base); .tk-status-tag__dot { background-color: var(--tk-text-placeholder); } }
}

.tk-detail-card { background: var(--tk-bg-surface); border: 1px solid var(--tk-border-color-base); border-radius: var(--tk-radius-lg); overflow: hidden; }
.tk-detail-tabs { :deep(.el-tabs__header) { margin: 0; padding: 0 var(--tk-spacing-10); } :deep(.el-tabs__content) { padding: var(--tk-spacing-10); } }
.tk-tab-count { font-family: var(--tk-font-mono); font-size: 10px; padding: 0 6px; color: var(--tk-text-secondary); background: var(--tk-bg-fill); border-radius: var(--tk-radius-round); margin-left: 4px; }

.tk-descriptions { display: grid; grid-template-columns: repeat(2, 1fr); gap: var(--tk-spacing-6) var(--tk-spacing-12); }
.tk-desc-item { display: flex; flex-direction: column; gap: 4px; padding-bottom: var(--tk-spacing-4); border-bottom: 1px dashed var(--tk-border-color-lighter);
  &__label { font-family: var(--tk-font-mono); font-size: var(--tk-font-size-xs); color: var(--tk-text-secondary); letter-spacing: var(--tk-letter-wide); text-transform: uppercase; }
  &__value { font-size: var(--tk-font-size-sm); color: var(--tk-text-primary); word-break: break-all;
    &--mono { font-family: var(--tk-font-mono); font-variant-numeric: tabular-nums; }
  }
}

.tk-mono-id { font-family: var(--tk-font-mono); font-size: var(--tk-font-size-xs); color: var(--tk-text-secondary); font-variant-numeric: tabular-nums; }
.tk-mono-text { font-family: var(--tk-font-mono); font-size: var(--tk-font-size-xs); color: var(--tk-text-secondary); }
.tk-empty-inline { text-align: center; color: var(--tk-text-placeholder); padding: var(--tk-spacing-12); font-size: var(--tk-font-size-sm); }
.tk-logs-footer { margin-top: var(--tk-spacing-5); }

.tk-code-section { display: flex; flex-direction: column; gap: var(--tk-spacing-3);
  &__label { font-family: var(--tk-font-mono); font-size: var(--tk-font-size-xs); color: var(--tk-text-secondary); letter-spacing: var(--tk-letter-wide); text-transform: uppercase; }
}
.tk-json-view { background: var(--tk-bg-fill-blank); border: 1px solid var(--tk-border-color-base); border-radius: var(--tk-radius-md); padding: var(--tk-spacing-6); font-family: var(--tk-font-mono); font-size: var(--tk-font-size-sm); color: var(--tk-text-primary); white-space: pre-wrap; word-break: break-all; line-height: var(--tk-line-height-snug); margin: 0; max-height: 400px; overflow: auto; }

@media (max-width: 960px) {
  .tk-stat-strip { grid-template-columns: repeat(2, 1fr); }
  .tk-descriptions { grid-template-columns: 1fr; }
}
</style>
