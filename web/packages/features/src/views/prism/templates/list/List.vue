// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { ConfirmDialog, formatDuration } from '@tickraft/core'
import PrismPageHeader from '../../components/PrismPageHeader.vue'
import { CONDITION_SYMBOL, SEVERITY_OPTIONS, SEVERITY_TAG_TYPE } from '../../constants'
import { MONITOR_TYPE_OPTIONS, TEMPLATES } from '../presets'
import type { AlertTemplate } from '../presets'
import type { AlertSeverity } from '../../../../api/prism'

const router = useRouter()
const { t } = useI18n()

/** Local mutable copy of the preset catalog (supports demo delete) */
const templates = ref<AlertTemplate[]>(TEMPLATES.map((tpl) => ({ ...tpl })))

/** Filter state */
const filter = reactive({
  name: '',
  monitorType: '' as '' | AlertTemplate['monitorType'],
  severity: '' as '' | AlertSeverity,
})

/** Pagination state */
const pagination = reactive({ current: 1, size: 9 })

/** Delete confirmation dialog state */
const deleteVisible = ref(false)
const deleteTarget = ref<AlertTemplate | null>(null)
const deleting = ref(false)

/** Apply (derive rule) confirmation dialog state */
const applyVisible = ref(false)
const applyTarget = ref<AlertTemplate | null>(null)

/** Total count for the header badge (full local catalog) */
const totalCount = computed(() => templates.value.length)

/** Per-severity counts for the summary chips (full local catalog) */
const severityCounts = computed(() => {
  const counts: Record<AlertSeverity, number> = { critical: 0, warning: 0, info: 0 }
  for (const tpl of templates.value) {
    counts[tpl.severity] += 1
  }
  return counts
})

/** Filtered templates (by name/key, monitor type, severity) */
const filtered = computed(() => {
  const q = filter.name.trim().toLowerCase()
  return templates.value.filter((tpl) => {
    if (q && !t(tpl.nameKey).toLowerCase().includes(q) && !tpl.key.toLowerCase().includes(q)) return false
    if (filter.monitorType && tpl.monitorType !== filter.monitorType) return false
    if (filter.severity && tpl.severity !== filter.severity) return false
    return true
  })
})

/** Current page slice */
const pageData = computed(() => {
  const start = (pagination.current - 1) * pagination.size
  return filtered.value.slice(start, start + pagination.size)
})

/** Monitor type label */
function monitorTypeLabel(type: AlertTemplate['monitorType']): string {
  return t(`prism.templates.monitorType.${type}`)
}

/** Toggle a severity chip filter */
function toggleSeverityChip(sev: AlertSeverity): void {
  filter.severity = filter.severity === sev ? '' : sev
  pagination.current = 1
}

/** Run search (collect filter values already bound via v-model) */
function handleSearch(): void {
  pagination.current = 1
}

/** Reset all filters */
function handleReset(): void {
  filter.name = ''
  filter.monitorType = ''
  filter.severity = ''
  pagination.current = 1
  ElMessage.info(t('prism.templates.resetToast'))
}

/** Click "apply template" — open confirmation */
function handleApplyClick(tpl: AlertTemplate): void {
  applyTarget.value = tpl
  applyVisible.value = true
}

/** Confirm applying a template — navigate to rule edit (create mode) with template_id */
function handleApplyConfirm(): void {
  if (!applyTarget.value) return
  const id = applyTarget.value.id
  applyVisible.value = false
  router.push({ path: '/prism/rule/edit', query: { templateId: String(id) } })
}

/** Copy template key to clipboard */
async function handleCopyKey(tpl: AlertTemplate): Promise<void> {
  try {
    await navigator.clipboard.writeText(tpl.key)
    ElMessage.success(t('prism.templates.copyToast', { key: tpl.key }))
  } catch {
    ElMessage.warning(t('prism.templates.copyFailed'))
  }
}

/** Click delete — open confirmation */
function handleDeleteClick(tpl: AlertTemplate): void {
  deleteTarget.value = tpl
  deleteVisible.value = true
}

