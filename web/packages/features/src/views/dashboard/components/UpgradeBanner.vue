// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

<script setup lang="ts">
/**
 * UpgradeBanner - unified dashboard upgrade banner.
 *
 * Replaces the scattered `<FeatureGuard locked>` placeholders across CE
 * business pages with a single, visually appealing card on the dashboard
 * that promotes upgrading to the professional edition. Pro-only features
 * (SSH/MySQL/Redis executors, DNS/SSL probes, multi-user, audit logs, etc.)
 * are highlighted here rather than shown as locked placeholders inline.
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

/**
 * Upgrade banner visibility toggle.
 *
 * The banner is hidden while the professional edition is not yet live.
 * Flip this flag to `true` once the paid edition ships and the pricing page
 * is reachable. Keeping it as a single module constant avoids scattering
 * feature-gating logic across the codebase.
 */
const UPGRADE_BANNER_ENABLED = false

/** External link to the subscription / pricing page */
const PRICING_URL = 'https://tickraft.com/pricing'

const { t } = useI18n()

/** Feature highlight shown in the banner; i18n keys are resolved via computed to stay reactive on language switch */
interface UpgradeFeature {
  icon: string
  labelKey: string
}

const features = computed<UpgradeFeature[]>(() => [
  { icon: 'i-ep-connection', labelKey: 'dashboard.upgrade.featureSshExecutor' },
  { icon: 'i-ep-data-base', labelKey: 'dashboard.upgrade.featureDbExecutor' },
  { icon: 'i-ep-position', labelKey: 'dashboard.upgrade.featureAdvancedProber' },
  { icon: 'i-user', labelKey: 'dashboard.upgrade.featureMultiUser' },
  { icon: 'i-ep-document', labelKey: 'dashboard.upgrade.featureAuditLog' },
  { icon: 'i-ep-bell', labelKey: 'dashboard.upgrade.featureMultiChannel' },
])
</script>

<template>
  <section
    v-if="UPGRADE_BANNER_ENABLED"
    class="tk-upgrade-banner"
    role="region"
    :aria-label="t('dashboard.upgrade.title')"
  >
    <div class="tk-upgrade-banner__glow" />
    <div class="tk-upgrade-banner__content">
      <div class="tk-upgrade-banner__head">
        <div class="tk-upgrade-banner__heading">
          <span class="tk-upgrade-banner__badge">PRO</span>
          <h2 class="tk-upgrade-banner__title">
            {{ t('dashboard.upgrade.title') }}
          </h2>
        </div>
        <p class="tk-upgrade-banner__desc">
          {{ t('dashboard.upgrade.description') }}
        </p>
      </div>

      <ul class="tk-upgrade-banner__features">
        <li
          v-for="item in features"
          :key="item.labelKey"
          class="tk-upgrade-banner__feature"
        >
          <span class="tk-upgrade-banner__feature-icon">
            <i :class="item.icon" />
          </span>
          <span class="tk-upgrade-banner__feature-label">{{ t(item.labelKey) }}</span>
        </li>
      </ul>

      <div class="tk-upgrade-banner__actions">
        <a
          class="tk-upgrade-banner__cta"
          :href="PRICING_URL"
          target="_blank"
          rel="noopener"
        >
          {{ t('dashboard.upgrade.cta') }}
          <i class="i-ep-arrow-right" />
        </a>
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.tk-upgrade-banner {
  position: relative;
  display: block;
  border-radius: var(--tk-border-radius-lg);
  overflow: hidden;
  background:
    linear-gradient(135deg, color-mix(in srgb, var(--tk-primary-color) 12%, transparent) 0%, transparent 60%),
    var(--tk-bg-surface);
  border: 1px solid color-mix(in srgb, var(--tk-primary-color) 30%, var(--tk-border-color));
  padding: var(--tk-spacing-lg);
  isolation: isolate;
}

.tk-upgrade-banner__glow {
  position: absolute;
  top: -60px;
  right: -60px;
  width: 220px;
  height: 220px;
  border-radius: 50%;
  background: radial-gradient(
    circle,
    color-mix(in srgb, var(--tk-primary-color) 22%, transparent) 0%,
    transparent 70%
  );
  pointer-events: none;
  z-index: -1;
}

