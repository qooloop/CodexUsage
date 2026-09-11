<script setup lang="ts">
/**
 * StatItem - 单行统计项（设计图：本会话里的"会话累计 Token 714万"）
 * icon | label（占主） | value（右对齐）
 */
defineProps<{
  icon: 'token' | 'cache' | 'request' | 'status' | 'time' | 'clock' | 'calendar';
  label: string;
  value: string;
  statusColor?: 'good' | 'warn' | 'bad' | 'idle'; // 仅 status 模式
  hint?: string; // 问号提示
}>();
</script>

<template>
  <div class="stat-item">
    <span class="ico">
      <svg v-if="icon === 'token'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="9" />
        <path d="M12 3v18M3 12h18" />
      </svg>
      <svg v-else-if="icon === 'cache'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 12a9 9 0 1 1-3.5-7.1" />
        <path d="M21 3v6h-6" />
      </svg>
      <svg v-else-if="icon === 'request'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M2 12h20M2 12l4-4M2 12l4 4" />
      </svg>
      <svg v-else-if="icon === 'status'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M3 20h18M5 20V10M10 20V4M15 20v-8M20 20V14" />
      </svg>
      <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="9" />
        <polyline points="12 7 12 12 15 14" />
      </svg>
    </span>
    <span class="label">{{ label }}</span>
    <span class="value tabular" v-if="icon === 'status'">
      <span class="status-dot" :class="statusColor ?? 'good'"></span>
      <span class="status-text">{{ value }}</span>
      <span v-if="hint" class="hint">?</span>
    </span>
    <span v-else class="value tabular">{{ value }}</span>
  </div>
</template>

<style scoped>
.stat-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 4px;
  border-bottom: 1px solid var(--border);
  font-size: var(--fs-13);
}
.stat-item:last-child { border-bottom: 0; }
.ico {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  color: var(--text-tertiary);
  flex-shrink: 0;
}
.label {
  flex: 1;
  color: var(--text-secondary);
}
.value {
  color: var(--text-primary);
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.status-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}
.status-dot.good {
  background: var(--accent);
  box-shadow: 0 0 6px rgba(53, 230, 165, 0.5);
}
.status-dot.warn {
  background: var(--warning);
}
.status-dot.bad {
  background: var(--danger);
}
.status-dot.idle {
  background: var(--text-tertiary);
}
.status-text {
  margin-left: 2px;
}
.hint {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-tertiary);
  font-size: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-left: 2px;
}
</style>