/** Confirm deleting a template (removes from local catalog) */
function handleDeleteConfirm(): void {
  if (!deleteTarget.value) return
  deleting.value = true
  const target = deleteTarget.value
  const idx = templates.value.findIndex((t) => t.id === target.id)
  if (idx !== -1) templates.value.splice(idx, 1)
  deleteVisible.value = false
  deleting.value = false
  ElMessage.success(t('prism.templates.deletedToast', { name: t(target.nameKey) }))
}

/** "Create template" placeholder (not yet available in open-source edition) */
function handleCreate(): void {
  ElMessage.info(t('prism.templates.createTip'))
}

/** Pagination change */
function handlePageChange(current: number): void {
  pagination.current = current
}

function handleSizeChange(size: number): void {
  pagination.size = size
  pagination.current = 1
}
</script>

<template>
  <div class="tk-prism-templates tk-page-container">
    <PrismPageHeader
      :title="t('prism.templates.title')"
      :subtitle="t('prism.templates.subtitle')"
      :count="totalCount"
      :count-label="t('prism.templates.countLabel')"
    >
      <template #actions>
        <el-button
          type="primary"
          @click="handleCreate"
        >
          + {{ t('prism.templates.create') }}
        </el-button>
      </template>

      <template #chips>
        <button
          v-for="opt in SEVERITY_OPTIONS"
          :key="opt.value"
          type="button"
          class="tk-prism-templates__chip"
          :class="[
            `tk-prism-templates__chip--${opt.value}`,
            { 'is-active': filter.severity === opt.value },
          ]"
          @click="toggleSeverityChip(opt.value)"
        >
          <span class="tk-prism-templates__chip-dot" />
          <span class="tk-prism-templates__chip-label">{{ t(`prism.severity.${opt.value}`) }}</span>
          <span class="tk-prism-templates__chip-count">{{ severityCounts[opt.value] }}</span>
        </button>
      </template>
    </PrismPageHeader>

    <!-- Filter toolbar -->
    <div class="tk-prism-templates__toolbar">
      <div class="tk-prism-templates__filters">
        <div class="tk-prism-templates__field">
          <label class="tk-prism-templates__field-label">{{ t('prism.templates.filterName') }}</label>
          <el-input
            v-model="filter.name"
            :placeholder="t('prism.templates.searchPlaceholder')"
            clearable
            class="tk-prism-templates__input"
            @keyup.enter="handleSearch"
          />
        </div>
        <div class="tk-prism-templates__field">
          <label class="tk-prism-templates__field-label">{{ t('prism.templates.filterType') }}</label>
          <el-select
            v-model="filter.monitorType"
            :placeholder="t('prism.templates.allType')"
            clearable
            class="tk-prism-templates__input"
          >
            <el-option
              v-for="opt in MONITOR_TYPE_OPTIONS"
              :key="opt.value"
              :label="t(opt.labelKey)"
              :value="opt.value"
            />
          </el-select>
        </div>
        <div class="tk-prism-templates__field">
          <label class="tk-prism-templates__field-label">{{ t('prism.templates.filterSeverity') }}</label>
          <el-select
            v-model="filter.severity"
            :placeholder="t('prism.templates.allSeverity')"
            clearable
            class="tk-prism-templates__input"
          >
            <el-option
              v-for="opt in SEVERITY_OPTIONS"
              :key="opt.value"
              :label="t(`prism.severity.${opt.value}`)"
              :value="opt.value"
            />
          </el-select>
        </div>
      </div>
      <div class="tk-prism-templates__actions">
        <el-button @click="handleReset">
          {{ t('prism.templates.reset') }}
        </el-button>
        <el-button
          type="primary"
          @click="handleSearch"
        >
          {{ t('prism.templates.search') }}
        </el-button>
      </div>
    </div>

    <!-- Templates grid -->
    <div
      v-if="pageData.length"
      class="tk-prism-templates__grid"
    >
      <article
        v-for="tpl in pageData"
        :key="tpl.id"
        class="tk-prism-templates__card"
        :class="`tk-prism-templates__card--${tpl.severity}`"
      >
        <div class="tk-prism-templates__card-top">
          <span class="tk-prism-templates__card-id">
            <span class="tk-prism-templates__card-hash">#</span>{{ tpl.id }}
          </span>
          <span
            class="tk-prism-templates__card-key"
            :title="tpl.key"
          >{{ tpl.key }}</span>
        </div>

        <h3 class="tk-prism-templates__card-name">
          {{ t(tpl.nameKey) }}
        </h3>
        <p class="tk-prism-templates__card-desc">
          {{ t(tpl.descriptionKey) }}
        </p>

        <!-- condition expression -->
        <div class="tk-prism-templates__expr">
          <span class="tk-prism-templates__expr-metric">{{ tpl.metric }}</span>
          <span class="tk-prism-templates__expr-op">{{ CONDITION_SYMBOL[tpl.condition] }}</span>
          <span class="tk-prism-templates__expr-thresh">{{ tpl.threshold }}</span>
          <span class="tk-prism-templates__expr-for">{{ t('prism.templates.forKeyword') }}</span>
          <span class="tk-prism-templates__expr-dur">{{ formatDuration(tpl.duration * 1000) }}</span>
        </div>

        <!-- meta row -->
        <div class="tk-prism-templates__card-meta">
          <el-tag
            type="info"
            effect="plain"
            size="small"
          >
            {{ monitorTypeLabel(tpl.monitorType) }}
          </el-tag>
          <el-tag
            :type="SEVERITY_TAG_TYPE[tpl.severity]"
            effect="light"
            size="small"
          >
            {{ t(`prism.severity.${tpl.severity}`) }}
          </el-tag>
          <span
            class="tk-prism-templates__usage"
            :class="{ 'is-active': tpl.usage > 0 }"
          >
            <span class="tk-prism-templates__usage-num">{{ tpl.usage }}</span>
            {{ t('prism.templates.usage') }}
          </span>
        </div>

        <!-- footer actions -->
        <div class="tk-prism-templates__card-footer">
          <el-button
            type="primary"
            size="small"
            @click="handleApplyClick(tpl)"
          >
            {{ t('prism.templates.apply') }}
          </el-button>
          <div class="tk-prism-templates__card-actions">
            <el-button
              link
              @click="handleCopyKey(tpl)"
            >
              {{ t('prism.templates.copy') }}
            </el-button>
            <el-button
              link
              type="danger"
              @click="handleDeleteClick(tpl)"
            >
              {{ t('prism.templates.delete') }}
            </el-button>
          </div>
        </div>
      </article>
    </div>

    <!-- empty state -->
    <div
      v-else
      class="tk-prism-templates__empty"
    >
      <el-icon class="tk-prism-templates__empty-icon">
        <Document />
      </el-icon>
      <p class="tk-prism-templates__empty-text">
        {{ t('prism.templates.empty') }}
      </p>
      <p class="tk-prism-templates__empty-hint">
        {{ t('prism.templates.emptyHint') }}
      </p>
    </div>

    <!-- pagination -->
    <div
      v-if="filtered.length"
      class="tk-prism-templates__pagination"
    >
      <el-pagination
        :current-page="pagination.current"
        :page-size="pagination.size"
        :page-sizes="[9, 12, 24]"
        :total="filtered.length"
        layout="total, sizes, prev, pager, next"
        background
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
      />
    </div>

    <!-- Apply (derive rule) confirmation -->
    <ConfirmDialog
      v-model="applyVisible"
      :title="t('prism.templates.applyTitle')"
      :content="applyTarget ? t('prism.templates.applyContent', { name: t(applyTarget.nameKey) }) : ''"
      :confirm-text="t('prism.templates.applyConfirm')"
      @confirm="handleApplyConfirm"
    />

    <!-- Delete confirmation -->
    <ConfirmDialog
      v-model="deleteVisible"
      :title="t('prism.templates.deleteTitle')"
      :content="deleteTarget ? t('prism.templates.deleteContent', { name: t(deleteTarget.nameKey) }) : ''"
      type="danger"
      :loading="deleting"
      :confirm-text="t('prism.templates.delete')"
      @confirm="handleDeleteConfirm"
    />
  </div>
