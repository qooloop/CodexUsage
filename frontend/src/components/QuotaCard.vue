<script setup lang="ts">
/**
 * QuotaCard - 配额卡片（设计图 3 张卡：5h / 7d / 今日 Token）
 * 顶部：图标 + 标签
 * 中部：大数字 + 单位
 * 底部：进度条 + 重置时间
 */
import { computed } from 'vue';
import { fmtBig, fmtQuota } from '@/lib/format';

interface Props {
  icon: 'clock' | 'calendar' | 'token';
  label: string;
  value: number | null; // 剩余% 或 token 数
  total?: number; // 当 value 表示"已用"时，可算出"剩余"
  /** 进度条方向：剩余（默认）从左到右，已用（token）从左到右 */
  fillPercent: number; // 0~100，进度条已填充宽度
  fillColor?: 'good' | 'warn' | 'bad' | 'blue';
  footLeft: string; // 卡片脚左（如"重置 13:46"）
  footRight?: string; // 可选右箭头（点击事件由父级传）
}

const props = withDefaults(defineProps<Props>(), {
  fillColor: 'good',
  footRight: '›',
});

const valueText = computed(() => {
  if (props.value == null) return '—';
  if (props.icon === 'token') {
    return fmtBig(props.value); // 统一口径：21万 / 23.5万 / 1.20亿
  }
  return fmtQuota(props.value);
});
</script>

<template>
  <div class="quota-card">
    <div class="head">
      <span class="ico" :class="icon">
        <svg v-if="icon === 'clock'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="9" />
          <polyline points="12 7 12 12 15 14" />
        </svg>
        <svg v-else-if="icon === 'calendar'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="3" y="4" width="18" height="18" rx="2" />
          <line x1="3" y1="10" x2="21" y2="10" />
          <line x1="8" y1="2" x2="8" y2="6" />
          <line x1="16" y1="2" x2="16" y2="6" />
        </svg>
        <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v13c0 4 16 4 16 0V5M4 11c0 4 16 4 16 0"/>
        </svg>
      </span>
      <span class="label">{{ label }}</span>
    </div>

    <div class="value tabular">{{ valueText }}</div>

    <div class="bar">
      <div class="bar-fill" :class="fillColor" :style="{ width: fillPercent + '%' }"></div>
    </div>

    <div class="foot">
      <span class="foot-left">{{ footLeft }}</span>
      <span v-if="footRight" class="foot-right">{{ footRight }}</span>
    </div>
  </div>
</template>

<style scoped>
.quota-card {
  background: linear-gradient(180deg, rgba(22, 34, 52, 0.65) 0%, rgba(16, 26, 40, 0.55) 100%);
  border-radius: var(--radius-card);
  padding: 16px 20px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
  border: 1px solid var(--border);
  box-shadow: var(--shadow-inner);
  transition: all var(--t-base) var(--ease);
  position: relative;
  overflow: hidden;
}
.quota-card::before {
  /* 顶部高光：设计图卡片顶部有一道亮边 */
  content: '';
  position: absolute;
  top: 0;
  left: 12px;
  right: 12px;
  height: 1px;
  background: linear-gradient(90deg, transparent 0%, rgba(255, 255, 255, 0.10) 50%, transparent 100%);
}
.quota-card:hover {
  background: var(--bg-hover);
  border-color: var(--border-strong);
}
.head {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-secondary);
  font-size: 14px;
}
.ico {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  color: var(--accent);
}
.ico.calendar {
  color: var(--accent-blue);
}
.ico.token {
  color: var(--accent);
}
.label {
  font-size: 14px;
  color: var(--text-secondary);
}
.value {
  font-size: 34px;
  font-weight: 700;
  line-height: 1.1;
  color: var(--text-primary);
  letter-spacing: -0.5px;
  margin-top: 2px;
}
.bar {
  width: 100%;
  height: 9px;
  background: rgba(112, 143, 172, 0.20);
  border-radius: var(--radius-progress);
  overflow: hidden;
  margin-top: 2px;
}
.bar-fill {
  height: 100%;
  border-radius: var(--radius-progress);
  transition: width var(--t-slow) var(--ease);
}
.bar-fill.good {
  background: linear-gradient(90deg, #31ef78 0%, #12e8bb 100%);
  box-shadow: 0 0 8px rgba(53, 230, 165, 0.3);
}
.bar-fill.warn {
  background: linear-gradient(90deg, var(--warning) 0%, var(--warning-dim) 100%);
  box-shadow: 0 0 8px rgba(246, 196, 83, 0.3);
}
.bar-fill.bad {
  background: linear-gradient(90deg, var(--danger) 0%, var(--danger-dim) 100%);
  box-shadow: 0 0 8px rgba(255, 97, 116, 0.3);
}
.bar-fill.blue {
  background: linear-gradient(90deg, #319cf6 0%, #46b2ff 100%);
  box-shadow: 0 0 8px rgba(70, 189, 244, 0.3);
}
.foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: var(--fs-12);
  color: var(--text-secondary);
  margin-top: 2px;
}
.foot-right {
  color: var(--text-secondary);
  font-size: 16px;
  line-height: 1;
  opacity: 0.6;
}
.quota-card:hover .foot-right {
  opacity: 1;
  color: var(--text-secondary);
}
</style>