.tk-upgrade-banner__content {
  display: flex;
  flex-direction: column;
  gap: var(--tk-spacing-md);
}

.tk-upgrade-banner__head {
  display: flex;
  flex-direction: column;
  gap: var(--tk-spacing-xs);
}

.tk-upgrade-banner__heading {
  display: flex;
  align-items: center;
  gap: var(--tk-spacing-sm);
  flex-wrap: wrap;
}

.tk-upgrade-banner__badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  font-family: var(--tk-font-family-mono, monospace);
  font-size: var(--tk-font-size-xs);
  font-weight: var(--tk-font-weight-bold);
  color: #fff;
  background: linear-gradient(135deg, var(--tk-primary-color) 0%, var(--tk-primary-color-dark, var(--tk-primary-color)) 100%);
  border-radius: var(--tk-radius-sm);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.tk-upgrade-banner__title {
  margin: 0;
  font-family: var(--tk-font-family);
  font-size: var(--tk-font-size-xl);
  font-weight: var(--tk-font-weight-bold);
  color: var(--tk-text-primary);
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.tk-upgrade-banner__desc {
  margin: 0;
  font-size: var(--tk-font-size-sm);
  color: var(--tk-text-secondary);
  line-height: var(--tk-line-height-normal);
  max-width: 640px;
}

.tk-upgrade-banner__features {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--tk-spacing-sm);
  list-style: none;
  margin: 0;
  padding: 0;
}

.tk-upgrade-banner__feature {
  display: flex;
  align-items: center;
  gap: var(--tk-spacing-sm);
  padding: var(--tk-spacing-sm) var(--tk-spacing-md);
  background-color: var(--tk-bg-surface);
  border: 1px solid var(--tk-border-color-lighter);
  border-radius: var(--tk-radius-md);
  transition: border-color var(--tk-transition-fast), background-color var(--tk-transition-fast);

  &:hover {
    border-color: color-mix(in srgb, var(--tk-primary-color) 40%, var(--tk-border-color));
    background-color: color-mix(in srgb, var(--tk-primary-color) 5%, var(--tk-bg-surface));
  }
}

.tk-upgrade-banner__feature-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  border-radius: var(--tk-radius-sm);
  background-color: var(--tk-primary-color-bg);
  color: var(--tk-primary-color);

  i { font-size: 15px; }
}

.tk-upgrade-banner__feature-label {
  font-size: var(--tk-font-size-sm);
  font-weight: var(--tk-font-weight-medium);
  color: var(--tk-text-primary);
  line-height: 1.3;
}

.tk-upgrade-banner__actions {
  display: flex;
  align-items: center;
  gap: var(--tk-spacing-sm);
}

.tk-upgrade-banner__cta {
  display: inline-flex;
  align-items: center;
  gap: var(--tk-spacing-xs);
  padding: var(--tk-spacing-sm) var(--tk-spacing-lg);
  font-family: var(--tk-font-family);
  font-size: var(--tk-font-size-sm);
  font-weight: var(--tk-font-weight-semibold);
  color: #fff;
  background: var(--tk-gradient-primary, var(--tk-primary-color));
  border: 1px solid var(--tk-primary-color);
  border-radius: var(--tk-radius-md);
  text-decoration: none;
  cursor: pointer;
  transition: box-shadow var(--tk-transition-fast), transform var(--tk-transition-fast);

  &:hover {
    box-shadow: var(--tk-glow-primary, 0 0 0 3px color-mix(in srgb, var(--tk-primary-color) 20%, transparent));
    color: #fff;
  }

  &:active {
    transform: translateY(1px);
  }

  &:focus-visible {
    outline: 2px solid var(--tk-primary-color);
    outline-offset: 2px;
  }

  i { font-size: 14px; }
}

@media (max-width: 1199px) {
  .tk-upgrade-banner__features {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 639px) {
  .tk-upgrade-banner__features {
    grid-template-columns: 1fr;
  }
}

@media (prefers-reduced-motion: reduce) {
  .tk-upgrade-banner__cta,
  .tk-upgrade-banner__feature {
    transition: none;
  }
}
</style>