</template>

<style scoped lang="scss">
.tk-prism-templates {
  // ===== severity chips =====
  &__chip {
    display: inline-flex;
    align-items: center;
    gap: var(--tk-spacing-sm);
    height: 36px;
    padding: 0 var(--tk-spacing-lg);
    background-color: var(--tk-bg-surface);
    border: 1px solid var(--tk-border-color);
    border-radius: var(--tk-radius-md);
    cursor: pointer;
    transition: border-color var(--tk-animation-duration-fast) var(--tk-ease-out),
      background-color var(--tk-animation-duration-fast) var(--tk-ease-out);

    &:hover {
      border-color: var(--tk-border-strong);
    }

    &.is-active {
      border-color: var(--tk-primary-color);
      background-color: var(--tk-primary-color-bg);

      .tk-prism-templates__chip-label,
      .tk-prism-templates__chip-count {
        color: var(--tk-primary-color);
      }
    }
  }

  &__chip-dot {
    width: 8px;
    height: 8px;
    border-radius: var(--tk-border-radius-circle);
    flex-shrink: 0;
  }

  &__chip--critical &__chip-dot {
    background-color: var(--tk-danger-color);
  }

  &__chip--warning &__chip-dot {
    background-color: var(--tk-warning-color);
  }

  &__chip--info &__chip-dot {
    background-color: var(--tk-info-color);
  }

  &__chip-label {
    font-family: var(--tk-font-family-mono);
    font-size: var(--tk-font-size-xs);
    font-weight: var(--tk-font-weight-semibold);
    color: var(--tk-text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  &__chip-count {
    font-family: var(--tk-font-family-mono);
    font-size: var(--tk-font-size-base);
    font-weight: var(--tk-font-weight-bold);
    color: var(--tk-text-primary);
    font-variant-numeric: tabular-nums;
    min-width: 18px;
    text-align: right;
  }

  // ===== filter toolbar =====
  &__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--tk-spacing-md);
    padding: var(--tk-spacing-lg);
    background-color: var(--tk-bg-surface);
    border: 1px solid var(--tk-border-color);
    border-radius: var(--tk-radius-lg);
    flex-wrap: wrap;
    margin-bottom: var(--tk-spacing-lg);
  }

  &__filters {
    display: flex;
    align-items: center;
    gap: var(--tk-spacing-md);
    flex-wrap: wrap;
  }

  &__field {
    display: flex;
    flex-direction: column;
    gap: var(--tk-spacing-xs);
  }

  &__field-label {
    font-family: var(--tk-font-family-mono);
    font-size: var(--tk-font-size-xs);
    font-weight: var(--tk-font-weight-semibold);
    color: var(--tk-text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  &__input {
    min-width: 180px;
  }

  &__actions {
    display: flex;
    gap: var(--tk-spacing-sm);
  }

  // ===== templates grid =====
  &__grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
    gap: var(--tk-spacing-lg);
  }

  &__card {
    position: relative;
    display: flex;
    flex-direction: column;
    padding: var(--tk-spacing-lg);
    background-color: var(--tk-bg-surface);
    border: 1px solid var(--tk-border-color);
    border-radius: var(--tk-radius-lg);
    transition: border-color var(--tk-animation-duration-fast) var(--tk-ease-out),
      transform var(--tk-animation-duration-fast) var(--tk-ease-out);
    overflow: hidden;

    &:hover {
      border-color: var(--tk-border-strong);
      transform: translateY(-2px);
    }

    &::before {
      content: "";
      position: absolute;
      left: 0;
      top: 0;
      bottom: 0;
      width: 3px;
    }
  }

  &__card--critical::before {
    background-color: var(--tk-danger-color);
  }

  &__card--warning::before {
    background-color: var(--tk-warning-color);
  }

  &__card--info::before {
    background-color: var(--tk-info-color);
  }

  &__card-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--tk-spacing-sm);
    margin-bottom: var(--tk-spacing-md);
  }

  &__card-id {
    font-family: var(--tk-font-family-mono);
    font-size: var(--tk-font-size-xs);
    color: var(--tk-text-secondary);
    font-variant-numeric: tabular-nums;
  }

  &__card-hash {
    color: var(--tk-text-placeholder);
    margin-right: 1px;
  }

  &__card-key {
    font-family: var(--tk-font-family-mono);
    font-size: var(--tk-font-size-xs);
    font-weight: var(--tk-font-weight-semibold);
    color: var(--tk-text-secondary);
    padding: 2px var(--tk-spacing-sm);
    background-color: var(--tk-neutral-100);
    border: 1px solid var(--tk-border-color-light);
    border-radius: var(--tk-radius-sm);
    max-width: 180px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__card-name {
    font-size: var(--tk-font-size-lg);
    font-weight: var(--tk-font-weight-bold);
    color: var(--tk-text-primary);
    line-height: 1.25;
    margin: 0 0 var(--tk-spacing-sm);
  }

  &__card-desc {
    font-size: var(--tk-font-size-xs);
    color: var(--tk-text-secondary);
    line-height: var(--tk-line-height-normal);
    margin: 0 0 var(--tk-spacing-lg);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  // ===== condition expression =====
  &__expr {
    display: flex;
    align-items: center;
    gap: var(--tk-spacing-sm);
    flex-wrap: wrap;
    padding: var(--tk-spacing-md) var(--tk-spacing-lg);
    background-color: var(--tk-neutral-50);
    border: 1px solid var(--tk-border-color-light);
    border-radius: var(--tk-radius-md);
    font-family: var(--tk-font-family-mono);
    font-variant-numeric: tabular-nums;
    margin-bottom: var(--tk-spacing-lg);
  }

  &__expr-metric {
    font-size: var(--tk-font-size-sm);
    font-weight: var(--tk-font-weight-semibold);
    color: var(--tk-text-primary);
  }

  &__expr-op {
    font-size: var(--tk-font-size-sm);
    color: var(--tk-text-secondary);
    font-weight: var(--tk-font-weight-medium);
  }

  &__expr-thresh {
    font-size: var(--tk-font-size-sm);
    color: var(--tk-text-primary);
    font-weight: var(--tk-font-weight-bold);
  }

  &__expr-for {
    font-family: var(--tk-font-family);
    font-size: var(--tk-font-size-xs);
    color: var(--tk-text-placeholder);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-left: var(--tk-spacing-xs);
  }

  &__expr-dur {
    font-size: var(--tk-font-size-xs);
    color: var(--tk-text-regular);
    font-weight: var(--tk-font-weight-medium);
  }

  // ===== meta row =====
  &__card-meta {
    display: flex;
    align-items: center;
    gap: var(--tk-spacing-sm);
    margin-bottom: var(--tk-spacing-md);
    flex-wrap: wrap;
  }

  &__usage {
    display: inline-flex;
    align-items: center;
    gap: var(--tk-spacing-xs);
    font-family: var(--tk-font-family-mono);
    font-size: var(--tk-font-size-xs);
    font-weight: var(--tk-font-weight-semibold);
    color: var(--tk-text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  &__usage-num {
    color: var(--tk-text-primary);
    font-variant-numeric: tabular-nums;
  }

  &__usage.is-active &__usage-num {
    color: var(--tk-primary-color);
  }

  // ===== footer =====
  &__card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--tk-spacing-sm);
    padding-top: var(--tk-spacing-md);
    border-top: 1px solid var(--tk-border-color-light);
    margin-top: auto;
  }

  &__card-actions {
    display: flex;
    align-items: center;
    gap: var(--tk-spacing-sm);
  }

  // ===== empty state =====
  &__empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--tk-spacing-sm);
    padding: var(--tk-spacing-5xl) var(--tk-spacing-lg);
    background-color: var(--tk-bg-surface);
    border: 1px dashed var(--tk-border-color);
    border-radius: var(--tk-radius-lg);
    color: var(--tk-text-secondary);
  }

  &__empty-icon {
    font-size: 32px;
    color: var(--tk-text-placeholder);
  }

  &__empty-text {
    font-size: var(--tk-font-size-sm);
    color: var(--tk-text-secondary);
    margin: 0;
  }

  &__empty-hint {
    font-size: var(--tk-font-size-xs);
    color: var(--tk-text-placeholder);
    margin: 0;
  }

  // ===== pagination =====
  &__pagination {
    display: flex;
    justify-content: flex-end;
    padding: var(--tk-spacing-lg) 0 0;
  }
}
</style>
